package protocol

import (
	"encoding/json"
	"testing"
)

func BenchmarkCollectedClientData_Unmarshal(b *testing.B) {
	data := []byte(`{"type":"webauthn.get","challenge":"E4PTcIH_HfX1pC6Sigk1SC9NAlgeztN0439vi8z_c9k","origin":"https://webauthn.io","crossOrigin":false}`)
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var ccd CollectedClientData
		if err := json.Unmarshal(data, &ccd); err != nil {
			b.Fatal(err)
		}
	}
}
