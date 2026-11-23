// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

package steamworks

import (
	"testing"
	"unsafe"
)

// The Workshop's calls are thin: each one hands Steam the ISteamUGC pointer
// and its arguments, and gives back what Steam answered. The native
// functions are swapped for fakes, so no Steam library is needed; the
// library name tests check that the shipped libraries have every name.

const ugc = steamUGC(0xBEEF)

// swap replaces *slot with fake until the test ends.
func swap[F any](t *testing.T, slot *F, fake F) {
	t.Helper()
	saved := *slot
	t.Cleanup(func() { *slot = saved })
	*slot = fake
}

func TestSteamUGC_TheAccessorNamesItsVersion(t *testing.T) {
	swap(t, &ptrAPI_SteamUGC, func() uintptr { return 0xBEEF })
	if SteamUGC() != ugc {
		t.Error("SteamUGC")
	}
	if flatAPI_SteamUGC != "SteamAPI_SteamUGC_v021" {
		t.Errorf("accessor %q", flatAPI_SteamUGC)
	}
}

func TestSteamUGC_MakingAndUpdatingAnItem(t *testing.T) {
	var seen []any
	note := func(values ...any) { seen = append(seen, values...) }
	swap(t, &ptrAPI_ISteamUGC_CreateItem, func(self uintptr, app AppId_t, kind EWorkshopFileType) SteamAPICall_t {
		note(self, app, kind)
		return 11
	})
	swap(t, &ptrAPI_ISteamUGC_StartItemUpdate, func(self uintptr, app AppId_t, item PublishedFileId_t) UGCUpdateHandle_t {
		note(app, item)
		return 22
	})
	text := func(self uintptr, h UGCUpdateHandle_t, s string) bool { note(h, s); return true }
	swap(t, &ptrAPI_ISteamUGC_SetItemTitle, text)
	swap(t, &ptrAPI_ISteamUGC_SetItemDescription, text)
	swap(t, &ptrAPI_ISteamUGC_SetItemContent, text)
	swap(t, &ptrAPI_ISteamUGC_SetItemPreview, text)
	swap(t, &ptrAPI_ISteamUGC_SetItemVisibility, func(self uintptr, h UGCUpdateHandle_t, v ERemoteStoragePublishedFileVisibility) bool {
		note(h, v)
		return true
	})
	swap(t, &ptrAPI_ISteamUGC_SubmitItemUpdate, func(self uintptr, h UGCUpdateHandle_t, s string) SteamAPICall_t {
		note(h, s)
		return 33
	})
	swap(t, &ptrAPI_ISteamUGC_GetItemUpdateProgress, func(self uintptr, h UGCUpdateHandle_t, processed, total *uint64) EItemUpdateStatus {
		*processed, *total = 5, 10
		return EItemUpdateStatusUploadingContent
	})

	if call := ugc.CreateItem(5534970, EWorkshopFileTypeCommunity); call != 11 {
		t.Errorf("CreateItem = %d", call)
	}
	update := ugc.StartItemUpdate(5534970, 1001)
	ok := ugc.SetItemTitle(update, "Big Tanks") && ugc.SetItemDescription(update, "More armour.") &&
		ugc.SetItemContent(update, "/mods/big-tanks") && ugc.SetItemPreview(update, "/mods/preview.png") &&
		ugc.SetItemVisibility(update, ERemoteStoragePublishedFileVisibilityPublic)
	if !ok {
		t.Error("a SetItem call said no")
	}
	if call := ugc.SubmitItemUpdate(update, "Version 2026-11-16"); call != 33 {
		t.Errorf("SubmitItemUpdate = %d", call)
	}
	if status, processed, total := ugc.GetItemUpdateProgress(update); status != EItemUpdateStatusUploadingContent || processed != 5 || total != 10 {
		t.Errorf("GetItemUpdateProgress = %d %d %d", status, processed, total)
	}
	want := []any{
		uintptr(0xBEEF), AppId_t(5534970), EWorkshopFileTypeCommunity,
		AppId_t(5534970), PublishedFileId_t(1001),
		UGCUpdateHandle_t(22), "Big Tanks", UGCUpdateHandle_t(22), "More armour.",
		UGCUpdateHandle_t(22), "/mods/big-tanks", UGCUpdateHandle_t(22), "/mods/preview.png",
		UGCUpdateHandle_t(22), ERemoteStoragePublishedFileVisibilityPublic,
		UGCUpdateHandle_t(22), "Version 2026-11-16",
	}
	if len(seen) != len(want) {
		t.Fatalf("seen %v", seen)
	}
	for i := range want {
		if seen[i] != want[i] {
			t.Errorf("argument %d: %v, want %v", i, seen[i], want[i])
		}
	}
}

