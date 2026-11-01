// Copyright (c) 2021 Quan guanyu
// randomness is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package randomness

import (
	"math"
	"math/bits"
)

// BinaryDerivative 二元推导检测， k=7
func BinaryDerivative(data []byte) *TestResult {
	p, q := BinaryDerivativeTestBytes(data, 7)
	return &TestResult{Name: "二元推导检测(k=7)", P: p, Q: q, Pass: p >= Alpha}
}

// BinaryDerivativeTest 二元推导检测， k=7
func BinaryDerivativeTest(bits []bool, k int) (float64, float64) {
	return BinaryDerivativeProto(bits, k)
}

// BinaryDerivativeTestBytes 二元推导检测
// bits: 待检测序列
// k: 重复次数，k=3,7
func BinaryDerivativeTestBytes(data []byte, k int) (float64, float64) {
	return binaryDerivativeWords(packBytesMSB(data), len(data)*8, k)
}

// BinaryDerivativeProto 二元推导检测
// bits: 待检测序列
// k: 重复次数，k=3,7
func BinaryDerivativeProto(bits []bool, k int) (float64, float64) {
	n := len(bits)
	if n < 7 {
		panic("please provide valid test bits")
	}

	if n >= 4096 {
		return binaryDerivativeWords(packBools(bits), n, k)
	}

	S := 0
	var V float64 = 0
	_bits := make([]bool, len(bits))
	copy(_bits, bits)

	// Step 1, 2
	for i := 0; i < k; i++ {
		for j := 0; j < n-i-1; j++ {
			_bits[j] = xor(_bits[j], _bits[j+1])
		}
	}

	// Step 3
	for i := 0; i < n-k; i++ {
		if _bits[i] {
			S++
		} else {
			S--
		}
	}
	// Step 4, 提前对V除以2的平方根，避免求P Q时再求解
	V = float64(S) / math.Sqrt(2*float64(n-k))

	// Step 5
	P := math.Erfc(math.Abs(V))

	// Step 6
	Q := math.Erfc(V) / 2
	return P, Q
}

func binaryDerivativeWords(w []uint64, n, k int) (float64, float64) {
	if n < 7 {
		panic("please provide valid test bits")
	}
	for i := 0; i < k; i++ {
		n = derivativeStep(w, n)
	}
	ones := 0
	nwords := (n + 63) >> 6
	for i := 0; i < nwords; i++ {
		x := w[i]
		if i == nwords-1 {
			if rem := n & 63; rem != 0 {
				x &= (uint64(1) << uint(rem)) - 1
			}
		}
		ones += bits.OnesCount64(x)
	}
	S := 2*ones - n
	V := float64(S) / math.Sqrt(2*float64(n))
	P := math.Erfc(math.Abs(V))
	Q := math.Erfc(V) / 2
	return P, Q
}

// derivativeStep sets bit j to bit j XOR bit j+1 and shortens the vector by one.
func derivativeStep(w []uint64, n int) int {
	if n <= 1 {
		return 0
	}
	nwords := (n + 63) >> 6
	for i := 0; i < nwords; i++ {
		x := w[i]
		var nextLow uint64
		if i+1 < nwords {
			nextLow = w[i+1] & 1
		}
		w[i] = (x ^ (x >> 1)) ^ (nextLow << 63)
	}
	newN := n - 1
	word := newN >> 6
	off := uint(newN & 63)
	if word < len(w) {
		if off != 0 {
			w[word] &= (uint64(1) << off) - 1
			word++
		}
		for ; word < len(w); word++ {
			w[word] = 0
		}
	}
	return newN
}
