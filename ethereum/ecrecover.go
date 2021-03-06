// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package ethereum

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// PrecompiledEcRecover is implemented following geth's implementation 
// and thus guarantee to always return the same output as
//
// https://github.com/ethereum/go-ethereum/blob/7504dbd6eb3f62371f86b06b03ffd665690951f2/core/vm/common.go#L92
func allZero(b []byte) bool {
	for _, byte := range b {
		if byte != 0 {
			return false
		}
	}
	return true
}

const ecRecoverInputLength = 128

// PrecompiledEcRecover is implemented following the precompiles implementation
// and thus guarantee to always return the same output as
//
// https://github.com/ethereum/go-ethereum/blob/52f2461774bcb8cdd310f86b4bc501df5b783852/core/vm/contracts.go#L78
func PrecompiledEcRecover(input []byte) ([]byte, error) {
	input = common.RightPadBytes(input, ecRecoverInputLength)
	// "input" is (hash, v, r, s), each 32 bytes
	// but for ecrecover we want (r, s, v)

	r := new(big.Int).SetBytes(input[64:96])
	s := new(big.Int).SetBytes(input[96:128])
	v := input[63] - 27

	// tighter sig s values input homestead only apply to tx sigs
	if !allZero(input[32:63]) || !crypto.ValidateSignatureValues(v, r, s, false) {
		return nil, nil
	}
	// v needs to be at the end for libsecp256k1
	pubKey, err := crypto.Ecrecover(input[:32], append(input[64:128], v))
	// make sure the public key is a valid one
	if err != nil {
		return nil, nil
	}

	// the first byte of pubkey is bitcoin heritage
	return common.LeftPadBytes(crypto.Keccak256(pubKey[1:])[12:], 32), nil
}

// EcRecover an Ethereum secp256k1 ECDSA signature, by casting the inputs as a precompile request argument
// then returns the recovered ethereum address.
// Hash is a 32 bytes slices
// VRS is a 65 bytes slice.
//
// It can be used to verify that a go implementation of ethereum signature and formatting complies with 
// ecrecover standard through unit-tests
func EcRecover(Hash common.Hash, VRS []byte) (common.Address, error) {
	// All values are initialized to zero
	input := make([]byte, 128)
	// Copying the V element signature
	input[63] = VRS[0]
	// Copying the Hash value
	copy(input[:32], Hash[:32])
	// Copying the R and S elements of the signature
	copy(input[64:128], VRS[1:65])
	
	output, err := PrecompiledEcRecover(input)
	if err != nil {
		return common.Address{}, err
	}

	// Extract the address from the returned slice
	return common.BytesToAddress(output[12:32]), nil
}
