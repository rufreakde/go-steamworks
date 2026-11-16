// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

package steamworks

import "github.com/ebitengine/purego"

// This file binds ISteamUGC, the Steam Workshop: making and updating an
// item, subscribing, the state and the folder of an installed item, the
// details of items, and the Workshop's legal agreement. A call that answers
// later returns a SteamAPICall_t; its result comes through PollCallbacks
// (callresult.go). Added for modwars-workspace issue #802.

type (
	PublishedFileId_t uint64
	UGCUpdateHandle_t uint64
	UGCQueryHandle_t  uint64
	SteamAPICall_t    uint64
)

// The value of a handle or a call that Steam did not make.
const (
	UGCUpdateHandleInvalid UGCUpdateHandle_t = 0xffffffffffffffff
	UGCQueryHandleInvalid  UGCQueryHandle_t  = 0xffffffffffffffff
	SteamAPICallInvalid    SteamAPICall_t    = 0
)

type EWorkshopFileType int32

// EWorkshopFileTypeCommunity is a normal Workshop item, the one a Mod is.
const EWorkshopFileTypeCommunity EWorkshopFileType = 0

type ERemoteStoragePublishedFileVisibility int32

const (
	ERemoteStoragePublishedFileVisibilityPublic      ERemoteStoragePublishedFileVisibility = 0
	ERemoteStoragePublishedFileVisibilityFriendsOnly ERemoteStoragePublishedFileVisibility = 1
	ERemoteStoragePublishedFileVisibilityPrivate     ERemoteStoragePublishedFileVisibility = 2
	ERemoteStoragePublishedFileVisibilityUnlisted    ERemoteStoragePublishedFileVisibility = 3
)

// EItemState is a set of bits (GetItemState).
type EItemState uint32

const (
	EItemStateSubscribed      EItemState = 1
	EItemStateLegacyItem      EItemState = 2
	EItemStateInstalled       EItemState = 4
	EItemStateNeedsUpdate     EItemState = 8
	EItemStateDownloading     EItemState = 16
	EItemStateDownloadPending EItemState = 32
	EItemStateDisabledLocally EItemState = 64
)

type EItemUpdateStatus int32

const (
	EItemUpdateStatusInvalid              EItemUpdateStatus = 0
	EItemUpdateStatusPreparingConfig      EItemUpdateStatus = 1
	EItemUpdateStatusPreparingContent     EItemUpdateStatus = 2
	EItemUpdateStatusUploadingContent     EItemUpdateStatus = 3
	EItemUpdateStatusUploadingPreviewFile EItemUpdateStatus = 4
	EItemUpdateStatusCommittingChanges    EItemUpdateStatus = 5
)

// installFolderMax is the size of the buffer GetItemInstallInfo writes the
// item's folder into.
const installFolderMax = 4096

// ItemInstallInfo is where an installed item lies on the disk.
type ItemInstallInfo struct {
	SizeOnDisk uint64
	Folder     string
	TimeStamp  uint32
}

// ISteamUGC is the Steam Workshop.
type ISteamUGC interface {
	// CreateItem makes a new item of the game; its result is a CreateItemResult.
	CreateItem(appID AppId_t, fileType EWorkshopFileType) SteamAPICall_t
	// StartItemUpdate starts a change of an item; the SetItem calls fill it,
	// SubmitItemUpdate sends it.
	StartItemUpdate(appID AppId_t, item PublishedFileId_t) UGCUpdateHandle_t
	SetItemTitle(update UGCUpdateHandle_t, title string) bool
	SetItemDescription(update UGCUpdateHandle_t, description string) bool
	SetItemVisibility(update UGCUpdateHandle_t, visibility ERemoteStoragePublishedFileVisibility) bool
	// SetItemContent names the folder whose files become the item.
	SetItemContent(update UGCUpdateHandle_t, folder string) bool
	SetItemPreview(update UGCUpdateHandle_t, file string) bool
	// SubmitItemUpdate sends the change; its result is a SubmitItemUpdateResult.
	SubmitItemUpdate(update UGCUpdateHandle_t, changeNote string) SteamAPICall_t
	// GetItemUpdateProgress says how far a submitted change is.
	GetItemUpdateProgress(update UGCUpdateHandle_t) (status EItemUpdateStatus, processed, total uint64)
	// SubscribeItem and UnsubscribeItem: their results are SubscribeResults.
	SubscribeItem(item PublishedFileId_t) SteamAPICall_t
	UnsubscribeItem(item PublishedFileId_t) SteamAPICall_t
	// GetSubscribedItems lists the items the player subscribed to.
	GetSubscribedItems(includeLocallyDisabled bool) []PublishedFileId_t
	GetItemState(item PublishedFileId_t) EItemState
	// GetItemInstallInfo says where an installed item is; false when it is not installed.
	GetItemInstallInfo(item PublishedFileId_t) (ItemInstallInfo, bool)
	// GetItemDownloadInfo says how far a download is; false when none runs.
	GetItemDownloadInfo(item PublishedFileId_t) (downloaded, total uint64, ok bool)
	// DownloadItem starts or hurries a download; a DownloadItemResult follows.
	DownloadItem(item PublishedFileId_t, highPriority bool) bool
	// RequestItemDetails asks for the details of items; the result is a
	// QueryCompleted, whose details GetQueryDetails reads.
	RequestItemDetails(items []PublishedFileId_t) SteamAPICall_t
	// GetQueryDetails reads the details of the items of a completed query
	// and releases it.
	GetQueryDetails(query QueryCompleted) []UGCDetails
	// GetWorkshopEULAStatus asks whether the player accepted the Workshop's
	// legal agreement; the result is a WorkshopEULAStatus.
	GetWorkshopEULAStatus() SteamAPICall_t
	// ShowWorkshopEULA shows the agreement in the Steam overlay.
	ShowWorkshopEULA() bool
}

