package strcursor

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestRuneCursorBasic(t *testing.T) {
	buf := bytes.Buffer{}
	for i := 0; i < 100; i++ {
		buf.WriteString(`はろ〜、World!`)
	}
	rdr := bytes.NewReader(buf.Bytes())
	cur := NewRuneCursor(rdr)

	{
		r := cur.PeekN(5)
		if !assert.Equal(t, 'W', r, "cur.PeekN(5) succeeds") {
			return
		}
	}

	{
		r := cur.Peek()
		if !assert.NotEqual(t, utf8.RuneError, r, "cur.Peek() should succeed") {
			return
		}

		for i := 0; i < 10; i++ {
			if !assert.Equal(t, r, cur.Peek(), "cur.Peek() should keep working") {
				return
			}
		}
	}

	runecount := utf8.RuneCount(buf.Bytes())
	count := 0
	for r := cur.Cur(); r != utf8.RuneError; r = cur.Cur() {
		count++
	}
	if !assert.Equal(t, count, runecount, "Read expected count of runes") {
		return
	}

	if !assert.True(t, cur.Done(), "cur.Done() should be true") {
		return
	}
}

func TestRuneCursorConsume(t *testing.T) {
	rdr := strings.NewReader(`はろ〜、World!`)
	cur := NewRuneCursor(rdr)

	if !assert.True(t, cur.HasPrefix(`はろ〜`), "cur.HasPrefix() succeeds") {
		return
	}

	if !assert.True(t, cur.Consume(`はろ〜`), "cur.Consume() succeeds") {
		return
	}

	if !assert.False(t, cur.HasPrefix(`はろ〜`), "cur.HasPrefix() fails") {
		return
	}
}

func TestRuneCursorNewLines(t *testing.T) {
	rdr := strings.NewReader(`Alice
Bob
Charlie`)
	cur := NewRuneCursor(rdr)

	if !assert.Equal(t, 1, cur.LineNumber(), "cur.LineNumber() is 1") {
		return
	}
	if !assert.Equal(t, 1, cur.Column(), "cur.Column() is 1") {
		return
	}

	if !assert.True(t, cur.Consume("Al"), "cur.Consume() succeeds") {
		return
	}
	if !assert.Equal(t, 3, cur.Column(), "cur.Column() is 3") {
		return
	}
	if !assert.True(t, cur.Consume("ice\n"), "cur.Consume() succeeds") {
		return
	}

	if !assert.Equal(t, 2, cur.LineNumber(), "cur.LineNumber() is 2") {
		return
	}
	if !assert.Equal(t, 1, cur.Column(), "cur.Column() is 1") {
		return
	}
}
