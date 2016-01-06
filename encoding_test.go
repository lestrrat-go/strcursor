// +build encoding

package strcursor

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/encoding/japanese"
)

func TestEncoding(t *testing.T) {
	txt := `はろう、World!`
	euc, err := japanese.EUCJP.NewEncoder().String(txt)
	if !assert.NoError(t, err, "Encoder.String works") {
		return
	}

	// This doesn't make sense, I know -- putting a UTF8 BOM on a
	// EUC-JP string. But this is a test to just read some bytes from
	// a ByteCursor, use encoding.Decoder.Reader to get a reader,
	// and have RuneCursor read the runes, so anything was OK
	buf := bytes.Buffer{}
	buf.Write([]byte{0xfe, 0xff})
	buf.WriteString(euc)

	bcur := NewByteCursor(&buf)
	if !assert.True(t, bcur.Consume([]byte{0xfe, 0xff}), "Consume works") {
		return
	}

	rcur := NewRuneCursor(japanese.EUCJP.NewDecoder().Reader(bcur))
	if !assert.True(t, rcur.Consume(txt), "Consume works") {
		return
	}
}