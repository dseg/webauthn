//go:build goexperiment.jsonv2

package protocol

import (
	"encoding/base64"
	"encoding/json/jsontext"
	"fmt"
	"strings"
)

// MarshalJSONTo is the encoding/json/v2 equivalent of [URLEncodedBase64.MarshalJSON]:
// nil emits a JSON null, any other value (including an empty slice) emits a base64-url
// quoted string. json/v2 calls this in preference to MarshalJSON, avoiding the per-call
// []byte buffer allocation that v1 imposes.
func (e URLEncodedBase64) MarshalJSONTo(enc *jsontext.Encoder) error {
	if e == nil {
		return enc.WriteToken(jsontext.Null)
	}

	// Base64-url alphabet is JSON-safe, so we hand-assemble "<b64>" into the
	// encoder's own scratch buffer and emit it as a single raw value. This
	// skips the string returned by EncodeToString and the jsontext.Token wrap.
	b := enc.AvailableBuffer()
	b = append(b, '"')
	b = base64.RawURLEncoding.AppendEncode(b, e)
	b = append(b, '"')

	return enc.WriteValue(b)
}

// UnmarshalJSONFrom is the encoding/json/v2 equivalent of [URLEncodedBase64.UnmarshalJSON].
// json/v2 calls this in preference to UnmarshalJSON, avoiding the per-call []byte buffer
// allocation that v1 imposes. A null token is a no-op (matching the v1 method); a string
// token is base64-url-decoded into the receiver. Trailing '=' padding is tolerated even
// though raw URL encoding does not produce it.
func (e *URLEncodedBase64) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}

	switch tok.Kind() {
	case 'n':
		return nil
	case '"':
		s := strings.TrimRight(tok.String(), "=")

		out, err := base64.RawURLEncoding.DecodeString(s)
		if err != nil {
			return err
		}

		*e = out

		return nil
	default:
		return fmt.Errorf("URLEncodedBase64: expected JSON string, got %v", tok.Kind())
	}
}
