package strcursor

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestByteCursorReader(t *testing.T) {
	pattern := `abcdefghijk`
	buf := bytes.Buffer{}
	for i := 0; i < 100; i++ {
		buf.WriteString(pattern)
	}
	var cur Cursor
	cur = NewByteCursor(&buf)

	if !assert.True(t, cur.ConsumeString(pattern), "Consume succeeds") {
		return
	}

	rdr, ok := cur.(io.Reader)
	if !assert.True(t, ok, "ByteCursor should be an io.Reader") {
		return
	}

	lpat := len(pattern)
	readbuf := make([]byte, lpat)
	for i := 0; i < 99; i++ {
		n, err := rdr.Read(readbuf)
		if !assert.Equal(t, lpat, n, "Read %d bytes", n) {
			return
		}
		if err != nil {
			// It's okay to err, but only at the last iteration
			if !assert.Equal(t, 98, "should be err only if it's the last iteration") {
				return
			}
		}

		if !assert.Equal(t, []byte(pattern), readbuf, "read buffer should match") {
			return
		}
	}
}

func TestByteCursorBasic(t *testing.T) {
	buf := bytes.Buffer{}
	for i := 0; i < 100; i++ {
		buf.WriteString(`はろ〜、World!`)
	}
	rdr := bytes.NewReader(buf.Bytes())
	cur := NewByteCursor(rdr)

	{
		r := cur.PeekN(5)
		if !assert.Equal(t, rune(byte(0x82)), r, "cur.PeekN(5) succeeds") {
			return
		}
	}

	{
		r := cur.Peek()
		if !assert.NotEqual(t, utf8.RuneError, r, "cur.Peek() should succeed") {
			return
		}

		for i := 0; i < 18; i++ {
			if !assert.Equal(t, r, cur.Peek(), "cur.Peek() should keep working") {
				return
			}
		}
	}

	count := 0
	for r := cur.Cur(); r != utf8.RuneError; r = cur.Cur() {
		count++
	}

	if !assert.Equal(t, count, buf.Len(), "cur.Done() should be true") {
		return
	}

	if !assert.True(t, cur.Done(), "cur.Done() should be true") {
		return
	}

}

func TestByteCursorConsume(t *testing.T) {
	rdr := strings.NewReader(`はろ〜、World!`)
	cur := NewByteCursor(rdr)

	if !assert.True(t, cur.HasPrefixString(`はろ〜`), "cur.HasPrefix() succeeds") {
		return
	}

	if !assert.True(t, cur.ConsumeString(`はろ〜`), "cur.Consume() succeeds") {
		return
	}

	if !assert.False(t, cur.HasPrefixString(`はろ〜`), "cur.HasPrefix() fails") {
		return
	}
}
