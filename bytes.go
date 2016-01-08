package strcursor

import (
	"bytes"
	"errors"
	"io"

	"github.com/lestrrat/go-pdebug"
)

const nilbyte = 0x0

type ByteCursor struct {
	buf    []byte    // scratch bufer, read in from the io.Reader
	buflen int       // size of scratch buffer
	bufpos int       // amount consumed within the scratch buffer
	column int       // column number
	in     io.Reader // input source
	lineno int       // line number
}

func NewByteCursor(in io.Reader, nn ...int) *ByteCursor {
	var n int
	if len(nn) > 0 {
		n = nn[0]
	}
	// This buffer is used when reading from the underlying io.Reader.
	// It is necessary to read from the io.Reader because otherwise
	// we can't call utf8.DecodeRune on it
	if n <= 0 {
		// by default, read up to 40 bytes = maximum 10 runes worth of data
		n = 40
	}

	buf := make([]byte, n)
	return &ByteCursor{
		buf:    buf,
		buflen: n,
		bufpos: n, // set to maximum to force filling up the bufer on first read
		column: 1,
		in:     in,
		lineno: 1,
	}
}

func (c *ByteCursor) fillBuffer(n int) error {
	if c.buflen < n {
		return errors.New("fillBuffer request exceeds buffer size")
	}

	if c.buflen-c.bufpos >= n {
		return nil
	}

	if c.bufpos < c.buflen {
		// first, rescue the remaining bytes, if any. only do the copying
		// when we have something left to consume in the buffer
		copy(c.buf, c.buf[c.bufpos:])
	}
	// reset bufpos.
	if c.bufpos >= c.buflen {
		// If bufpos is for some reason > c.buflen, just set it to 0
		c.bufpos = 0
	} else {
		// Otherwise, the remaining bytes up to buflen is the content
		// that is yet to be consumed
		c.bufpos = c.buflen - c.bufpos
	}

	nread, err := c.in.Read(c.buf[c.bufpos:])
	if nread == 0 && err != nil {
		// Oh, we're done. really done.
		c.buf = []byte{}
		c.buflen = 0
		return err
	}

	c.buflen = nread + c.bufpos
	if c.buflen < n {
		return errors.New("fillBuffer request exceeds buffer size")
	}

	return nil
}

func (c *ByteCursor) Done() bool {
	if err := c.fillBuffer(1); err != nil {
		return true
	}
	return false
}

func (c *ByteCursor) Advance(n int) error {
	if err := c.fillBuffer(n); err != nil {
		return err
	}

	c.bufpos += n
	return nil
}

func (c *ByteCursor) Cur() byte {
	b := c.Peek()
	c.Advance(1)
	return b
}

func (c *ByteCursor) Peek() byte {
	return c.PeekN(1)
}

func (c *ByteCursor) PeekN(n int) byte {
	if err := c.fillBuffer(n); err != nil {
		return nilbyte
	}

	return c.buf[c.bufpos+n-1]
}

func (c *ByteCursor) hasPrefix(s []byte, consume bool) bool {
	n := len(s)
	if err := c.fillBuffer(n); err != nil {
		return false
	}

	if !bytes.HasPrefix(c.buf[c.bufpos:], []byte(s)) {
		return false
	}

	if consume {
		c.bufpos += n
	}
	return true
}

func (c *ByteCursor) HasPrefix(s []byte) bool {
	return c.hasPrefix(s, false)
}

func (c *ByteCursor) Consume(s []byte) bool {
	return c.hasPrefix(s, true)
}

func (c *ByteCursor) HasPrefixString(s string) bool {
	return c.hasPrefix([]byte(s), false)
}

func (c *ByteCursor) ConsumeString(s string) bool {
	return c.hasPrefix([]byte(s), true)
}

// Read fulfills the io.Reader interface
func (c *ByteCursor) Read(buf []byte) (int, error) {
	if pdebug.Enabled {
		g := pdebug.IPrintf("START ByteCursor.Read %d bytes", len(buf))
		defer func() {
			g.IRelease("END ByteCursor.Read '%s'", buf)
		}()
	}

	nread := 0
	// Do we have a read ahead buffer?
	if c.bufpos < c.buflen {
		l := len(buf)
		if l >= c.buflen-c.bufpos {
			// their buffer is greater thant what we have.
			// just copy our contents over, and perform a limited read from
			// the underlying io.Reader
			copy(buf, c.buf[c.bufpos:])
			nread = c.buflen - c.bufpos
			buf = buf[nread:]   // advance so io.Read starts at the right place
			c.bufpos = c.buflen // avoid copying next time
		} else {
			copy(buf, c.buf[c.bufpos:c.bufpos+l])
			c.bufpos += l
			return l, nil
		}
	}

	n, err := c.in.Read(buf)
	return n + nread, err
}