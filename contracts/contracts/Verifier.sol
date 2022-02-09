// This file is MIT Licensed.
//
// Copyright 2017 Christian Reitwiessner
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
pragma solidity ^0.8.0;

library Pairing {
    struct G1Point {
        uint256 X;
        uint256 Y;
    }
    // Encoding of field elements is: X[0] * z + X[1]
    struct G2Point {
        uint256[2] X;
        uint256[2] Y;
    }

    /// @return the generator of G1
    function P1() internal pure returns (G1Point memory) {
        return G1Point(1, 2);
    }

    /// @return the generator of G2
    function P2() internal pure returns (G2Point memory) {
        return
            G2Point(
                [
                    10857046999023057135944570762232829481370756359578518086990519993285655852781,
                    11559732032986387107991004021392285783925812861821192530917403151452391805634
                ],
                [
                    8495653923123431417604973247489272438418190587263600148770280649306958101930,
                    4082367875863433681332203403145435568316851327593401208105741076214120093531
                ]
            );
    }

    /// @return the negation of p, i.e. p.addition(p.negate()) should be zero.
    function negate(G1Point memory p) internal pure returns (G1Point memory) {
        // The prime q in the base field F_q for G1
        uint256 q = 21888242871839275222246405745257275088696311157297823662689037894645226208583;
        if (p.X == 0 && p.Y == 0) return G1Point(0, 0);
        return G1Point(p.X, q - (p.Y % q));
    }

    /// @return r the sum of two points of G1
    function addition(G1Point memory p1, G1Point memory p2)
        internal
        view
        returns (G1Point memory r)
    {
        uint256[4] memory input;
        input[0] = p1.X;
        input[1] = p1.Y;
        input[2] = p2.X;
        input[3] = p2.Y;
        bool success;
        assembly {
            success := staticcall(sub(gas(), 2000), 6, input, 0xc0, r, 0x60)
            // Use "invalid" to make gas estimation work
            switch success
            case 0 {
                invalid()
            }
        }
        require(success);
    }

    /// @return r the product of a point on G1 and a scalar, i.e.
    /// p == p.scalar_mul(1) and p.addition(p) == p.scalar_mul(2) for all points p.
    function scalar_mul(G1Point memory p, uint256 s)
        internal
        view
        returns (G1Point memory r)
    {
        uint256[3] memory input;
        input[0] = p.X;
        input[1] = p.Y;
        input[2] = s;
        bool success;
        assembly {
            success := staticcall(sub(gas(), 2000), 7, input, 0x80, r, 0x60)
            // Use "invalid" to make gas estimation work
            switch success
            case 0 {
                invalid()
            }
        }
        require(success);
    }

    /// @return the result of computing the pairing check
    /// e(p1[0], p2[0]) *  .... * e(p1[n], p2[n]) == 1
    /// For example pairing([P1(), P1().negate()], [P2(), P2()]) should
    /// return true.
    function pairing(G1Point[] memory p1, G2Point[] memory p2)
        internal
        view
        returns (bool)
    {
        require(p1.length == p2.length);
        uint256 elements = p1.length;
        uint256 inputSize = elements * 6;
        uint256[] memory input = new uint256[](inputSize);
        for (uint256 i = 0; i < elements; i++) {
            input[i * 6 + 0] = p1[i].X;
            input[i * 6 + 1] = p1[i].Y;
            input[i * 6 + 2] = p2[i].X[1];
            input[i * 6 + 3] = p2[i].X[0];
            input[i * 6 + 4] = p2[i].Y[1];
            input[i * 6 + 5] = p2[i].Y[0];
        }
        uint256[1] memory out;
        bool success;
        assembly {
            success := staticcall(
                sub(gas(), 2000),
                8,
                add(input, 0x20),
                mul(inputSize, 0x20),
                out,
                0x20
            )
            // Use "invalid" to make gas estimation work
            switch success
            case 0 {
                invalid()
            }
        }
        require(success);
        return out[0] != 0;
    }

    /// Convenience method for a pairing check for two pairs.
    function pairingProd2(
        G1Point memory a1,
        G2Point memory a2,
        G1Point memory b1,
        G2Point memory b2
    ) internal view returns (bool) {
        G1Point[] memory p1 = new G1Point[](2);
        G2Point[] memory p2 = new G2Point[](2);
        p1[0] = a1;
        p1[1] = b1;
        p2[0] = a2;
        p2[1] = b2;
        return pairing(p1, p2);
    }

    /// Convenience method for a pairing check for three pairs.
    function pairingProd3(
        G1Point memory a1,
        G2Point memory a2,
        G1Point memory b1,
        G2Point memory b2,
        G1Point memory c1,
        G2Point memory c2
    ) internal view returns (bool) {
        G1Point[] memory p1 = new G1Point[](3);
        G2Point[] memory p2 = new G2Point[](3);
        p1[0] = a1;
        p1[1] = b1;
        p1[2] = c1;
        p2[0] = a2;
        p2[1] = b2;
        p2[2] = c2;
        return pairing(p1, p2);
    }

    /// Convenience method for a pairing check for four pairs.
    function pairingProd4(
        G1Point memory a1,
        G2Point memory a2,
        G1Point memory b1,
        G2Point memory b2,
        G1Point memory c1,
        G2Point memory c2,
        G1Point memory d1,
        G2Point memory d2
    ) internal view returns (bool) {
        G1Point[] memory p1 = new G1Point[](4);
        G2Point[] memory p2 = new G2Point[](4);
        p1[0] = a1;
        p1[1] = b1;
        p1[2] = c1;
        p1[3] = d1;
        p2[0] = a2;
        p2[1] = b2;
        p2[2] = c2;
        p2[3] = d2;
        return pairing(p1, p2);
    }
}

