package mailaddress

import "testing"

// BenchmarkParse-4          300000              5322 ns/op
func BenchmarkParse(b *testing.B) {
	in := "Martin <martin@example.com>"
	for n := 0; n < b.N; n++ {
		_, _ = Parse(in)
	}
}
