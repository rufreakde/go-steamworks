// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

//go:build !windows

package steamworks

// callbackPack is Steam's packing of its callback structs on Linux, macOS
// and FreeBSD (VALVE_CALLBACK_PACK_SMALL).
const callbackPack = 4
