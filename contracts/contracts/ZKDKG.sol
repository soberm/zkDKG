//SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./ShareVerifier.sol";
import "./KeyVerifier.sol";

// TODO Implement staking
contract ZKDKG {
    uint public constant STAKE = 0 ether;

    uint public immutable noParticipants;
    uint public immutable minimumThreshold;
    uint public immutable userThreshold;
    uint16 public immutable periodLength;

    Phase public phase;
    uint64 public phaseEnd;

    mapping(address => Participant) public participants;
    address[] public addresses;

    mapping(address => bytes32) public commitmentHashes;
    mapping(address => bytes32) public shareHashes;
    uint[] public firstCoefficients;

    // Compressed encoding of the point at infinity (0, 1)
    uint private constant INFINITY = 1;

    ShareVerifier private shareVerifier;
    KeyVerifier private keyVerifier;

    mapping(address => Dispute) private disputes;

    struct Participant {
        uint64 index;
        uint publicKey;
    }

    struct Dispute {
        uint64 disputerIndex;
        uint64 disputeeIndex;
        uint64 end;
        uint share;
    }

    enum Phase {
        UNINITIALIZED,
        REGISTER,
        BROADCAST_SUBMIT,
        BROADCAST_DISPUTE
    }

    event DisputeShare(uint64 disputerIndex, uint64 disputeeIndex);
    event BroadcastSharesLog(address sender, uint64 broadcasterIndex);
    event RegistrationEndLog();
    event DistributionEndLog();
    event PublicKeySubmission();
    event Abortion();
    event Reset();
    event Exclusion(uint64 index);

    constructor(
        address _shareVerifier,
        address _keyVerifier,
        uint _noParticipants,
        uint _userThreshold,
        uint16 _periodLength
    ) {
        uint _minimumThreshold = (_noParticipants + 1) / 2;

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

        // Avoid higher costs for the last participant that calls register
        phase = Phase.REGISTER;
        phaseEnd = type(uint64).max;
    }

    function register(uint publicKey) public payable {
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
        participants[msg.sender] = Participant(uint64(addresses.length), publicKey);

        if (addresses.length == noParticipants) {
            phaseEnd = uint64(block.timestamp) + periodLength;
            phase = Phase.BROADCAST_SUBMIT;

            emit RegistrationEndLog();
        }
    }

    function broadcastShares(uint[] memory commitments, uint[] memory shares) external registered {
        require(phase == Phase.BROADCAST_SUBMIT, "broadcast period has not started yet");
        require(block.timestamp <= phaseEnd, "broadcast period has expired");
        require(commitmentHashes[msg.sender] == 0, "already broadcasted before");
        require(shares.length == addresses.length - 1, "invalid number of shares");
        require(commitments.length == minimumThreshold, "invalid number of commitments");

        firstCoefficients.push(commitments[0]);
        commitmentHashes[msg.sender] = keccak256(abi.encodePacked(commitments));
        shareHashes[msg.sender] = keccak256(abi.encodePacked(shares));

        emit BroadcastSharesLog(msg.sender, participants[msg.sender].index);

        if (firstCoefficients.length == noParticipants) {
            phaseEnd = uint64(block.timestamp) + periodLength;
            phase = Phase.BROADCAST_DISPUTE;

            emit DistributionEndLog();
        }
    }

    function disputeShare(uint64 disputeeIndex, uint[] calldata shares) external registered {
        address disputeeAddr = addresses[disputeeIndex - 1];
        
        require(phase == Phase.BROADCAST_DISPUTE && block.timestamp <= phaseEnd, "not in dispute period");
        require(!isDisputed(disputes[disputeeAddr]), "disputee already disputed");
        require(shareHashes[disputeeAddr] == keccak256(abi.encodePacked(shares)), "invalid shares");

        uint64 disputerIndex = participants[msg.sender].index;

        // TODO Check that disputer public key is valid

        uint shareIndex = disputerIndex;

        // The shares of each dealer don't include a share for themselves, so there's a "closed gap" in each shares array
        if (shareIndex > disputeeIndex) {
            shareIndex--;
        }
        shareIndex--; // Participant indices are one-based

        phaseEnd = uint64(block.timestamp) + periodLength;

        disputes[disputeeAddr] = Dispute(
            disputerIndex,
            disputeeIndex,
            phaseEnd,
            shares[shareIndex]
        );
        
        emit DisputeShare(disputerIndex, disputeeIndex);
    }

    function defendShare(ShareVerifier.Proof calldata proof) external registered {
        Dispute memory dispute = disputes[msg.sender];

        require(isDisputed(disputes[msg.sender]), "not being disputed");
        require(block.timestamp <= dispute.end, "defense period expired");

        address disputee = addresses[dispute.disputeeIndex - 1];
        address disputer = addresses[dispute.disputerIndex - 1];

        uint hash = truncateHash(keccak256(
            bytes.concat(
                commitmentHashes[disputee],
                bytes32(participants[disputee].publicKey),
                bytes32(participants[disputer].publicKey),
                bytes32(uint(dispute.disputerIndex)),
                bytes32(dispute.share)
            )
        ));


        uint[2] memory input = [
            hash,
            1
        ];

        require(shareVerifier.verifyTx(proof, input), "invalid proof");

        delete disputes[msg.sender];
        excludeNode(dispute.disputerIndex);
    }

    function submitPublicKey(uint[2] calldata _publicKey, KeyVerifier.Proof calldata proof) external registered {
        require(phase >= Phase.BROADCAST_DISPUTE, "not in submission phase");
        require(block.timestamp > phaseEnd, "dispute period still ongoing");

        checkExpiredDisputes();

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
    function payNodes() private {
        reset();
    }

    function reset() private {
        for (uint i = 0; i < addresses.length; i++) {
            address addr = addresses[i];

            delete participants[addr];
            delete shareHashes[addr];
            delete commitmentHashes[addr];
            delete disputes[addr];
        }

        delete addresses;
        delete firstCoefficients;

        phase = Phase.REGISTER;

        emit Reset();
    }

    function excludeNode(uint64 disputerIndex) internal {
        firstCoefficients[disputerIndex - 1] = INFINITY;
        delete participants[addresses[disputerIndex - 1]];

        emit Exclusion(disputerIndex);

        if (participantsCount() < userThreshold) {
            emit Abortion();
        }
    }

    function expiredDisputes(uint timestamp) public view returns (bool[] memory) {
        bool[] memory indices = new bool[](addresses.length);
        for (uint64 i = 0; i < indices.length; i++) {
            address addr = addresses[i];
            Dispute memory dispute = disputes[addr];

            if (isDisputed(dispute) && dispute.end <= timestamp) {
                indices[i] = true;
            }
        }
        return indices;
    }

    function checkExpiredDisputes() internal {
        bool[] memory indices = expiredDisputes(block.timestamp);

        for (uint64 i = 0; i < indices.length; i++) {
            if (indices[i]) {
                excludeNode(i + 1);
            }
        }
    }

    function participantsCount() internal view returns (uint) {
        uint count = 0;
        for (uint i = 0; i < addresses.length; i++) {
            if (participants[addresses[i]].index != 0) {
                count++;
            }
        }
        return count;
    }

    function isRegistered(address _addr) public view returns (bool) {
        if (addresses.length == 0) return false;
        uint index = participants[_addr].index;
        return index != 0 && addresses[index - 1] == _addr;
    }

    function isDisputed(Dispute memory dispute) internal pure returns (bool) {
        return dispute.disputeeIndex != 0;
    }

    function publicKeys() external view returns (uint[] memory) {
        uint[] memory results = new uint[](addresses.length);
        for (uint i = 0; i < addresses.length; i++) {
            results[i] = participants[addresses[i]].publicKey;
        }
        return results;
    }

    function truncateHash(bytes32 _hash) internal pure returns (uint) {
        // Truncate the first 3 bits s.t. value range is limited to 253 bits (longest bit length that only contains BabyJubJub field values)
        return uint(_hash & hex"1FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF");
    }

    modifier registered() {
        require(isRegistered(msg.sender), "not registered");
        _;
    }
}
