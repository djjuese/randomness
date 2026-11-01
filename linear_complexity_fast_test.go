package randomness

import (
	"math/rand"
	"testing"
)

// linearComplexityNaive is the original bit-at-a-time Berlekamp-Massey.
// Kept to lock the word-parallel implementation to the same lengths.
func linearComplexityNaive(a []bool, M int) int {
	N_ := 0
	L := 0
	m := -1
	B_ := make([]int, M)
	C := make([]int, M)
	P := make([]int, M)
	T := make([]int, M)
	C[0] = 1
	B_[0] = 1

	for N_ < M {
		d := b2i(a[N_])
		for i := 1; i <= L; i++ {
			if C[i] == 1 {
				d ^= b2i(a[N_-i])
			}
		}
		if d == 1 {
			copy(T, C)
			for i := range P {
				P[i] = 0
			}
			shift := N_ - m
			for j := 0; j < M; j++ {
				if B_[j] == 1 && j+shift < M {
					P[j+shift] = 1
				}
			}
			for i := 0; i < M; i++ {
				C[i] ^= P[i]
			}
			if L <= N_/2 {
				L = N_ + 1 - L
				m = N_
				copy(B_, T)
			}
		}
		N_++
	}
	return L
}

func TestLinearComplexityFastMatchesNaive(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	sizes := []int{1, 2, 3, 7, 8, 9, 31, 32, 33, 63, 64, 65, 100, 127, 128, 129, 200, 500, 1000}
	for _, M := range sizes {
		trials := 40
		if M >= 500 {
			trials = 8
		}
		for trial := 0; trial < trials; trial++ {
			a := make([]bool, M)
			switch trial {
			case 0:
				// all zeros
			case 1:
				for i := range a {
					a[i] = true
				}
			case 2:
				for i := range a {
					a[i] = i%2 == 0
				}
			default:
				for i := range a {
					a[i] = rng.Intn(2) == 1
				}
			}
			got := linearComplexity(a, M)
			want := linearComplexityNaive(a, M)
			if got != want {
				t.Fatalf("M=%d trial=%d got %d want %d", M, trial, got, want)
			}
		}
	}
}

func TestRank32MatchesRowEchelon(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	for n := 0; n < 200; n++ {
		matrix := make([][]int, 32)
		var rows [32]uint32
		for i := 0; i < 32; i++ {
			matrix[i] = make([]int, 32)
			for j := 0; j < 32; j++ {
				b := 0
				if n > 2 && rng.Intn(2) == 1 {
					b = 1
				}
				if n == 1 {
					b = 1
				}
				matrix[i][j] = b
				if b == 1 {
					rows[i] |= 1 << uint(j)
				}
			}
		}
		if rank(matrix, 32) != rank32(&rows) {
			t.Fatalf("case %d rank mismatch", n)
		}
	}
}
