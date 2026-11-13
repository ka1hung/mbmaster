package mbmaster

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// Real captures: id 43, fc 3, start 1, qty 50 — a valid 105-byte frame
// followed by one noise byte (FE/FF) from the RS-485 line.
var fieldCaptures = []string{
	"2B0364000851A1B4960000887DA73D000351AB7F61090C0103580757BD57610143014F034A984A978F97E304080048004A00BC014E002D0028007F00D40038003E008A0100000802570056007A00950FBE098C0FBE000002FE01F497E9000000000309034502DE9E7CFE",
	"2B0364000851A1B9960000887DA73D000351AB8385090C0103580757BD57610143014F034A984A978F97E304080048004A00BC014E002D0028007F00D40038003E008A0100000802570056007A00950FBE098C0FBE000002FE01F497E9000000000309034502DEED43FE",
	"2B0364000851A1BE960000887DA73D000351AB87A9090C0103580757BD57610143014F034A984A978F97E304080048004A00BC014E002D0028007F00D40038003E008A0100000802570056007A00950FBE098C0FBE000002FE01F497E9000000000309034502DEA119FF",
}

var req43 = []byte{0x2B, 0x03, 0x00, 0x00, 0x00, 0x32}

func TestExtractTrailingNoise(t *testing.T) {
	for _, s := range fieldCaptures {
		raw := mustHex(t, s)
		f, ok := ExtractFrame(raw, req43)
		if !ok {
			t.Fatalf("frame not found in % X", raw)
		}
		if len(f) != 105 || !bytes.Equal(f, raw[:105]) {
			t.Fatalf("got %d bytes, want first 105", len(f))
		}
	}
}

func TestExtractLeadingNoise(t *testing.T) {
	good := mustHex(t, fieldCaptures[0])[:105]
	raw := append([]byte{0xFF, 0x2B, 0x00}, good...) // includes a fake id byte
	f, ok := ExtractFrame(raw, req43)
	if !ok || !bytes.Equal(f, good) {
		t.Fatalf("leading noise not skipped: ok=%v", ok)
	}
}

func TestExtractIncomplete(t *testing.T) {
	good := mustHex(t, fieldCaptures[0])[:105]
	if _, ok := ExtractFrame(good[:104], req43); ok {
		t.Fatal("incomplete frame accepted")
	}
}

func TestExtractCorrupted(t *testing.T) {
	raw := mustHex(t, fieldCaptures[0])[:105]
	raw[50] ^= 0x01
	if _, ok := ExtractFrame(raw, req43); ok {
		t.Fatal("corrupted frame accepted")
	}
}

func TestExtractWrongSlave(t *testing.T) {
	raw := mustHex(t, fieldCaptures[0])[:105]
	other := []byte{0x2C, 0x03, 0x00, 0x00, 0x00, 0x32}
	if _, ok := ExtractFrame(raw, other); ok {
		t.Fatal("frame from another slave accepted")
	}
}

func TestExtractWrongQty(t *testing.T) {
	raw := mustHex(t, fieldCaptures[0])[:105]
	req := []byte{0x2B, 0x03, 0x00, 0x00, 0x00, 0x31} // asked for 49
	if _, ok := ExtractFrame(raw, req); ok {
		t.Fatal("frame with wrong byte count accepted")
	}
}

func TestExtractExceptionWithNoise(t *testing.T) {
	exc := CrcAppend([]byte{0x2B, 0x83, 0x02})
	raw := append(append([]byte{0x00}, exc...), 0xFE)
	f, ok := ExtractFrame(raw, req43)
	if !ok || !bytes.Equal(f, exc) {
		t.Fatalf("exception frame not found: ok=%v f=% X", ok, f)
	}
	if err := checkException(f); err == nil {
		t.Fatal("exception not reported")
	}
}

func TestExtractWriteEcho(t *testing.T) {
	req := []byte{0x01, 0x06, 0x00, 0x10, 0x12, 0x34}
	echo := CrcAppend(append([]byte(nil), req...))
	f, ok := ExtractFrame(append(echo, 0xFF), req)
	if !ok || !bytes.Equal(f, echo) {
		t.Fatalf("fc6 echo not found: ok=%v", ok)
	}
	bad := CrcAppend([]byte{0x01, 0x06, 0x00, 0x10, 0x12, 0x35})
	if _, ok := ExtractFrame(bad, req); ok {
		t.Fatal("mismatched fc6 echo accepted")
	}
}

func TestExtractCoils(t *testing.T) {
	req := []byte{0x05, 0x01, 0x00, 0x00, 0x00, 0x0A} // 10 coils -> 2 bytes
	resp := CrcAppend([]byte{0x05, 0x01, 0x02, 0xAA, 0x03})
	f, ok := ExtractFrame(append(resp, 0xFE), req)
	if !ok || !bytes.Equal(f, resp) {
		t.Fatalf("coil frame not found: ok=%v", ok)
	}
}

func TestResponseLen(t *testing.T) {
	cases := []struct {
		req  []byte
		want int
	}{
		{[]byte{1, 3, 0, 0, 0x01, 0x00}, 5 + 512}, // qty 256: old precedence bug gave 0x0105
		{[]byte{1, 3, 0, 0, 0, 50}, 105},
		{[]byte{1, 1, 0, 0, 0, 8}, 6},
		{[]byte{1, 1, 0, 0, 0, 9}, 7},
		{[]byte{1, 16, 0, 0, 0, 2}, 8},
		{[]byte{1, 0x2B, 0, 0, 0, 0}, 0},
	}
	for _, c := range cases {
		if got := ResponseLen(c.req); got != c.want {
			t.Errorf("ResponseLen(% X) = %d, want %d", c.req, got, c.want)
		}
	}
}
