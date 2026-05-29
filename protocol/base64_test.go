package protocol

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLEncodedBase64_MarshalJSON(t *testing.T) {
	testCases := []struct {
		name     string
		have     URLEncodedBase64
		expected string
	}{
		{
			name:     "ShouldMarshalData",
			have:     URLEncodedBase64("test data"),
			expected: `"dGVzdCBkYXRh"`,
		},
		{
			name:     "ShouldMarshalNil",
			have:     nil,
			expected: `null`,
		},
		{
			name:     "ShouldMarshalEmpty",
			have:     URLEncodedBase64{},
			expected: `""`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.have.MarshalJSON()
			require.NoError(t, err)
			assert.Equal(t, tc.expected, string(data))
		})
	}
}

func TestURLEncodedBase64_UnmarshalJSON_Error(t *testing.T) {
	testCases := []struct {
		name string
		data string
		err  string
	}{
		{
			name: "ShouldFailInvalidBase64",
			data: `"not valid base64!!!"`,
			err:  "illegal base64 data at input byte 3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var e URLEncodedBase64

			assert.EqualError(t, e.UnmarshalJSON([]byte(tc.data)), tc.err)
		})
	}
}

// urlEncodedBase64Bench wraps URLEncodedBase64 fields of three representative sizes; under
// GOEXPERIMENT=jsonv2 these benchmarks exercise URLEncodedBase64.UnmarshalJSONFrom /
// MarshalJSONTo, under v1 they exercise the byte-slice UnmarshalJSON / MarshalJSON.
type urlEncodedBase64Bench struct {
	A URLEncodedBase64 `json:"a"`
	B URLEncodedBase64 `json:"b"`
	C URLEncodedBase64 `json:"c"`
}

func newURLEncodedBase64Bench() urlEncodedBase64Bench {
	return urlEncodedBase64Bench{
		A: make([]byte, 32),
		B: make([]byte, 64),
		C: make([]byte, 256),
	}
}

func BenchmarkURLEncodedBase64Marshal(b *testing.B) {
	in := newURLEncodedBase64Bench()

	out, err := json.Marshal(in)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(out)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(in); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkURLEncodedBase64Unmarshal(b *testing.B) {
	in := newURLEncodedBase64Bench()

	buf, err := json.Marshal(in)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(buf)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var out urlEncodedBase64Bench

		if err := json.Unmarshal(buf, &out); err != nil {
			b.Fatal(err)
		}
	}
}

func TestBase64UnmarshalJSON(t *testing.T) {
	type testData struct {
		StringData  string           `json:"string_data"`
		EncodedData URLEncodedBase64 `json:"encoded_data"`
	}

	testCases := []struct {
		name     string
		message  string
		expected testData
		err      string
	}{
		{
			name:    "ShouldHandleBase64Data",
			message: "\"" + base64.RawURLEncoding.EncodeToString([]byte("test base64 data")) + "\"",
			expected: testData{
				StringData:  "test string",
				EncodedData: URLEncodedBase64("test base64 data"),
			},
			err: "",
		},
		{
			name:    "ShouldHandleNull",
			message: "null",
			expected: testData{
				StringData:  "test string",
				EncodedData: nil,
			},
			err: "",
		},
	}

	for _, tc := range testCases {
		raw := fmt.Sprintf(`{"string_data": "test string", "encoded_data": %s}`, tc.message)
		actual := &testData{}

		if tc.err != "" {
			assert.EqualError(t, json.NewDecoder(strings.NewReader(raw)).Decode(actual), tc.err)
		} else {
			assert.NoError(t, json.NewDecoder(strings.NewReader(raw)).Decode(actual))
		}

		assert.Equal(t, tc.expected.EncodedData, actual.EncodedData)
		assert.Equal(t, tc.expected.StringData, actual.StringData)
	}
}