func TestSteamUGC_SubscribingAndTheInstalledItems(t *testing.T) {
	swap(t, &ptrAPI_ISteamUGC_SubscribeItem, func(self uintptr, item PublishedFileId_t) SteamAPICall_t { return SteamAPICall_t(item) + 1 })
	swap(t, &ptrAPI_ISteamUGC_UnsubscribeItem, func(self uintptr, item PublishedFileId_t) SteamAPICall_t { return SteamAPICall_t(item) + 2 })
	swap(t, &ptrAPI_ISteamUGC_GetItemState, func(self uintptr, item PublishedFileId_t) uint32 {
		return uint32(EItemStateSubscribed | EItemStateInstalled)
	})
	swap(t, &ptrAPI_ISteamUGC_DownloadItem, func(self uintptr, item PublishedFileId_t, high bool) bool { return high && item == 7 })
	swap(t, &ptrAPI_ISteamUGC_GetItemDownloadInfo, func(self uintptr, item PublishedFileId_t, downloaded, total *uint64) bool {
		*downloaded, *total = 3, 9
		return item == 7
	})
	if ugc.SubscribeItem(7) != 8 || ugc.UnsubscribeItem(7) != 9 {
		t.Error("Subscribe / Unsubscribe")
	}
	if state := ugc.GetItemState(7); state&EItemStateInstalled == 0 || state&EItemStateSubscribed == 0 {
		t.Errorf("GetItemState = %d", state)
	}
	if !ugc.DownloadItem(7, true) || ugc.DownloadItem(7, false) {
		t.Error("DownloadItem")
	}
	if downloaded, total, ok := ugc.GetItemDownloadInfo(7); !ok || downloaded != 3 || total != 9 {
		t.Errorf("GetItemDownloadInfo = %d %d %v", downloaded, total, ok)
	}
}

func TestSteamUGC_GetSubscribedItems(t *testing.T) {
	swap(t, &ptrAPI_ISteamUGC_GetNumSubscribedItems, func(self uintptr, disabled bool) uint32 {
		if disabled {
			return 3
		}
		return 0
	})
	swap(t, &ptrAPI_ISteamUGC_GetSubscribedItems, func(self uintptr, items *PublishedFileId_t, max uint32, disabled bool) uint32 {
		list := unsafe.Slice(items, max)
		list[0], list[1] = 101, 102
		return 2 // one went away between the two calls
	})
	if got := ugc.GetSubscribedItems(false); got != nil {
		t.Errorf("none subscribed = %v", got)
	}
	if got := ugc.GetSubscribedItems(true); len(got) != 2 || got[0] != 101 || got[1] != 102 {
		t.Errorf("GetSubscribedItems = %v", got)
	}
}

func TestSteamUGC_GetItemInstallInfo(t *testing.T) {
	swap(t, &ptrAPI_ISteamUGC_GetItemInstallInfo, func(self uintptr, item PublishedFileId_t, size *uint64, folder *byte, folderSize uint32, stamp *uint32) bool {
		if item != 7 {
			return false
		}
		copy(unsafe.Slice(folder, folderSize), "/steam/workshop/content/5534970/7\x00junk")
		*size, *stamp = 4096, 1763280000
		return folderSize == installFolderMax
	})
	info, ok := ugc.GetItemInstallInfo(7)
	if !ok || info != (ItemInstallInfo{SizeOnDisk: 4096, Folder: "/steam/workshop/content/5534970/7", TimeStamp: 1763280000}) {
		t.Errorf("GetItemInstallInfo = %+v, %v", info, ok)
	}
	if _, ok := ugc.GetItemInstallInfo(8); ok {
		t.Error("an item that is not installed")
	}
}

