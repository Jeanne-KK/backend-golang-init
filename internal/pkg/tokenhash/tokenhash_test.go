package tokenhash

import "testing"

func TestSHA256HexDeterministic(t *testing.T) {
	const raw = "refresh-token-raw-value"
	got := SHA256Hex(raw)
	want := "2488f939baccd313e9256b45e0830986f5aa92114da97b6a034e4a1b45611d83"
	if got != want {
		t.Fatalf("SHA256Hex: got %q want %q", got, want)
	}
	if SHA256Hex(raw) != got {
		t.Fatal("expected same hash on second call")
	}
}
