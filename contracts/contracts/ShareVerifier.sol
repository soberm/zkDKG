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

    function verifyingKey() internal pure returns (VerifyingKey memory vk) {
        vk.alpha = Pairing.G1Point(
            uint256(
                0x1afbc55a0b7536d5cd87f5f11c2ca19b81bf224a641ae92ac9c6610bfc153642
            ),
            uint256(
                0x18e4ea541b23db7a845852abaa5c80b1e8895121a09342db7944d443fb412a08
            )
        );
        vk.beta = Pairing.G2Point(
            [
                uint256(
                    0x075acbc5eba1308dadfd44c937d943d3cfcad40266c4d7e3ecad433eac383a08
                ),
                uint256(
                    0x0bf7a6a3f7e37626a1f1e3c629d03afbde741082b117978861459f299b49a1eb
                )
            ],
            [
                uint256(
                    0x18bd5cf214b1e8b61dcddf1cb8acaa50d4490a23dd7d330aa1e57f5846acc219
                ),
                uint256(
                    0x0d9b9fd6676c54dae268aa7e6a81111a2df31b1e9b3367690d95ccb75fda10db
                )
            ]
        );
        vk.gamma = Pairing.G2Point(
            [
                uint256(
                    0x20113342f738ac96ad8a33f40942b131d2d2a41f3a3254ea6c5b5e8287058d43
                ),
                uint256(
                    0x2d48a5791b6f46e3dc456b5cb5c9ed3488c9aaf3a3905bcdb7ee236d4713df51
                )
            ],
            [
                uint256(
                    0x1a1bb03cc8d716566ed8be8833a31949ff4c421c35d5bfd2ea82a230a05dfdb0
                ),
                uint256(
                    0x26749d92288dcf276b246f38053c48d96ff37329245fdc57bf7f62ffab45be2d
                )
            ]
        );
        vk.delta = Pairing.G2Point(
            [
                uint256(
                    0x2c673ebc438e40e55b12a62b5b4e1f3c5b23ca2a59ae80a2c6522bf1a4a67731
                ),
                uint256(
                    0x072b0235cf0ad0cd820deb63aa7fe1249a6eaaa9f29b752c8d37d5802ae91fae
                )
            ],
            [
                uint256(
                    0x24409a64c9ce12f4e49581a0d1f66eea7a4288c12178be584021ef67167c272f
                ),
                uint256(
                    0x01946ab111efdf6de54c64c550b75098ef4f569582f82b9cce27fa0435429075
                )
            ]
        );
        vk.gamma_abc = new Pairing.G1Point[](6);
        vk.gamma_abc[0] = Pairing.G1Point(
            uint256(
                0x1e40a7c0d1faf23d8eca40f3bd0efa3825852497b8b71a6f66af87b330c4e827
            ),
            uint256(
                0x03655b916baadee5352759c52c889b86aa3069c30c48d66d86b90dbc937763e3
            )
        );
        vk.gamma_abc[1] = Pairing.G1Point(
            uint256(
                0x03fbaf8f839627f703777a3938486d86c51f92650107a1e863b15a9f5eaa277b
            ),
            uint256(
                0x1f07f7979b2e6c7163b428a10e9352b88b1bd7410b42d4ea049905203a740665
            )
        );
        vk.gamma_abc[2] = Pairing.G1Point(
            uint256(
                0x038a67c791dde9efc7265d9f1ea10a6951f04e1a856f226da49247078026d103
            ),
            uint256(
                0x26d8d7e86e4e561253f9666ce56160cf5578ddf96b437ff675e98aed426db0f0
            )
        );
        vk.gamma_abc[3] = Pairing.G1Point(
            uint256(
                0x073d07faf44e9cbd0248ce7a6c957008923c250ee21c757e82d7b3423ded8fdf
            ),
            uint256(
                0x06ec76e348f3aa96abb66084c64efb8af767e8be58c262d7912ad760209b0b64
            )
        );
        vk.gamma_abc[4] = Pairing.G1Point(
            uint256(
                0x29851bff020d7d6eb46da34e235b77da475980eb9cd80f0f97e29d1a9bb76500
            ),
            uint256(
                0x1302dcaf78c1e589cdba75eea68dc2ce9b450c43804ff629f1be1df271ccb6a5
            )
        );
        vk.gamma_abc[5] = Pairing.G1Point(
            uint256(
                0x1b6bd7a96e57da95172f6fbe3cc55e755d5884c11437029445953105c1d37062
            ),
            uint256(
                0x1d29383a1eed81838b08a67138a42df7b088fc1fb27abab0afbab045a65656fa
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

    function verifyTx(Proof memory proof, uint256[5] memory input)
        public
        view
        returns (bool r)
    {
        uint256[] memory inputValues = new uint256[](5);

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
