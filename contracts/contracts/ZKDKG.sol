//SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./ShareVerifier.sol";
import "./KeyVerifier.sol";

// TODO Implement staking
contract ZKDKG {
    uint public constant STAKE = 0 ether;

    uint16 public immutable noParticipants;
    uint16 public immutable minimumThreshold;
    uint16 public immutable userThreshold;
    uint16 public immutable periodLength;

    Phase public phase;
    uint64 public phaseEnd;

    mapping(address => Participant) public participants;
    address[] public addresses;
    address[] private disqualified;

    mapping(address => bytes32) public commitmentHashes;
    mapping(address => bytes32) public shareHashes;
    uint[] public firstCoefficients;

    // Compressed encoding of the point at infinity (0, 1)
    uint private constant INFINITY = 1;

    // Order of the underlying prime finite field of the used curve, Baby Jubjub
    uint private constant FIELD_ORDER = 21888242871839275222246405745257275088548364400416034343698204186575808495617;

    // Parameters defining the Twisted Edwards curve
    uint private constant FIELD_A = 168700;
    uint private constant FIELD_D = 168696;

    // Used for setting expiries to a far distant point in time
    uint64 private constant POINT_IN_FUTURE = 7258118400;

    bool private immutable isEvaluation;

    uint16 private noBroadcasts = 0;

    ShareVerifier private shareVerifier;
    KeyVerifier private keyVerifier;

    mapping(address => Dispute) private disputes;
    address[] private disputed;

    struct Participant {
        uint16 index;

        uint[2] publicKey;
        /**
         * The public key could be stored in compressed form, reducing the array to a single unsigned integer.
         * For gas efficiency and code complexity reasons this is not done here.
         *
         * If the point were in compressed Twisted Edwards form, decompression would require submod, divmod and sqrt operations
         * due to the need to compute x = sqrt((1 - y^2) / (1 - dy^2)).
         * submod(x, y, p) can be rewritten as addmod(p - (y % p), x, p) and divmod is equivalent to the multiplication of the modular inverse
         * (see https://github.com/witnet/elliptic-curve-solidity/blob/b6886bb08333ccf6883ac42827d62c1bfdb37d44/contracts/EllipticCurve.sol#L22).
         * The "hack" to compute the square root of an integer without fractional exponents doesn't apply for the Baby Jubjub curve.
         * It relies on the identity that sqrt(r) = r^(1/2) is congruent r^((p + 1) / 4) modulo p if p mod 4 = 3, where r and p are unsigned integers.
         * The order p_B of the Baby Jubjub curve doesn't satisfy this condition, therefore a more sophisticated algorithm (like the Tonelli-Shanks algorithm)
         * would have to be used.
         *
         * If the point were in compressed Weierstrass form, only a method for computing the square root would be additionally required.
         * As seen before, due to the defining characteristics of the Baby Jubjub curve, this is not easily achievable.
         * Also, an extra conversion would have to be made on the Zokrates side because it needs points of a Twisted Edwards curve for
         * its computations.
         *
         * Therefore, the point is being passed in uncompressed form.
         * To circumvent this overhead one could require participants to submit a proof to the register call or implement a square root algorithm.
         * We consider both of these solutions to be too complex for the problem at hand.
        **/
    }

    struct Dispute {
        uint16 disputerIndex;
        uint16 disputeeIndex;
        uint64 end;
        uint share;
    }

    enum Phase {
        UNINITIALIZED,
        REGISTER,
        BROADCAST_SUBMIT,
        BROADCAST_DISPUTE
    }

    event DisputeShare(uint16 disputerIndex, uint16 disputeeIndex);
    event BroadcastSharesLog(address sender, uint16 broadcasterIndex);
    event RegistrationEndLog();
    event DistributionEndLog();
    event PublicKeySubmission();
    event Abortion();
    event Reset();
    event Exclusion(uint16 index);

    constructor(
        address _shareVerifier,
        address _keyVerifier,
        uint16 _noParticipants,
        uint16 _userThreshold,
        uint16 _periodLength
    ) {
        uint16 _minimumThreshold = (_noParticipants + 1) / 2;

        require(
            _userThreshold >= _minimumThreshold && _userThreshold <= _noParticipants,
            "user threshold has to be between the mathematical minimum threshold and the number of participants (inclusively)"
        );

        shareVerifier = ShareVerifier(_shareVerifier);
        keyVerifier = KeyVerifier(_keyVerifier);

        noParticipants = _noParticipants;
        minimumThreshold = _minimumThreshold;
        userThreshold = _userThreshold;
        periodLength = _periodLength;
        isEvaluation = _periodLength == 0;

        firstCoefficients = new uint[](_noParticipants);

        // Avoid higher costs for the last participant that calls register
        phase = Phase.REGISTER;
        phaseEnd = type(uint64).max;
    }

    function register(uint[2] calldata publicKey) public payable {
        require(msg.value == STAKE, "value too low");

        if (phase == Phase.REGISTER) {
            require(!isRegistered(msg.sender), "already registered");
        } else if (phase == Phase.BROADCAST_DISPUTE) {
            require(block.timestamp > phaseEnd, "dispute period still ongoing");
            reset();
        } else {
            revert("registration phase is over");
        }

        addresses.push(msg.sender);
        participants[msg.sender] = Participant(uint16(addresses.length), publicKey);

        if (addresses.length == noParticipants) {
            phaseEnd = isEvaluation ? POINT_IN_FUTURE : uint64(block.timestamp) + periodLength;
            phase = Phase.BROADCAST_SUBMIT;

            emit RegistrationEndLog();
        }
    }

    function broadcastShares(uint[] memory commitments, uint[] memory shares) external registered {
        require(phase == Phase.BROADCAST_SUBMIT, "broadcast period has not started yet");
        require(block.timestamp <= phaseEnd, "broadcast period has expired");
        require(commitmentHashes[msg.sender] == 0, "already broadcasted before");
        require(shares.length == noParticipants - 1, "invalid number of shares");
        require(commitments.length == minimumThreshold, "invalid number of commitments");

        uint16 index = participants[msg.sender].index;

        firstCoefficients[index - 1] = commitments[0];
        commitmentHashes[msg.sender] = keccak256(abi.encodePacked(commitments));
        shareHashes[msg.sender] = keccak256(abi.encodePacked(shares));

        emit BroadcastSharesLog(msg.sender, index);

        if (++noBroadcasts == noParticipants) {
            phaseEnd = isEvaluation ? POINT_IN_FUTURE : uint64(block.timestamp) + periodLength;
            phase = Phase.BROADCAST_DISPUTE;

            emit DistributionEndLog();
        }
    }

    function disputeShare(uint16 disputeeIndex, uint[] calldata shares) external registered {
        address disputeeAddr = addresses[disputeeIndex - 1];
        
        require(phase == Phase.BROADCAST_DISPUTE && block.timestamp <= phaseEnd, "not in dispute period");
        require(!isDisputed(disputes[disputeeAddr]), "disputee already disputed");
        require(shareHashes[disputeeAddr] == keccak256(abi.encodePacked(shares)), "invalid shares");

        uint16 disputerIndex = participants[msg.sender].index;

        require(isPublicKeyValid(), "sender's public key not on curve");

        uint shareIndex = disputerIndex;

        // The shares of each dealer don't include a share for themselves, so there's a "closed gap" in each shares array
        if (shareIndex > disputeeIndex) {
            shareIndex--;
        }
        shareIndex--; // Participant indices are one-based

        phaseEnd = isEvaluation ? POINT_IN_FUTURE : uint64(block.timestamp) + periodLength;

        disputes[disputeeAddr] = Dispute(
            disputerIndex,
            disputeeIndex,
            phaseEnd,
            shares[shareIndex]
        );

        disputed.push(disputeeAddr);
        
        emit DisputeShare(disputerIndex, disputeeIndex);
    }

    function defendShare(ShareVerifier.Proof calldata proof) external {
        Dispute memory dispute = disputes[msg.sender];

        require(isDisputed(dispute), "not being disputed");
        require(block.timestamp <= dispute.end, "defense period expired");

        address disputee = addresses[dispute.disputeeIndex - 1];
        address disputer = addresses[dispute.disputerIndex - 1];

        uint hash = truncateHash(keccak256(
            bytes.concat(
                commitmentHashes[disputee],
                bytes32(participants[disputee].publicKey[0]),
                bytes32(participants[disputee].publicKey[1]),
                bytes32(participants[disputer].publicKey[0]),
                bytes32(participants[disputer].publicKey[1]),
                bytes32(uint(dispute.disputerIndex)),
                bytes32(dispute.share)
            )
        ));

        uint[2] memory input = [
            hash,
            1
        ];

        require(shareVerifier.verifyTx(proof, input), "invalid proof");

        // For the purpose of the evaluation script, end the dispute / defend phase once one share has been successfully defended
        if (isEvaluation) {
            phaseEnd = uint64(block.timestamp);
        }

        delete disputes[msg.sender];

        uint index = findDisputeIndex(msg.sender);
        if (index != disputed.length - 1) {
            disputed[index] = disputed[disputed.length - 1];
        }
        disputed.pop();

        excludeNode(dispute.disputerIndex);
    }

    function submitPublicKey(uint[2] calldata _publicKey, KeyVerifier.Proof calldata proof) external {
        require(phase == Phase.BROADCAST_DISPUTE, "not in submission phase");
        require(block.timestamp > phaseEnd, "dispute period still ongoing");

        removeExpiredDisputes();

        uint hash = truncateHash(keccak256(abi.encodePacked(firstCoefficients)));

        uint[3] memory input = [
            hash,
            _publicKey[0],
            _publicKey[1]
        ];

        require(keyVerifier.verifyTx(proof, input), "invalid proof");

        emit PublicKeySubmission();

        payNodes();
    }

    /**
     * TODO Generate incentive by distributing rewards in a different way
     *
     * It has to be ensured that:
     *  1) the uninvolved nodes get at least their tx costs for registering and broadcasting refunded
     *  2) the disputer gets their tx costs refunded + a reward fee, which is in total higher than the
     *     rewards of the other nodes
     *  3) one rational node is sufficient to guarantee finalization
     *
     * - If the lost stake is distributed equally, the disputer has no incentive to pay tx
     * costs for the dispute call [2] (although no one benefits from an undisputed invalid share)
     * - If the disputer receives the whole stake all other nodes don't get
     * their tx costs refunded. [1]
     * - A distribution s.t [1] and [2] (and therefore [3] if everybody can call the function) are satisfied.
     * Such a distribution is hard to calculate in the general case due to tx costs being variable and
     * the dependence on the stored stake, i.e. it is not possible to distribute the reward s.t. [1] and [2] hold
     * for every stake.
     */
    function payNodes() private {}

    function reset() private {

        // Exclude disqualified participants from next round, update indices
        for (uint i = 0; i < disqualified.length; i++) {
            address addr = disqualified[i];
            uint16 index = participants[addr].index;

            delete participants[addr];
            delete shareHashes[addr];
            delete commitmentHashes[addr];

            if (index != addresses.length) {
                Participant memory lastParticipant = participants[addresses[addresses.length - 1]];
                addresses[index - 1] = addresses[addresses.length - 1];
                lastParticipant.index = index;
            }
            addresses.pop();
        }

        for (uint i = 0; i < addresses.length; i++) {
            address addr = addresses[i];

            delete shareHashes[addr];
            delete commitmentHashes[addr];
        }

        delete disqualified;
        delete disputed;
        delete firstCoefficients;
        delete noBroadcasts;

        phase = Phase.REGISTER;

        emit Reset();
    }

    function excludeNode(uint16 index) private {
        address addr = addresses[index - 1];

        firstCoefficients[index - 1] = INFINITY;
        disqualified.push(addr);

        emit Exclusion(index);

        if (noParticipants - disqualified.length < userThreshold) {
            emit Abortion();
        }
    }

    function expiredDisputes() external view returns (uint16[] memory) {
        uint16[] memory indices = new uint16[](disputed.length);
        for (uint16 i = 0; i < disputed.length; i++) {
            indices[i] = participants[disputed[i]].index;
        }
        return indices;
    }

    function removeExpiredDisputes() private {
        for (uint16 i = 0; i < disputed.length; i++) {
            address addr = disputed[i];
            excludeNode(participants[addr].index);
            delete disputes[addr];
        }
    }

    /// @dev Check whether point (x,y) is on the Baby Jubjub curve.
    /// @param x x coordinate
    /// @param y y coordinate
    /// @return true if (x,y) is on the curve, false otherwise
    function isOnCurve(uint x, uint y) private pure returns (bool) {
        if (x >= FIELD_ORDER || y >= FIELD_ORDER) {
            return false;
        }

        uint xx = mulmod(x, x, FIELD_ORDER);
        uint yy = mulmod(y, y, FIELD_ORDER);

        uint lhs = mulmod(FIELD_A, xx, FIELD_ORDER);
        lhs = addmod(lhs, yy, FIELD_ORDER);

        uint rhs = 1 + mulmod(mulmod(FIELD_D, xx, FIELD_ORDER), yy, FIELD_ORDER);

        return lhs == rhs;
    }

    function isPublicKeyValid() private view returns (bool) {
        uint[2] memory pk = participants[msg.sender].publicKey;

        return isOnCurve(pk[0], pk[1]);
    }

    function isRegistered(address _addr) public view returns (bool) {
        if (addresses.length == 0) return false;
        uint index = participants[_addr].index;
        return index != 0 && addresses[index - 1] == _addr;
    }

    function isDisputed(Dispute memory dispute) private pure returns (bool) {
        return dispute.disputeeIndex != 0;
    }

    function publicKeys() external view returns (uint[2][] memory) {
        uint[2][] memory results = new uint[2][](addresses.length);
        for (uint i = 0; i < addresses.length; i++) {
            results[i] = participants[addresses[i]].publicKey;
        }
        return results;
    }

    function truncateHash(bytes32 _hash) private pure returns (uint) {
        // Truncate the hash s.t. its value range is limited to exactly all field elements
        return uint(_hash) % FIELD_ORDER;
    }

    function findDisputeIndex(address addr) private view returns (uint) {
        for (uint i = 0; i < disputed.length; i++) {
            if (disputed[i] == addr) {
                return i;
            }
        }
        return disputed.length;
    }

    modifier registered() {
        require(isRegistered(msg.sender), "not registered");
        _;
    }
}
