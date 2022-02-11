//SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./ShareVerifier.sol";
import "./KeyVerifier.sol";

contract ZKDKG {
    uint256 public constant MIN_STAKE = 0 ether;

    mapping(address => uint256) private indices;
    address[] private addresses;
    mapping(address => bytes32) public commitmentHashes;
    mapping(address => bytes32) private shareHashes;
    uint256[2][] private firstCoefficients;

    uint256[2] private publicKey;

    ShareVerifier private shareVerifier;
    KeyVerifier private keyVerifier;

    event DisputeShare(bool result);

    constructor(address _shareVerifier, address _keyVerifier) {
        shareVerifier = ShareVerifier(_shareVerifier);
        keyVerifier = KeyVerifier(_keyVerifier);
    }

    function register() public payable {
        require(msg.value == MIN_STAKE, "value too low");
        indices[msg.sender] = addresses.length;
        addresses.push(msg.sender);
    }

    function broadcastShares(
        uint256[2][] memory commitments,
        uint256[] memory shares
    ) external {
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
            indices[dealer] + 1,
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
        publicKey = _publicKey;
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
