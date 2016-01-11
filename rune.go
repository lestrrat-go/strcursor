package strcursor

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"unicode/utf8"

	"github.com/lestrrat/go-pdebug"
)

// RuneCursor is a cursor for consumers that are interested in series of
// runes (not bytes)
type RuneCursor struct {
	buf       []byte    // scratch bufer, read in from the io.Reader
	buflen    int       // size of scratch buffer
	bufpos    int       // amount consumed within the scratch buffer
	column    int       // column number
	in        io.Reader // input source
	line      bytes.Buffer
	lineno    int      // line number
	nread     int      // number of bytes consumed so far
	rabuf     *runebuf // Read-ahead buffer.
	lastrabuf *runebuf // the end of read-ahead buffer.
	rabuflen  int      // Number of runes in read-ahead buffer
}

type runebuf struct {
	val   rune
	width int
	next  *runebuf
}

// NewRuneCursor creates a cursor that deals exclusively with runes
func NewRuneCursor(in io.Reader, nn ...int) *RuneCursor {
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
	return &RuneCursor{
		buf:    buf,
		buflen: n,
		bufpos: n, // set to maximum to force filling up the bufer on first read
		column: 1,
		in:     in,
		line:   bytes.Buffer{},
		lineno: 1,
		nread:  0,
		rabuf:  nil,
	}
}

var runebufPool = sync.Pool{
	New: allocRunebuf,
}

func allocRunebuf() interface{} {
	return &runebuf{}
}

func getRunebuf() *runebuf {
	return runebufPool.Get().(*runebuf)
}

func releaseRunebuf(rb *runebuf) {
	rb.next = nil
	runebufPool.Put(rb)
}

// decode the contents of c.buf into runes, and append to the
// read-ahead rune buffer
func (c *RuneCursor) decodeIntoRuneBuffer() error {
	if pdebug.Enabled {
		old := c.rabuflen
		defer func() {
			pdebug.Printf("RuneCursor.decodeIntoRuneBuffer %d -> %d runes", old, c.rabuflen)
		}()
	}

	last := c.lastrabuf
	var err error
	for c.bufpos < c.buflen {
		r, w := utf8.DecodeRune(c.buf[c.bufpos:])
		if r == utf8.RuneError {
			err = errors.New("failed to decode")
			break
		}
		c.bufpos += w
		c.rabuflen++
		cur := getRunebuf()
		cur.val = r
		cur.width = w
		if last == nil {
			c.rabuf = cur
		} else {
			last.next = cur
		}
		last = cur
	}
	c.lastrabuf = last

	if err != nil {
		return err
	}

	return nil
}

func (c *RuneCursor) fillRuneBuffer(n int) error {
	// Check if we have a read-ahead rune buffer
	if c.rabuflen >= n {
		return nil
	}

	if c.buflen == 0 {
		return io.EOF
	}

	// Fill the buffer until we have n runes. However, make sure to
	// detect if we have a failure loop
	prevrabuflen := c.rabuflen
	for {
		// do we have the underlying byte buffer? if we have at least 1 byte,
		// we may be able to decode it
		c.decodeIntoRuneBuffer()
		// we still have a chance to read from the underlying source
		// and succeed in decoding, so we won't return here, even if there
		// was an error

		// we got enough. return success
		if c.rabuflen >= n {
			return nil
		}

		// Hmm, still didn't read anything? try reading from the underlying
		// io.Reader.
		if c.bufpos < c.buflen {
			// first, rescue the remaining bytes, if any. only do the copying
			// when we have something left to consume in the buffer
			copy(c.buf, c.buf[c.bufpos:])
		}

		// reset bufpos.
		nread, err := c.in.Read(c.buf[c.buflen-c.bufpos:])
		if nread == 0 && err != nil {
			// Oh, we're done. really done.
			c.buf = []byte{}
			c.buflen = 0
			return err
		}
		c.buflen = nread + (c.buflen - c.bufpos)
		c.bufpos = 0
		// well, we read something. see if we can fill the rune buffer
		c.decodeIntoRuneBuffer()

		if prevrabuflen == c.rabuflen {
			c.buf = []byte{}
			c.buflen = 0
			return errors.New("failed to fill read buffer")
		}

		prevrabuflen = c.rabuflen
	}

	return errors.New("unrecoverable error")
}

// Done returns true if there are no more runes left.
func (c *RuneCursor) Done() bool {
	if err := c.fillRuneBuffer(1); err != nil {
		return true
	}
	return false
}

// Cur returns the first rune and consumes it.
func (c *RuneCursor) Cur() rune {
	if err := c.fillRuneBuffer(1); err != nil {
		return utf8.RuneError
	}

	// Okay, we got something. Pop off the stack, and we're done
	head := c.rabuf
	c.Advance(1)
	return head.val
}

// Peek returns the first rune without consuming it.
func (c *RuneCursor) Peek() rune {
	return c.PeekN(1)
}

// PeekN returns the n-th rune without consuming it.
func (c *RuneCursor) PeekN(n int) rune {
	if err := c.fillRuneBuffer(n); err != nil {
		return utf8.RuneError
	}

	cur := c.rabuf
	for i := 1; i < n; i++ {
		cur = cur.next
		if cur == nil {
			break
		}
	}
	// Note: This should not happen, because c.fillRuneBuffer should
	// guarantee us that we have at least n elements. But we do this
	// to avoid potential panics
	if cur == nil {
		return utf8.RuneError
	}
	return cur.val
}

// Advance advances the cursor n runes
func (c *RuneCursor) Advance(n int) error {
	head := c.rabuf
	for i := 0; i < n; i++ {
		if head == nil {
			return errors.New("failed to pop enough runes")
		}
		c.nread += head.width
		if head.val == '\n' {
			c.lineno++
			c.line.Reset()
			c.column = 1
		} else {
			c.column++
		}
		n := head
		c.line.WriteRune(n.val)
		head = head.next
		releaseRunebuf(n)
		c.rabuflen--
	}
	c.rabuf = head
	if c.rabuf == nil {
		c.lastrabuf = nil
	}
	return nil
}

func (c *RuneCursor) hasPrefix(s string, n int, consume bool) bool {
	// First, make sure we have enough read ahead buffer
	if err := c.fillRuneBuffer(n); err != nil {
		return false
	}

	count := 0
	for cur := c.rabuf; cur != nil; cur = cur.next {
		r, w := utf8.DecodeRuneInString(s)
		if r == utf8.RuneError {
			return false
		}
		s = s[w:]
		if cur.val != r {
			return false
		}
		count++
		if len(s) == 0 {
			// match! if we have the consume flag set, change the pointers
			if consume {
				c.Advance(count)
			}
			return true
		}
	}
	return false
}

// HasPrefix takes a string returns true if the rune buffer contains
// the exact sequence of runes. This method does NOT consume upon a match
func (c *RuneCursor) HasPrefix(s string) bool {
	n := utf8.RuneCountInString(s)
	return c.hasPrefix(s, n, false)
}

// Consume takes a string and advances the cursor that many runes
// if the rune buffer contains the exact sequence of runes
func (c *RuneCursor) Consume(s string) bool {
	n := utf8.RuneCountInString(s)
	return c.hasPrefix(s, n, true)
}

// Line returns the what we have processed in the current line
func (c *RuneCursor) Line() string {
	return c.line.String()
}

// LineNumber returns the current line number
func (c *RuneCursor) LineNumber() int {
	return c.lineno
}

// Column returns the current column number
func (c *RuneCursor) Column() int {
	return c.column
}