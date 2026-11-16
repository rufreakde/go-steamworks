// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

package steamworks

import "testing"

// ActivateGameOverlayInviteDialogConnectString hands Steam the ISteamFriends
// pointer it was made from and the connect string, unchanged: Steam puts that
// string into the invite, and a friend's game gets it back on a join. The
// native function is swapped for a recorder, so no Steam library is needed.
// TestTheMacOSLibraryHasEveryRegisteredName and its Linux twin check that the
// shipped libraries have the name.
func TestActivateGameOverlayInviteDialogConnectStringPassesTheConnectString(t *testing.T) {
	type call struct {
		self    uintptr
		connect string
	}
	var calls []call
	saved := ptrAPI_ISteamFriends_ActivateGameOverlayInviteDialogConnectString
	t.Cleanup(func() { ptrAPI_ISteamFriends_ActivateGameOverlayInviteDialogConnectString = saved })
	ptrAPI_ISteamFriends_ActivateGameOverlayInviteDialogConnectString = func(self uintptr, connect string) {
		calls = append(calls, call{self, connect})
	}

	const connect = "play.example.org:27016/lobby-42"
	steamFriends(0xBEEF).ActivateGameOverlayInviteDialogConnectString(connect)

	if len(calls) != 1 || calls[0] != (call{0xBEEF, connect}) {
		t.Errorf("calls = %+v, want one call with the ISteamFriends pointer and %q", calls, connect)
	}
}

// The binding asks the Steam library for Valve's own flat name of the call.
func TestActivateGameOverlayInviteDialogConnectStringHasValvesName(t *testing.T) {
	const want = "SteamAPI_ISteamFriends_ActivateGameOverlayInviteDialogConnectString"
	if flatAPI_ISteamFriends_ActivateGameOverlayInviteDialogConnectString != want {
		t.Errorf("flat name = %q, want %q", flatAPI_ISteamFriends_ActivateGameOverlayInviteDialogConnectString, want)
	}
}
