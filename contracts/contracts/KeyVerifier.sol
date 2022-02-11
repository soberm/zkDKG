// This file is MIT Licensed.
//
// Copyright 2017 Christian Reitwiessner
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
pragma solidity ^0.8.0;

import "./Pairing.sol";

contract KeyVerifier {
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

    function verifyingKey() internal pure returns (VerifyingKey memory vk) {
        vk.alpha = Pairing.G1Point(
            uint256(
                0x1cc14a73323bbccfd680a8fc1e3da4dfbd2d1cdac2a172a722a17a3ba6881bce
            ),
            uint256(
                0x13950099efd5144064868abf1035361e11c04a0f85ed4bebfc3a2ef08f273329
            )
        );
        vk.beta = Pairing.G2Point(
            [
                uint256(
                    0x04459e62d9bad6985007943a90ef5b058dfc924d458908ff8eecfb6bcda67b39
                ),
                uint256(
                    0x11b7f4a35330ad54db55fd1484f6c299ccad56047568162fe3d035dc0d59b770
                )
            ],
            [
                uint256(
                    0x20872917071c2a583972e13be5ba2e900215b6c6c3114d957380115b0d5aee64
                ),
                uint256(
                    0x18e8e86e11326d786352e6b491a52c25584a6b93ae6c54203f8fbd54f6e75e27
                )
            ]
        );
        vk.gamma = Pairing.G2Point(
            [
                uint256(
                    0x0869756dcdf24d3bfc5adbac714c9895d744a7f7f18d73a09cd8a55e9de6ab25
                ),
                uint256(
                    0x229ccff28463d8899dbaa5e0f713f78a4cef0748af921cb703ab7c96ac85619f
                )
            ],
            [
                uint256(
                    0x11928d46cb21859a5036f55ae3d3a700add35bd245e707ea20e639e027e95398
                ),
                uint256(
                    0x0362c9cd10c01264e7dcee1ac9f319a98c73b5ee96c5803ec919c0b2382ba500
                )
            ]
        );
        vk.delta = Pairing.G2Point(
            [
                uint256(
                    0x0dbc15d53a4c7e79541285f1f245da9e3fe8dbf7e2961eb1419009bfb8c7fb16
                ),
                uint256(
                    0x04cc809aae306cba6c0da3bfe78588023d965e828b56676d0df8543fad8ceebe
                )
            ],
            [
                uint256(
                    0x2ff43f4200d3081064194ee1d9a9269f2ff0d5bbe6e2570cb9cc2c2f5b41d536
                ),
                uint256(
                    0x1d8b7d5869eae8938f93e1081bb19eeac5ced2aa97a9d42aa4576dcb50789b5f
                )
            ]
        );
        vk.gamma_abc = new Pairing.G1Point[](5);
        vk.gamma_abc[0] = Pairing.G1Point(
            uint256(
                0x25be5d023c8dc824c7b6098151f1f43b255346f65f5b2690493ce85eb9dbda21
            ),
            uint256(
                0x0a4eee5f2169b0f2b70f74dbede14b52ea9cfa1f431e331037a9e9b40131376c
            )
        );
        vk.gamma_abc[1] = Pairing.G1Point(
            uint256(
                0x1497fef31d3ccf9c49452656ff0e92081498f627218f75314eb7f91e495285a2
            ),
            uint256(
                0x03c36c1a1ad66c814d137808b8c51cf4be03f2ab75b1aeaf7d80b05dc2e7021b
            )
        );
        vk.gamma_abc[2] = Pairing.G1Point(
            uint256(
                0x1f433603c021274b91fea2ae3a3924cab60d913b6707ebadce52f86027a54d8f
            ),
            uint256(
                0x15df70fad3c62e448f74719439a1d2b719c8e396ecb93c0c4251fed044f0de3f
            )
        );
        vk.gamma_abc[3] = Pairing.G1Point(
            uint256(
                0x1935b8fa9274b5ed212483e4fc1b578028bd61d157dce556533d56938067a0e5
            ),
            uint256(
                0x239d039416e1679d0aa6acb3accc2effb29430dcf7c73057396ea2427af28226
            )
        );
        vk.gamma_abc[4] = Pairing.G1Point(
            uint256(
                0x20c85d8a7599f9c70946edfe11ca683fd2632e886b5357cfaf8eb98c4badae6b
            ),
            uint256(
                0x17f8d3fe11412f2ef3e8aa0cfd6aca3a6fdff0080cfd381bb86a077092a76152
            )
        );
    }

    function verify(uint256[] memory input, Proof memory proof)
        internal
        view
        returns (uint256)
    {
        uint256 snark_scalar_field = 21888242871839275222246405745257275088548364400416034343698204186575808495617;
        VerifyingKey memory vk = verifyingKey();
        require(input.length + 1 == vk.gamma_abc.length);
        // Compute the linear combination vk_x
        Pairing.G1Point memory vk_x = Pairing.G1Point(0, 0);
        for (uint256 i = 0; i < input.length; i++) {
            require(input[i] < snark_scalar_field);
            vk_x = Pairing.addition(
                vk_x,
                Pairing.scalar_mul(vk.gamma_abc[i + 1], input[i])
            );
        }
        vk_x = Pairing.addition(vk_x, vk.gamma_abc[0]);
        if (
            !Pairing.pairingProd4(
                proof.a,
                proof.b,
                Pairing.negate(vk_x),
                vk.gamma,
                Pairing.negate(proof.c),
                vk.delta,
                Pairing.negate(vk.alpha),
                vk.beta
            )
        ) return 1;
        return 0;
    }

    function verifyTx(Proof memory proof, uint256[4] memory input)
        public
        view
        returns (bool r)
    {
        uint256[] memory inputValues = new uint256[](4);

        for (uint256 i = 0; i < input.length; i++) {
            inputValues[i] = input[i];
        }
        if (verify(inputValues, proof) == 0) {
            return true;
        } else {
            return false;
        }
    }
}
