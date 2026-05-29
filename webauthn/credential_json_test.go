package webauthn

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJSONMarshalUnmarshalCredential is the JSON analogue of
// TestMarshalUnmarshalCredential in credential_gen_test.go.
func TestJSONMarshalUnmarshalCredential(t *testing.T) {
	cases := []struct {
		name string
		v    Credential
	}{
		{"Zero", Credential{}},
		{"Populated", newPopulatedCredential()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(&tc.v)
			require.NoError(t, err)

			var decoded Credential

			require.NoError(t, json.Unmarshal(data, &decoded))
			assertJSONRoundTripEqual(t, &tc.v, &decoded)
		})
	}
}

// TestJSONEncodeDecodeCredential is the JSON analogue of
// TestEncodeDecodeCredential in credential_gen_test.go.
func TestJSONEncodeDecodeCredential(t *testing.T) {
	cases := []struct {
		name string
		v    Credential
	}{
		{"Zero", Credential{}},
		{"Populated", newPopulatedCredential()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			require.NoError(t, json.NewEncoder(&buf).Encode(&tc.v))

			var decoded Credential

			require.NoError(t, json.NewDecoder(&buf).Decode(&decoded))
			assertJSONRoundTripEqual(t, &tc.v, &decoded)
		})
	}
}

// TestJSONMarshalUnmarshalCredentialAttestation is the JSON analogue of
// TestMarshalUnmarshalCredentialAttestation in credential_gen_test.go.
func TestJSONMarshalUnmarshalCredentialAttestation(t *testing.T) {
	cases := []struct {
		name string
		v    CredentialAttestation
	}{
		{"Zero", CredentialAttestation{}},
		{"Populated", newPopulatedCredentialAttestation()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(&tc.v)
			require.NoError(t, err)

			var decoded CredentialAttestation

			require.NoError(t, json.Unmarshal(data, &decoded))
			assert.Equal(t, tc.v, decoded)
		})
	}
}

// TestJSONEncodeDecodeCredentialAttestation is the JSON analogue of
// TestEncodeDecodeCredentialAttestation in credential_gen_test.go.
func TestJSONEncodeDecodeCredentialAttestation(t *testing.T) {
	cases := []struct {
		name string
		v    CredentialAttestation
	}{
		{"Zero", CredentialAttestation{}},
		{"Populated", newPopulatedCredentialAttestation()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			require.NoError(t, json.NewEncoder(&buf).Encode(&tc.v))

			var decoded CredentialAttestation

			require.NoError(t, json.NewDecoder(&buf).Decode(&decoded))
			assert.Equal(t, tc.v, decoded)
		})
	}
}

// TestJSONMarshalUnmarshalCredentials is the JSON analogue of
// TestMarshalUnmarshalCredentials in credential_gen_test.go.
func TestJSONMarshalUnmarshalCredentials(t *testing.T) {
	cases := []struct {
		name string
		v    Credentials
	}{
		{"Empty", Credentials{}},
		{"Single", Credentials{newPopulatedCredential()}},
		{"Multiple", Credentials{newPopulatedCredential(), newPopulatedCredential(), newPopulatedCredential()}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.v)
			require.NoError(t, err)

			var decoded Credentials

			require.NoError(t, json.Unmarshal(data, &decoded))
			assertJSONRoundTripEqual(t, tc.v, decoded)
		})
	}
}

// TestJSONEncodeDecodeCredentials is the JSON analogue of
// TestEncodeDecodeCredentials in credential_gen_test.go.
func TestJSONEncodeDecodeCredentials(t *testing.T) {
	cases := []struct {
		name string
		v    Credentials
	}{
		{"Empty", Credentials{}},
		{"Single", Credentials{newPopulatedCredential()}},
		{"Multiple", Credentials{newPopulatedCredential(), newPopulatedCredential(), newPopulatedCredential()}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			require.NoError(t, json.NewEncoder(&buf).Encode(tc.v))

			var decoded Credentials

			require.NoError(t, json.NewDecoder(&buf).Decode(&decoded))
			assertJSONRoundTripEqual(t, tc.v, decoded)
		})
	}
}

func BenchmarkJSONMarshalCredential(b *testing.B) {
	v := newPopulatedCredential()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(&v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONUnmarshalCredential(b *testing.B) {
	v := newPopulatedCredential()

	data, err := json.Marshal(&v)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var d Credential

		if err := json.Unmarshal(data, &d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONEncodeCredential(b *testing.B) {
	v := newPopulatedCredential()

	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()

		if err := enc.Encode(&v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONDecodeCredential(b *testing.B) {
	v := newPopulatedCredential()

	data, err := json.Marshal(&v)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var d Credential

		if err := json.NewDecoder(bytes.NewReader(data)).Decode(&d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONMarshalCredentialAttestation(b *testing.B) {
	v := newPopulatedCredentialAttestation()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(&v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONUnmarshalCredentialAttestation(b *testing.B) {
	v := newPopulatedCredentialAttestation()

	data, err := json.Marshal(&v)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var d CredentialAttestation

		if err := json.Unmarshal(data, &d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONEncodeCredentialAttestation(b *testing.B) {
	v := newPopulatedCredentialAttestation()

	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()

		if err := enc.Encode(&v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONDecodeCredentialAttestation(b *testing.B) {
	v := newPopulatedCredentialAttestation()

	data, err := json.Marshal(&v)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var d CredentialAttestation

		if err := json.NewDecoder(bytes.NewReader(data)).Decode(&d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONMarshalCredentials(b *testing.B) {
	v := Credentials{newPopulatedCredential(), newPopulatedCredential()}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONUnmarshalCredentials(b *testing.B) {
	v := Credentials{newPopulatedCredential(), newPopulatedCredential()}

	data, err := json.Marshal(v)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var d Credentials

		if err := json.Unmarshal(data, &d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONEncodeCredentials(b *testing.B) {
	v := Credentials{newPopulatedCredential(), newPopulatedCredential()}

	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()

		if err := enc.Encode(v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONDecodeCredentials(b *testing.B) {
	v := Credentials{newPopulatedCredential(), newPopulatedCredential()}

	data, err := json.Marshal(v)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var d Credentials

		if err := json.NewDecoder(bytes.NewReader(data)).Decode(&d); err != nil {
			b.Fatal(err)
		}
	}
}

func newPopulatedCredentialAttestation() CredentialAttestation {
	return CredentialAttestation{
		ClientDataJSON:     []byte(`{"type":"webauthn.create","challenge":"abc","origin":"https://example.com"}`),
		ClientDataHash:     bytes.Repeat([]byte{0xde}, 32),
		AuthenticatorData:  bytes.Repeat([]byte{0xaa}, 64),
		PublicKeyAlgorithm: -7,
		Object:             bytes.Repeat([]byte{0xbb}, 128),
	}
}

// assertJSONRoundTripEqual asserts that two values produce equivalent JSON. It is preferred over
// assert.Equal for types whose Go representation does not round-trip through JSON byte-for-byte:
// CredentialFlags carries an unexported raw byte that is dropped on UnmarshalJSON, and
// SessionData.Extensions is map[string]any, so numeric values come back as float64 rather than the
// int64 the original held.
func assertJSONRoundTripEqual(t *testing.T, original, decoded any) {
	t.Helper()

	expected, err := json.Marshal(original)
	require.NoError(t, err)

	actual, err := json.Marshal(decoded)
	require.NoError(t, err)

	assert.JSONEq(t, string(expected), string(actual))
}
