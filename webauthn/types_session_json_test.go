package webauthn

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJSONMarshalUnmarshalSessionData is the JSON analogue of
// TestMarshalUnmarshalSessionData in types_session_gen_test.go.
func TestJSONMarshalUnmarshalSessionData(t *testing.T) {
	cases := []struct {
		name string
		v    SessionData
	}{
		{"Zero", SessionData{}},
		{"Populated", newPopulatedSessionData()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(&tc.v)
			require.NoError(t, err)

			var decoded SessionData

			require.NoError(t, json.Unmarshal(data, &decoded))
			assertJSONRoundTripEqual(t, &tc.v, &decoded)

			// Time round-trips through RFC3339 and may lose the monotonic component, so compare with Equal.
			assert.True(t, tc.v.Expires.Equal(decoded.Expires))
		})
	}
}

// TestJSONEncodeDecodeSessionData is the JSON analogue of
// TestEncodeDecodeSessionData in types_session_gen_test.go.
func TestJSONEncodeDecodeSessionData(t *testing.T) {
	cases := []struct {
		name string
		v    SessionData
	}{
		{"Zero", SessionData{}},
		{"Populated", newPopulatedSessionData()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			require.NoError(t, json.NewEncoder(&buf).Encode(&tc.v))

			var decoded SessionData

			require.NoError(t, json.NewDecoder(&buf).Decode(&decoded))
			assertJSONRoundTripEqual(t, &tc.v, &decoded)
			assert.True(t, tc.v.Expires.Equal(decoded.Expires))
		})
	}
}

func BenchmarkJSONMarshalSessionData(b *testing.B) {
	v := newPopulatedSessionData()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(&v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONUnmarshalSessionData(b *testing.B) {
	v := newPopulatedSessionData()

	data, err := json.Marshal(&v)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var d SessionData

		if err := json.Unmarshal(data, &d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONEncodeSessionData(b *testing.B) {
	v := newPopulatedSessionData()

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

func BenchmarkJSONDecodeSessionData(b *testing.B) {
	v := newPopulatedSessionData()

	data, err := json.Marshal(&v)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var d SessionData

		if err := json.NewDecoder(bytes.NewReader(data)).Decode(&d); err != nil {
			b.Fatal(err)
		}
	}
}
