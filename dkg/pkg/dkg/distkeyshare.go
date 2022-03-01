package dkg

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
)

type DistKeyShare struct {
	Commits     []kyber.Point
	Share       *share.PriShare
	PrivatePoly []kyber.Scalar
}

func (d *DistKeyShare) Public() kyber.Point {
	return d.Commits[0]
}

func (d *DistKeyShare) PriShare() *share.PriShare {
	return d.Share
}

func (d *DistKeyShare) Commitments() []kyber.Point {
	return d.Commits
}
