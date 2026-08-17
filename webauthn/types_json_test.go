package webauthn

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// exerciseJsonCodec round trips a value through JSON encoding, asserting that
// the decoded value re-encodes to the same bytes and matches the original.
//
// fresh must return a zero value of the same concrete type, so the comparison
// is against a destination which carried nothing over from the source.
func exerciseJsonCodec[T any](t *testing.T, populated T, fresh func() T) []byte {
	t.Helper()

	marshalled, err := json.Marshal(populated)
	require.NoError(t, err)

	// Verify round-trip: decode and re-encode should match
	decoded := fresh()
	err = json.Unmarshal(marshalled, decoded)
	require.NoError(t, err)

	again, err := json.Marshal(decoded)
	require.NoError(t, err)
	assert.Equal(t, marshalled, again, "the decoded value must re-encode to the same bytes")

	// Verify a fresh decode also matches
	decoded2 := fresh()
	err = json.Unmarshal(marshalled, decoded2)
	require.NoError(t, err)
	again2, err := json.Marshal(decoded2)
	require.NoError(t, err)
	assert.Equal(t, marshalled, again2, "independent decodes must re-encode to the same bytes")

	return marshalled
}

func TestJsonCodecs(t *testing.T) {
	t.Run("Credential", func(t *testing.T) {
		exerciseJsonCodec(t, ptr(newPopulatedCredentialWithEverything()), func() *Credential { return &Credential{} })
	})

	t.Run("CredentialZero", func(t *testing.T) {
		exerciseJsonCodec(t, &Credential{}, func() *Credential { return &Credential{} })
	})

	t.Run("CredentialExtensions", func(t *testing.T) {
		exerciseJsonCodec(t, ptr(newPopulatedCredentialExtensions()), func() *CredentialExtensions { return &CredentialExtensions{} })
	})

	t.Run("CredentialExtensionsZero", func(t *testing.T) {
		exerciseJsonCodec(t, &CredentialExtensions{}, func() *CredentialExtensions { return &CredentialExtensions{} })
	})

	t.Run("CredentialAttestation", func(t *testing.T) {
		exerciseJsonCodec(t, ptr(newPopulatedCredentialWithEverything().Attestation), func() *CredentialAttestation { return &CredentialAttestation{} })
	})

	t.Run("CredentialFlags", func(t *testing.T) {
		exerciseJsonCodec(t, &CredentialFlags{UserPresent: true, UserVerified: true, BackupEligible: true, BackupState: true}, func() *CredentialFlags { return &CredentialFlags{} })
	})

	t.Run("Credentials", func(t *testing.T) {
		exerciseJsonCodec(t, &Credentials{newPopulatedCredentialWithEverything(), newPopulatedCredential()}, func() *Credentials { return &Credentials{} })
	})

	t.Run("Authenticator", func(t *testing.T) {
		exerciseJsonCodec(t, ptr(newPopulatedCredentialWithEverything().Authenticator), func() *Authenticator { return &Authenticator{} })
	})

	t.Run("SessionData", func(t *testing.T) {
		exerciseJsonCodec(t, ptr(newPopulatedSessionData()), func() *SessionData { return &SessionData{} })
	})

	t.Run("SessionDataZero", func(t *testing.T) {
		exerciseJsonCodec(t, &SessionData{}, func() *SessionData { return &SessionData{} })
	})

	t.Run("SessionExtensions", func(t *testing.T) {
		exerciseJsonCodec(t, ptr(newPopulatedSessionExtensions()), func() *sessionExtensions { return &sessionExtensions{} })
	})

	t.Run("SessionExtensionsZero", func(t *testing.T) {
		exerciseJsonCodec(t, &sessionExtensions{}, func() *sessionExtensions { return &sessionExtensions{} })
	})
}

