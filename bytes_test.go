package strcursor

import (
	"bytes"
	"io/ioutil"
	"strings"
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

func TestByteCursorReader(t *testing.T) {
	rdr := strings.NewReader(string([]byte{0xfe,0xff}) + "はろ〜、World!")
	cur := NewByteCursor(rdr)

	if !assert.True(t, cur.Consume([]byte{0xfe, 0xff}), "Consume should succeed") {
		return
	}

	buf, err := ioutil.ReadAll(cur)
	if !assert.NoError(t, err, "ioutil.ReadAll should succeed") {
		return
	}

	if !assert.Equal(t, buf, []byte(`はろ〜、World!`), "Read content matches") {
		return
	}
}
