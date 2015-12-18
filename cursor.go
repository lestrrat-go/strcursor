// Package strcursor contains objects to make inspecting UTF-8 strings
// on a character basis easier.
package strcursor

import (
	"math/rand"
	"time"
	"unicode/utf8"

	"github.com/lestrrat/go-strcursor/internal/debug"
)

// Cursor allows you to inspect small chunks of characters efficiently
type Cursor struct {
	off       int    // raw offset, or how many bytes we have already consumed
	buf       []byte // raw byte buffer. this should be immutable during the lifecycle of this buffer
	bufmax    int
	cache     []rune     // list of runes we have already decoded, but haven't consumed
	cacheoff  int        // cache offset, or how many bytes we have already decoded
	random    *rand.Rand // random source for purge
	nextpurge int        // purge the underlying cache slice buffer after this many bytes read
}

// PurgeThreshold is the global threshold for when the Cursor should
// start purging rune caches
var PurgeThreshold = 1024 * 1024 * 10 // 10MB

// New creates a new cursor
func New(b []byte) *Cursor {
	l := len(b)
	buf := &Cursor{buf: b, bufmax: l}
	if l > PurgeThreshold { // start purging slices if buffer is this big
		buf.random = rand.New(rand.NewSource(time.Now().UnixNano()))
		buf.nextpurge = buf.random.Intn(l)
	}
	return buf
}

// Bytes returns the unconsumed bytes
func (b *Cursor) Bytes() []byte {
	return b.buf[b.off:]
}

// fills the rune cache. returns true if we have enough to serve n runes
func (b *Cursor) fill(n int) bool {
	if b.Done() {
		if debug.Enabled {
			debug.Printf("  -> already done, false")
		}
		return false
	}

	if len(b.cache) >= n {
		return true
	}

	for len(b.cache) < n {
		if b.cacheoff >= b.bufmax {
			if debug.Enabled {
				debug.Printf("  -> cache offset (%d) >= bufmax (%d), false", b.cacheoff, b.bufmax)
			}
			return false
		}

		r, w := utf8.DecodeRune(b.buf[b.cacheoff:])
		b.cacheoff += w
		b.cache = append(b.cache, r)
	}
	return true
}

// Peek returns the n-th rune in the buffer (base 1, so if you want the
// 8th rune, you use n = 8, not n = 7)
func (b *Cursor) Peek(n int) rune {
	if n <= 0 {
		return utf8.RuneError
	}

	if !b.fill(n) {
		return utf8.RuneError
	}

	return b.cache[n-1]
}

// Advance moves the cursor n characters so that that many characters are
// deemed "consumed" already. Advance must receive a number >= 0. If you
// pass a number < 1, there are not enough characters to be consumed, or you
// have already reached the end of the buffer, this method returns false
func (b *Cursor) Advance(n int) bool {
	if debug.Enabled {
		debug.Printf("Cursor.Advance(%d)", n)
	}

	if n <= 0 { // this is an error
		if debug.Enabled {
			debug.Printf("  -> n <= 0, false")
		}
		return false
	}

	if !b.fill(n) {
		if debug.Enabled {
			debug.Printf("  -> fill failed, false")
		}
		return false
	}

	for i := 0; i < n; i++ {
		l := utf8.RuneLen(b.cache[i])
		b.off += l
	}

	if b.nextpurge > 0 && b.off > b.nextpurge {
		// XXX slow, but saves a tree
		b.cache = append([]rune(nil), b.cache[n:]...)
		if b.bufmax/5 <= b.off {
			b.nextpurge = 0
		} else {
			b.nextpurge = b.off + b.random.Intn(b.bufmax-b.off)
		}
	} else {
		b.cache = b.cache[n:]
	}

	return true
}

// Next consumes the next character. If the operation fails, utf8.RuneError is
// returned
func (b *Cursor) Next() rune {
	if !b.fill(1) {
		return utf8.RuneError
	}
	r := b.cache[0]
	b.Advance(1)
	return r
}

// Done returns true if we have consumed all of the characters
func (b *Cursor) Done() bool {
	return b.off >= b.bufmax
}

// HasPrefix checks if the given string exists at the beginning of the
// Cursor. This method does NOT advance the cursor, so it's basically a
// string-wise Peek() + equality check
func (b *Cursor) HasPrefix(s string) bool {
	if debug.Enabled {
		debug.Printf("Cursor.HasPrefix(%s)", s)
	}
	l := utf8.RuneCountInString(s)
	if !b.fill(l) {
		return false
	}

	if string(b.cache[:l]) != s {
		if debug.Enabled {
			debug.Printf("  -> prefix (%s) != s (%s), false", string(b.cache[:l]), s)
		}
		return false
	}
	return true
}

// ConsumePrefix checks if the given string exists at the beginning of the
// Cursor, and if it does, consumes the buffer by advancing the cursor
// as required.
func (b *Cursor) ConsumePrefix(s string) bool {
	if debug.Enabled {
		debug.Printf("Cursor.ConsumePrefix(%s)", s)
	}

	l := utf8.RuneCountInString(s)
	if !b.fill(l) {
		return false
	}

	if string(b.cache[:l]) != s {
		if debug.Enabled {
			debug.Printf("  -> prefix (%s) != s (%s), false", string(b.cache[:l]), s)
		}
		return false
	}
	b.Advance(l)
	return true
}