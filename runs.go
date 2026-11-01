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

// Runs 游程总数检测
func Runs(data []byte) *TestResult {
	p, q := RunsTestBytes(data)
	return &TestResult{Name: "游程总数检测", P: p, Q: q, Pass: p >= Alpha}
}

// RunsTestBytes 游程总数检测
func RunsTestBytes(data []byte) (float64, float64) {
	return runsFromWords(packBytesMSB(data), len(data)*8)
}

// RunsTest 游程总数检测
func RunsTest(bits []bool) (float64, float64) {
	n := len(bits)
	if n == 0 {
		panic("please provide test bits")
	}
	if n >= 4096 {
		return runsFromWords(packBools(bits), n)
	}

	var Pi float64 = 0
	var V_obs int = 1
	var P, Q float64 = 0, 0

	// Step 1, 2
	for i := 0; i < n-1; i++ {
		if bits[i] != bits[i+1] {
			V_obs++
		}
		if bits[i] {
			Pi++
		}
	}
	if bits[n-1] {
		Pi++
	}
	Pi /= float64(n)

	// Step 3, 第四、五步的除math.Sqrt(2)，放到这里提前处理，减少math.Sqrt的调用。
	V := (float64(V_obs) - 2.0*float64(n)*Pi*(1.0-Pi)) / (2.0 * math.Sqrt(float64(2*n)) * Pi * (1.0 - Pi))

	// Step 4
	P = math.Erfc(math.Abs(V))

	// Step 5
	Q = math.Erfc(V) / 2.0
	return P, Q
}

func runsFromWords(w []uint64, n int) (float64, float64) {
	if n == 0 {
		panic("please provide test bits")
	}
	ones := 0
	transitions := 0
	nwords := (n + 63) >> 6
	var prev uint64
	for i := 0; i < nwords; i++ {
		x := w[i]
		width := 64
		if i == nwords-1 {
			if rem := n & 63; rem != 0 {
				width = rem
				x &= (uint64(1) << uint(rem)) - 1
			}
		}
		ones += bits.OnesCount64(x)
		if width > 1 {
			mask := uint64(1)<<uint(width-1) - 1
			transitions += bits.OnesCount64((x ^ (x >> 1)) & mask)
		}
		if i > 0 && (x&1) != prev {
			transitions++
		}
		if width == 64 {
			prev = (x >> 63) & 1
		} else if width > 0 {
			prev = (x >> uint(width-1)) & 1
		}
	}
	Pi := float64(ones) / float64(n)
	V_obs := transitions + 1
	V := (float64(V_obs) - 2.0*float64(n)*Pi*(1.0-Pi)) / (2.0 * math.Sqrt(float64(2*n)) * Pi * (1.0 - Pi))
	P := math.Erfc(math.Abs(V))
	Q := math.Erfc(V) / 2.0
	return P, Q
}
