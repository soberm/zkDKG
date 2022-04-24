//SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./ShareVerifier.sol";
import "./KeyVerifier.sol";

contract ZKDKG {
    uint16 public constant KEY_DISPUTE_PERIOD = 0 minutes;
    uint16 public constant SHARES_DISPUTE_PERIOD = 0 minutes;
    uint256 public constant MIN_STAKE = 0 ether;

    struct Participant {
        uint256 index;
        uint256[2] publicKey;
    }

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

    uint64 private keyDisputableUntil;
    uint64 private sharesDisputableUntil;

    event DisputeShare(bool result);
    event BroadcastSharesLog(address sender, uint256 broadcasterIndex);
    event RegistrationEndLog();
    event DistributionEndLog();

    constructor(
        address _shareVerifier,
        address _keyVerifier,
        uint256 _noParticipants
    ) {
        shareVerifier = ShareVerifier(_shareVerifier);
        keyVerifier = KeyVerifier(_keyVerifier);
        noParticipants = _noParticipants;
    }

    function register(uint256[2] memory publicKey) public payable {
        require(msg.value == MIN_STAKE, "value too low");
        require(!isRegistered(msg.sender), "already registered");
        require(addresses.length != noParticipants, "participants full");

        participants[msg.sender] = Participant(addresses.length, publicKey);
        addresses.push(msg.sender);

        if (addresses.length == noParticipants) {
            emit RegistrationEndLog();
        }
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

    function broadcastShares(
        uint256[2][] memory commitments,
        uint256[] memory shares
    ) external {
        require(
            shares.length == addresses.length - 1,
            "invalid number of shares"
        );
        require(isRegistered(msg.sender), "not registered");

        require(
            commitments.length == threshold(),
            "invalid number of commitments"
        );

        firstCoefficients.push(commitments[0]);
        commitmentHashes[msg.sender] = keccak256(abi.encodePacked(commitments));
        shareHashes[msg.sender] = keccak256(abi.encodePacked(shares));

        emit BroadcastSharesLog(msg.sender, participants[msg.sender].index);

        if (firstCoefficients.length == noParticipants) {
            sharesDisputableUntil = uint64(block.timestamp) + SHARES_DISPUTE_PERIOD;
            emit DistributionEndLog();
        }
    }

    function disputeShare(
        uint256 dealerIndex,
        uint256[] memory shares,
        ShareVerifier.Proof memory proof
    ) external {
        address dealer = addresses[dealerIndex];
        require(
            shareHashes[dealer] == keccak256(abi.encodePacked(shares)),
            "invalid shares"
        );

        if (firstCoefficients.length == noParticipants) {
            require(block.timestamp <= sharesDisputableUntil, "dispute period has expired");
        }

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
        bool result = shareVerifier.verifyTx(proof, input);
        emit DisputeShare(result);
    }

    function submitPublicKey(uint256[2] memory _publicKey) external {
        require(isRegistered(msg.sender), "not registered");
        require(submitter == address(0), "already submitted");
        submitter = msg.sender;
        masterPublicKey = _publicKey;
        keyDisputableUntil = uint64(block.timestamp) + KEY_DISPUTE_PERIOD;
    }

    function disputePublicKey(KeyVerifier.Proof memory proof) external {
        require(block.timestamp <= keyDisputableUntil, "dispute period has expired");

        uint256[2] memory hash = hashToUint128(
            keccak256(abi.encode(firstCoefficients))
        );
        uint256[4] memory input = [
            hash[0],
            hash[1],
            masterPublicKey[0],
            masterPublicKey[1]
        ];
        require(keyVerifier.verifyTx(proof, input), "invalid proof");
        delete masterPublicKey;
        delete submitter;
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
