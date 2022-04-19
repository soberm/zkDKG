// This file is MIT Licensed.
//
// Copyright 2017 Christian Reitwiessner
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
pragma solidity ^0.8.0;

import "./Pairing.sol";

contract ShareVerifier {
    using Pairing for *;
    struct VerifyingKey {
        Pairing.G1Point alpha;
        Pairing.G2Point beta;
        Pairing.G2Point gamma;
        Pairing.G2Point delta;
        Pairing.G1Point[] gamma_abc;
    }
    struct Proof {
        Pairing.G1Point a;
        Pairing.G2Point b;
        Pairing.G1Point c;
    }
    function verifyingKey() pure internal returns (VerifyingKey memory vk) {
        vk.alpha = Pairing.G1Point(uint256(0x0bfebcfe1e1059c2382030e655ff68c44f2a16c246386a3f9bad5a5175a6ee27), uint256(0x2e6a34cc8b7a9fc1ab4b3445d5c75e28d8dc6e407acfebf792348a89d2d654fe));
        vk.beta = Pairing.G2Point([uint256(0x19f3601ed386b19d4db89d6f19a02d4fc8b5dce077f9d58ddf3c9b0e67f9fac2), uint256(0x005333411fba9c0f6349c2a74d3046c4e46c1cc92842c89e4728b9d708af26db)], [uint256(0x0d841531fbbc197829d7e8ccf52b5b75fde24d534659e54c2e11c83721130d44), uint256(0x28bf91ffd58237870e5479a585a454621334e91ea951594236be44d9573d95c8)]);
        vk.gamma = Pairing.G2Point([uint256(0x1185d30adc06a2c37b073d6923e8f306b394e95fa7c9cce177dfc1e91577407a), uint256(0x06d792b298fba673d558a90ac8b8a502a4876b8f350ce3b56fe4b478fc3af067)], [uint256(0x1198d6b347f461488c820e11fa8cb4539b38f815d3d64d7ce4e95085ec650f8d), uint256(0x1e85c9345d2523b62e6a5e0e8f5ccd4a961c775ee4aae4eddc4a66162c1e5583)]);
        vk.delta = Pairing.G2Point([uint256(0x0a8b1cb01a0691b375ebc9ad47729859620102e601bda1836a0d642b3daa6a43), uint256(0x1505b869f223c0d7549ea0047c006c205a2a19ad65f33240c336e6bac76edb9f)], [uint256(0x10892d562bb51f9719d31bb61f7f7cac23f0e8ae59375605625e326307237481), uint256(0x1004afe1a2372b6d0354551679da47a27d220a5158c429688ec7c7c6f67778ae)]);
        vk.gamma_abc = new Pairing.G1Point[](4);
        vk.gamma_abc[0] = Pairing.G1Point(uint256(0x26758100aef294054b422b407d9419acb1f0ee6fe807c8fca660f82dc24ff342), uint256(0x020e70fd7f32256d89ae5f40f78bcd155a9f92ab6a00a45f5b9ffb762f3ed88a));
        vk.gamma_abc[1] = Pairing.G1Point(uint256(0x1f23c6354f34bdd8370611f865ccc572af94a9a3b83e3ab7c05e3a78d6339779), uint256(0x06cc954cbd1f439f8cb8f7bf35eedf17d8681343c7dbdcc72efaeaf8547438b3));
        vk.gamma_abc[2] = Pairing.G1Point(uint256(0x1f489e609026f867aed85ace25b550d29da4164d6d522cc2ec39a00dd7f51d72), uint256(0x15e329ed5738d4fa1f90a37eece1beb57d51e903a5da8b6881cd873671f03849));
        vk.gamma_abc[3] = Pairing.G1Point(uint256(0x10bcdad485962fda3f2a433c055be2b143700fdb341a183c764bebad43eafe3b), uint256(0x1b9cb65348dd666102f88450324a5ce7bb08efe8f6a5c17b967c826c768c7b71));
    }
    function verify(uint[] memory input, Proof memory proof) internal view returns (uint) {
        uint256 snark_scalar_field = 21888242871839275222246405745257275088548364400416034343698204186575808495617;
        VerifyingKey memory vk = verifyingKey();
        require(input.length + 1 == vk.gamma_abc.length);
        // Compute the linear combination vk_x
        Pairing.G1Point memory vk_x = Pairing.G1Point(0, 0);
        for (uint i = 0; i < input.length; i++) {
            require(input[i] < snark_scalar_field);
            vk_x = Pairing.addition(vk_x, Pairing.scalar_mul(vk.gamma_abc[i + 1], input[i]));
        }
        vk_x = Pairing.addition(vk_x, vk.gamma_abc[0]);
        if(!Pairing.pairingProd4(
             proof.a, proof.b,
             Pairing.negate(vk_x), vk.gamma,
             Pairing.negate(proof.c), vk.delta,
             Pairing.negate(vk.alpha), vk.beta)) return 1;
        return 0;
    }
    function verifyTx(
            Proof memory proof, uint[3] memory input
        ) public view returns (bool r) {
        uint[] memory inputValues = new uint[](3);
        
        for(uint i = 0; i < input.length; i++){
            inputValues[i] = input[i];
        }
        if (verify(inputValues, proof) == 0) {
            return true;
        } else {
            return false;
        }
    }
}
