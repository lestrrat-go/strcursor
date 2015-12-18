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

	for i := 0; i < 5; i++ {
		if !assert.True(t, utf8.ValidRune(b.Next()), "utf8.Valid(b.Next()) is true") {
			return
		}
	}

	if !assert.Equal(t, "ASC", b.Consume(3)) {
		return
	}

	if !assert.Equal(t, "IIと日本語が入り交じった文章です", b.Consume(100)) {
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
