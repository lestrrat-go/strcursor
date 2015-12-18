package strcursor_test

import (
	"testing"
	"unicode/utf8"

	"github.com/lestrrat/go-strcursor"
	"github.com/stretchr/testify/assert"
)

func TestBuffer(t *testing.T) {
	strcursor.PurgeThreshold = 5

	s := "hello, 日本! これは ASCIIと日本語が入り交じった文章です"
	l := utf8.RuneCountInString(s)
	b := strcursor.New([]byte(s))

	if !assert.Equal(t, '日', b.Peek(8), "Peek succeeds") {
		return
	}

	if !assert.True(t, b.HasPrefix("hello, 日本!"), "HasPrefix matches") {
		return
	}

	if !assert.False(t, b.HasPrefix("hallo, 日本!"), "HasPrefix fails") {
		return
	}

	if !assert.Equal(t, 'す', b.Peek(l), "Peek (max) succeeds") {
		return
	}

	if !assert.True(t, b.Advance(7), "Advance(7) succeeds") {
		return
	}

	if !assert.Equal(t, '日', b.Peek(1), "Peek after Advance succeeds") {
		return
	}

	if !assert.True(t, b.HasPrefix("日本!"), "HasPrefix matches") {
		return
	}

	if !assert.False(t, b.ConsumePrefix("日本語!"), "ConsumePrefix (not matching) fails") {
		return
	}

	if !assert.True(t, b.ConsumePrefix("日本!"), "ConsumePrefix (matching) succeeds") {
		return
	}

	if !assert.True(t, b.HasChars(5), "HasChars(5) returns true") {
		return
	}

	for i := 0; i < 5; i++ {
		if !assert.True(t, utf8.ValidRune(b.Next()), "utf8.Valid(b.Next()) is true") {
			return
		}
	}

	if !assert.Equal(t, "ASC", b.Consume(3)) {
		return
	}

	if !assert.True(t, b.HasChars(17), "HasChars(17) returns true") {
		return
	}

	if !assert.False(t, b.HasChars(18), "HasChars(18) returns false") {
		return
	}

	if !assert.True(t, b.HasPrefixBytes([]byte{'I', 'I'}), "HasPrefixBytes returns true") {
		return
	}

	if !assert.Equal(t, []byte{'I', 'I'}, b.PeekBytes(2), "PeekBytes matches") {
		return
	}

	if !assert.Equal(t, []byte{'I', 'I'}, b.ConsumeBytes(2), "ConsumeBytes matches") {
		return
	}

	if !assert.Equal(t, "と日本語が入り交じった文章です", b.Consume(100)) {
		return
	}

	if !assert.Equal(t, "", b.Consume(100)) {
		return
	}

	if !assert.True(t, b.Done(), "Done is true") {
		return
	}

	// From here on, every operation should be invalid, but should still not
	// cause any panic
	if !assert.Equal(t, utf8.RuneError, b.Peek(1), "Peek after being done returns utf8.RuneError") {
		return
	}

	if !assert.False(t, b.Advance(1), "Next after being done returns false") {
		return
	}

	if !assert.Equal(t, utf8.RuneError, b.Next(), "Next after being done returns utf8.RuneError") {
		return
	}

	if !assert.Len(t, b.Bytes(), 0, "Bytes() should return 0 length slice") {
		return
	}
}

func TestLineno(t *testing.T) {
	b := strcursor.New([]byte(`Alice
Bob
Charlie
David
Ellis`))

	// Before doing anything
	if !assert.Equal(t, 1, b.LineNumber(), "LineNumber == 1") {
		return
	}
	if !assert.Equal(t, 1, b.Column(), "Column == 1") {
		return
	}

	b.Next() // 'A'
	if !assert.Equal(t, 1, b.LineNumber(), "LineNumber still 1") {
		return
	}
	if !assert.Equal(t, 2, b.Column(), "Column == 2") {
		return
	}

	b.Consume(7) // 'lice\nBo'
	if !assert.Equal(t, 2, b.LineNumber(), "LineNumber == 2") {
		return
	}
	if !assert.Equal(t, 3, b.Column(), "Column == 3") {
		return
	}

	if !assert.Equal(t, "b\nCharlie\n", b.Consume(10), "Consume(10)") {
		return
	}
	if !assert.Equal(t, 3, b.LineNumber(), "LineNumber == 3") {
		return
	}
	if !assert.Equal(t, 9, b.Column(), "Column == 9") {
		return
	}

	if !assert.Equal(t, "David\nEllis", b.Consume(12), "Consume(12)") {
		return
	}

	if !assert.Equal(t, 5, b.LineNumber(), "LineNumber == 5") {
		return
	}

	if !assert.Equal(t, 5, b.LineNumber(), "LineNumber == 5 (second time)") {
		return
	}

}
