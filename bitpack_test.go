package randomness

import (
	"math"
	"math/rand"
	"testing"
)

func naiveRuns(bits []bool) (int, int) {
	n := len(bits)
	ones := 0
	transitions := 0
	for i := 0; i < n-1; i++ {
		if bits[i] != bits[i+1] {
			transitions++
		}
		if bits[i] {
			ones++
		}
	}
	if bits[n-1] {
		ones++
	}
	return ones, transitions + 1
}

func naiveDerivative(bits []bool, k int) int {
	n := len(bits)
	cp := append([]bool(nil), bits...)
	for i := 0; i < k; i++ {
		for j := 0; j < n-i-1; j++ {
			cp[j] = cp[j] != cp[j+1]
		}
	}
	S := 0
	for i := 0; i < n-k; i++ {
		if cp[i] {
			S++
		} else {
			S--
		}
	}
	return S
}

func TestPackedRunsAndDerivative(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	lengths := []int{1, 2, 8, 9, 63, 64, 65, 127, 128, 129, 200, 1000, 4096, 5000, 100000}
	for _, n := range lengths {
		for trial := 0; trial < 5; trial++ {
			raw := make([]byte, (n+7)/8)
			rng.Read(raw)
			bits := B2bitArr(raw)[:n]
			if n >= 2 {
				ones, vObs := naiveRuns(bits)
				if ones != 0 && ones != n {
					p1, q1 := runsFromWords(packBools(bits), n)
					Pi := float64(ones) / float64(n)
					V := (float64(vObs) - 2.0*float64(n)*Pi*(1-Pi)) / (2.0 * math.Sqrt(float64(2*n)) * Pi * (1 - Pi))
					wantP := math.Erfc(math.Abs(V))
					if math.Abs(p1-wantP) > 1e-12 {
						t.Fatalf("runs n=%d trial=%d got %v want %v", n, trial, p1, wantP)
					}
					if n%8 == 0 {
						p2, q2 := runsFromWords(packBytesMSB(raw), len(raw)*8)
						if p1 != p2 || q1 != q2 {
							t.Fatalf("runs bytes mismatch n=%d", n)
						}
					}
				}
			}
			if n >= 7 {
				for _, k := range []int{3, 7} {
					if k >= n {
						continue
					}
					want := naiveDerivative(bits, k)
					length := n
					w := packBools(bits)
					for i := 0; i < k; i++ {
						length = derivativeStep(w, length)
					}
					ones := 0
					nwords := (length + 63) >> 6
					for i := 0; i < nwords; i++ {
						x := w[i]
						if i == nwords-1 && length&63 != 0 {
							x &= (uint64(1) << uint(length&63)) - 1
						}
						ones += bitsOnes(x)
					}
					S := 2*ones - length
					if S != want {
						t.Fatalf("derivative start=%d k=%d got S %d want %d", n, k, S, want)
					}
				}
			}
		}
	}
}

func bitsOnes(x uint64) int {
	c := 0
	for x != 0 {
		c += int(x & 1)
		x >>= 1
	}
	return c
}