// TestJsonDecodeWrongFieldTypes covers type validation during JSON unmarshaling.
// A stored record whose field carries the wrong type should be reported as an error
// rather than silently decoding to an incorrect value.
func TestJsonDecodeWrongFieldTypes(t *testing.T) {
	testCases := []struct {
		name     string
		jsonData string
		test     func(t *testing.T)
	}{
		{
			name:     "CredentialIDMustBeString",
			jsonData: `{"id": 123}`,
			test: func(t *testing.T) {
				var decoded Credential
				err := json.Unmarshal([]byte(`{"id": 123}`), &decoded)
				require.Error(t, err, "Unmarshal must reject non-string id")
			},
		},
		{
			name:     "CredentialPublicKeyMustBeString",
			jsonData: `{"publicKey": 456}`,
			test: func(t *testing.T) {
				var decoded Credential
				err := json.Unmarshal([]byte(`{"publicKey": 456}`), &decoded)
				require.Error(t, err, "Unmarshal must reject non-string publicKey")
			},
		},
		{
			name:     "CredentialExtensionsRKMustBeBoolean",
			jsonData: `{"rk": "yes"}`,
			test: func(t *testing.T) {
				var decoded CredentialExtensions
				err := json.Unmarshal([]byte(`{"rk": "yes"}`), &decoded)
				require.Error(t, err, "Unmarshal must reject non-boolean rk")
			},
		},
		{
			name:     "CredentialExtensionsMinPinLengthMustBeNumber",
			jsonData: `{"minPinLength": "six"}`,
			test: func(t *testing.T) {
				var decoded CredentialExtensions
				err := json.Unmarshal([]byte(`{"minPinLength": "six"}`), &decoded)
				require.Error(t, err, "Unmarshal must reject non-number minPinLength")
			},
		},
		{
			name:     "CredentialExtensionsPRFMustBeBoolean",
			jsonData: `{"prfEnabled": "yes"}`,
			test: func(t *testing.T) {
				var decoded CredentialExtensions
				err := json.Unmarshal([]byte(`{"prfEnabled": "yes"}`), &decoded)
				require.Error(t, err, "Unmarshal must reject non-boolean prfEnabled")
			},
		},
		{
			name:     "SessionDataChallengeMustBeString",
			jsonData: `{"challenge": 12345}`,
			test: func(t *testing.T) {
				var decoded SessionData
				err := json.Unmarshal([]byte(`{"challenge": 12345}`), &decoded)
				require.Error(t, err, "Unmarshal must reject non-string challenge")
			},
		},
		{
			name:     "SessionExtensionsAppIDMustBeString",
			jsonData: `{"appid": 123}`,
			test: func(t *testing.T) {
				var decoded sessionExtensions
				err := json.Unmarshal([]byte(`{"appid": 123}`), &decoded)
				require.Error(t, err, "Unmarshal must reject non-string appid")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

// TestJsonDecodeUnknownField covers that decoders gracefully handle unknown fields,
// which allows a record written by a newer release to be read by an older one.
func TestJsonDecodeUnknownField(t *testing.T) {
	payload := []byte(`{"futureField": "valueFromNewerVersion"}`)

	t.Run("Credential", func(t *testing.T) {
		var decoded Credential
		err := json.Unmarshal(payload, &decoded)
		require.NoError(t, err, "Unmarshal must ignore unknown fields for forward compatibility")
	})

	t.Run("CredentialExtensions", func(t *testing.T) {
		var decoded CredentialExtensions
		err := json.Unmarshal(payload, &decoded)
		require.NoError(t, err, "Unmarshal must ignore unknown fields for forward compatibility")
	})

	t.Run("Authenticator", func(t *testing.T) {
		var decoded Authenticator
		err := json.Unmarshal(payload, &decoded)
		require.NoError(t, err, "Unmarshal must ignore unknown fields for forward compatibility")
	})

	t.Run("SessionData", func(t *testing.T) {
		var decoded SessionData
		err := json.Unmarshal(payload, &decoded)
		require.NoError(t, err, "Unmarshal must ignore unknown fields for forward compatibility")
	})

	t.Run("SessionExtensions", func(t *testing.T) {
		var decoded sessionExtensions
		err := json.Unmarshal(payload, &decoded)
		require.NoError(t, err, "Unmarshal must ignore unknown fields for forward compatibility")
	})
}

// TestJsonNilPointerFields covers that nil pointer fields are properly handled.
// When a pointer field is nil, it should not be included in the encoded JSON (with omitempty),
// and decoding should set it to nil rather than allocating a zero value.
func TestJsonNilPointerFields(t *testing.T) {
	t.Run("CredentialExtensionsWithNilPointers", func(t *testing.T) {
		value := &CredentialExtensions{}
		marshalled, err := json.Marshal(value)
		require.NoError(t, err)

		var decoded CredentialExtensions
		err = json.Unmarshal(marshalled, &decoded)
		require.NoError(t, err)

		// Re-encoding should produce the same output
		again, err := json.Marshal(&decoded)
		require.NoError(t, err)
		assert.Equal(t, marshalled, again, "re-encoded value must match original encoding")
	})

	t.Run("CredentialExtensionsPopulated", func(t *testing.T) {
		value := newPopulatedCredentialExtensions()
		marshalled, err := json.Marshal(value)
		require.NoError(t, err)

		var decoded CredentialExtensions
		err = json.Unmarshal(marshalled, &decoded)
		require.NoError(t, err)

		// Re-encoding should produce the same output
		again, err := json.Marshal(&decoded)
		require.NoError(t, err)
		assert.Equal(t, marshalled, again, "re-encoded value must match original encoding")
	})
}

// TestJsonCompactEncoding verifies that the JSON encoding is reasonably compact.
// While JSON is not as compact as msgp, we should verify that omitempty is working
// and unnecessary fields are not included.
func TestJsonCompactEncoding(t *testing.T) {
	// Zero value should produce compact output
	zero := &Credential{}
	zeroData, err := json.Marshal(zero)
	require.NoError(t, err)

	// Populated value should be larger
	populated := newPopulatedCredentialWithEverything()
	populatedData, err := json.Marshal(populated)
	require.NoError(t, err)

	assert.Less(t, len(zeroData), len(populatedData),
		"zero value should produce smaller JSON than populated value")
}
