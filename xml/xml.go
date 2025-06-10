package xml

import (
	"encoding/xml"
	"io"

	"github.com/thinkgos/encoding/codec"
)

// Codec is a Codec implementation with xml.
type Codec struct{}

// ContentType always Returns "application/xml; charset=utf-8".
func (*Codec) ContentType(_ any) string {
	return "application/xml; charset=utf-8"
}
func (*Codec) Marshal(v any) ([]byte, error) {
	return xml.Marshal(v)
}
func (*Codec) Unmarshal(data []byte, v any) error {
	return xml.Unmarshal(data, v)
}
func (*Codec) NewEncoder(w io.Writer) codec.Encoder {
	return xml.NewEncoder(w)
}
func (*Codec) NewDecoder(r io.Reader) codec.Decoder {
	return xml.NewDecoder(r)
}
