// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

package steamworks

import (
	"encoding/binary"
	"math"
	"testing"
)

// The offsets and sizes below are worked out by hand from the SDK 1.65
// headers: a field sits at the next multiple of min(its own alignment, the
// pack), and the struct's size is a multiple of its largest such alignment.

func TestLayout_CreateItemResult(t *testing.T) {
	for pack, want := range map[int]struct {
		offsets []int
		size    int
	}{
		4: {[]int{0, 4, 12}, 16},
		8: {[]int{0, 8, 16}, 24},
	} {
		got := layoutOf(createItemResultFields, pack)
		if !equalInts(got.offsets, want.offsets) || got.size != want.size {
			t.Errorf("pack %d: %+v, want %+v", pack, got, want)
		}
	}
}

func TestLayout_SubmitItemUpdateResult(t *testing.T) {
	for _, pack := range []int{4, 8} {
		got := layoutOf(submitItemUpdateResultFields, pack)
		if !equalInts(got.offsets, []int{0, 4, 8}) || got.size != 16 {
			t.Errorf("pack %d: %+v", pack, got)
		}
	}
}

func TestLayout_SubscribeResult(t *testing.T) {
	for pack, want := range map[int]struct {
		offsets []int
		size    int
	}{
		4: {[]int{0, 4}, 12},
		8: {[]int{0, 8}, 16},
	} {
		got := layoutOf(subscribeResultFields, pack)
		if !equalInts(got.offsets, want.offsets) || got.size != want.size {
			t.Errorf("pack %d: %+v, want %+v", pack, got, want)
		}
	}
}

func TestLayout_QueryCompleted(t *testing.T) {
	for pack, want := range map[int]struct {
		offsets []int
		size    int
	}{
		4: {[]int{0, 8, 12, 16, 20, 21}, 280},
		8: {[]int{0, 8, 12, 16, 20, 21}, 280},
	} {
		got := layoutOf(queryCompletedFields, pack)
		if !equalInts(got.offsets, want.offsets) || got.size != want.size {
			t.Errorf("pack %d: %+v, want %+v", pack, got, want)
		}
	}
}

func TestLayout_EULAStatus(t *testing.T) {
	for _, pack := range []int{4, 8} {
		got := layoutOf(eulaStatusFields, pack)
		if !equalInts(got.offsets, []int{0, 4, 8, 12, 16, 17}) || got.size != 20 {
			t.Errorf("pack %d: %+v", pack, got)
		}
	}
}

func TestLayout_ItemInstalledAndDownloadResult(t *testing.T) {
	for pack, want := range map[int]struct {
		installed, download         []int
		installedSize, downloadSize int
	}{
		4: {[]int{0, 4, 12, 20}, []int{0, 4, 12}, 28, 16},
		8: {[]int{0, 8, 16, 24}, []int{0, 8, 16}, 32, 24},
	} {
		installed, download := layoutOf(itemInstalledFields, pack), layoutOf(downloadItemResultFields, pack)
		if !equalInts(installed.offsets, want.installed) || installed.size != want.installedSize {
			t.Errorf("ItemInstalled_t pack %d: %+v", pack, installed)
		}
		if !equalInts(download.offsets, want.download) || download.size != want.downloadSize {
			t.Errorf("DownloadItemResult_t pack %d: %+v", pack, download)
		}
	}
}

// SteamUGCDetails_t, the largest one. Its uint64 fields after the char
// arrays move by 4 bytes between the two packs.
func TestLayout_UGCDetails(t *testing.T) {
	small, large := layoutOf(ugcDetailsFields, 4), layoutOf(ugcDetailsFields, 8)
	// Up to the description, both packs agree.
	for _, i := range []int{0, 1, 2, 3, 4, 5, 6} {
		if small.offsets[i] != large.offsets[i] {
			t.Errorf("field %d: %d and %d", i, small.offsets[i], large.offsets[i])
		}
	}
	// m_rgchTitle at 24 (8 + 4*4), m_rgchDescription at 153, then the owner:
	// 8153 rounds up to 8156 with pack 4 and to 8160 with pack 8.
	if small.offsets[5] != 24 || small.offsets[6] != 153 || small.offsets[7] != 8156 || large.offsets[7] != 8160 {
		t.Errorf("offsets %v / %v", small.offsets[:8], large.offsets[:8])
	}
	if small.size != 9772 || large.size != 9784 {
		t.Errorf("sizes %d and %d", small.size, large.size)
	}
}

