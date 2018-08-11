package main

import (
	"github.com/ethereum/go-ethereum/common/math"
	"math/big"
	"fmt"
	"encoding/hex"
	"crypto/sha256"
	"github.com/btcsuite/golangcrypto/ripemd160"
	b58 "github.com/jbenet/go-base58"
	"hash"
)

var G *Point
var P *big.Int
var B *FieldElement
var Ripemd160 hash.Hash

type Network byte

const (
	MAINNET Network = 0
	TESTNET Network = 111
)

type FieldElement struct {
	Value, Field *big.Int
}

type Point struct{
	X, Y *FieldElement
}

func (p *Point) GetAddress(network Network) string {
	result := p.encodeUncompressedSecBytes()

	// Compute HASH160 (SHA256 + RIPEMD160)
	v256result := sha256.Sum256(result)
	Ripemd160.Reset()
	Ripemd160.Write(v256result[:])
	v160result := Ripemd160.Sum(nil)
	// Payload format: [1 byte (network) + 20 byte (key) + 4 byte (checksum)]
	resultPayload := make([]byte, 1 + 20 + 4)
	resultPayload[0] = byte(network)
	copy(resultPayload[1:22], v160result)

	// Perform Checksum
	firstShaChecksum := sha256.Sum256(resultPayload[:21])
	doubleShaChecksum := sha256.Sum256(firstShaChecksum[:])
	copy(resultPayload[21:], doubleShaChecksum[:4])

	str := b58.Encode(resultPayload)
	return str
}

func (p *Point) encodeUncompressedSecBytes() []byte {
	result := make([]byte, 32 + 1)

	// Check if Y is odd
	v2:= new(big.Int).SetUint64(2)
	mod := new(big.Int).Mod(p.Y.Value, v2)
	isEven := mod.Text(10) == "2"

	if isEven {
		result[0] = byte(2)
	} else {
		result[0] = byte(3)
	}
	copy(result[1:], p.Y.Value.Bytes())
	return result
}

func (p *Point) EncodeUncompressedSec() string {
    result := p.encodeUncompressedSecBytes()
	return hex.EncodeToString(result)
}

func (p *Point) ToString() string {
	return fmt.Sprintf("Point(%s, %s)", p.X.Value.Text(10), p.Y.Value.Text(10))
}

func (p *Point) IsOnSecpCurve() bool {
	// y² = x³ + b
	y2 := p.Y.Mul(p.Y)

	x3 := p.X.Mul(p.X).Mul(p.X)
	xRes := x3.Add(B)

	return xRes.Equal(y2)
}

func (p *Point) MultiplyScalar() *Point {
	//c = (3px2 + a) / 2py
	//rx = c2 - 2px
	//ry = c (px - rx) - py
	vx2 := p.X.Mul(p.X)
	v3x2 := vx2.Mul(NewSecp256k1FieldElement(3))
	v2py := p.Y.Mul(NewSecp256k1FieldElement(2))
	c := v3x2.Div(v2py)

	vc2 := c.Mul(c)
	v2px := p.X.Mul(NewSecp256k1FieldElement(2))
	rx := vc2.Minus(v2px)

	vpxrx := p.X.Minus(rx)
	vcpxrx := c.Mul(vpxrx)
	ry := vcpxrx.Minus(p.Y)

	return &Point{
		rx, ry,
	}
}

func NewSecp256k1FieldElement(value int64) *FieldElement {
	res := math.U256(big.NewInt(value))

	return &FieldElement{
		Value:res,
		Field: P,
	}
}

func (f1 *FieldElement) Mul (f2 *FieldElement) *FieldElement {
	if f1.Field != f2.Field {
		panic("Fields are different")
	}
	res := new(big.Int).Mul(f1.Value, f2.Value)
	res.Mod(res, f1.Field)
	return &FieldElement{
		res, P,
	}
}


func (f1 *FieldElement) Div (f2 *FieldElement) *FieldElement {
	if f1.Field != f2.Field {
		panic("Fields are different")
	}
	// a * b^(p-2)
	v3 := new(big.Int).SetUint64(2)
	vp2 := new(big.Int).Sub(P, v3)
	vbp2 := new(big.Int).Exp(f2.Value, vp2, P)
	vbp2fe := &FieldElement{vbp2, P}
	return f1.Mul(vbp2fe)
}

func (f1 *FieldElement) Minus (f2 *FieldElement) *FieldElement {
	if f1.Field != f2.Field {
		panic("Fields are different")
	}
	var result *big.Int
	res := new(big.Int).Sub(f1.Value, f2.Value)
	if !res.IsUint64() {
		resAbs := new(big.Int).Abs(res)
		resAbsMod := new(big.Int).Mod(resAbs, P)
		result = new(big.Int).Sub(P, resAbsMod)
	} else {
		result = new(big.Int).Mod(res, f1.Field)
	}
	return &FieldElement{result, f1.Field}
}


func (f1 *FieldElement) Add (f2 *FieldElement) *FieldElement {
	if f1.Field != f2.Field {
		panic("Fields are different")
	}
	res := new(big.Int).Add(f1.Value, f2.Value)
	res.Mod(res, f1.Field)
	return &FieldElement{
		res, P,
	}
}

func (f1 *FieldElement) Equal (f2 *FieldElement) bool {
	if f1.Field != f2.Field {
		panic("Fields are different")
	}
	res := f1.Value.Cmp(f2.Value)
	return res == 0
}


func init() {
	P, _ = new(big.Int).SetString("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 0)

	bValue, _ := new(big.Int).SetString("0x0000000000000000000000000000000000000000000000000000000000000007", 0)
	B = &FieldElement{
		bValue,
		P,
	}

	Gx, _ := new(big.Int).SetString("0x79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798", 0)
	Gy, _ := new(big.Int).SetString("0x483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8", 0)

	G = &Point{
		X: &FieldElement{
			Gx, P,
		},
		Y: &FieldElement{
			Gy, P,
		},
	}

	Ripemd160 = ripemd160.New()
}