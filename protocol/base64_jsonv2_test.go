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

// Compile-time check: URLEncodedBase64 implements the v2 interfaces.
var (
	_ jsonv2.MarshalerTo     = URLEncodedBase64(nil)
	_ jsonv2.UnmarshalerFrom = (*URLEncodedBase64)(nil)
)

func TestURLEncodedBase64_MarshalJSONTo(t *testing.T) {
	testCases := []struct {
		name     string
		have     URLEncodedBase64
		expected string
	}{
		{"ShouldMarshalData", URLEncodedBase64("test data"), `"dGVzdCBkYXRh"`},
		{"ShouldMarshalNil", nil, `null`},
		{"ShouldMarshalEmpty", URLEncodedBase64{}, `""`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := jsonv2.Marshal(tc.have)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, strings.TrimSpace(string(data)))
		})
	}
}

// TestURLEncodedBase64_MarshalJSONTo_Direct exercises the method directly with a
// jsontext.Encoder, independent of the json/v2 reflection path.
func TestURLEncodedBase64_MarshalJSONTo_Direct(t *testing.T) {
	var buf strings.Builder

	enc := jsontext.NewEncoder(&buf)
	require.NoError(t, URLEncodedBase64("hello world").MarshalJSONTo(enc))
	assert.Equal(t, `"aGVsbG8gd29ybGQ"`, strings.TrimSpace(buf.String()))
}

func TestURLEncodedBase64_UnmarshalJSONFrom(t *testing.T) {
	type testData struct {
		StringData  string           `json:"string_data"`
		EncodedData URLEncodedBase64 `json:"encoded_data"`
	}

	encoded := base64.RawURLEncoding.EncodeToString([]byte("test base64 data"))

	testCases := []struct {
		name     string
		message  string
		expected testData
		err      string
	}{
		{
			name:    "ShouldHandleBase64Data",
			message: `"` + encoded + `"`,
			expected: testData{
				StringData:  "test string",
				EncodedData: URLEncodedBase64("test base64 data"),
			},
		},
		{
			name:    "ShouldTolerateTrailingPadding",
			message: `"` + encoded + `==="`,
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
			name:    "ShouldHandleEmptyString",
			message: `""`,
			expected: testData{
				StringData:  "test string",
				EncodedData: URLEncodedBase64{},
			},
		},
		{
			name:    "ShouldFailInvalidBase64",
			message: `"not valid base64!!!"`,
			err:     "illegal base64 data at input byte 3",
		},
		{
			name:    "ShouldFailNonStringToken",
			message: "42",
			err:     "URLEncodedBase64: expected JSON string, got number",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			raw := fmt.Sprintf(`{"string_data": "test string", "encoded_data": %s}`, tc.message)
			actual := &testData{}

			err := jsonv2.UnmarshalRead(strings.NewReader(raw), actual)
			if tc.err != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected.EncodedData, actual.EncodedData)
			assert.Equal(t, tc.expected.StringData, actual.StringData)
		})
	}
}

// TestURLEncodedBase64_UnmarshalJSONFrom_Direct exercises the method directly with a
// jsontext.Decoder, independent of the json/v2 reflection path.
func TestURLEncodedBase64_UnmarshalJSONFrom_Direct(t *testing.T) {
	encoded := base64.RawURLEncoding.EncodeToString([]byte("hello world"))
	dec := jsontext.NewDecoder(strings.NewReader(`"` + encoded + `"`))

	var e URLEncodedBase64

	require.NoError(t, e.UnmarshalJSONFrom(dec))
	assert.Equal(t, URLEncodedBase64("hello world"), e)
}