contract Verifier {
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
                0x1ef94c309140c47db7455104899d3ef95ceb913408575d9083c8b9f7363c851a
            ),
            uint256(
                0x23fe66c2ce76cbd35b7ff14c517a847fcd92e663664d020ea11142bed146ab74
            )
        );
        vk.beta = Pairing.G2Point(
            [
                uint256(
                    0x2c822b5344e43a6f2edfbd63f2f3db33ea48e9f87c19ee82614e25ef17b03b8f
                ),
                uint256(
                    0x170af333b99403e79d5ce96a7dad802ad186484cdad8a49bff856bef8eda2d4b
                )
            ],
            [
                uint256(
                    0x22ef2eda48d1665c357932691a7804432cd06cb0de3f2893542df728bab6a494
                ),
                uint256(
                    0x1612e6333d189c3d40c131893846033e63780e6817535294de6df03670208667
                )
            ]
        );
        vk.gamma = Pairing.G2Point(
            [
                uint256(
                    0x0c5a2d552d1ec6dccc658033955794874b0f58ff4cf5a14f8cfa64e4818e4a10
                ),
                uint256(
                    0x1079dd0da2773121e345404a0035e5d2d73b137dc608fb7aa08aed3714767628
                )
            ],
            [
                uint256(
                    0x1074f80322543606a6d01e714ec7e76564db960ad6e38d4d879d9b20c9ebb2bd
                ),
                uint256(
                    0x012a74399bb12da154155319a18b60810131c8d0730fb6d38236831e3d499192
                )
            ]
        );
        vk.delta = Pairing.G2Point(
            [
                uint256(
                    0x188d9d327514ded6866fd4e0c7631ff1abcb488e24272ebd001af37afa81bf2f
                ),
                uint256(
                    0x13d57f3eeda37aa7bec2bba95b73bb4f4df8040f71818dd2fc3028539e4c577f
                )
            ],
            [
                uint256(
                    0x26af6457eff9232a8aa1565889380497cebd599c6ccd6d157dc31750a2baa87b
                ),
                uint256(
                    0x149351804f26415f26bbb00576368da32fbbd6d4654ffecb5d50164d989ce958
                )
            ]
        );
        vk.gamma_abc = new Pairing.G1Point[](6);
        vk.gamma_abc[0] = Pairing.G1Point(
            uint256(
                0x18ab7647233b2738283198f3ce0af20266f42c241b1256972d2692062ed7e481
            ),
            uint256(
                0x0aa8edb67889da4fd139c3e645f0286164c4a85742a6d3459ec824eb9f5913fc
            )
        );
        vk.gamma_abc[1] = Pairing.G1Point(
            uint256(
                0x243bd8e003b495ddf298a2c217f7c27091cbb08164598ee7e6a8952b1cb63073
            ),
            uint256(
                0x078bcc1029d390c29296d2b4fd94f1470575cea7dea8c81bba35563eb82b5d41
            )
        );
        vk.gamma_abc[2] = Pairing.G1Point(
            uint256(
                0x0e1c43511ffddb7abbad0465ff7b48d523e79f540cf96e7f69a5deb75838c96e
            ),
            uint256(
                0x157ff4454f9e439d314752c7ac6a19d1bc1af481a4e029e4084795edd604b83b
            )
        );
        vk.gamma_abc[3] = Pairing.G1Point(
            uint256(
                0x0e927fca429b48c2999cf71ff175b05677164faf3997d81749749989ff5dc6bb
            ),
            uint256(
                0x1c7417a3b908e8566a2ddeaa8c58c2ddceeb2b9b7cc870aa73f1aa4c7d68bb54
            )
        );
        vk.gamma_abc[4] = Pairing.G1Point(
            uint256(
                0x029b32f5984a1bf6313c2db7ba55ad19375e7c90aaecdc167b6cb28309865949
            ),
            uint256(
                0x2ff815f0490353f58ee5bb8d875b5da2b09943d9dd2db88b68978c2d6a70adde
            )
        );
        vk.gamma_abc[5] = Pairing.G1Point(
            uint256(
                0x2ad0ec4bb074d843a50449a2cae2d7b881c3169d0d52b012968d14d3ace94b42
            ),
            uint256(
                0x07a83cb3543d6562b60c8ce74c921f0b7b48558f0b55546ffe2d49df617ee42b
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
