package dkg

import (
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

func PointToBig(point kyber.Point) ([2]*big.Int, error) {
	b, err := point.MarshalBinary()
	if err != nil {
		return [2]*big.Int{}, fmt.Errorf("marshal point: %w", err)
	}

	if len(b) != 64 {
		return [2]*big.Int{}, fmt.Errorf("invalid length")
	}

	return [2]*big.Int{
		new(big.Int).SetBytes(b[:32]),
		new(big.Int).SetBytes(b[32:]),
	}, nil
}

func PointsToBig(points []kyber.Point) ([][2]*big.Int, error) {
	values := make([][2]*big.Int, 0)
	for _, point := range points {
		v, err := PointToBig(point)
		if err != nil {
			return nil, fmt.Errorf("point to big: %w", err)
		}
		values = append(values, v)
	}
	return values, nil
}

func BigToPoint(suite suites.Suite, p [2]*big.Int) (kyber.Point, error) {
	point := suite.Point().Base()

	x := make([]byte, 32)
	copy(x[len(x)-len(p[0].Bytes()):], p[0].Bytes())

	y := make([]byte, 32)
	copy(y[len(y)-len(p[1].Bytes()):], p[1].Bytes())

	err := point.UnmarshalBinary(append(x, y...))
	if err != nil {
		return nil, fmt.Errorf("unmarshal binary: %w", err)
	}
	return point, nil
}

func BigToPoints(suite suites.Suite, p [][2]*big.Int) ([]kyber.Point, error) {
	points := make([]kyber.Point, 0)
	for _, values := range p {
		point, err := BigToPoint(suite, values)
		if err != nil {
			return nil, fmt.Errorf("big to point: %w", err)
		}
		points = append(points, point)
	}
	return points, nil
}
