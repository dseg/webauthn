package webauthn

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-webauthn/webauthn/protocol"
)

// TestJSONMarshalUnmarshalAuthenticator is the JSON analogue of
// TestMarshalUnmarshalAuthenticator in authenticator_gen_test.go.
func TestJSONMarshalUnmarshalAuthenticator(t *testing.T) {
	cases := []struct {
		name string
		v    Authenticator
	}{
		{"Zero", Authenticator{}},
		{"Populated", newPopulatedAuthenticator()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(&tc.v)
			require.NoError(t, err)

			var decoded Authenticator

			require.NoError(t, json.Unmarshal(data, &decoded))
			assert.Equal(t, tc.v, decoded)
		})
	}
}

// TestJSONEncodeDecodeAuthenticator is the JSON analogue of
// TestEncodeDecodeAuthenticator in authenticator_gen_test.go.
func TestJSONEncodeDecodeAuthenticator(t *testing.T) {
	cases := []struct {
		name string
		v    Authenticator
	}{
		{"Zero", Authenticator{}},
		{"Populated", newPopulatedAuthenticator()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			require.NoError(t, json.NewEncoder(&buf).Encode(&tc.v))

			var decoded Authenticator

			require.NoError(t, json.NewDecoder(&buf).Decode(&decoded))
			assert.Equal(t, tc.v, decoded)
		})
	}
}

func BenchmarkJSONMarshalAuthenticator(b *testing.B) {
	v := newPopulatedAuthenticator()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(&v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONUnmarshalAuthenticator(b *testing.B) {
	v := newPopulatedAuthenticator()

	data, err := json.Marshal(&v)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var d Authenticator

		if err := json.Unmarshal(data, &d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONEncodeAuthenticator(b *testing.B) {
	v := newPopulatedAuthenticator()

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

func BenchmarkJSONDecodeAuthenticator(b *testing.B) {
	v := newPopulatedAuthenticator()

	data, err := json.Marshal(&v)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var d Authenticator

		if err := json.NewDecoder(bytes.NewReader(data)).Decode(&d); err != nil {
			b.Fatal(err)
		}
	}
}

func newPopulatedAuthenticator() Authenticator {
	return Authenticator{
		AAGUID:       bytes.Repeat([]byte{0x11}, 16),
		SignCount:    42,
		CloneWarning: true,
		Attachment:   protocol.Platform,
	}
}
