package dkg

import (
	"client/internal/pkg/group/curve25519"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"
)

func HexToScalar(suite suites.Suite, hexScalar string) (kyber.Scalar, error) {
	b, err := hex.DecodeString(hexScalar)
	if byteErr, ok := err.(hex.InvalidByteError); ok {
		return nil, fmt.Errorf("invalid hex character %q in scalar", byte(byteErr))
	} else if err != nil {
		return nil, errors.New("invalid hex data for scalar")
	}
	s := suite.Scalar()
	if err := s.UnmarshalBinary(b); err != nil {
		return nil, fmt.Errorf("unmarshal scalar binary: %w", err)
	}
	return s, nil
}

func PointToBigUncompressed(point kyber.Point) ([2]*big.Int) {
	p := point.(*curve25519.ProjPoint)
	x, y := p.GetXY()
	return [2]*big.Int{&x.V, &y.V}
}

func PointToBig(point kyber.Point) (*big.Int, error) {
	b, err := point.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return new(big.Int).SetBytes(b), nil
}

func PointsToBig(points []kyber.Point) ([]*big.Int, error) {
	values := make([]*big.Int, 0)
	for _, point := range points {
		v, err := PointToBig(point)
		if err != nil {
			return nil, fmt.Errorf("point to big: %w", err)
		}
		values = append(values, v)
	}
	return values, nil
}

func BigToPoint(suite suites.Suite, p *big.Int) (kyber.Point, error) {
	point := suite.Point().Base()

	buf := make([]byte, 32)
	err := point.UnmarshalBinary(p.FillBytes(buf))
	if err != nil {
		return nil, err
	}
	return point, nil
}

func BigToPoints(suite suites.Suite, p []*big.Int) ([]kyber.Point, error) {
	points := make([]kyber.Point, 0)
	for _, value := range p {
		point, err := BigToPoint(suite, value)
		if err != nil {
			return nil, fmt.Errorf("big to point: %w", err)
		}
		points = append(points, point)
	}
	return points, nil
}
