package mbmaster

import (
	"bytes"
	"errors"
)

// ErrCRC is returned (wrapped, with the raw bytes) when no valid response
// frame could be found in the bytes received from the slave.
var ErrCRC = errors.New("CRC Error")

// ResponseLen returns the length (including CRC) of a normal RTU response to
// req, where req is the request frame without CRC. It returns 0 when the
// function code is not one the length can be derived for.
func ResponseLen(req []byte) int {
	if len(req) < 6 {
		return 0
	}
	qty := int(req[4])<<8 | int(req[5])
	switch req[1] {
	case 0x01, 0x02:
		return 5 + (qty+7)/8 // id + fc + bc + data + crc(2)
	case 0x03, 0x04:
		return 5 + qty*2
	case 0x05, 0x06, 0x0F, 0x10:
		return 8 // id + fc + addr(2) + value/qty(2) + crc(2)
	}
	return 0
}

// ExtractFrame searches buf for a valid RTU response to req (request without
// CRC) and returns a copy of it, CRC included. Noise bytes before and after
// the frame (e.g. an RS-485 driver glitch when the transceiver releases the
// bus) are ignored.
//
// A normal response must match the slave id, function code, the expected
// byte count / echoed fields, and the CRC; an exception response must match
// id, fc|0x80 and the CRC. Requiring all of these makes it practically
// impossible for line noise to be mistaken for a frame.
func ExtractFrame(buf, req []byte) ([]byte, bool) {
	if len(req) < 2 {
		return nil, false
	}
	id, fc := req[0], req[1]
	want := ResponseLen(req)

	for i := 0; i+5 <= len(buf); i++ {
		if buf[i] != id {
			continue
		}
		switch buf[i+1] {
		case fc | 0x80:
			if f := buf[i : i+5]; CrcCheck(f) {
				return append([]byte(nil), f...), true
			}
		case fc:
			n := want
			if n == 0 { // unknown fc: assume id + fc + bc + data + crc
				n = 5 + int(buf[i+2])
			}
			if i+n > len(buf) {
				continue
			}
			if f := buf[i : i+n]; bodyMatches(f, req, n) && CrcCheck(f) {
				return append([]byte(nil), f...), true
			}
		}
	}
	return nil, false
}

// bodyMatches checks the header fields of a normal response against req.
func bodyMatches(f, req []byte, n int) bool {
	switch req[1] {
	case 0x01, 0x02, 0x03, 0x04:
		return int(f[2]) == n-5
	case 0x05, 0x06, 0x0F, 0x10:
		return len(req) >= 6 && bytes.Equal(f[2:6], req[2:6])
	}
	return true
}