// The details of items: the request asks for the long description, and the
// details are read with this system's packing; the query is released.
func TestSteamUGC_TheDetailsOfItems(t *testing.T) {
	var released []UGCQueryHandle_t
	var longDescription bool
	swap(t, &ptrAPI_ISteamUGC_CreateQueryUGCDetailsRequest, func(self uintptr, items *PublishedFileId_t, n uint32) UGCQueryHandle_t {
		if n == 2 && unsafe.Slice(items, n)[1] == 102 {
			return 55
		}
		return UGCQueryHandleInvalid
	})
	swap(t, &ptrAPI_ISteamUGC_SetReturnLongDescription, func(self uintptr, q UGCQueryHandle_t, on bool) bool {
		longDescription = on
		return true
	})
	swap(t, &ptrAPI_ISteamUGC_SendQueryUGCRequest, func(self uintptr, q UGCQueryHandle_t) SteamAPICall_t { return 66 })
	swap(t, &ptrAPI_ISteamUGC_ReleaseQueryUGCRequest, func(self uintptr, q UGCQueryHandle_t) bool {
		released = append(released, q)
		return true
	})
	swap(t, &ptrAPI_ISteamUGC_GetQueryUGCResult, func(self uintptr, q UGCQueryHandle_t, index uint32, details *byte) bool {
		if index == 1 {
			return false // Steam has no details for this one
		}
		l := layoutOf(ugcDetailsFields, callbackPack)
		b := unsafe.Slice(details, l.size)
		put(b, l.offsets[0], uint64(101))
		put(b, l.offsets[1], int32(EResultOK))
		put(b, l.offsets[5], "Big Tanks")
		put(b, l.offsets[7], uint64(76561198000000001))
		return true
	})

	if call := ugc.RequestItemDetails([]PublishedFileId_t{101, 102}); call != 66 || !longDescription {
		t.Errorf("RequestItemDetails = %d, long description %v", call, longDescription)
	}
	details := ugc.GetQueryDetails(QueryCompleted{Handle: 55, Result: EResultOK, Returned: 2, Total: 2})
	if len(details) != 1 || details[0].Title != "Big Tanks" || details[0].Owner != 76561198000000001 || details[0].PublishedFileID != 101 {
		t.Errorf("GetQueryDetails = %+v", details)
	}
	if len(released) != 1 || released[0] != 55 {
		t.Errorf("released %v", released)
	}
	// A query that failed is released and gives nothing.
	if details := ugc.GetQueryDetails(QueryCompleted{Handle: 56, Result: EResultFail}); details != nil || len(released) != 2 {
		t.Errorf("a failed query: %v, released %v", details, released)
	}
	// No items, or a request Steam did not make: no call.
	if ugc.RequestItemDetails(nil) != SteamAPICallInvalid || ugc.RequestItemDetails([]PublishedFileId_t{9}) != SteamAPICallInvalid {
		t.Error("a request with nothing to ask")
	}
}

// A query Steam made but could not send is released at once.
func TestSteamUGC_AQueryThatCannotBeSentIsReleased(t *testing.T) {
	var released []UGCQueryHandle_t
	swap(t, &ptrAPI_ISteamUGC_CreateQueryUGCDetailsRequest, func(self uintptr, items *PublishedFileId_t, n uint32) UGCQueryHandle_t { return 57 })
	swap(t, &ptrAPI_ISteamUGC_SetReturnLongDescription, func(self uintptr, q UGCQueryHandle_t, on bool) bool { return true })
	swap(t, &ptrAPI_ISteamUGC_SendQueryUGCRequest, func(self uintptr, q UGCQueryHandle_t) SteamAPICall_t { return SteamAPICallInvalid })
	swap(t, &ptrAPI_ISteamUGC_ReleaseQueryUGCRequest, func(self uintptr, q UGCQueryHandle_t) bool {
		released = append(released, q)
		return true
	})
	if call := ugc.RequestItemDetails([]PublishedFileId_t{1}); call != SteamAPICallInvalid || len(released) != 1 || released[0] != 57 {
		t.Errorf("call %d, released %v", call, released)
	}
}

func TestSteamUGC_TheWorkshopAgreement(t *testing.T) {
	swap(t, &ptrAPI_ISteamUGC_GetWorkshopEULAStatus, func(self uintptr) SteamAPICall_t { return 77 })
	swap(t, &ptrAPI_ISteamUGC_ShowWorkshopEULA, func(self uintptr) bool { return self == 0xBEEF })
	if ugc.GetWorkshopEULAStatus() != 77 || !ugc.ShowWorkshopEULA() {
		t.Error("the Workshop agreement")
	}
}

func TestTheAppIDAndTheOverlayWebPage(t *testing.T) {
	swap(t, &ptrAPI_ISteamUtils_GetAppID, func(self uintptr) AppId_t { return 5534970 })
	var opened string
	var mode EActivateGameOverlayToWebPageMode
	swap(t, &ptrAPI_ISteamFriends_ActivateGameOverlayToWebPage, func(self uintptr, url string, m EActivateGameOverlayToWebPageMode) {
		opened, mode = url, m
	})
	if got := steamUtils(1).GetAppID(); got != 5534970 {
		t.Errorf("GetAppID = %d", got)
	}
	steamFriends(1).ActivateGameOverlayToWebPage("steam://url/CommunityFilePage/77", EActivateGameOverlayToWebPageModeModal)
	if opened != "steam://url/CommunityFilePage/77" || mode != EActivateGameOverlayToWebPageModeModal {
		t.Errorf("opened %q in mode %d", opened, mode)
	}
}

