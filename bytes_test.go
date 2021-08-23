package strcursor_test

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/lestrrat-go/strcursor"
	"github.com/stretchr/testify/assert"
)

func init() {
	c := strcursor.NewByteCursor(bytes.NewReader([]byte("Test")))
	var _ strcursor.Cursor = c
	var _ io.Reader = c
}

func TestByteCursor(t *testing.T) {
	t.Run("Consume", func(t *testing.T) {
		rdr := strings.NewReader(`はろ〜、World!`)
		cur := strcursor.NewByteCursor(rdr)

		if !assert.True(t, cur.HasPrefixString(`はろ〜`), "cur.HasPrefix() succeeds") {
			return
		}

		if !assert.True(t, cur.ConsumeString(`はろ〜`), "cur.Consume() succeeds") {
			return
		}

		if !assert.False(t, cur.HasPrefixString(`はろ〜`), "cur.HasPrefix() fails") {
			return
		}
	})
	t.Run("Reader", func(t *testing.T) {
		const count = 100
		const pattern = `abcdefghijk`
		const epilogue = `---epilogue---`
		var buf bytes.Buffer
		for i := 0; i < count; i++ {
			fmt.Fprintf(&buf, "%02d-%s", i, pattern)
		}
		// Add an epilogue that needs to be rescued via Unused()
		buf.WriteString(epilogue)

		rdr := strcursor.NewByteCursor(&buf, 19)

		lpat := len(pattern) + 3
		readbuf := make([]byte, lpat)
		for i := 0; i < count; i++ {
			nextPattern := fmt.Sprintf("%02d-%s", i, pattern)
			if i%2 == 0 {
				if !assert.True(t, rdr.ConsumeString(nextPattern), `rdr.Consume should succeed`) {
					return
				}
			} else {
				n, err := rdr.Read(readbuf)
				if !assert.NoError(t, err, `rdr.Read should suceed`) {
					return
				}

				if !assert.Equal(t, lpat, n, "Read %d bytes", n) {
					return
				}

				if !assert.Equal(t, []byte(nextPattern), readbuf, "read buffer should match") {
					return
				}
			}
		}

		unused := rdr.Unused()
		epilogueRead, err := io.ReadAll(unused)
		if !assert.NoError(t, err, `io.ReadAll should succeed`) {
			return
		}
		if !assert.Equal(t, epilogue, string(epilogueRead), `epilogue should match`) {
			return
		}
	})
	t.Run("API", func(t *testing.T) {
		var buf bytes.Buffer
		for i := 0; i < 100; i++ {
			buf.WriteString(`はろ〜、World!`)
		}
		rdr := bytes.NewReader(buf.Bytes())
		cur := strcursor.NewByteCursor(rdr)

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
	})
}
