// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

//go:build darwin || (linux && amd64)

package steamworks

import "testing"

// The shipped Steam library is opened and every Workshop name is looked up
// in it, the way the first use of Steam does. Nothing is called: no Steam
// client is needed.
func TestRegisterUGC_FindsEveryNameInTheShippedLibrary(t *testing.T) {
	lib, err := loadLib()
	if err != nil {
		t.Skipf("the shipped library does not open here: %v", err)
	}
	saved := ptrAPI_ISteamUGC_CreateItem
	t.Cleanup(func() { ptrAPI_ISteamUGC_CreateItem = saved })
	if err := registerSafely(func() { registerUGC(lib) }); err != nil {
		t.Fatalf("registerUGC: %v", err)
	}
	if ptrAPI_ISteamUGC_CreateItem == nil || ptrAPI_ManualDispatch_GetAPICallResult == nil || ptrAPI_SteamUGC == nil {
		t.Error("a binding is still empty")
	}
}
