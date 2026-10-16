// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021 The go-steamworks Authors

//go:build !windows && !darwin && !(linux && amd64)

package steamworks

import (
	"fmt"
	"runtime"
)

// loadLib reports that no Steam library is embedded for this platform. The
// Steamworks SDK ships libsteam_api only for darwin, linux/amd64 and windows
// (gen.go), so the package still compiles here and Init returns this error.
func loadLib() (uintptr, error) {
	return 0, fmt.Errorf("steamworks: no Steam library for %s/%s", runtime.GOOS, runtime.GOARCH)
}
