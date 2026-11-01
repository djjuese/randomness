package fft

import (
	"math/rand"
	"testing"
)

// oldTransform is the previous two-multiply butterfly, used to lock the
// single-multiply form to the same complex values.
func oldTransform(f FFT, x []complex128) {
	inputPermutation(x, f.perm)
	n := 1
	s := f.N
	for p := 1; p <= f.p; p++ {
		s >>= 1
		for b := 0; b < s; b++ {
			o := 2 * b * n
			for k := 0; k < n; k++ {
				i := k + o
				j := i + n
				x[i], x[j] = x[i]+f.E[k*s]*x[j], x[i]+f.E[s*(k+n)]*x[j]
			}
		}
		n <<= 1
	}
}

func TestTransformMatchesOldButterfly(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for _, N := range []int{8, 16, 32, 64, 256, 1024, 4096} {
		f, err := New(N)
		if err != nil {
			t.Fatal(err)
		}
		x := make([]complex128, N)
		y := make([]complex128, N)
		for i := range x {
			// Mix gaussian values and the ±1 inputs used by the DFT test.
			if i%3 == 0 {
				if rng.Intn(2) == 0 {
					x[i] = 1
				} else {
					x[i] = -1
				}
			} else {
				x[i] = complex(rng.NormFloat64(), rng.NormFloat64())
			}
			y[i] = x[i]
		}
		f.Transform(x)
		oldTransform(f, y)
		for i := range x {
			if x[i] != y[i] {
				t.Fatalf("N=%d i=%d new=%v old=%v", N, i, x[i], y[i])
			}
		}
	}
}
