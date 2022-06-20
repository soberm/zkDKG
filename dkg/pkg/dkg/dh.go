package dkg

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"
)

func DhExchange(suite suites.Suite, ownPrivate kyber.Scalar, remotePublic kyber.Point) kyber.Point {
	sk := suite.Point()
	sk.Mul(ownPrivate, remotePublic)
	return sk
}
