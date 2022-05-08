package dkg

import (
	"encoding/json"
	"math/big"
	"strings"
)

type ZKProof struct {
	A PairingG1Point
	B PairingG2Point
	C PairingG1Point
}

type Proof struct {
	Inputs []*big.Int
	Proof  *ZKProof
}

func (proof *Proof) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		return err
	}

	var inputHex []string
	err = json.Unmarshal(*objMap["inputs"], &inputHex)
	if err != nil {
		return err
	}

	inputs := make([]*big.Int, 0)
	for _, input := range inputHex {
		value, _ := new(big.Int).SetString(strings.TrimPrefix(input, "0x"), 16)
		inputs = append(inputs, value)
	}

	proof.Inputs = inputs

	var zkProof *ZKProof
	err = json.Unmarshal(*objMap["proof"], &zkProof)
	if err != nil {
		return err
	}

	proof.Proof = zkProof

	return nil
}

func (proof *ZKProof) UnmarshalJSON(b []byte) error {
	var objMapProof map[string]*json.RawMessage
	err := json.Unmarshal(b, &objMapProof)
	if err != nil {
		return err
	}

	var pointA PairingG1Point
	err = json.Unmarshal(*objMapProof["a"], &pointA)
	if err != nil {
		return err
	}

	var pointB PairingG2Point
	err = json.Unmarshal(*objMapProof["b"], &pointB)
	if err != nil {
		return err
	}

	var pointC PairingG1Point
	err = json.Unmarshal(*objMapProof["c"], &pointC)
	if err != nil {
		return err
	}

	proof.A = pointA
	proof.B = pointB
	proof.C = pointC

	return nil
}

func (point *PairingG1Point) UnmarshalJSON(b []byte) error {
	var hexValues []string
	err := json.Unmarshal(b, &hexValues)
	if err != nil {
		return err
	}

	point.X, _ = new(big.Int).SetString(strings.TrimPrefix(hexValues[0], "0x"), 16)
	point.Y, _ = new(big.Int).SetString(strings.TrimPrefix(hexValues[1], "0x"), 16)

	return nil
}

func (point *PairingG2Point) UnmarshalJSON(b []byte) error {
	var hexValues [][]string
	err := json.Unmarshal(b, &hexValues)
	if err != nil {
		return err
	}

	x, _ := new(big.Int).SetString(strings.TrimPrefix(hexValues[0][0], "0x"), 16)
	xi, _ := new(big.Int).SetString(strings.TrimPrefix(hexValues[0][1], "0x"), 16)

	y, _ := new(big.Int).SetString(strings.TrimPrefix(hexValues[1][0], "0x"), 16)
	yi, _ := new(big.Int).SetString(strings.TrimPrefix(hexValues[1][1], "0x"), 16)

	point.X = [2]*big.Int{x, xi}
	point.Y = [2]*big.Int{y, yi}

	return nil
}
