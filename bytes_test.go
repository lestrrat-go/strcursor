package strcursor

import (
	"bytes"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestByteCursorBasic(t *testing.T) {
	buf := bytes.Buffer{}
	for i := 0; i < 100; i++ {
		buf.WriteString(`はろ〜、World!`)
	}
	rdr := bytes.NewReader(buf.Bytes())
	cur := NewByteCursor(rdr)

	{
		r := cur.PeekN(5)
		if !assert.Equal(t, byte(0x82), r, "cur.PeekN(5) succeeds") {
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
	for r := cur.Cur(); r != nilbyte; r = cur.Cur() {
		count++
	}

	if !assert.Equal(t, count, buf.Len(), "cur.Done() should be true") {
		return
	}

	if !assert.True(t, cur.Done(), "cur.Done() should be true") {
		return
	}

}