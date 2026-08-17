//go:build goexperiment.jsonv2

package webauthn

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"

	"github.com/go-webauthn/webauthn/protocol"
)

// UnmarshalJSONFrom is the encoding/json/v2 equivalent of [Credential.UnmarshalJSON]
func (c *Credential) UnmarshalJSONFrom(decoder *jsontext.Decoder) error {
	type credentialAlias Credential

	var tmp credentialAlias

	if err := jsonv2.UnmarshalDecode(decoder, &tmp); err != nil {
		return err
	}

	*c = Credential(tmp)

	if c.AttestationFormat == "" && protocol.IsAttestationFormatString(c.AttestationType) {
		c.AttestationFormat = c.AttestationType
		c.AttestationType = ""
	}

	return nil
}
