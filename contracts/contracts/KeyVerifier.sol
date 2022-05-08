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
    function verifyingKey() pure internal returns (VerifyingKey memory vk) {
        vk.alpha = Pairing.G1Point(uint256(0x062eb6de50f3430a98b38867338d95fa9485a94b22b514971c441a3d8cf9ac68), uint256(0x13f454f70e0a02ab59e99c31ced549a780a7ca79577a23702aaafd631df26e59));
        vk.beta = Pairing.G2Point([uint256(0x139f21a3ab8ec071d2b1cd5744dd0c6bb598782c9019d3228dc1138c03c5bf83), uint256(0x18fadea8a894a5fe4f8343b9c08d1b8fdad3534bdb7149d6a168d166a30c6db9)], [uint256(0x0a5a15e1cc26e4c344d3b5d1809da35a715954359c3ae5187c1ee9b307a4d631), uint256(0x300355c899cdea05a03c70073de504251259cb1488d6618148f82b1e20944bec)]);
        vk.gamma = Pairing.G2Point([uint256(0x134317c20a85de088adddac62afc90e0807883dd9779521d4d6212ba52918b90), uint256(0x1b4303c67519f6baf8f642a53db42107a6ddba8913ad7769e9bafe14d2e050d3)], [uint256(0x19a3e65ed4826aa92cc643b938dd65ae7d6f059a7606623cd1b2a2ca36f8acb3), uint256(0x1d2e6d55f55c3a71329c905d0c52b1c01ff5cfae89eaaa2228dba936ff195584)]);
        vk.delta = Pairing.G2Point([uint256(0x17c56311c406eff2bc48167cecf90d45af0fe8b67780e485fd1d6c32075fe513), uint256(0x1f0c5da3844b826ad787c5ef23b39ca20cabfa0bfdff87a2b84e1c0191e647e9)], [uint256(0x2db5095f73e196ee608b937ba3608827a4ebbf459b0dfc065a1deefe32c531e5), uint256(0x0a0bd33d641d1dc0f56e1c58d8b8d380d960ba8cf9d87503fb92be1dc3fd5328)]);
        vk.gamma_abc = new Pairing.G1Point[](6);
        vk.gamma_abc[0] = Pairing.G1Point(uint256(0x0c4b7bca36deb7d4add3a4b1245ec2d3b3cf8c6d0f84df39561eacdce2c8a78d), uint256(0x2f4c90fb6e4191d200b8ff87f58aa56f2bff7402c33fa04882a030376d727676));
        vk.gamma_abc[1] = Pairing.G1Point(uint256(0x1e2cf395d7be3bf5e3e954d1294699699bc94a279be6401e22fc9eee80ccad9e), uint256(0x0990f960a6a4e73493fdfaf644cc761f95bf86a65eb8b14de3e5cb3d69244e62));
        vk.gamma_abc[2] = Pairing.G1Point(uint256(0x04c55f83f7f562ac54eeec77162209fadd699fda6456f17a7a404d50632b62b4), uint256(0x2aea63c2ca89ab68681d6d51153ecda3e7a876b2293fa4a1d8befb6915a79ab0));
        vk.gamma_abc[3] = Pairing.G1Point(uint256(0x03f6006ba7e3988646127a3edea25bb8df608b2af81cdc23e6555fcdba321526), uint256(0x23e93b8d27a8771fcd64ef6e889cc13fc44d292d2d276f98eb79a184fb2c1c36));
        vk.gamma_abc[4] = Pairing.G1Point(uint256(0x03e43f561c2e0da01dd22a0a2e26bbb56131283d5710edf7ad74c43a8219716c), uint256(0x258f678b7325cf352e9bf41f7d14003930bd8173fd4b7d2a3bcadc5b32ccf9fd));
        vk.gamma_abc[5] = Pairing.G1Point(uint256(0x0435cc3941c3e288bbfed0964846aa3598923fedd28dabdc4ea11e1a8f1f9d19), uint256(0x15c738eeb1c2f520acf935ec423613264d357b1698ffda001e9d808d089ce9bd));
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
            Proof memory proof, uint[5] memory input
        ) public view returns (bool r) {
        uint[] memory inputValues = new uint[](5);
        
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
