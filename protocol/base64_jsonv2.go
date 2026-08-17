//go:build goexperiment.jsonv2

package protocol

import (
	"encoding/base64"
	"encoding/json/jsontext"
)

// MarshalJSONTo is the encoding/json/v2 equivalent of [URLEncodedBase64.MarshalJSON].
//
// Note: No UnmarshalJSONFrom is defined for production use. This is intentional.
// json/v2 already delegates to the v1 [URLEncodedBase64.UnmarshalJSON] (see the
// jsonUnmarshalerType dispatch in encoding/json/v2/arshal_methods.go), and that
// delegation is not worth replacing: a native UnmarshalJSONFrom built on
// decoder.ReadValue benchmarks byte-for-byte identical to the delegation (113 ns /
// 2 allocs on the go1.27 line), while a decoder.ReadToken variant is slower
// (151 ns / 3 allocs, the extra alloc being Token.String).
func (e URLEncodedBase64) MarshalJSONTo(encoder *jsontext.Encoder) error {
	if e == nil {
		return encoder.WriteToken(jsontext.Null)
	}

	b := encoder.AvailableBuffer()
	b = append(b, '"')
	b = base64.RawURLEncoding.AppendEncode(b, e)
	b = append(b, '"')

	return encoder.WriteValue(b)
}
