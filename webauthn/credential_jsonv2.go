//go:build goexperiment.jsonv2

package webauthn

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"

	"github.com/go-webauthn/webauthn/protocol"
)

// UnmarshalJSONFrom is the encoding/json/v2 equivalent of [Credential.UnmarshalJSON]:
// it decodes the next JSON value from dec into the receiver and applies the same
// backward-compatibility migration — if the decoded record has no AttestationFormat
// and the AttestationType value is a recognised attestation FORMAT identifier (i.e.
// "packed", "tpm", "none"), the value is moved to AttestationFormat and
// AttestationType is cleared so callers can re-derive the true attestation type by
// calling [Credential.Verify]. The credentialAlias indirection avoids infinite
// recursion: the alias type drops Credential's method set, so jsonv2.UnmarshalDecode
// uses the default struct decoder rather than re-entering UnmarshalJSONFrom.
func (c *Credential) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	type credentialAlias Credential

	var tmp credentialAlias

	if err := jsonv2.UnmarshalDecode(dec, &tmp); err != nil {
		return err
	}

	*c = Credential(tmp)

	if c.AttestationFormat == "" && protocol.IsAttestationFormatString(c.AttestationType) {
		c.AttestationFormat = c.AttestationType
		c.AttestationType = ""
	}

	return nil
}
