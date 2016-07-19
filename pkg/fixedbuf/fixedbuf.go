package fixedbuf

type FixedBuf struct {
	size int
	buf  []byte
}

func New(size int) *FixedBuf {
	b := &FixedBuf{
		size: size,
		buf:  make([]byte, 0, size),
	}
	return b
}

func (b *FixedBuf) Write(d []byte) (n int, err error) {
	dLen := len(d)
	bLen := len(b.buf)
	free := b.size - bLen

	if dLen <= free {
		b.buf = b.buf[0 : bLen+dLen]
		copy(b.buf[bLen:bLen+dLen], d)
	} else if (bLen + dLen) > b.size {
		if dLen <= b.size {
			bytesToKeep := b.size - dLen
			copy(b.buf, b.buf[bLen-bytesToKeep:])
			b.buf = b.buf[:b.size]
			copy(b.buf[bytesToKeep:], d)
		} else {
			b.buf = b.buf[0:b.size]
			copy(b.buf, d[dLen-b.size:])
		}
	}

	return dLen, nil
}

func (b *FixedBuf) Bytes() []byte {
	d := make([]byte, len(b.buf))
	copy(d, b.buf)
	return d
}
