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
			name:              "ShouldMigrateLegacyRecordWithAndroidKeyFormat",
			input:             `{"id":"MTIz","publicKey":"YWJj","attestationType":"android-key"}`,
			attestationType:   "",
			attestationFormat: "android-key",
		},
		{
			name:              "ShouldMigrateLegacyRecordWithTPMFormat",
			input:             `{"id":"MTIz","publicKey":"YWJj","attestationType":"tpm"}`,
			attestationType:   "",
			attestationFormat: "tpm",
		},
		{
			name:              "ShouldMigrateLegacyRecordWithAndroidSafetyNetFormat",
			input:             `{"id":"MTIz","publicKey":"YWJj","attestationType":"android-safetynet"}`,
			attestationType:   "",
			attestationFormat: "android-safetynet",
		},
		{
			name:              "ShouldMigrateLegacyRecordWithAppleFormat",
			input:             `{"id":"MTIz","publicKey":"YWJj","attestationType":"apple"}`,
			attestationType:   "",
			attestationFormat: "apple",
		},
		{
			name:              "ShouldMigrateLegacyRecordWithCompoundFormat",
			input:             `{"id":"MTIz","publicKey":"YWJj","attestationType":"compound"}`,
			attestationType:   "",
			attestationFormat: "compound",
		},
		{
			name:              "ShouldPreserveNewRecordWithBothFields",
			input:             `{"id":"MTIz","publicKey":"YWJj","attestationType":"basic_full","attestationFormat":"packed"}`,
			attestationType:   "basic_full",
			attestationFormat: "packed",
		},
		{
			name:              "ShouldPreserveNewRecordWithSurrogate",
			input:             `{"id":"MTIz","publicKey":"YWJj","attestationType":"basic_surrogate","attestationFormat":"packed"}`,
			attestationType:   "basic_surrogate",
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

	// Matches the v1 Credential.UnmarshalJSON semantics: fields absent from the JSON
	// payload are reset to their zero value on the destination, not merged.
	t.Run("ShouldResetPreExistingFields", func(t *testing.T) {
		c := Credential{AttestationType: "leftover", AttestationFormat: "leftover-format"}

		require.NoError(t, jsonv2.Unmarshal([]byte(`{"id":"MTIz"}`), &c))
		assert.Empty(t, c.AttestationType)
		assert.Empty(t, c.AttestationFormat)
	})

	// Confirms the v2 []byte path actually uses standard base64 (not URL-safe): "////" and "++++"
	// decode to bytes that contain characters which differ between the two alphabets, so any future
	// drift to a URL-safe default would surface here instead of silently passing.
	t.Run("ShouldDecodeBytesContainingNonURLSafeBase64Chars", func(t *testing.T) {
		var c Credential

		require.NoError(t, jsonv2.Unmarshal(
			[]byte(`{"id":"////","publicKey":"++++"}`), &c,
		))
		assert.Equal(t, []byte{0xff, 0xff, 0xff}, c.ID)
		assert.Equal(t, []byte{0xfb, 0xef, 0xbe}, c.PublicKey)
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

func BenchmarkJSONv2MarshalCredential(b *testing.B) {
	v := newPopulatedCredential()

	out, err := jsonv2.Marshal(&v)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(out)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := jsonv2.Marshal(&v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONv2UnmarshalCredential(b *testing.B) {
	v := newPopulatedCredential()

	data, err := jsonv2.Marshal(&v)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var d Credential

		if err := jsonv2.Unmarshal(data, &d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONv2MarshalCredentialAttestation(b *testing.B) {
	v := newPopulatedCredentialAttestation()

	out, err := jsonv2.Marshal(&v)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(out)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := jsonv2.Marshal(&v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONv2UnmarshalCredentialAttestation(b *testing.B) {
	v := newPopulatedCredentialAttestation()

	data, err := jsonv2.Marshal(&v)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var d CredentialAttestation

		if err := jsonv2.Unmarshal(data, &d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONv2MarshalCredentials(b *testing.B) {
	v := Credentials{newPopulatedCredential(), newPopulatedCredential()}

	out, err := jsonv2.Marshal(v)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(out)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := jsonv2.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONv2UnmarshalCredentials(b *testing.B) {
	v := Credentials{newPopulatedCredential(), newPopulatedCredential()}

	data, err := jsonv2.Marshal(v)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var d Credentials

		if err := jsonv2.Unmarshal(data, &d); err != nil {
			b.Fatal(err)
		}
	}
}
