package protocol

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-webauthn/webauthn/protocol/webauthncose"
)

// TestJSONMarshalUnmarshalCredentialParameter is the JSON analogue of
// TestMarshalUnmarshalCredentialParameter in options_msgp_gen_test.go.
func TestJSONMarshalUnmarshalCredentialParameter(t *testing.T) {
	cases := []struct {
		name string
		v    CredentialParameter
	}{
		{"Zero", CredentialParameter{}},
		{"Populated", newPopulatedCredentialParameter()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.v)
			require.NoError(t, err)

			var decoded CredentialParameter

			require.NoError(t, json.Unmarshal(data, &decoded))
			assert.Equal(t, tc.v, decoded)
		})
	}
}

// TestJSONEncodeDecodeCredentialParameter is the JSON analogue of
// TestEncodeDecodeCredentialParameter in options_msgp_gen_test.go.
func TestJSONEncodeDecodeCredentialParameter(t *testing.T) {
	cases := []struct {
		name string
		v    CredentialParameter
	}{
		{"Zero", CredentialParameter{}},
		{"Populated", newPopulatedCredentialParameter()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			require.NoError(t, json.NewEncoder(&buf).Encode(tc.v))

			var decoded CredentialParameter

			require.NoError(t, json.NewDecoder(&buf).Decode(&decoded))
			assert.Equal(t, tc.v, decoded)
		})
	}
}

func BenchmarkJSONMarshalCredentialParameter(b *testing.B) {
	v := newPopulatedCredentialParameter()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONUnmarshalCredentialParameter(b *testing.B) {
	v := newPopulatedCredentialParameter()

	data, err := json.Marshal(v)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var d CredentialParameter

		if err := json.Unmarshal(data, &d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONEncodeCredentialParameter(b *testing.B) {
	v := newPopulatedCredentialParameter()

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

func BenchmarkJSONDecodeCredentialParameter(b *testing.B) {
	v := newPopulatedCredentialParameter()

	data, err := json.Marshal(v)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var d CredentialParameter

		if err := json.NewDecoder(bytes.NewReader(data)).Decode(&d); err != nil {
			b.Fatal(err)
		}
	}
}

func newPopulatedCredentialParameter() CredentialParameter {
	return CredentialParameter{
		Type:      PublicKeyCredentialType,
		Algorithm: webauthncose.AlgES256,
	}
}