var (
	ptrAPI_SteamUGC                               func() uintptr
	ptrAPI_ISteamUGC_CreateItem                   func(uintptr, AppId_t, EWorkshopFileType) SteamAPICall_t
	ptrAPI_ISteamUGC_StartItemUpdate              func(uintptr, AppId_t, PublishedFileId_t) UGCUpdateHandle_t
	ptrAPI_ISteamUGC_SetItemTitle                 func(uintptr, UGCUpdateHandle_t, string) bool
	ptrAPI_ISteamUGC_SetItemDescription           func(uintptr, UGCUpdateHandle_t, string) bool
	ptrAPI_ISteamUGC_SetItemVisibility            func(uintptr, UGCUpdateHandle_t, ERemoteStoragePublishedFileVisibility) bool
	ptrAPI_ISteamUGC_SetItemContent               func(uintptr, UGCUpdateHandle_t, string) bool
	ptrAPI_ISteamUGC_SetItemPreview               func(uintptr, UGCUpdateHandle_t, string) bool
	ptrAPI_ISteamUGC_SubmitItemUpdate             func(uintptr, UGCUpdateHandle_t, string) SteamAPICall_t
	ptrAPI_ISteamUGC_GetItemUpdateProgress        func(uintptr, UGCUpdateHandle_t, *uint64, *uint64) EItemUpdateStatus
	ptrAPI_ISteamUGC_SubscribeItem                func(uintptr, PublishedFileId_t) SteamAPICall_t
	ptrAPI_ISteamUGC_UnsubscribeItem              func(uintptr, PublishedFileId_t) SteamAPICall_t
	ptrAPI_ISteamUGC_GetNumSubscribedItems        func(uintptr, bool) uint32
	ptrAPI_ISteamUGC_GetSubscribedItems           func(uintptr, *PublishedFileId_t, uint32, bool) uint32
	ptrAPI_ISteamUGC_GetItemState                 func(uintptr, PublishedFileId_t) uint32
	ptrAPI_ISteamUGC_GetItemInstallInfo           func(uintptr, PublishedFileId_t, *uint64, *byte, uint32, *uint32) bool
	ptrAPI_ISteamUGC_GetItemDownloadInfo          func(uintptr, PublishedFileId_t, *uint64, *uint64) bool
	ptrAPI_ISteamUGC_DownloadItem                 func(uintptr, PublishedFileId_t, bool) bool
	ptrAPI_ISteamUGC_CreateQueryUGCDetailsRequest func(uintptr, *PublishedFileId_t, uint32) UGCQueryHandle_t
	ptrAPI_ISteamUGC_SetReturnLongDescription     func(uintptr, UGCQueryHandle_t, bool) bool
	ptrAPI_ISteamUGC_SendQueryUGCRequest          func(uintptr, UGCQueryHandle_t) SteamAPICall_t
	ptrAPI_ISteamUGC_GetQueryUGCResult            func(uintptr, UGCQueryHandle_t, uint32, *byte) bool
	ptrAPI_ISteamUGC_ReleaseQueryUGCRequest       func(uintptr, UGCQueryHandle_t) bool
	ptrAPI_ISteamUGC_GetWorkshopEULAStatus        func(uintptr) SteamAPICall_t
	ptrAPI_ISteamUGC_ShowWorkshopEULA             func(uintptr) bool
	ptrAPI_ManualDispatch_GetAPICallResult        func(HSteamPipe, SteamAPICall_t, *byte, int32, int32, *bool) bool
)

