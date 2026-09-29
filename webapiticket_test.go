// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

package steamworks

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unsafe"
)

// ticketResponse builds a buffer laid out exactly like Valve's own
// GetTicketForWebApiResponse_t: two 4-byte handles/results, a 4-byte
// length, then the 2560-byte ticket field. No live Steam client or native
// call is involved, so these tests run in any environment.
func ticketResponse(handle HAuthTicket, result EResult, length int32, ticket []byte) []byte {
	buf := make([]byte, ticketForWebAPIResponseSize)
	binary.LittleEndian.PutUint32(buf[0:], uint32(handle))
	binary.LittleEndian.PutUint32(buf[4:], uint32(result))
	binary.LittleEndian.PutUint32(buf[8:], uint32(length))
	copy(buf[12:], ticket)
	return buf
}

func TestDecodeTicketForWebAPIResponse(t *testing.T) {
	want := []byte{0x14, 0x00, 0x00, 0x00, 0xAB, 0xCD, 0x00, 0xEF}
	buf := ticketResponse(42, EResultOK, int32(len(want)), want)

	got := decodeTicketForWebAPIResponse(uintptr(unsafe.Pointer(&buf[0])))

	if got.Handle != 42 || got.Result != EResultOK {
		t.Errorf("handle = %d, result = %d; want 42, EResultOK", got.Handle, got.Result)
	}
	if !bytes.Equal(got.Ticket, want) {
		t.Errorf("ticket = %x, want %x", got.Ticket, want)
	}
	// The ticket is a copy: Steam frees its own memory after the poll.
	buf[12] = 0xFF
	if got.Ticket[0] != 0x14 {
		t.Error("the decoded ticket still points into Steam's memory")
	}
}

// A failed request carries a result that is not OK and no ticket.
func TestDecodeTicketForWebAPIResponseFailed(t *testing.T) {
	buf := ticketResponse(7, 2, 0, nil) // 2: k_EResultFail

	got := decodeTicketForWebAPIResponse(uintptr(unsafe.Pointer(&buf[0])))

	if got.Handle != 7 || got.Result == EResultOK || len(got.Ticket) != 0 {
		t.Errorf("got %+v, want a failed result with no ticket", got)
	}
}

// A length outside the field must never make the decoder read beyond it.
func TestDecodeTicketForWebAPIResponseBoundsTheLength(t *testing.T) {
	full := bytes.Repeat([]byte{0xAA}, maxTicketForWebAPILength)
	for name, length := range map[string]int32{"too long": maxTicketForWebAPILength + 1000, "negative": -5} {
		buf := ticketResponse(1, EResultOK, length, full)
		got := decodeTicketForWebAPIResponse(uintptr(unsafe.Pointer(&buf[0])))
		if len(got.Ticket) > maxTicketForWebAPILength {
			t.Errorf("%s: the ticket has %d bytes, more than the field holds", name, len(got.Ticket))
		}
		if length < 0 && len(got.Ticket) != 0 {
			t.Errorf("%s: the ticket has %d bytes, want none", name, len(got.Ticket))
		}
	}
}

// TestTicketForWebAPIResponseLayout pins the offsets this file reads at to
// Valve's real GetTicketForWebApiResponse_t: every field is 4 bytes or a
// byte array, so the struct has no padding under any packing Valve uses
// (4 on Linux and macOS, 8 on Windows).
func TestTicketForWebAPIResponseLayout(t *testing.T) {
	if ticketForWebAPIResponseSize != 12+2560 {
		t.Errorf("ticketForWebAPIResponseSize = %d, want 2572", ticketForWebAPIResponseSize)
	}
	if k_iCallbackGetTicketForWebApiResponse != 168 {
		t.Errorf("k_iCallbackGetTicketForWebApiResponse = %d, want 168 (k_iSteamUserCallbacks 100 + 68)", k_iCallbackGetTicketForWebApiResponse)
	}
}
