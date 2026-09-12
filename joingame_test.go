// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

package steamworks

import (
	"testing"
	"unsafe"
)

// TestDecodeGameRichPresenceJoinRequested exercises the pure decode
// logic against a hand-built buffer laid out exactly like Valve's own
// GameRichPresenceJoinRequested_t (CSteamID followed by a fixed-size
// char array) — no live Steam client or native call involved, so this
// runs in any environment.
func TestDecodeGameRichPresenceJoinRequested(t *testing.T) {
	const wantConnect = "203.0.113.5:27015/lobby-42"
	const wantSteamID = CSteamID(76561197960287930)

	buf := make([]byte, unsafe.Sizeof(CSteamID(0))+maxRichPresenceValueLength)
	*(*CSteamID)(unsafe.Pointer(&buf[0])) = wantSteamID
	copy(buf[unsafe.Sizeof(CSteamID(0)):], wantConnect)

	got := decodeGameRichPresenceJoinRequested(uintptr(unsafe.Pointer(&buf[0])))

	if got.SteamIDFriend != wantSteamID {
		t.Errorf("SteamIDFriend = %v, want %v", got.SteamIDFriend, wantSteamID)
	}
	if got.Connect != wantConnect {
		t.Errorf("Connect = %q, want %q", got.Connect, wantConnect)
	}
}

// TestDecodeGameRichPresenceJoinRequestedEmptyConnect confirms an
// all-zero connect field (the case where an app-launch-only join, no
// connect string set) decodes to an empty string rather than a
// null-byte-laden garbage string.
func TestDecodeGameRichPresenceJoinRequestedEmptyConnect(t *testing.T) {
	buf := make([]byte, unsafe.Sizeof(CSteamID(0))+maxRichPresenceValueLength)

	got := decodeGameRichPresenceJoinRequested(uintptr(unsafe.Pointer(&buf[0])))

	if got.Connect != "" {
		t.Errorf("Connect = %q, want empty string", got.Connect)
	}
	if got.SteamIDFriend != CSteamID(0) {
		t.Errorf("SteamIDFriend = %v, want 0", got.SteamIDFriend)
	}
}

// TestCallbackMsgLayoutMatchesNativeSize pins callbackMsg's own size to
// Valve's real, documented CallbackMsg_t layout (two 4-byte fields, a
// pointer, a trailing 4-byte field, padded to the struct's own 8-byte
// alignment) on any 64-bit target — 24 bytes. A future accidental field
// reorder/addition that changed this size would silently corrupt every
// read this file does directly against Steam's native memory, so this
// test exists to catch that class of regression immediately instead of
// via a hard-to-diagnose live crash.
func TestCallbackMsgLayoutMatchesNativeSize(t *testing.T) {
	if got, want := unsafe.Sizeof(callbackMsg{}), uintptr(24); got != want {
		t.Errorf("unsafe.Sizeof(callbackMsg{}) = %d, want %d (Valve's real CallbackMsg_t size on a 64-bit target)", got, want)
	}
}
