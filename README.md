# go-strcursor

[![Build Status](https://travis-ci.org/lestrrat/go-strcursor.svg?branch=master)](https://travis-ci.org/lestrrat/go-strcursor)
[![GoDoc](https://godoc.org/github.com/lestrrat/go-strcursor?status.svg)](https://godoc.org/github.com/lestrrat/go-strcursor)

Some types of structured text requires you to start parsing using byte semantics,
only to require character semantics after discovering the text's characteristics
such as its encoding.

A good example is XML. You must take into consideration the BOM, then the
XML declaration. The XML declaration is guaranteed to be in ASCII, but
after that you need to look at things character by character after decoding
the content in the specified encoding.

This is a bit tricky because these parsers usually require you to "peek"
into the target buffer. You have to be able to examine the incoming
bytes without consuming it. This in itself is a relatively simple task
but when you have to decode it, you will need to incorporate these
bytes that were read ahead along with those that have not been read yet.
This is important because what you already read ahead might be part of a
multi-byte rune.

If you are working with an Reader type that supports "Unread" operations,
you can get this for 1 byte/character. But this is not enough.

To solve this issue, this package provides `ByteCursor` and `RuneCursor`
objects. Given an `io.Reader`, you can wrap it with a `ByteCursor`,
which gives you byte semantics, with fixed amount of read ahead (by default
40 bytes)

You can first use the `ByteCursor` to parse/consume enough bytes to determine
the encoding:

```go
  // Create a ByteCursor from an io.Reader
  bcur := NewByteCursor(input)

  // Let's say your document starts with line with the
  // encoding name
  encbuf := bytes.Buffer{}
  i := 0
  for c := bcur.PeekN(i+1); c != '\n'; c = bcur.PeekN(i+1) {
    // Maybe validate c...
    encbuf.WriteByte(c)
    i++
  }
  if i < 1 {
    return errors.New("no encoding")
  }

  cur.Advance(i) // Consume `i` bytes
  // now encbuf contains the encoding name
```

Then load this encoding from `golang.org/x/text/encoding` or some
such by name. Let's say this is EUC-JP. Then you can use `bcur`
to as argument to the decoder:

```go
  // bcur implements io.Reader, so you can safely
  // pass it to `Reader()` method
  decoded := japanese.EUCJP.NewDecoder().Reader(bcur)
```

...And feed this to `RuneCursor`, where you can get UTF-8 runes

```go
  // Create a RuneCursor
  rcur := NewRuneCursor(decoded)

  r := rcur.Peek() // This is now a rune, not a byte.
                   // It's also decoded to UTF-8!
```

And there was much rejoicing.




