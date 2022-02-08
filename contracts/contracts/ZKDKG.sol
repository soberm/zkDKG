//SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./Verifier.sol";

contract ZKDKG {
    uint256 public constant MIN_STAKE = 0 ether;

    mapping(address => uint256) private indices;
    address[] private addresses;
    mapping(address => bytes32) public commitmentHashes;
    mapping(address => bytes32) private shareHashes;

    Verifier private verifier;

    event DisputeShare(bool result);

    constructor(address _verifier) {
        verifier = Verifier(_verifier);
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
        commitmentHashes[msg.sender] = keccak256(abi.encodePacked(commitments));
        shareHashes[msg.sender] = keccak256(abi.encodePacked(shares)); // TODO: Store in merkle tree
    }

    function disputeShare(
        address dealer,
        uint256 share,
        Verifier.Proof memory proof
    ) external {
        uint256[2] memory hash = hashToUint128(commitmentHashes[dealer]);
        uint256[5] memory input = [
            hash[0],
            hash[1],
            indices[dealer] + 1,
            share,
            1
        ];
        bool result = verifier.verifyTx(proof, input);
        emit DisputeShare(result);
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
