//go:build goexperiment.jsonv2

package protocol

import (
	"encoding/base64"
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time checks: URLEncodedBase64 marshals via the v2-native MarshalJSONTo, and
// unmarshals via the v1 UnmarshalJSON that json/v2 delegates to (no UnmarshalJSONFrom).
var (
	_ jsonv2.MarshalerTo = URLEncodedBase64(nil)
	_ jsonv2.Unmarshaler = (*URLEncodedBase64)(nil)
)

func TestURLEncodedBase64_MarshalJSONTo(t *testing.T) {
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
			var buf strings.Builder
			enc := jsontext.NewEncoder(&buf)
			require.NoError(t, tc.have.MarshalJSONTo(enc))
			// jsontext.Encoder terminates each top-level values with \n, need to remove it manually
			assert.Equal(t, tc.expected, strings.TrimRight(buf.String(), "\n"))
		})
	}
}

func BenchmarkJSONv2URLEncodedBase64Marshal(b *testing.B) {
	in := newURLEncodedBase64Bench()

	out, err := jsonv2.Marshal(in)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(out)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := jsonv2.Marshal(in); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONv2URLEncodedBase64Unmarshal(b *testing.B) {
	in := newURLEncodedBase64Bench()

	buf, err := jsonv2.Marshal(in)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(buf)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var out urlEncodedBase64Bench

		if err := jsonv2.Unmarshal(buf, &out); err != nil {
			b.Fatal(err)
		}
	}
}

// TestURLEncodedBase64_UnmarshalJSONv2 verifies URLEncodedBase64 unmarshals through
// encoding/json/v2. There is no UnmarshalJSONFrom; json/v2 delegates to the v1
// UnmarshalJSON and wraps its errors (hence assert.ErrorContains, not EqualError).
func TestURLEncodedBase64_UnmarshalJSONv2(t *testing.T) {
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
			message: `"` + base64.RawURLEncoding.EncodeToString([]byte("test base64 data")) + `"`,
			expected: testData{
				StringData:  "test string",
				EncodedData: URLEncodedBase64("test base64 data"),
			},
		},
		{
			name:    "ShouldHandleNull",
			message: "null",
			expected: testData{
				StringData:  "test string",
				EncodedData: nil,
			},
		},
		{
			name:    "ShouldFailInvalidBase64",
			message: `"not valid base64!!!"`,
			err:     "illegal base64 data at input byte 3",
		},
	}

	for _, tc := range testCases {
		raw := fmt.Sprintf(`{"string_data": "test string", "encoded_data": %s}`, tc.message)
		actual := &testData{}

		dec := jsontext.NewDecoder(strings.NewReader(raw))
		err := jsonv2.UnmarshalDecode(dec, actual)

		if tc.err != "" {
			assert.ErrorContains(t, err, tc.err)
			continue
		}

		require.NoError(t, err)
		assert.Equal(t, tc.expected.EncodedData, actual.EncodedData)
		assert.Equal(t, tc.expected.StringData, actual.StringData)
	}
}
