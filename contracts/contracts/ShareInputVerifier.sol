// This file is MIT Licensed.
//
// Copyright 2017 Christian Reitwiessner
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
pragma solidity ^0.8.0;

import "./Pairing.sol";

contract ShareInputVerifier {
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
        vk.alpha = Pairing.G1Point(uint256(0x21d62715eb2b6e8a3d21d060a67274f704198d8a99030081c498e88d19f5f436), uint256(0x0833ff03319fef14160d6ed71f83a6fef80ced8ffd1ef98bd27eaf41ab76a07f));
        vk.beta = Pairing.G2Point([uint256(0x1950cbff300ce9e5f4dd316205a9ce52f852a53022b1e784912b01e1d6ac7360), uint256(0x0663e37c86a97f7ca05cf8ab15fd465c41e2bde66f59dda6c6780d0c886fc062)], [uint256(0x2935618426f9567b39d1421deeff69891f10ff92cd39cdbeb6777d798fb57015), uint256(0x219485d996b26d1fffeb55d3be3c79dd3df8df9dea39eea74469e4f7db4940af)]);
        vk.gamma = Pairing.G2Point([uint256(0x0766ab8b866de17992592b42cb1da38cdaeeee2f17cf7cdceb3dd39e2893a95d), uint256(0x1adf0bb1fdd100633364f21aa995963f71c854df24c8cbb06a1f2ff39d7e6b3b)], [uint256(0x264c838d900f580d6cd6ba52be896051dc43b6a7bce74f47066aa517152630be), uint256(0x1f4a8b511c85a15aceb272027a5682c64099b948e6bee6c4ec8de08954bc756d)]);
        vk.delta = Pairing.G2Point([uint256(0x26f02cba61b321b7a5cd235ab2a26efc97d57ba600dc7035962b680e286862f5), uint256(0x071ac441114e30d119e34103ae71be46e63565f38329e23bbc2698b17e8da422)], [uint256(0x1b007aed20d27e5d22b244a543257f511cbced2834bb85f3271d90b6395fec8a), uint256(0x1c244c294352243b5f10479b18955fb07ebe61bc422dad1b3a943091f8c4917d)]);
        vk.gamma_abc = new Pairing.G1Point[](4);
        vk.gamma_abc[0] = Pairing.G1Point(uint256(0x19e9ba854e06bbbdb8457769cdc7655fd4731e0f9af4518617188e2c0dc18dd4), uint256(0x265fa67bc059c7f16f9ebb80f2a7fd41eb4438be9a7e7e322a53c7325b9602a5));
        vk.gamma_abc[1] = Pairing.G1Point(uint256(0x2d102d846c17ae86978089be8ce7eec6e078111c9163b6fffae24cb79295dce8), uint256(0x1192b524a198714311acb74f5f430bf4ed4822307c061d0674ee03b08f01992c));
        vk.gamma_abc[2] = Pairing.G1Point(uint256(0x2440c6886f0ada11b5cb6c7468692d53b213af87fb78ab7d9bf2b8006c4b114c), uint256(0x17247e10a9519edccf60db1b1a51bb38eb3894c9ba87c3fb90caaa2601a18b5e));
        vk.gamma_abc[3] = Pairing.G1Point(uint256(0x1fe452d7b45a74e795c4dfa8f155e831b8edf2046e74957048b2e7e665386971), uint256(0x0adafb47158787822462b871b4d6a1e5f7827c684d8237faf4a20bf7b184ee51));
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