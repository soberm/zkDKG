package main

import (
	"client/pkg/dkg"
	"fmt"

	"go.dedis.ch/kyber/v3/group/curve25519"
	"go.dedis.ch/kyber/v3/share"
)

func main() {
	curve := &curve25519.ProjectiveCurve{}
	curve.Init(dkg.ParamBabyJubJub(), false)

	suite := curve25519.SuiteCurve25519{ProjectiveCurve: *curve}

	base := suite.Point().Base()
	value := suite.Scalar().SetInt64(3)

	commit := suite.Point().Mul(value, base)

	test := suite.Point().Add(base, commit)

	fmt.Printf("%+v\n", test)

	priPoly := share.NewPriPoly(&suite, 2, value, suite.RandomStream())
	fmt.Printf("%+v\n", priPoly)

	pubPoly := priPoly.Commit(nil)
	_, commits := pubPoly.Info()
	for _, point := range commits {
		fmt.Printf("Commitment: %+v\n", point)
	}
	share := priPoly.Eval(0)
	fmt.Printf("Share: %+v", share.V.String())
}