// registerUGC binds ISteamUGC and the call results (registerFunctions).
func registerUGC(lib uintptr) {
	purego.RegisterLibFunc(&ptrAPI_ManualDispatch_GetAPICallResult, lib, flatAPI_ManualDispatch_GetAPICallResult)
	purego.RegisterLibFunc(&ptrAPI_SteamUGC, lib, flatAPI_SteamUGC)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_CreateItem, lib, flatAPI_ISteamUGC_CreateItem)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_StartItemUpdate, lib, flatAPI_ISteamUGC_StartItemUpdate)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_SetItemTitle, lib, flatAPI_ISteamUGC_SetItemTitle)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_SetItemDescription, lib, flatAPI_ISteamUGC_SetItemDescription)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_SetItemVisibility, lib, flatAPI_ISteamUGC_SetItemVisibility)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_SetItemContent, lib, flatAPI_ISteamUGC_SetItemContent)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_SetItemPreview, lib, flatAPI_ISteamUGC_SetItemPreview)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_SubmitItemUpdate, lib, flatAPI_ISteamUGC_SubmitItemUpdate)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_GetItemUpdateProgress, lib, flatAPI_ISteamUGC_GetItemUpdateProgress)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_SubscribeItem, lib, flatAPI_ISteamUGC_SubscribeItem)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_UnsubscribeItem, lib, flatAPI_ISteamUGC_UnsubscribeItem)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_GetNumSubscribedItems, lib, flatAPI_ISteamUGC_GetNumSubscribedItems)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_GetSubscribedItems, lib, flatAPI_ISteamUGC_GetSubscribedItems)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_GetItemState, lib, flatAPI_ISteamUGC_GetItemState)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_GetItemInstallInfo, lib, flatAPI_ISteamUGC_GetItemInstallInfo)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_GetItemDownloadInfo, lib, flatAPI_ISteamUGC_GetItemDownloadInfo)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_DownloadItem, lib, flatAPI_ISteamUGC_DownloadItem)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_CreateQueryUGCDetailsRequest, lib, flatAPI_ISteamUGC_CreateQueryUGCDetailsRequest)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_SetReturnLongDescription, lib, flatAPI_ISteamUGC_SetReturnLongDescription)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_SendQueryUGCRequest, lib, flatAPI_ISteamUGC_SendQueryUGCRequest)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_GetQueryUGCResult, lib, flatAPI_ISteamUGC_GetQueryUGCResult)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_ReleaseQueryUGCRequest, lib, flatAPI_ISteamUGC_ReleaseQueryUGCRequest)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_GetWorkshopEULAStatus, lib, flatAPI_ISteamUGC_GetWorkshopEULAStatus)
	purego.RegisterLibFunc(&ptrAPI_ISteamUGC_ShowWorkshopEULA, lib, flatAPI_ISteamUGC_ShowWorkshopEULA)
}

// SteamUGC is the Steam Workshop of the player.
func SteamUGC() ISteamUGC { return steamUGC(ptrAPI_SteamUGC()) }

type steamUGC uintptr

func (s steamUGC) CreateItem(appID AppId_t, fileType EWorkshopFileType) SteamAPICall_t {
	return ptrAPI_ISteamUGC_CreateItem(uintptr(s), appID, fileType)
}

func (s steamUGC) StartItemUpdate(appID AppId_t, item PublishedFileId_t) UGCUpdateHandle_t {
	return ptrAPI_ISteamUGC_StartItemUpdate(uintptr(s), appID, item)
}

func (s steamUGC) SetItemTitle(update UGCUpdateHandle_t, title string) bool {
	return ptrAPI_ISteamUGC_SetItemTitle(uintptr(s), update, title)
}

func (s steamUGC) SetItemDescription(update UGCUpdateHandle_t, description string) bool {
	return ptrAPI_ISteamUGC_SetItemDescription(uintptr(s), update, description)
}

func (s steamUGC) SetItemVisibility(update UGCUpdateHandle_t, visibility ERemoteStoragePublishedFileVisibility) bool {
	return ptrAPI_ISteamUGC_SetItemVisibility(uintptr(s), update, visibility)
}

func (s steamUGC) SetItemContent(update UGCUpdateHandle_t, folder string) bool {
	return ptrAPI_ISteamUGC_SetItemContent(uintptr(s), update, folder)
}

func (s steamUGC) SetItemPreview(update UGCUpdateHandle_t, file string) bool {
	return ptrAPI_ISteamUGC_SetItemPreview(uintptr(s), update, file)
}

