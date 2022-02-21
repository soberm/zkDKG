//SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./ShareVerifier.sol";
import "./KeyVerifier.sol";

contract ZKDKG {
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

    uint256[2] public masterPublicKey;

    ShareVerifier private shareVerifier;
    KeyVerifier private keyVerifier;

    event DisputeShare(bool result);

    constructor(address _shareVerifier, address _keyVerifier) {
        shareVerifier = ShareVerifier(_shareVerifier);
        keyVerifier = KeyVerifier(_keyVerifier);
    }

    function register(uint256[2] memory publicKey) public payable {
        require(msg.value == MIN_STAKE, "value too low");
        require(participants[msg.sender].index == 0, "already registered");
        addresses.push(msg.sender);
        participants[msg.sender] = Participant(addresses.length, publicKey);
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

    function broadcastShares(
        uint256[2][] memory commitments,
        uint256[] memory shares
    ) external {
        require(shares.length == addresses.length, "invalid number of shares");
        require(participants[msg.sender].index != 0, "not registered");

        uint256 threshold = (addresses.length + 1) / 2;

        require(
            commitments.length == threshold,
            "invalid number of commitments"
        );

        firstCoefficients.push(commitments[0]);
        commitmentHashes[msg.sender] = keccak256(abi.encodePacked(commitments));
        shareHashes[msg.sender] = keccak256(abi.encodePacked(shares)); // TODO: Store in merkle tree
    }

    function disputeShare(
        address dealer,
        uint256 share,
        ShareVerifier.Proof memory proof
    ) external {
        uint256[2] memory hash = hashToUint128(commitmentHashes[dealer]);
        uint256[5] memory input = [
            hash[0],
            hash[1],
            participants[dealer].index,
            share,
            1
        ];
        bool result = shareVerifier.verifyTx(proof, input);
        emit DisputeShare(result);
    }

    function derivePublicKey(
        uint256[2] memory _publicKey,
        KeyVerifier.Proof memory proof
    ) external {
        uint256[2] memory hash = hashToUint128(
            keccak256(abi.encode(firstCoefficients))
        );
        uint256[4] memory input = [
            hash[0],
            hash[1],
            _publicKey[0],
            _publicKey[1]
        ];
        require(keyVerifier.verifyTx(proof, input), "invalid proof");
        masterPublicKey = _publicKey;
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
