package strcursor

import (
	"bytes"
	"errors"
	"io"
	"unicode/utf8"
)

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

// Unused returns the unused portion of the underlying buffer.
// Users should stop using the cursor before calling this method.
func (c *ByteCursor) Unused() io.Reader {
	ret := &Unused{rdr: c.in}
	if buf := c.buf[c.bufpos:]; len(buf) > 0 {
		ret.unused = buf
	}
	return ret
}

func (c ByteCursor) Column() int {
	return c.column
}

func (c ByteCursor) Line() string {
	return ""
}

func (c ByteCursor) LineNumber() int {
	return c.lineno
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
	offset := c.buflen - c.bufpos
	// "memclear"
	for i := offset; i < c.buflen; i++ {
		c.buf[i] = 0x0
	}
	c.bufpos = 0

	nread, err := c.in.Read(c.buf[offset:])
	if nread == 0 && err != nil {
		// Oh, we're done. really done.
		c.buf = []byte{}
		c.buflen = 0
		return err
	}

	c.buflen = nread + offset
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

	if i := bytes.IndexByte(c.buf[c.bufpos:c.bufpos+n], '\n'); i > -1 {
		c.lineno++
		c.column = n - i + 1
		c.line.Reset()
		c.line.Write(c.buf[c.bufpos+i : c.bufpos+n])
	} else {
		c.column += n
		c.line.Write(c.buf[c.bufpos : c.bufpos+n])
	}
	c.bufpos += n
	return nil
}

func (c *ByteCursor) Cur() rune {
	b := c.Peek()
	c.Advance(1)
	return b
}

func (c *ByteCursor) Peek() rune {
	return c.PeekN(1)
}

func (c *ByteCursor) PeekN(n int) rune {
	if err := c.fillBuffer(n); err != nil {
		return utf8.RuneError
	}

	return rune(c.buf[c.bufpos+n-1])
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
