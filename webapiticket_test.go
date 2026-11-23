// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

package steamworks

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unsafe"
)

// ticketResponseSize is the size of Valve's GetTicketForWebApiResponse_t:
// a handle, a result and a length of 4 bytes each, then the 2560-byte
// ticket field.
const ticketResponseSize = 12 + 2560

// ticketResponse builds a buffer laid out exactly like Valve's own
// GetTicketForWebApiResponse_t. No live Steam client or native call is
// involved, so these tests run in any environment.
func ticketResponse(handle HAuthTicket, result EResult, length int32, ticket []byte) []byte {
	buf := make([]byte, ticketResponseSize)
	binary.LittleEndian.PutUint32(buf[0:], uint32(handle))
	binary.LittleEndian.PutUint32(buf[4:], uint32(result))
	binary.LittleEndian.PutUint32(buf[8:], uint32(length))
	copy(buf[12:], ticket)
	return buf
}

func TestDecodeTicketForWebAPIResponse(t *testing.T) {
	want := []byte{0x14, 0x00, 0x00, 0x00, 0xAB, 0xCD, 0x00, 0xEF}
	buf := ticketResponse(42, EResultOK, int32(len(want)), want)

	got := decodeTicketForWebAPIResponse(unsafe.Pointer(&buf[0]))

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
	buf := ticketResponse(7, EResultFail, 0, nil)

	got := decodeTicketForWebAPIResponse(unsafe.Pointer(&buf[0]))

	if got.Handle != 7 || got.Result == EResultOK || len(got.Ticket) != 0 {
		t.Errorf("got %+v, want a failed result with no ticket", got)
	}
}

// A length outside the field must never make the decoder read beyond it.
func TestDecodeTicketForWebAPIResponseBoundsTheLength(t *testing.T) {
	const fieldSize = 2560
	full := bytes.Repeat([]byte{0xAA}, fieldSize)
	for name, length := range map[string]int32{"too long": fieldSize + 1000, "negative": -5} {
		buf := ticketResponse(1, EResultOK, length, full)
		got := decodeTicketForWebAPIResponse(unsafe.Pointer(&buf[0]))
		if len(got.Ticket) > fieldSize {
			t.Errorf("%s: the ticket has %d bytes, more than the field holds", name, len(got.Ticket))
		}
		if length < 0 && len(got.Ticket) != 0 {
			t.Errorf("%s: the ticket has %d bytes, want none", name, len(got.Ticket))
		}
	}
}

// The struct this file reads through must have Valve's layout: every field
// is 4 bytes or a byte array, so it has no padding under any packing Valve
// uses (4 on Linux and macOS, 8 on Windows).
func TestTicketForWebAPIResponseLayout(t *testing.T) {
	if got := unsafe.Sizeof(GetTicketForWebApiResponse{}); got != ticketResponseSize {
		t.Errorf("unsafe.Sizeof(GetTicketForWebApiResponse{}) = %d, want %d", got, ticketResponseSize)
	}
	if got := unsafe.Offsetof(GetTicketForWebApiResponse{}.Ticket); got != 12 {
		t.Errorf("offset of Ticket = %d, want 12", got)
	}
	if GetTicketForWebApiResponseCallbackID != 168 {
		t.Errorf("GetTicketForWebApiResponseCallbackID = %d, want 168 (k_iSteamUserCallbacks 100 + 68)", GetTicketForWebApiResponseCallbackID)
	}
}

// message makes the CallbackMsg Steam would give for payload.
func message(callbackID int32, payload []byte) CallbackMsg {
	msg := CallbackMsg{CallbackID: callbackID, ParamSize: int32(len(payload))}
	if len(payload) > 0 {
		msg.Param = unsafe.Pointer(&payload[0])
	}
	return msg
}

func TestCallbacksTakeKeepsWhatThisBindingKnows(t *testing.T) {
	ticket := []byte{1, 2, 3}
	var found Callbacks

	found.take(message(GetTicketForWebApiResponseCallbackID, ticketResponse(9, EResultOK, 3, ticket)))
	found.take(message(gameRichPresenceJoinRequestedCallbackID, joinRequestPayload(5, "old.example:1/a")))
	found.take(message(gameRichPresenceJoinRequestedCallbackID, joinRequestPayload(6, "new.example:1/b")))
	found.take(message(GetTicketForWebApiResponseCallbackID, ticketResponse(10, EResultFail, 0, nil)))

	if len(found.WebAPITickets) != 2 || found.WebAPITickets[0].Handle != 9 || found.WebAPITickets[1].Handle != 10 {
		t.Fatalf("tickets = %+v, want the two tickets in Steam's order", found.WebAPITickets)
	}
	if !bytes.Equal(found.WebAPITickets[0].Ticket, ticket) {
		t.Errorf("first ticket = %x, want %x", found.WebAPITickets[0].Ticket, ticket)
	}
	if found.JoinRequested == nil || found.JoinRequested.Connect != "new.example:1/b" || found.JoinRequested.SteamIDFriend != 6 {
		t.Errorf("JoinRequested = %+v, want the most recent request", found.JoinRequested)
	}
}

func TestCallbacksTakeLeavesOutWhatItCannotRead(t *testing.T) {
	var found Callbacks

	found.take(message(1101, []byte{1, 2, 3, 4}))                                              // a callback this binding does not know
	found.take(CallbackMsg{CallbackID: GetTicketForWebApiResponseCallbackID, ParamSize: 2572}) // no payload
	found.take(message(GetTicketForWebApiResponseCallbackID, make([]byte, 100)))               // a ticket payload that is too small
	found.take(message(gameRichPresenceJoinRequestedCallbackID, make([]byte, 4)))              // too small for the friend's id

	if found.JoinRequested != nil || len(found.WebAPITickets) != 0 {
		t.Errorf("found = %+v, want nothing", found)
	}
}

// Issue #785: GameOverlayActivated_t says whether the overlay is open.
func TestCallbacks_OverlayActivated(t *testing.T) {
	var found Callbacks
	open := []byte{1, 1, 0, 0}
	found.take(CallbackMsg{CallbackID: gameOverlayActivatedCallbackID, Param: unsafe.Pointer(&open[0]), ParamSize: int32(len(open))})
	if found.OverlayActive == nil || !*found.OverlayActive {
		t.Fatalf("open: %v", found.OverlayActive)
	}
	closed := []byte{0}
	found.take(CallbackMsg{CallbackID: gameOverlayActivatedCallbackID, Param: unsafe.Pointer(&closed[0]), ParamSize: 1})
	if *found.OverlayActive {
		t.Error("closed")
	}
	var none Callbacks
	none.take(CallbackMsg{CallbackID: gameOverlayActivatedCallbackID, Param: unsafe.Pointer(&closed[0]), ParamSize: 0})
	if none.OverlayActive != nil {
		t.Error("no payload")
	}
}
