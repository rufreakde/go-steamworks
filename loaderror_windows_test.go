// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

package steamworks

import (
	"strings"
	"testing"
)

// No steam_api64.dll ships with this repository, so on a Windows machine
// without it the package must still load (this test binary starting at all
// proves init no longer panics) and Init must report the missing library
// as an ordinary error the caller can fall back from.
func TestInit_MissingSteamAPIDLLIsAnErrorNotAPanic(t *testing.T) {
	if loadErr == nil {
		t.Skip("steam_api64.dll is present on this machine; nothing to check")
	}
	err := Init()
	if err == nil {
		t.Fatal("Init() = nil with steam_api64.dll missing, want an error")
	}
	if !strings.Contains(err.Error(), "steam_api64.dll") {
		t.Errorf("Init() error %q does not name the missing library", err)
	}
}

func TestRestartAppIfNecessary_MissingSteamAPIDLLMeansNoRestart(t *testing.T) {
	if loadErr == nil {
		t.Skip("steam_api64.dll is present on this machine; nothing to check")
	}
	if RestartAppIfNecessary(480) {
		t.Error("RestartAppIfNecessary = true with no Steam library, want false")
	}
}
