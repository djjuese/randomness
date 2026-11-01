package randomness

import (
	"math/bits"
	"sync"
)

// lcWords holds packed Berlekamp-Massey state for one block.
// Bit i of a word is sequence index word*64+i.
// R is the reversed history: at the start of step N, R bit k is a[N-1-k].
type lcWords struct {
	S, C, B, T, R []uint64
}

var lcPool = sync.Pool{
	New: func() interface{} {
		return &lcWords{}
	},
}

func resizeZero(s []uint64, n int) []uint64 {
	if cap(s) < n {
		return make([]uint64, n)
	}
	s = s[:n]
	for i := range s {
		s[i] = 0
	}
	return s
}

func resize(s []uint64, n int) []uint64 {
	if cap(s) < n {
		return make([]uint64, n)
	}
	return s[:n]
}

// linearComplexity is the binary Berlekamp-Massey length of a[:M].
// The discrepancy is a word-wise parity against a shifted history register,
// instead of a bit-by-bit inner product.
func linearComplexity(a []bool, M int) int {
	nwords := (M + 63) >> 6
	w := lcPool.Get().(*lcWords)
	defer lcPool.Put(w)
	w.S = resizeZero(w.S, nwords)
	w.C = resizeZero(w.C, nwords)
	w.B = resizeZero(w.B, nwords)
	w.R = resizeZero(w.R, nwords)
	w.T = resize(w.T, nwords)

	for i := 0; i < M; i++ {
		if a[i] {
			w.S[i>>6] |= 1 << uint(i&63)
		}
	}
	w.C[0] = 1
	w.B[0] = 1

	L := 0
	m := -1
	for N := 0; N < M; N++ {
		bitN := int((w.S[N>>6] >> uint(N&63)) & 1)
		d := bitN
		if L > 0 {
			d ^= historyParity(w.C, w.R, L)
		}
		if d != 0 {
			copy(w.T, w.C)
			xorShiftInto(w.C, w.B, N-m, M)
			if L <= N/2 {
				L = N + 1 - L
				m = N
				copy(w.B, w.T)
			}
		}
		// R bit 0 becomes a[N]; older bits move up by one.
		shiftLeftInsert(w.R, bitN)
	}
	return L
}

// historyParity is the GF(2) dot product of C[1..L] with R[0..L-1].
func historyParity(C, R []uint64, L int) int {
	parity := 0
	full := L >> 6
	for i := 0; i < full; i++ {
		c := C[i] >> 1
		if i+1 < len(C) {
			c |= C[i+1] << 63
		}
		parity ^= bits.OnesCount64(c&R[i]) & 1
	}
	if rem := L & 63; rem != 0 {
		c := C[full] >> 1
		if full+1 < len(C) {
			c |= C[full+1] << 63
		}
		mask := uint64(1)<<uint(rem) - 1
		parity ^= bits.OnesCount64(c&R[full]&mask) & 1
	}
	return parity
}

func shiftLeftInsert(R []uint64, bit int) {
	carry := uint64(bit)
	for i := range R {
		next := R[i] >> 63
		R[i] = (R[i] << 1) | carry
		carry = next
	}
}

// xorShiftInto applies C ^= (B << shift), then drops bits at index >= M.
func xorShiftInto(C, B []uint64, shift, M int) {
	n := len(C)
	wordShift := shift >> 6
	bitShift := uint(shift & 63)
	if bitShift == 0 {
		for i := wordShift; i < n; i++ {
			C[i] ^= B[i-wordShift]
		}
	} else {
		for i := wordShift; i < n; i++ {
			v := B[i-wordShift] << bitShift
			src := i - wordShift - 1
			if src >= 0 {
				v |= B[src] >> (64 - bitShift)
			}
			C[i] ^= v
		}
	}
	rem := M & 63
	if rem != 0 {
		C[n-1] &= (uint64(1) << uint(rem)) - 1
	}
}
