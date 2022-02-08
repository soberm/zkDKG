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
                0x17bf9549e4054ff6117bf702d19da58aae3852439cd74bf5c1cf7d6e54f29002
            ),
            uint256(
                0x139aec266f34784e9a17ed64918f74975545cb1aa643019aecb1d3d07ac3f4b2
            )
        );
        vk.beta = Pairing.G2Point(
            [
                uint256(
                    0x0a6230ccf7bf44dcf8b4af487d07508739a8cf8385fb2a73f208bd120c1e28fe
                ),
                uint256(
                    0x0ad54093832f5b0f46bfb2a60fcf8cf50395c019a57c5f4ba89384db079aaace
                )
            ],
            [
                uint256(
                    0x00332c3ef704a061940cd80dfcf4b16bcca42563779e79979ca7adfc63cae898
                ),
                uint256(
                    0x2d15bc50a3df1abae4c5c283f5aa09d06c116ce5f6fbcc6e80edefcfb373bc85
                )
            ]
        );
        vk.gamma = Pairing.G2Point(
            [
                uint256(
                    0x1a8565f85030645201f9178e64abc3ff0cc47836115ee9c45cd12b5866c4708f
                ),
                uint256(
                    0x1f0581d1ef6f575e93bc040c2cbbaf96c773efad4e284bf04d26ca3f08985deb
                )
            ],
            [
                uint256(
                    0x17d1a8fb5824ecac7c1c90fce03be4c32bfd23ecee305150505a3db9fb188276
                ),
                uint256(
                    0x1a03fd04e51809938ce025f130e4498cb319cadffcf52442210d68a05b5749dd
                )
            ]
        );
        vk.delta = Pairing.G2Point(
            [
                uint256(
                    0x2aca631ac06da892e7a96b29ec29ec887276c6968ca544d8b9201de4d27f86da
                ),
                uint256(
                    0x2dc4ef58b72cc21d6746aa536fcf17e514530c04a101b64fce4f56f35f7b374a
                )
            ],
            [
                uint256(
                    0x1fb86eb7a8e84944976cb471b29308e67acd8589305681af10b63ee196969432
                ),
                uint256(
                    0x1f53cdbe111e056a20f3896459a98253ca4fbf6fcb915986b1e51711368221a2
                )
            ]
        );
        vk.gamma_abc = new Pairing.G1Point[](6);
        vk.gamma_abc[0] = Pairing.G1Point(
            uint256(
                0x2b5061f912a220fab1e22b0e2a022ac7e0be1fbae9244386a80129716ccd78d6
            ),
            uint256(
                0x0a995354dff2d2b47e16b7ea74f760766a0db55c6dd0f5a8506669cdf702c7b0
            )
        );
        vk.gamma_abc[1] = Pairing.G1Point(
            uint256(
                0x13a168e2b6c9b268da34d4903058c197cfb747bf48b8039caf9c37f3cb4a2328
            ),
            uint256(
                0x27645c8691f2b7468aede27f71c40a4c7940cbf5886d60fce9215dd9e3be480e
            )
        );
        vk.gamma_abc[2] = Pairing.G1Point(
            uint256(
                0x1373ae4764286835d7962d65ea017db07ea941c9e375874f52574c5b272b057b
            ),
            uint256(
                0x05dea6b0a496141629298bbce171e4c2cf1e4ad61be223ca4ceeeace4eb9c9c4
            )
        );
        vk.gamma_abc[3] = Pairing.G1Point(
            uint256(
                0x13fbb2eed08168a23b46201ec26b7c4d34e9ccc77e9c596452ddd76b8de40210
            ),
            uint256(
                0x0eca52114b2cfdd58554174dac48cc05da1dd85e250fb5c5ac3bf791b2602022
            )
        );
        vk.gamma_abc[4] = Pairing.G1Point(
            uint256(
                0x03f7a2eebd10c8c792a9f22593b6f886f5c110ad8c87f176d802b02ce1785036
            ),
            uint256(
                0x13017c9f2baaee63b57d2206438426964a62f470290cf7eb282a406137e41511
            )
        );
        vk.gamma_abc[5] = Pairing.G1Point(
            uint256(
                0x046be1ccbffdc29b4ce2c78929ae7f30afcf64d9b06f4f3f233ac3038ab5493b
            ),
            uint256(
                0x159e785110cfea69571f6651fd79ad9a6490ac2ae8e720ef1d379ec688e1c33e
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
