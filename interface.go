package strcursor

import (
	"bytes"
	"io"
)

type Cursor interface {
	io.Reader

	// Advance moves the position by the requested count of runes.
	Advance(int) error

	// Column returns the current column number
	Column() int

	Consume([]byte) bool

	// ConsumeString advances the cursor position by the length of the
	// input string, only if the input string appears as the next set of
	// data in the cursor. It returns true if the string was found
	ConsumeString(string) bool

	// Cur returns the current rune and consumes the rune (calls
	// Advance()). On error, it returns utf8.RuneError
	Cur() rune

	// Done returns true if we have exhausted this cursor
	Done() bool

	// HasPrefix checks if the cursor has the specified set of bytes as prefix
	HasPrefix([]byte) bool

	// HasPrefix checks if the cursor has the specified string as prefix
	HasPrefixString(string) bool

	// Line returns the current line that we are processing
	Line() string

	// LineNumber returns the current line number
	LineNumber() int

	// Peek returns the current rune, but does not advance the position
	// On error, it returns utf8.RuneError
	Peek() rune

	// PeekN returns the rune at requested position.
	// It does not advance the position.
	// On error, it returns utf8.RuneError
	PeekN(int) rune

	// Unused returns a new io.Reader that contains everything that has
	// not already been consumed
	Unused() io.Reader
}

// ByteCursor is a cursor for consumers that are interested in series of
// bytes
type ByteCursor struct {
	buf    []byte       // scratch bufer, read in from the io.Reader
	buflen int          // size of scratch buffer
	bufpos int          // amount consumed within the scratch buffer
	column int          // column number
	in     io.Reader    // input source
	line   bytes.Buffer // current line
	lineno int          // line number
}

// RuneCursor is a cursor for consumers that are interested in series of
// runes (not bytes)
type RuneCursor struct {
	buf       []byte       // scratch bufer, read in from the io.Reader
	buflen    int          // size of scratch buffer
	bufpos    int          // amount consumed within the scratch buffer
	column    int          // column number
	in        io.Reader    // input source
	line      bytes.Buffer // current line
	lineno    int          // line number
	nread     int          // number of bytes consumed so far
	rabuf     *runebuf     // Read-ahead buffer.
	lastrabuf *runebuf     // the end of read-ahead buffer.
	rabuflen  int          // Number of runes in read-ahead buffer
}

type runebuf struct {
	val   rune
	width int
	next  *runebuf
}
