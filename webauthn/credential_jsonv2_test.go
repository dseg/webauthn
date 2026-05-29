//go:build goexperiment.jsonv2

package webauthn

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time check: *Credential implements the v2 interface.
var _ jsonv2.UnmarshalerFrom = (*Credential)(nil)

// TestCredential_UnmarshalJSONFrom mirrors TestCredential_UnmarshalJSON so the legacy
// AttestationType → AttestationFormat migration is exercised under json/v2.
func TestCredential_UnmarshalJSONFrom(t *testing.T) {
	testCases := []struct {
		name              string
		input             string
		attestationType   string
		attestationFormat string
	}{
		{
			name:              "ShouldMigrateLegacyRecordWithPackedFormat",
			input:             `{"id":"MTIz","publicKey":"YWJj","attestationType":"packed"}`,
			attestationType:   "",
			attestationFormat: "packed",
		},
		{
			name:              "ShouldMigrateLegacyRecordWithNoneFormat",
			input:             `{"id":"MTIz","publicKey":"YWJj","attestationType":"none"}`,
			attestationType:   "",
			attestationFormat: "none",
		},
		{
			name:              "ShouldMigrateLegacyRecordWithFIDOU2FFormat",
			input:             `{"id":"MTIz","publicKey":"YWJj","attestationType":"fido-u2f"}`,
			attestationType:   "",
			attestationFormat: "fido-u2f",
		},
		{
			name:              "ShouldPreserveNewRecordWithBothFields",
			input:             `{"id":"MTIz","publicKey":"YWJj","attestationType":"basic_full","attestationFormat":"packed"}`,
			attestationType:   "basic_full",
			attestationFormat: "packed",
		},
		{
			name:              "ShouldPreserveTypeValueThatIsNotAFormat",
			input:             `{"id":"MTIz","publicKey":"YWJj","attestationType":"basic_full"}`,
			attestationType:   "basic_full",
			attestationFormat: "",
		},
		{
			name:              "ShouldHandleEmptyBothFields",
			input:             `{"id":"MTIz","publicKey":"YWJj"}`,
			attestationType:   "",
			attestationFormat: "",
		},
		{
			name:              "ShouldHandleUnknownTypeString",
			input:             `{"id":"MTIz","publicKey":"YWJj","attestationType":"something-unrecognised"}`,
			attestationType:   "something-unrecognised",
			attestationFormat: "",
		},
		{
			name:              "ShouldNotMigrateWhenFormatAlreadyPresent",
			input:             `{"id":"MTIz","publicKey":"YWJj","attestationType":"packed","attestationFormat":"tpm"}`,
			attestationType:   "packed",
			attestationFormat: "tpm",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var c Credential

			require.NoError(t, jsonv2.Unmarshal([]byte(tc.input), &c))
			assert.Equal(t, tc.attestationType, c.AttestationType)
			assert.Equal(t, tc.attestationFormat, c.AttestationFormat)
		})
	}

	t.Run("ShouldRejectMalformedJSON", func(t *testing.T) {
		var c Credential

		assert.Error(t, jsonv2.Unmarshal([]byte(`{not-json`), &c))
	})
}

// TestCredential_UnmarshalJSONFrom_Direct exercises the method directly with a
// jsontext.Decoder, independent of the json/v2 reflection entry points.
func TestCredential_UnmarshalJSONFrom_Direct(t *testing.T) {
	dec := jsontext.NewDecoder(strings.NewReader(
		`{"id":"MTIz","publicKey":"YWJj","attestationType":"packed"}`,
	))

	var c Credential

	require.NoError(t, c.UnmarshalJSONFrom(dec))
	assert.Equal(t, "", c.AttestationType)
	assert.Equal(t, "packed", c.AttestationFormat)
}

// TestCredential_UnmarshalJSONFrom_RoundTrip confirms the populated credential survives
// a json/v2 Marshal → Unmarshal cycle without losing the AttestationType/Format split.
func TestCredential_UnmarshalJSONFrom_RoundTrip(t *testing.T) {
	original := newPopulatedCredential()

	data, err := jsonv2.Marshal(&original)
	require.NoError(t, err)

	var decoded Credential

	require.NoError(t, jsonv2.Unmarshal(data, &decoded))

	assert.Equal(t, original.AttestationType, decoded.AttestationType)
	assert.Equal(t, original.AttestationFormat, decoded.AttestationFormat)

	redata, err := jsonv2.Marshal(&decoded)
	require.NoError(t, err)
	assert.JSONEq(t, string(data), string(redata))
}
