// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021 The go-steamworks Authors

package steamworks

import (
	"fmt"
	"syscall"
)

func loadLib() (uintptr, error) {
	dllName := "steam_api64.dll"
	handle, err := syscall.LoadLibrary(dllName)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", dllName, err)
	}
	return uintptr(handle), nil
}
