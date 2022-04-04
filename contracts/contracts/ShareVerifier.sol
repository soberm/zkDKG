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
                0x037a2cfd1377bf2b74eb8ec572a0505e597e46d9d59e023cf6e9f933e67f04ca
            ),
            uint256(
                0x2b2ed48170d00dc49d50a6f295ecb58e16c6cbdace74d3b9a71dea2b33dbdcdf
            )
        );
        vk.beta = Pairing.G2Point(
            [
                uint256(
                    0x03862fae092ea23bd77721628101fb9763e98adebe51e7515134e31645df2d70
                ),
                uint256(
                    0x0ebe1a51a21d24b9568e0b04009972cf31e2ddb233877a64edecb23f9885a127
                )
            ],
            [
                uint256(
                    0x1a22d5bcfed8804a36ed54eb7a943fd0a99d68c755257eaf3eface4743ec8831
                ),
                uint256(
                    0x1e61e5e2af2416ca4ffd43f1ea81c0c3e21e9d31d05e1ed4fabb6da90a4d4980
                )
            ]
        );
        vk.gamma = Pairing.G2Point(
            [
                uint256(
                    0x200f6b858639643b251fb2d673b6c0c5a665cea1503e8f49ad17c40e106f30ef
                ),
                uint256(
                    0x1ff5624534c42e25d0748bbfe9fad0c57cb8efe4b5c9e04c8be7920b284c3b08
                )
            ],
            [
                uint256(
                    0x2d38498623db1873c77aab462e958649be0091f6591f4db90ca4d411d041ae7c
                ),
                uint256(
                    0x2fb0d517b98eb5123f8c0f1e378c345b0580ddd69b0a980e35f4b5785f9c572e
                )
            ]
        );
        vk.delta = Pairing.G2Point(
            [
                uint256(
                    0x26c8db0c7ba642f70a40e5b309cb168ea12fe16742cfbd0b7951d4c28ef950f0
                ),
                uint256(
                    0x26c7bf1be26a4ea4c2ae4e997a95a6eb066b8dc32ea61afd6b08ed1c6670afd0
                )
            ],
            [
                uint256(
                    0x1522f25182cf84b0a14105715615b8a083f3d127c4ccafb470196801578bf775
                ),
                uint256(
                    0x2de78ba4e313f413ffec666bb0a9fbb37a4d6bbbf11ba8f9c030631cb4f9dfe6
                )
            ]
        );
        vk.gamma_abc = new Pairing.G1Point[](10);
        vk.gamma_abc[0] = Pairing.G1Point(
            uint256(
                0x04ca190335456ad74429b906540a4353f45e071790c08b500168fbde8156c439
            ),
            uint256(
                0x090ffaef86c8f0295faf57c7c7fb21083936f702bff1b190fa9189462373f7e9
            )
        );
        vk.gamma_abc[1] = Pairing.G1Point(
            uint256(
                0x0ebc25188505c2993787c985207ed0aac7e8845ac8478a32640e4b9a99b0477c
            ),
            uint256(
                0x2cdb27afbadf35f8872429fe6bf44e9b9ab7784900655425549e881b17d2d108
            )
        );
        vk.gamma_abc[2] = Pairing.G1Point(
            uint256(
                0x2b2e389b307856b1c24f61c4dc6a6f3006f2e77d368adb1d3ebb50e94e82e122
            ),
            uint256(
                0x26eb69c225999fb0bcdb916c266083395ef3f162331ed5e296111e048983e198
            )
        );
        vk.gamma_abc[3] = Pairing.G1Point(
            uint256(
                0x00621eec37b7407d39f03f2132a95160cfedd48e8988aac61eeb7beabd94fc41
            ),
            uint256(
                0x1bc1dbe9d8ca5a0418be269f8277b3174688f39404c878f01f47b1653d4503a1
            )
        );
        vk.gamma_abc[4] = Pairing.G1Point(
            uint256(
                0x2c9f1af90887d81c48016bde8c37645d62ac964e73de1598edf773796fa354ba
            ),
            uint256(
                0x1e0cf5921ebbfba5e70c233546d3116ddf3030dc14b10737f5056af5402ce95b
            )
        );
        vk.gamma_abc[5] = Pairing.G1Point(
            uint256(
                0x2e380d144fb3e7eda8ba2aa2f4b91cb7afacaffac9df11ea61514b5d38fc0d54
            ),
            uint256(
                0x0fff23557eab45d69b155910b2e0a192a338d64fb8749d56eb745db0b5083ef3
            )
        );
        vk.gamma_abc[6] = Pairing.G1Point(
            uint256(
                0x22ff969c993c6c6102fdcd834b9deae200a7997a65612e424256b6b6697835a4
            ),
            uint256(
                0x2be90be708ce067d4e92d99f621f524828a10c4473567a6e47b85861d8581048
            )
        );
        vk.gamma_abc[7] = Pairing.G1Point(
            uint256(
                0x14fc5fccb26a37fccbfe5ca782b737ffb4ecfd04235c2d70df54fe962bb850a4
            ),
            uint256(
                0x035daaeb2b8e1939e2898c1406e2db75fd3c81f81cc92fc8ff00b710989979b5
            )
        );
        vk.gamma_abc[8] = Pairing.G1Point(
            uint256(
                0x21dda2d02e1e28fc147579b6637e407413a3815883a02516976b9219feb5ed0a
            ),
            uint256(
                0x2652e8639216b22c36dcf0146144e3a4709bce77bebee9a6420342de7d271101
            )
        );
        vk.gamma_abc[9] = Pairing.G1Point(
            uint256(
                0x0fed1e02325917f6455c5659e1d7f40027e720df8cce6471070ceba386dcb87b
            ),
            uint256(
                0x2a5b50d41f6a2e1ee3e10bf81afb6f69e510c4797e1a0be7c439357922a2c761
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

    function verifyTx(Proof memory proof, uint256[9] memory input)
        public
        view
        returns (bool r)
    {
        uint256[] memory inputValues = new uint256[](9);

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
