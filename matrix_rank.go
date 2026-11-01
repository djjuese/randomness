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

// MatrixRank 矩阵秩检测,M=Q=32
func MatrixRank(data []byte) *TestResult {
	p, q := MatrixRankTestBytes(data, 32, 32)
	return &TestResult{Name: "矩阵秩检测", P: p, Q: q, Pass: p >= Alpha}
}

// MatrixRankTest 矩阵秩检测,M=Q=32
func MatrixRankTest(bits []bool) (float64, float64) {
	return MatrixRankProto(bits, 32, 32)
}

// MatrixRankTestBytes 矩阵秩检测
func MatrixRankTestBytes(data []byte, M, Q int) (float64, float64) {
	return MatrixRankProto(B2bitArr(data), M, Q)
}

// MatrixRankProto 矩阵秩检测
// bits: 待检测序列
// M: 矩阵行数
// Q: 矩阵列隶属
func MatrixRankProto(bits []bool, M, Q int) (float64, float64) {
	n := len(bits)

	N := n / (M * Q)
	if N == 0 {
		panic("please provide valid test bits")
	}
	//int n_disc = n % (M * Q);
	if M == 32 && Q == 32 {
		return matrixRank32(bits, N)
	}

	var Fm, Fm1, Fr = 0, 0, 0
	var matrix = make([][]int, M)
	for i := 0; i < M; i++ {
		matrix[i] = make([]int, Q)
	}
	var r int

	offset := 0
	for i := 0; i < N; i++ {
		for j := 0; j < M; j++ {
			for k := 0; k < Q; k++ {
				if bits[offset] {
					matrix[j][k] = 1
				} else {
					matrix[j][k] = 0
				}
				offset++
			}
		}
		r = rank(matrix, M)

		if r == min(M, Q) {
			Fm++
		} else if r == (min(M, Q) - 1) {
			Fm1++
		} else {
			Fr++
		}
	}
	return matrixRankPValue(N, Fm, Fm1, Fr)
}

// matrixRank32 ranks each 32x32 bit matrix with packed rows.
// Column k lives in bit k, matching MatrixRankProto's fill order.
func matrixRank32(bits []bool, N int) (float64, float64) {
	var Fm, Fm1, Fr int
	idx := 0
	for i := 0; i < N; i++ {
		var rows [32]uint32
		for r := 0; r < 32; r++ {
			var row uint32
			base := idx
			for c := 0; c < 32; c++ {
				if bits[base+c] {
					row |= 1 << uint(c)
				}
			}
			rows[r] = row
			idx += 32
		}
		switch rank32(&rows) {
		case 32:
			Fm++
		case 31:
			Fm1++
		default:
			Fr++
		}
	}
	return matrixRankPValue(N, Fm, Fm1, Fr)
}

func matrixRankPValue(N, Fm, Fm1, Fr int) (float64, float64) {
	n := float64(N)
	V := math.Pow(float64(Fm)-0.2888*n, 2.0)/(0.2888*n) +
		math.Pow(float64(Fm1)-0.5776*n, 2.0)/(0.5776*n) +
		math.Pow(float64(Fr)-0.1336*n, 2.0)/(0.1336*n)
	P := igamc(1, V/2.0)
	return P, P
}

// rank32 is the GF(2) rank of a 32x32 matrix. It counts pivots, which
// matches the nonzero-row count of the previous row-echelon routine.
func rank32(rows *[32]uint32) int {
	rank := 0
	for col := 0; col < 32; col++ {
		mask := uint32(1) << uint(col)
		pivot := -1
		for r := rank; r < 32; r++ {
			if rows[r]&mask != 0 {
				pivot = r
				break
			}
		}
		if pivot < 0 {
			continue
		}
		if pivot != rank {
			rows[rank], rows[pivot] = rows[pivot], rows[rank]
		}
		for r := rank + 1; r < 32; r++ {
			if rows[r]&mask != 0 {
				rows[r] ^= rows[rank]
			}
		}
		rank++
	}
	return rank
}
