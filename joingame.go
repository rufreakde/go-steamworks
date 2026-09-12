// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

package steamworks

import "unsafe"

// This file adds real support for GameRichPresenceJoinRequested_t — the
// callback Valve fires when a friend accepts a Join Game invite while
// this process is already running (ISteamFriends' own doc). Upstream
// go-steamworks has no callback-registration mechanism of any kind
// (only the package-level RunCallbacks, which pumps Valve's internal
// dispatch queue but hands nothing back to Go code) — see
// modwars-workspace issue #178 and its own research doc
// (docs/research/2026-09-12-steam-rich-presence-join-game.md) for the
// full investigation this fork resolves.
//
// The real, documented, non-C++ delivery path this file uses is
// Valve's own ManualDispatch flat API (SteamAPI_ManualDispatch_*,
// steam_api_flat.h) — NOT SteamAPI_RegisterCallback, which needs a real
// CCallbackBase-derived C++ vtable object, a materially harder and
// riskier shape to fake correctly from Go than a plain function
// pointer. Valve built ManualDispatch specifically as a callback-free
// polling alternative for bindings like this one that cannot construct
// a C++ vtable — confirmed against a real, shipped, non-Valve C++
// consumer of exactly this same alternative (aseprite/aseprite's own
// src/steam/steam.cpp) and cross-referenced against Steamworks.NET's
// mirrored steam_api_flat.h. Every symbol this file calls
// (SteamAPI_GetHSteamPipe, SteamAPI_ManualDispatch_Init/RunFrame/
// GetNextCallback/FreeLastCallback) was confirmed present via `nm`
// against this package's own vendored libsteam_api.dylib/so — not
// assumed from documentation alone.

// callbackMsg mirrors Valve's own CallbackMsg_t (steam_api_flat.h)
// field-for-field: a plain C struct, not a C++ vtable object. Field
// order/types must match the native struct exactly, since this is read
// directly out of Steam's own native memory via unsafe.Pointer:
//
//	struct CallbackMsg_t {
//	    HSteamUser m_hSteamUser;
//	    int        m_iCallback;
//	    uint8_t   *m_pubParam;
//	    int        m_cubParam;
//	};
//
// On a 64-bit target this Go struct's natural layout (two int32s, then
// a naturally-aligned uintptr, then a trailing int32 padded out to the
// struct's own 8-byte alignment) already matches the C struct's real
// size and field offsets with no explicit padding field needed.
type callbackMsg struct {
	steamUser HSteamUser
	callback  int32
	pubParam  uintptr
	cubParam  int32
}

// k_iCallbackGameRichPresenceJoinRequested is
// GameRichPresenceJoinRequested_t's own k_iCallback
// (k_iSteamFriendsCallbacks (300) + 37 = 337) — the value
// callbackMsg.callback must equal before pubParam may be reinterpreted
// as a GameRichPresenceJoinRequested_t. Cross-referenced against
// SteamRE/open-steamworks' community-reconstructed FriendsCommon.h and
// Facepunch.Steamworks' generated CallbackType enum, which independently
// agree on 337.
const k_iCallbackGameRichPresenceJoinRequested = 337

// maxRichPresenceValueLength mirrors Valve's own
// k_cchMaxRichPresenceValueLength — the fixed size of
// GameRichPresenceJoinRequested_t.m_rgchConnect.
const maxRichPresenceValueLength = 256

// GameRichPresenceJoinRequested is the connect string + inviting friend
// carried by a real GameRichPresenceJoinRequested_t callback. A zero
// SteamIDFriend matches Valve's own documented "invalid if not directly
// via a friend" case.
type GameRichPresenceJoinRequested struct {
	SteamIDFriend CSteamID
	Connect       string
}

// EnableManualDispatch switches this process from go-steamworks's
// default RunCallbacks-based dispatch to Valve's own ManualDispatch
// polling API and returns the HSteamPipe every subsequent
// PollGameRichPresenceJoinRequested call needs. Call once, after a
// successful Init.
//
// Additive, not a replacement: existing callers who never need
// GameRichPresenceJoinRequested_t can keep calling package-level
// RunCallbacks exactly as before. A process should pick exactly one of
// RunCallbacks or ManualDispatch, never both — mirrors Valve's own doc,
// which describes ManualDispatch as a full replacement for RunCallbacks'
// internal dispatch, not an addition to it.
func EnableManualDispatch() HSteamPipe {
	ptrAPI_ManualDispatch_Init()
	return ptrAPI_GetHSteamPipe()
}

// PollGameRichPresenceJoinRequested pumps ManualDispatch's own dispatch
// frame once, drains every callback currently queued, and reports the
// most recent GameRichPresenceJoinRequested_t found among them, if any.
// Every other pending callback is freed and discarded unread — this
// binding does not implement any other callback type, so there is
// nothing else yet to safely hand back. Call on the same cadence
// RunCallbacks would otherwise be called on (this project's own
// internal/client/steamworks.go throttles that pump to avoid a
// per-frame allocation source; the same reasoning applies here).
//
// Must only be called after EnableManualDispatch.
func PollGameRichPresenceJoinRequested(pipe HSteamPipe) (GameRichPresenceJoinRequested, bool) {
	ptrAPI_ManualDispatch_RunFrame(pipe)

	var found GameRichPresenceJoinRequested
	ok := false
	for {
		var msg callbackMsg
		if !ptrAPI_ManualDispatch_GetNextCallback(pipe, uintptr(unsafe.Pointer(&msg))) {
			break
		}
		if msg.callback == k_iCallbackGameRichPresenceJoinRequested && msg.pubParam != 0 {
			found = decodeGameRichPresenceJoinRequested(msg.pubParam)
			ok = true
		}
		ptrAPI_ManualDispatch_FreeLastCallback(pipe)
	}
	return found, ok
}

// decodeGameRichPresenceJoinRequested reinterprets a callbackMsg's
// pubParam as GameRichPresenceJoinRequested_t's own real layout:
//
//	struct GameRichPresenceJoinRequested_t {
//	    CSteamID m_steamIDFriend;
//	    char     m_rgchConnect[k_cchMaxRichPresenceValueLength];
//	};
func decodeGameRichPresenceJoinRequested(pubParam uintptr) GameRichPresenceJoinRequested {
	steamIDFriend := *(*CSteamID)(unsafe.Pointer(pubParam))
	connectOffset := pubParam + unsafe.Sizeof(CSteamID(0))
	connectBytes := unsafe.Slice((*byte)(unsafe.Pointer(connectOffset)), maxRichPresenceValueLength)
	return GameRichPresenceJoinRequested{
		SteamIDFriend: steamIDFriend,
		Connect:       cStringToGo(connectBytes),
	}
}
