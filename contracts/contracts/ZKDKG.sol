//SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./ShareVerifier.sol";
import "./KeyVerifier.sol";
import "./ShareInputVerifier.sol";

contract ZKDKG {
    uint16 public constant KEY_DISPUTE_PERIOD = 30 seconds;
    uint16 public constant SHARES_BROADCAST_PERIOD = 10 seconds;
    uint16 public constant SHARES_DISPUTE_PERIOD = 0 minutes;

    uint256 public constant MIN_STAKE = 0 ether;

    uint private constant CURVE_ORDER = 21888242871839275222246405745257275088614511777268538073601725287587578984328;

    struct Participant {
        uint256 index;
        uint256[2] publicKey;
    }

    enum Phases {
        REGISTER,
        BROADCAST_SUBMIT,
        BROADCAST_DISPUTE,
        PK_DISPUTE
    }
    Phases public phase;

    mapping(address => Participant) public participants;
    address[] public addresses;

    mapping(address => bytes32) public commitmentHashes;
    mapping(address => bytes32) public shareHashes;
    uint256[2][] public firstCoefficients;

    address submitter;
    uint256[2] public masterPublicKey;
    uint256 public noParticipants;

    ShareVerifier private shareVerifier;
    KeyVerifier private keyVerifier;
    ShareInputVerifier private shareInputVerifier;

    uint64 private keyDisputableUntil;
    uint64 private sharesBroadcastableUntil;
    uint64 private sharesDisputableUntil;

    event DisputeShare();
    event BroadcastSharesLog(address sender, uint256 broadcasterIndex);
    event RegistrationEndLog();
    event DistributionEndLog();
    event PublicKeySubmission();

    constructor(
        address _shareVerifier,
        address _keyVerifier,
        address _shareInputVerifier,
        uint256 _noParticipants
    ) {
        shareVerifier = ShareVerifier(_shareVerifier);
        keyVerifier = KeyVerifier(_keyVerifier);
        shareInputVerifier = ShareInputVerifier(_shareInputVerifier);
        noParticipants = _noParticipants;
    }

    function register(uint256[2] memory publicKey) public payable {

        require(msg.value == MIN_STAKE, "value too low");

        if (phase == Phases.REGISTER) {
            require(!isRegistered(msg.sender), "already registered");
        } else if (phase == Phases.PK_DISPUTE) {
            require(block.timestamp > keyDisputableUntil, "dispute period still ongoing");
            reset();
        } else {
            revert("registration phase is over");
        }

        require(publicKey[0] < CURVE_ORDER && publicKey[1] < CURVE_ORDER, "invalid public key");

        participants[msg.sender] = Participant(addresses.length, publicKey);
        addresses.push(msg.sender);

        if (addresses.length == noParticipants) {
            sharesBroadcastableUntil = uint64(block.timestamp) + SHARES_BROADCAST_PERIOD;
            phase = Phases.BROADCAST_SUBMIT;

            emit RegistrationEndLog();
        }
    }

    // FIXME One account can call this multiple times
    function broadcastShares(
        uint256[2][] memory commitments,
        uint256[] memory shares
    ) external {
        require(phase >= Phases.BROADCAST_SUBMIT, "broadcast period has not started yet");
        require(block.timestamp <= sharesBroadcastableUntil, "broadcast period has expired");

        require(
            shares.length == addresses.length - 1,
            "invalid number of shares"
        );
        require(isRegistered(msg.sender), "not registered");

        require(
            commitments.length == threshold(),
            "invalid number of commitments"
        );

        for (uint i = 0; i < shares.length; i++) {
            require(shares[i] < CURVE_ORDER, "invalid share");
        }
        for (uint i = 0; i < commitments.length; i++) {
            require(commitments[i][0] < CURVE_ORDER && commitments[i][1] < CURVE_ORDER, "invalid commitment");
        }

        firstCoefficients.push(commitments[0]);
        commitmentHashes[msg.sender] = keccak256(abi.encodePacked(commitments));
        shareHashes[msg.sender] = keccak256(abi.encodePacked(shares));

        emit BroadcastSharesLog(msg.sender, participants[msg.sender].index);

        if (firstCoefficients.length == noParticipants) {
            sharesDisputableUntil = uint64(block.timestamp) + SHARES_DISPUTE_PERIOD;
            phase = Phases.BROADCAST_DISPUTE;

            emit DistributionEndLog();
        }
    }

    function disputeShare(
        uint256 dealerIndex,
        uint256[] memory shares,
        ShareVerifier.Proof memory proof
    ) external {
        require(phase >= Phases.BROADCAST_DISPUTE, "dispute period has not started yet");
        require(block.timestamp <= sharesDisputableUntil, "dispute period has expired");

        address dealer = addresses[dealerIndex];
        require(
            shareHashes[dealer] == keccak256(abi.encodePacked(shares)),
            "invalid shares"
        );

        uint256 disputerIndex = participants[msg.sender].index;
        if (disputerIndex > dealerIndex) {
            disputerIndex--;
        }

        uint oneBasedDisputerIndex = participants[msg.sender].index + 1;

        uint256[2] memory hash = hashToUint128(
            keccak256(
                bytes.concat(
                    commitmentHashes[dealer],
                    bytes32(participants[msg.sender].publicKey[0]),
                    bytes32(participants[msg.sender].publicKey[1]),
                    bytes32(participants[dealer].publicKey[0]),
                    bytes32(participants[dealer].publicKey[1]),
                    bytes32(oneBasedDisputerIndex),
                    bytes32(shares[disputerIndex])
                )
            )
        );

        uint256[3] memory input = [
            hash[0],
            hash[1],
            1
        ];
        require(shareVerifier.verifyTx(proof, input), "invalid proof");
        emit DisputeShare();
    }

    function disputeShareInput(
        address dealerAddress,
        ShareInputVerifier.Proof memory proof
    ) external {
        require(phase >= Phases.BROADCAST_DISPUTE, "dispute period has not started yet");
        require(block.timestamp <= sharesDisputableUntil, "dispute period has expired");

        uint256[2] memory hash = hashToUint128(
            keccak256(
                bytes.concat(
                    commitmentHashes[dealerAddress],
                    bytes32(participants[dealerAddress].publicKey[0]),
                    bytes32(participants[dealerAddress].publicKey[1])
                )
            )
        );

        uint256[3] memory input = [
            hash[0],
            hash[1],
            0
        ];
        require(shareInputVerifier.verifyTx(proof, input), "invalid proof");
        emit DisputeShare();
    }

    function submitPublicKey(uint256[2] memory _publicKey) external {
        require(phase == Phases.BROADCAST_DISPUTE, "not in submission phase");
        require(block.timestamp > sharesDisputableUntil, "dispute period still ongoing");
        require(isRegistered(msg.sender), "not registered");

        submitter = msg.sender;
        masterPublicKey = _publicKey;
        keyDisputableUntil = uint64(block.timestamp) + KEY_DISPUTE_PERIOD;
        phase = Phases.PK_DISPUTE;

        emit PublicKeySubmission();
    }

    function disputePublicKey(KeyVerifier.Proof memory proof) external {
        require(phase == Phases.PK_DISPUTE, "not in dispute phase");
        require(block.timestamp <= keyDisputableUntil, "dispute period has expired");

        uint256[2] memory hash = hashToUint128(
            keccak256(abi.encodePacked(firstCoefficients))
        );

        uint256[5] memory input = [
            masterPublicKey[0],
            masterPublicKey[1],
            hash[0],
            hash[1],
            1
        ];
        require(keyVerifier.verifyTx(proof, input), "invalid proof");
        delete masterPublicKey;
        delete submitter;
        delete keyDisputableUntil;

        phase = Phases.BROADCAST_DISPUTE;
    }

    function reset() private {
        for (uint i = 0; i < addresses.length; i++) {
            address addr = addresses[i];

            delete participants[addr];
            delete shareHashes[addr];
            delete commitmentHashes[addr];
        }

        delete addresses;
        delete firstCoefficients;
        delete masterPublicKey;

        delete submitter;
        delete keyDisputableUntil;
        delete sharesDisputableUntil;
        delete phase;
    }

    function isRegistered(address _addr) public view returns (bool) {
        if (addresses.length == 0) return false;
        return (addresses[participants[_addr].index] == _addr);
    }

    function countParticipants() external view returns (uint256) {
        return addresses.length;
    }

    function findParticipantByIndex(uint256 _index)
        public
        view
        returns (Participant memory)
    {
        require(_index >= 0 && _index < addresses.length, "not found");
        return participants[addresses[_index]];
    }

    function threshold() public view returns (uint256) {
        return (addresses.length + 1) / 2;
    }

    function hashToUint128(bytes32 _hash)
        public
        pure
        returns (uint256[2] memory)
    {
        uint256 hash = uint256(_hash);
        uint128 lhs = uint128(hash >> 128);
        uint128 rhs = uint128(hash);
        return [uint256(lhs), uint256(rhs)];
    }
}
