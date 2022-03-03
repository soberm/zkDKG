package dkg

func PadTrimLeft(b []byte, size int) []byte {
	l := len(b)
	if l == size {
		return b
	}
	if l > size {
		return b[l-size:]
	}
	tmp := make([]byte, size)
	copy(tmp[size-l:], b)
	return tmp
}
