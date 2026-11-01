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
)

func mutFactorC(L, K int) float64 {
	var v float64
	v = 0.7
	v -= 0.8 / float64(L)
	v += (4.0 + 32.0/float64(L)) * (math.Pow(float64(K), -3.0/float64(L)) / 15.0)
	return v
}

// MaurerUniversal Maurer通用统计检测方法
func MaurerUniversal(data []byte) *TestResult {
	p, q := MaurerUniversalTestBytes(data)
	return &TestResult{Name: "Maurer通用统计检测方法", P: p, Q: q, Pass: p >= Alpha}
}

// MaurerUniversalTestBytes Maurer通用统计检测方法
func MaurerUniversalTestBytes(data []byte) (float64, float64) {
	return MaurerUniversalTest(B2bitArr(data))
}

// MaurerUniversalTest Maurer通用统计检测方法
func MaurerUniversalTest(bits []bool) (float64, float64) {
	n := len(bits)
	if n == 0 {
		panic("please provide test bits")
	}
	L := 7
	Q := 1280
	T := make([]int, 1<<uint(L))
	mask := (1 << uint(L)) - 1

	var K int = n/L - Q
	//var  n_disc int = n % L;
	var sum float64 = 0.0
	var V float64 = 0
	var P float64 = 0
	var sigma float64 = 0
	expected_value := []float64{0, 0, 0, 0, 0, 0, 5.2177052, 6.1962507, 7.1836656,
		8.1764248, 9.1723243, 10.170032, 11.168765,
		12.168070, 13.167693, 14.167488, 15.167379}
	variance := []float64{0, 0, 0, 0, 0, 0, 2.954, 3.125, 3.238, 3.311, 3.356, 3.384,
		3.401, 3.410, 3.416, 3.419, 3.421}

	var tmp int = 0
	idx := 0
	readBlock := func() int {
		for j := 0; j < L; j++ {
			tmp <<= 1
			if bits[idx] {
				tmp++
			}
			idx++
		}
		return tmp & mask
	}
	for i := 1; i <= Q; i++ {
		T[readBlock()] = i
	}

	log2 := math.Log(2.0)
	for i := Q + 1; i <= Q+K; i++ {
		block := readBlock()
		sum += math.Log(float64(i)-float64(T[block])) / log2
		T[block] = i
	}

	sigma = math.Sqrt(variance[L]/float64(K)) * mutFactorC(L, K)
	V = (sum/float64(K) - expected_value[L]) / (sigma * math.Sqrt(2.0)) // 避免求p q时V再除以math.Sqrt(2.0)
	P = math.Erfc(math.Abs(V))
	q := math.Erfc(V) / 2

	return P, q
}
