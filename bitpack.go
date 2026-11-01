package randomness

// expandByte maps a data byte to 8 sequence bits in LSB-first order.
// Sequence bit 0 is the byte's MSB, matching B2bitArr.
var expandByte [256]uint64

func init() {
	for i := 0; i < 256; i++ {
		var v uint64
		for s := 0; s < 8; s++ {
			if i&(0x80>>uint(s)) != 0 {
				v |= 1 << uint(s)
			}
		}
		expandByte[i] = v
	}
}

// packBools packs a bool sequence. Bit i is word i/64, bit i%64.
func packBools(bits []bool) []uint64 {
	w := make([]uint64, (len(bits)+63)>>6)
	for i, b := range bits {
		if b {
			w[i>>6] |= 1 << uint(i&63)
		}
	}
	return w
}

// packBytesMSB packs bytes in the same bit order as B2bitArr.
func packBytesMSB(data []byte) []uint64 {
	nbits := len(data) * 8
	w := make([]uint64, (nbits+63)>>6)
	bit := 0
	for _, b := range data {
		w[bit>>6] |= expandByte[b] << uint(bit&63)
		bit += 8
	}
	return w
}