func TestLayout_PackOfThisSystem(t *testing.T) {
	if callbackPack != 4 && callbackPack != 8 {
		t.Fatalf("callbackPack = %d", callbackPack)
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// --- reading the fields out of a buffer

// put writes v at offset into b, little-endian, by its size.
func put(b []byte, offset int, v any) {
	switch v := v.(type) {
	case uint32:
		binary.LittleEndian.PutUint32(b[offset:], v)
	case int32:
		binary.LittleEndian.PutUint32(b[offset:], uint32(v))
	case uint64:
		binary.LittleEndian.PutUint64(b[offset:], v)
	case bool:
		if v {
			b[offset] = 1
		}
	case string:
		copy(b[offset:], v)
	case float32:
		binary.LittleEndian.PutUint32(b[offset:], math.Float32bits(v))
	}
}

func TestDecode_CallResults(t *testing.T) {
	for _, pack := range []int{4, 8} {
		l := layoutOf(createItemResultFields, pack)
		b := make([]byte, l.size)
		put(b, l.offsets[0], int32(EResultOK))
		put(b, l.offsets[1], uint64(3000000001))
		put(b, l.offsets[2], true)
		created, ok := decodeCreateItemResult(b, pack)
		if !ok || created != (CreateItemResult{Result: EResultOK, PublishedFileID: 3000000001, NeedsLegalAgreement: true}) {
			t.Errorf("pack %d: CreateItemResult %+v, %v", pack, created, ok)
		}

		l = layoutOf(submitItemUpdateResultFields, pack)
		b = make([]byte, l.size)
		put(b, l.offsets[0], int32(EResultTimeout))
		put(b, l.offsets[2], uint64(42))
		submitted, ok := decodeSubmitItemUpdateResult(b, pack)
		if !ok || submitted != (SubmitItemUpdateResult{Result: EResultTimeout, PublishedFileID: 42}) {
			t.Errorf("pack %d: SubmitItemUpdateResult %+v", pack, submitted)
		}

		l = layoutOf(subscribeResultFields, pack)
		b = make([]byte, l.size)
		put(b, l.offsets[0], int32(EResultOK))
		put(b, l.offsets[1], uint64(77))
		subscribed, ok := decodeSubscribeResult(b, pack)
		if !ok || subscribed != (SubscribeResult{Result: EResultOK, PublishedFileID: 77}) {
			t.Errorf("pack %d: SubscribeResult %+v", pack, subscribed)
		}

		l = layoutOf(queryCompletedFields, pack)
		b = make([]byte, l.size)
		put(b, l.offsets[0], uint64(9))
		put(b, l.offsets[1], int32(EResultOK))
		put(b, l.offsets[2], uint32(2))
		put(b, l.offsets[3], uint32(5))
		query, ok := decodeQueryCompleted(b, pack)
		if !ok || query != (QueryCompleted{Handle: 9, Result: EResultOK, Returned: 2, Total: 5}) {
			t.Errorf("pack %d: QueryCompleted %+v", pack, query)
		}

		l = layoutOf(eulaStatusFields, pack)
		b = make([]byte, l.size)
		put(b, l.offsets[0], int32(EResultOK))
		put(b, l.offsets[1], uint32(5534970))
		put(b, l.offsets[2], uint32(3))
		put(b, l.offsets[4], true)
		eula, ok := decodeEULAStatus(b, pack)
		if !ok || eula != (WorkshopEULAStatus{Result: EResultOK, AppID: 5534970, Version: 3, Accepted: true}) {
			t.Errorf("pack %d: WorkshopEULAStatus %+v", pack, eula)
		}

		l = layoutOf(itemInstalledFields, pack)
		b = make([]byte, l.size)
		put(b, l.offsets[0], uint32(5534970))
		put(b, l.offsets[1], uint64(77))
		installed, ok := decodeItemInstalled(b, pack)
		if !ok || installed != (ItemInstalled{AppID: 5534970, PublishedFileID: 77}) {
			t.Errorf("pack %d: ItemInstalled %+v", pack, installed)
		}

		l = layoutOf(downloadItemResultFields, pack)
		b = make([]byte, l.size)
		put(b, l.offsets[0], uint32(5534970))
		put(b, l.offsets[1], uint64(77))
		put(b, l.offsets[2], int32(EResultFail))
		download, ok := decodeDownloadItemResult(b, pack)
		if !ok || download != (DownloadItemResult{AppID: 5534970, PublishedFileID: 77, Result: EResultFail}) {
			t.Errorf("pack %d: DownloadItemResult %+v", pack, download)
		}
	}
}

// A buffer smaller than its struct is not read at all.
func TestDecode_AShortBufferIsRefused(t *testing.T) {
	short := make([]byte, 3)
	if _, ok := decodeCreateItemResult(short, 4); ok {
		t.Error("CreateItemResult")
	}
	if _, ok := decodeSubmitItemUpdateResult(short, 4); ok {
		t.Error("SubmitItemUpdateResult")
	}
	if _, ok := decodeSubscribeResult(short, 4); ok {
		t.Error("SubscribeResult")
	}
	if _, ok := decodeQueryCompleted(short, 4); ok {
		t.Error("QueryCompleted")
	}
	if _, ok := decodeEULAStatus(short, 4); ok {
		t.Error("WorkshopEULAStatus")
	}
	if _, ok := decodeItemInstalled(short, 4); ok {
		t.Error("ItemInstalled")
	}
	if _, ok := decodeDownloadItemResult(short, 4); ok {
		t.Error("DownloadItemResult")
	}
	if _, ok := decodeUGCDetails(short, 4); ok {
		t.Error("UGCDetails")
	}
}

func TestDecode_UGCDetails(t *testing.T) {
	for _, pack := range []int{4, 8} {
		l := layoutOf(ugcDetailsFields, pack)
		b := make([]byte, l.size)
		put(b, l.offsets[0], uint64(3000000001))
		put(b, l.offsets[1], int32(EResultOK))
		put(b, l.offsets[5], "Big Tanks")
		put(b, l.offsets[6], "More armour.")
		put(b, l.offsets[7], uint64(76561198000000001))
		put(b, l.offsets[9], uint32(1763280000))
		put(b, l.offsets[11], int32(ERemoteStoragePublishedFileVisibilityPublic))
		put(b, l.offsets[12], true)
		put(b, l.offsets[26], uint64(123456))
		details, ok := decodeUGCDetails(b, pack)
		want := UGCDetails{
			PublishedFileID: 3000000001, Result: EResultOK, Title: "Big Tanks", Description: "More armour.",
			Owner: 76561198000000001, TimeUpdated: 1763280000, Visibility: ERemoteStoragePublishedFileVisibilityPublic,
			Banned: true, TotalFilesSize: 123456,
		}
		if !ok || details != want {
			t.Errorf("pack %d: %+v", pack, details)
		}
	}
}

// A text field that fills its whole array with no end mark is read up to
// the array's end, never past it.
func TestCString_StaysInsideItsArray(t *testing.T) {
	b := []byte("abcdefgh")
	if got := cString(b, 2, 4); got != "cdef" {
		t.Errorf("cString = %q", got)
	}
	b[4] = 0
	if got := cString(b, 2, 4); got != "cd" {
		t.Errorf("cString = %q", got)
	}
}
