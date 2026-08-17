package protocol

import (
	"bytes"
	"testing"
)

func BenchmarkDecodeBody_AssertionResponse(b *testing.B) {
	data := []byte(testAssertionResponses["success"])
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var car CredentialAssertionResponse
		if err := decodeBody(bytes.NewReader(data), &car); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeBytes_AssertionResponse(b *testing.B) {
	data := []byte(testAssertionResponses["success"])
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var car CredentialAssertionResponse
		if err := decodeBytes(data, &car); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeBody_CreationResponse(b *testing.B) {
	data := []byte(testCredentialRequestResponses["success"])
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var ccr CredentialCreationResponse
		if err := decodeBody(bytes.NewReader(data), &ccr); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeBytes_CreationResponse(b *testing.B) {
	data := []byte(testCredentialRequestResponses["success"])
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var ccr CredentialCreationResponse
		if err := decodeBytes(data, &ccr); err != nil {
			b.Fatal(err)
		}
	}
}
