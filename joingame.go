// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

package steamworks

import "unsafe"

// This file adds GameRichPresenceJoinRequested_t: the callback Valve fires
// when a friend accepts a Join Game invite while this process is already
// running (ISteamFriends' own doc; modwars-workspace issue #178 and
// docs/research/2026-09-12-steam-rich-presence-join-game.md).
//
// It reads the callback through the ManualDispatch functions of this
// package (api.go), Valve's polling alternative for a binding that cannot
// build the C++ vtable SteamAPI_RegisterCallback needs.

// gameRichPresenceJoinRequestedCallbackID is the k_iCallback of
// GameRichPresenceJoinRequested_t (k_iSteamFriendsCallbacks (300) + 37).
// CallbackMsg.CallbackID must equal it before Param may be read as a
// GameRichPresenceJoinRequested_t.
const gameRichPresenceJoinRequestedCallbackID = 337

// gameOverlayActivatedCallbackID is the k_iCallback of GameOverlayActivated_t
// (k_iSteamFriendsCallbacks + 31): the Steam overlay opened or closed.
const gameOverlayActivatedCallbackID = 331

// maxRichPresenceValueLength mirrors Valve's own
// k_cchMaxRichPresenceValueLength: the fixed size of
// GameRichPresenceJoinRequested_t.m_rgchConnect.
const maxRichPresenceValueLength = 256

// GameRichPresenceJoinRequested is the connect string and the inviting
// friend of one GameRichPresenceJoinRequested_t callback. A zero
// SteamIDFriend is Valve's documented "invalid if not directly via a
// friend" case.
type GameRichPresenceJoinRequested struct {
	SteamIDFriend CSteamID
	Connect       string
}

// EnableManualDispatch switches this process from RunCallbacks to Valve's
// ManualDispatch polling and returns the HSteamPipe every PollCallbacks
// call needs. Call it once, after a successful Init. A process uses
// RunCallbacks or ManualDispatch, never both (Valve's own rule).
func EnableManualDispatch() HSteamPipe {
	ManualDispatch_Init()
	return GetHSteamPipe()
}

// decodeGameRichPresenceJoinRequested reads a callback's payload as
// GameRichPresenceJoinRequested_t and copies it out of Steam's memory:
//
//	struct GameRichPresenceJoinRequested_t {
//	    CSteamID m_steamIDFriend;
//	    char     m_rgchConnect[k_cchMaxRichPresenceValueLength];
//	};
//
// size is the payload's size as Steam gave it. The connect string is read
// only as far as the payload goes, and never past the field. ok is false
// for a payload too small to hold the friend's id.
func decodeGameRichPresenceJoinRequested(param unsafe.Pointer, size int) (request GameRichPresenceJoinRequested, ok bool) {
	idSize := int(unsafe.Sizeof(CSteamID(0)))
	if param == nil || size < idSize {
		return GameRichPresenceJoinRequested{}, false
	}
	connectLength := min(size-idSize, maxRichPresenceValueLength)
	connect := unsafe.Slice((*byte)(unsafe.Add(param, idSize)), connectLength)
	return GameRichPresenceJoinRequested{
		SteamIDFriend: *(*CSteamID)(param),
		Connect:       cStringToGo(connect),
	}, true
}
