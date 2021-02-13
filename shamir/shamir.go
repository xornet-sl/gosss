/*
Copyright © 2020 Vladimir Sukhonosov. Contacts: founder.sl@gmail.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

This file is partially copyrighted by HashiCorp Vault developers (under Mozilla Public License 2.0)
Original code is here: https://github.com/hashicorp/vault/tree/master/shamir
*/

package shamir

import (
	"crypto/rand"
	"errors"

	"github.com/xornet-sl/gosss/galois"
)

type polynomial struct {
	coefs []uint8
}

var newPolynomial = func(intercept, degree uint8) (*polynomial, error) {
	p := &polynomial{
		coefs: make([]uint8, degree+1),
	}
	p.coefs[0] = intercept
	_, err := rand.Read(p.coefs[1:])
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *polynomial) evaluate(x uint8) uint8 {
	if x == 0 {
		return p.coefs[0]
	}

	degree := len(p.coefs) - 1
	ret := p.coefs[degree]
	for i := degree - 1; i >= 0; i-- {
		ret = galois.Add(galois.Mul(ret, x), p.coefs[i])
	}
	return ret
}

func interpolate(xSamples, ySamples []uint8, x uint8) (uint8, error) {
	limit := len(xSamples)
	ret := uint8(0)
	basis := uint8(0)

	for i := 0; i < limit; i++ {
		basis = 1
		for j := 0; j < limit; j++ {
			if i == j {
				continue
			}
			num := galois.Add(x, xSamples[j])
			denom := galois.Add(xSamples[i], xSamples[j])
			term, err := galois.Div(num, denom)
			if err != nil {
				return 0, err
			}
			basis = galois.Mul(basis, term)
		}
		group := galois.Mul(ySamples[i], basis)
		ret = galois.Add(ret, group)
	}
	return ret, nil
}

func splitBlock(block []byte, buf [][]byte, xCoords []uint8, threshold uint8, polynomPerByte bool) error {
	var (
		err error
		p   *polynomial
	)
	if !polynomPerByte {
		p, err = newPolynomial(0, threshold-1)
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(block); i++ {
		if polynomPerByte {
			p, err = newPolynomial(0, threshold-1)
			if err != nil {
				return err
			}
		}
		p.coefs[0] = block[i]
		for partIdx := 0; partIdx < len(buf); partIdx++ {
			x := xCoords[partIdx]
			buf[partIdx][i] = p.evaluate(x)
		}
	}
	return nil
}

var fillXCoords = func(size uint8) ([]uint8, error) {
	var (
		err     error
		xCoords = make([]uint8, size)
		seenX   = make(map[uint8]bool)
	)
	for i := range xCoords {
		for {
			_, err = rand.Read(xCoords[i : i+1])
			if err != nil {
				return nil, err
			}
			if xCoords[i] == 0 || seenX[xCoords[i]] {
				continue
			}
			seenX[xCoords[i]] = true
			break
		}
	}
	return xCoords, nil
}

func prepareSplit(blockSize uint64, partsCount uint8) ([][]byte, []uint8, error) {
	buf := make([][]byte, partsCount)
	xCoords, err := fillXCoords(partsCount)
	if err != nil {
		return nil, nil, err
	}
	for partIdx := uint8(0); partIdx < partsCount; partIdx++ {
		buf[partIdx] = make([]byte, blockSize)
	}
	return buf, xCoords, nil
}

func validateSplitArgs(partsCount, threshold int) error {
	if partsCount > 255 || partsCount < 2 {
		return errors.New("Number of parts should be between 2 and 255")
	}
	if threshold > 255 || threshold < 2 {
		return errors.New("Threshold should be between 2 and 255")
	}
	if threshold > partsCount {
		return errors.New("Threshold should not be greater than number of parts")
	}
	return nil
}

// Split splits bytes into parts
func Split(secret []byte, partsCount, threshold int) ([][]byte, error) {
	var (
		err        error
		ret        [][]byte
		buf        [][]byte
		xCoords    []uint8
		secretSize = uint64(len(secret))
	)
	err = validateSplitArgs(partsCount, threshold)
	if err != nil {
		return nil, err
	}
	ret, xCoords, err = prepareSplit(secretSize+1, uint8(partsCount))
	if err != nil {
		return nil, err
	}
	buf = make([][]byte, partsCount)
	for partIdx := range ret {
		ret[partIdx][0] = xCoords[partIdx]
		buf[partIdx] = ret[partIdx][1:]
	}
	err = splitBlock(secret, buf, xCoords, uint8(threshold), true)
	return ret, err
}

func combineBlock(buf []byte, parts [][]byte, xSamples, ySamplesBuf []uint8) error {
	var err error
	partLen := len(parts[0])
	for i := 0; i < partLen; i++ {
		for partIdx := range parts {
			ySamplesBuf[partIdx] = parts[partIdx][i]
		}
		buf[i], err = interpolate(xSamples, ySamplesBuf, 0)
		if err != nil {
			return err
		}
	}
	return nil
}

// Combine combines parts back into bytes
func Combine(parts [][]byte) ([]byte, error) {
	var (
		ret         []byte
		partsBuf    [][]byte
		err         error
		partsCount  = len(parts)
		xSamples    = make([]uint8, partsCount)
		ySamplesBuf = make([]uint8, partsCount)
	)
	if partsCount < 2 || partsCount >= 255 {
		return nil, errors.New("number of parts should be 2 <= x < 255")
	}
	partLen := len(parts[0]) - 1
	if partLen < 1 {
		return make([]byte, 0), nil
	}
	ret = make([]byte, partLen)
	partsBuf = make([][]byte, partsCount)

	for partIdx := range parts {
		xSamples[partIdx] = parts[partIdx][0]
		partsBuf[partIdx] = parts[partIdx][1:]
	}
	err = combineBlock(ret, partsBuf, xSamples, ySamplesBuf)
	return ret, err
}
