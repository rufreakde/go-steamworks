// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

package steamworks

import (
	"testing"
	"unsafe"
)

// joinRequestPayload builds a buffer laid out exactly like Valve's own
// GameRichPresenceJoinRequested_t: a CSteamID, then a fixed char array. No
// live Steam client or native call is involved, so these tests run anywhere.
func joinRequestPayload(friend CSteamID, connect string) []byte {
	buf := make([]byte, unsafe.Sizeof(CSteamID(0))+maxRichPresenceValueLength)
	*(*CSteamID)(unsafe.Pointer(&buf[0])) = friend
	copy(buf[unsafe.Sizeof(CSteamID(0)):], connect)
	return buf
}

func TestDecodeGameRichPresenceJoinRequested(t *testing.T) {
	const wantConnect = "203.0.113.5:27015/lobby-42"
	const wantSteamID = CSteamID(76561197960287930)
	buf := joinRequestPayload(wantSteamID, wantConnect)

	got, ok := decodeGameRichPresenceJoinRequested(unsafe.Pointer(&buf[0]), len(buf))

	if !ok {
		t.Fatal("ok = false for a full payload")
	}
	if got.SteamIDFriend != wantSteamID {
		t.Errorf("SteamIDFriend = %v, want %v", got.SteamIDFriend, wantSteamID)
	}
	if got.Connect != wantConnect {
		t.Errorf("Connect = %q, want %q", got.Connect, wantConnect)
	}
}

// An all-zero connect field (a join with no connect string) decodes to an
// empty string, not to a string of null bytes.
func TestDecodeGameRichPresenceJoinRequestedEmptyConnect(t *testing.T) {
	buf := joinRequestPayload(0, "")

	got, ok := decodeGameRichPresenceJoinRequested(unsafe.Pointer(&buf[0]), len(buf))

	if !ok || got.Connect != "" || got.SteamIDFriend != 0 {
		t.Errorf("got %+v, ok %v; want an empty request, ok", got, ok)
	}
}

// The decoder never reads past the payload Steam gave, and never past the
// connect field.
func TestDecodeGameRichPresenceJoinRequestedStaysInsideThePayload(t *testing.T) {
	const friend = CSteamID(42)
	idSize := int(unsafe.Sizeof(CSteamID(0)))
	buf := joinRequestPayload(friend, "abcdef")

	short, ok := decodeGameRichPresenceJoinRequested(unsafe.Pointer(&buf[0]), idSize+3)
	if !ok || short.Connect != "abc" || short.SteamIDFriend != friend {
		t.Errorf("payload of id+3 bytes: got %+v, ok %v; want connect \"abc\"", short, ok)
	}

	long := append(buf, []byte("PAST THE FIELD")...)
	for i := idSize; i < idSize+maxRichPresenceValueLength; i++ {
		long[i] = 'x' // no terminator inside the field
	}
	got, ok := decodeGameRichPresenceJoinRequested(unsafe.Pointer(&long[0]), len(long))
	if !ok || len(got.Connect) != maxRichPresenceValueLength {
		t.Errorf("payload longer than the struct: connect has %d bytes, want %d", len(got.Connect), maxRichPresenceValueLength)
	}

	if _, ok := decodeGameRichPresenceJoinRequested(unsafe.Pointer(&buf[0]), idSize-1); ok {
		t.Error("ok = true for a payload smaller than the friend's id")
	}
	if _, ok := decodeGameRichPresenceJoinRequested(nil, 300); ok {
		t.Error("ok = true for a nil payload")
	}
}

// CallbackMsg is read straight out of Steam's memory. Valve's CallbackMsg_t
// is two 4-byte fields, a pointer and a 4-byte field: 24 bytes on a 64-bit
// target. A changed field would corrupt every read, so pin the size.
func TestCallbackMsgLayoutMatchesNativeSize(t *testing.T) {
	if got, want := unsafe.Sizeof(CallbackMsg{}), uintptr(24); got != want {
		t.Errorf("unsafe.Sizeof(CallbackMsg{}) = %d, want %d", got, want)
	}
	if got, want := unsafe.Offsetof(CallbackMsg{}.Param), uintptr(8); got != want {
		t.Errorf("offset of Param = %d, want %d", got, want)
	}
	if got, want := unsafe.Offsetof(CallbackMsg{}.ParamSize), uintptr(16); got != want {
		t.Errorf("offset of ParamSize = %d, want %d", got, want)
	}
}
