package strcursor

import (
	"io"
)

type Unused struct {
	unused []byte
	rdr    io.Reader
}

func (u *Unused) Read(b []byte) (int, error) {
	if len(u.unused) > 0 {
		copy(b, u.unused)
		u.unused = nil
		return len(u.unused), nil
	}

	return u.rdr.Read(b)
}