// SetItemTags hands Steam a SteamParamStringArray_t: the count, and a
// pointer to that many C strings, each ending in a zero byte.
func TestSteamUGC_SetItemTagsPassesEveryTagAsACString(t *testing.T) {
	var got []string
	var adminTags bool
	swap(t, &ptrAPI_ISteamUGC_SetItemTags, func(self uintptr, h UGCUpdateHandle_t, array *steamParamStringArray, allowAdmin bool) bool {
		adminTags = allowAdmin
		for _, pointer := range unsafe.Slice(array.strings, array.count) {
			var text []byte
			for n := 0; *(*byte)(unsafe.Add(unsafe.Pointer(pointer), n)) != 0; n++ {
				text = append(text, *(*byte)(unsafe.Add(unsafe.Pointer(pointer), n)))
			}
			got = append(got, string(text))
		}
		return self == uintptr(ugc) && h == 22
	})
	if !ugc.SetItemTags(22, []string{"Map", "2 Players", "Free for all"}) {
		t.Error("SetItemTags said no")
	}
	if len(got) != 3 || got[0] != "Map" || got[1] != "2 Players" || got[2] != "Free for all" || adminTags {
		t.Errorf("tags %q, admin %v", got, adminTags)
	}
	got = nil
	if !ugc.SetItemTags(22, nil) || len(got) != 0 {
		t.Errorf("no tags: %q", got)
	}
}

// Issue #914: the key-value tags of an item, and a query for them.
func TestSteamUGC_KeyValueTagsAndTheirQuery(t *testing.T) {
	var tagged, removed, required []string
	swap(t, &ptrAPI_ISteamUGC_AddItemKeyValueTag, func(self uintptr, h UGCUpdateHandle_t, key, value string) bool {
		tagged = append(tagged, key+"="+value)
		return h == 22
	})
	swap(t, &ptrAPI_ISteamUGC_RemoveItemKeyValueTags, func(self uintptr, h UGCUpdateHandle_t, key string) bool {
		removed = append(removed, key)
		return h == 22
	})
	if !ugc.RemoveItemKeyValueTags(22, "mod") || !ugc.AddItemKeyValueTag(22, "mod", "P-A#isle@map") || len(tagged) != 1 || tagged[0] != "mod=P-A#isle@map" || removed[0] != "mod" {
		t.Errorf("tagged %q, removed %q", tagged, removed)
	}

	var created struct {
		query    EUGCQuery
		matching EUGCMatchingUGCType
		creator  AppId_t
		page     uint32
	}
	handle := UGCQueryHandle_t(57)
	swap(t, &ptrAPI_ISteamUGC_CreateQueryAllUGCRequestPage, func(self uintptr, q EUGCQuery, m EUGCMatchingUGCType, creator, consumer AppId_t, page uint32) UGCQueryHandle_t {
		created.query, created.matching, created.creator, created.page = q, m, creator, page
		return handle
	})
	requiredOK := true
	swap(t, &ptrAPI_ISteamUGC_AddRequiredKeyValueTag, func(self uintptr, h UGCQueryHandle_t, key, value string) bool {
		required = append(required, key+"="+value)
		return requiredOK
	})
	sent := SteamAPICall_t(9)
	swap(t, &ptrAPI_ISteamUGC_SendQueryUGCRequest, func(self uintptr, h UGCQueryHandle_t) SteamAPICall_t { return sent })
	released := 0
	swap(t, &ptrAPI_ISteamUGC_ReleaseQueryUGCRequest, func(self uintptr, h UGCQueryHandle_t) bool { released++; return true })

	if call := ugc.QueryItemsByKeyValue(480, "mod", "P-A#isle@map"); call != 9 || created.query != EUGCQueryRankedByPublicationDate ||
		created.matching != EUGCMatchingUGCTypeItems || created.creator != 480 || created.page != 1 || required[0] != "mod=P-A#isle@map" || released != 0 {
		t.Errorf("the query: call %d, %+v, %q, released %d", call, created, required, released)
	}
	sent = SteamAPICallInvalid
	if ugc.QueryItemsByKeyValue(480, "mod", "x") != SteamAPICallInvalid || released != 1 {
		t.Errorf("a send Steam did not take: released %d", released)
	}
	requiredOK = false
	if ugc.QueryItemsByKeyValue(480, "mod", "x") != SteamAPICallInvalid || released != 2 {
		t.Errorf("a tag Steam did not take: released %d", released)
	}
	handle = UGCQueryHandleInvalid
	if ugc.QueryItemsByKeyValue(480, "mod", "x") != SteamAPICallInvalid {
		t.Error("a query Steam did not make")
	}
}
