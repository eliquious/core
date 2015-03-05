package core

import (
	"testing"
)

func BenchmarkUUID4(b *testing.B) {

	for i := 0; i < b.N; i++ {
		UUID4()
	}
}