func (s steamUGC) SubmitItemUpdate(update UGCUpdateHandle_t, changeNote string) SteamAPICall_t {
	return ptrAPI_ISteamUGC_SubmitItemUpdate(uintptr(s), update, changeNote)
}

func (s steamUGC) GetItemUpdateProgress(update UGCUpdateHandle_t) (EItemUpdateStatus, uint64, uint64) {
	var processed, total uint64
	status := ptrAPI_ISteamUGC_GetItemUpdateProgress(uintptr(s), update, &processed, &total)
	return status, processed, total
}

func (s steamUGC) SubscribeItem(item PublishedFileId_t) SteamAPICall_t {
	return ptrAPI_ISteamUGC_SubscribeItem(uintptr(s), item)
}

func (s steamUGC) UnsubscribeItem(item PublishedFileId_t) SteamAPICall_t {
	return ptrAPI_ISteamUGC_UnsubscribeItem(uintptr(s), item)
}

func (s steamUGC) GetSubscribedItems(includeLocallyDisabled bool) []PublishedFileId_t {
	count := ptrAPI_ISteamUGC_GetNumSubscribedItems(uintptr(s), includeLocallyDisabled)
	if count == 0 {
		return nil
	}
	items := make([]PublishedFileId_t, count)
	got := ptrAPI_ISteamUGC_GetSubscribedItems(uintptr(s), &items[0], count, includeLocallyDisabled)
	return items[:min(got, count)]
}

func (s steamUGC) GetItemState(item PublishedFileId_t) EItemState {
	return EItemState(ptrAPI_ISteamUGC_GetItemState(uintptr(s), item))
}

func (s steamUGC) GetItemInstallInfo(item PublishedFileId_t) (ItemInstallInfo, bool) {
	var info ItemInstallInfo
	folder := make([]byte, installFolderMax)
	if !ptrAPI_ISteamUGC_GetItemInstallInfo(uintptr(s), item, &info.SizeOnDisk, &folder[0], uint32(len(folder)), &info.TimeStamp) {
		return ItemInstallInfo{}, false
	}
	info.Folder = cString(folder, 0, len(folder))
	return info, true
}

func (s steamUGC) GetItemDownloadInfo(item PublishedFileId_t) (uint64, uint64, bool) {
	var downloaded, total uint64
	ok := ptrAPI_ISteamUGC_GetItemDownloadInfo(uintptr(s), item, &downloaded, &total)
	return downloaded, total, ok
}

func (s steamUGC) DownloadItem(item PublishedFileId_t, highPriority bool) bool {
	return ptrAPI_ISteamUGC_DownloadItem(uintptr(s), item, highPriority)
}

func (s steamUGC) RequestItemDetails(items []PublishedFileId_t) SteamAPICall_t {
	if len(items) == 0 {
		return SteamAPICallInvalid
	}
	query := ptrAPI_ISteamUGC_CreateQueryUGCDetailsRequest(uintptr(s), &items[0], uint32(len(items)))
	if query == UGCQueryHandleInvalid {
		return SteamAPICallInvalid
	}
	ptrAPI_ISteamUGC_SetReturnLongDescription(uintptr(s), query, true)
	call := ptrAPI_ISteamUGC_SendQueryUGCRequest(uintptr(s), query)
	if call == SteamAPICallInvalid {
		ptrAPI_ISteamUGC_ReleaseQueryUGCRequest(uintptr(s), query)
	}
	return call
}

func (s steamUGC) GetQueryDetails(query QueryCompleted) []UGCDetails {
	defer ptrAPI_ISteamUGC_ReleaseQueryUGCRequest(uintptr(s), query.Handle)
	if query.Result != EResultOK {
		return nil
	}
	// Steam writes one SteamUGCDetails_t into the buffer: its size for this
	// system's packing, and some room more.
	buffer := make([]byte, layoutOf(ugcDetailsFields, callbackPack).size+64)
	var details []UGCDetails
	for index := uint32(0); index < query.Returned; index++ {
		clear(buffer)
		if !ptrAPI_ISteamUGC_GetQueryUGCResult(uintptr(s), query.Handle, index, &buffer[0]) {
			continue
		}
		if item, ok := decodeUGCDetails(buffer, callbackPack); ok {
			details = append(details, item)
		}
	}
	return details
}

func (s steamUGC) GetWorkshopEULAStatus() SteamAPICall_t {
	return ptrAPI_ISteamUGC_GetWorkshopEULAStatus(uintptr(s))
}

func (s steamUGC) ShowWorkshopEULA() bool { return ptrAPI_ISteamUGC_ShowWorkshopEULA(uintptr(s)) }
