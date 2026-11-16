// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

package steamworks

import (
	"strings"
	"testing"
)

// purego panics when the loaded library lacks a name, as a Steam library of
// another SDK version does. registerSafely must make that an error.
func TestRegisterSafelyTurnsAPanicIntoAnError(t *testing.T) {
	err := registerSafely(func() { panic("purego: couldn't find symbol SteamAPI_SteamApps_v009") })

	if err == nil {
		t.Fatal("err = nil for a registration that panicked")
	}
	if !strings.Contains(err.Error(), "SteamAPI_SteamApps_v009") {
		t.Errorf("error %q does not name the missing function", err)
	}
}

func TestRegisterSafelyPassesAGoodRegistration(t *testing.T) {
	ran := false
	if err := registerSafely(func() { ran = true }); err != nil || !ran {
		t.Errorf("err = %v, ran = %v; want nil, true", err, ran)
	}
}
