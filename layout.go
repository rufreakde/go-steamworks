// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

package steamworks

import (
	"bytes"
	"encoding/binary"
)

// Steam packs its callback structs and the structs of isteamugc.h to 4
// bytes on Linux and macOS and to 8 bytes on Windows (steamclientpublic.h,
// VALVE_CALLBACK_PACK_SMALL / _LARGE). A Go struct is laid out by Go's own
// rules, which match neither on every system, so the Workshop's structs are
// read field by field at offsets worked out here for callbackPack.

// field is one field of a native struct: its size and its own alignment.
type field struct {
	size, align int
}

var (
	fieldBool   = field{1, 1}
	fieldInt32  = field{4, 4} // also an enum, uint32, float, AppId_t, RTime32
	fieldUint64 = field{8, 8} // also PublishedFileId_t, the handles, SteamAPICall_t
)

func fieldChars(n int) field { return field{n, 1} }

// layout is where each field of a struct lies, and the struct's size.
type layout struct {
	offsets []int
	size    int
}

// layoutOf lays fields out as the C compiler does under #pragma pack(pack):
// each field at the next multiple of min(its alignment, pack), the size a
// multiple of the largest such alignment. Pure.
func layoutOf(fields []field, pack int) layout {
	var l layout
	offset, largest := 0, 1
	for _, f := range fields {
		align := min(f.align, pack)
		largest = max(largest, align)
		offset = (offset + align - 1) / align * align
		l.offsets = append(l.offsets, offset)
		offset += f.size
	}
	l.size = (offset + largest - 1) / largest * largest
	return l
}

// The structs, field by field, as steam_api.json of SDK 1.65 lists them.
var (
	// CreateItemResult_t: m_eResult, m_nPublishedFileId,
	// m_bUserNeedsToAcceptWorkshopLegalAgreement.
	createItemResultFields = []field{fieldInt32, fieldUint64, fieldBool}
	// SubmitItemUpdateResult_t: m_eResult,
	// m_bUserNeedsToAcceptWorkshopLegalAgreement, m_nPublishedFileId.
	submitItemUpdateResultFields = []field{fieldInt32, fieldBool, fieldUint64}
	// RemoteStorageSubscribePublishedFileResult_t and its Unsubscribe twin:
	// m_eResult, m_nPublishedFileId.
	subscribeResultFields = []field{fieldInt32, fieldUint64}
	// SteamUGCQueryCompleted_t: m_handle, m_eResult, m_unNumResultsReturned,
	// m_unTotalMatchingResults, m_bCachedData, m_rgchNextCursor[256].
	queryCompletedFields = []field{fieldUint64, fieldInt32, fieldInt32, fieldInt32, fieldBool, fieldChars(256)}
	// WorkshopEULAStatus_t: m_eResult, m_nAppID, m_unVersion, m_rtAction,
	// m_bAccepted, m_bNeedsAction.
	eulaStatusFields = []field{fieldInt32, fieldInt32, fieldInt32, fieldInt32, fieldBool, fieldBool}
	// ItemInstalled_t: m_unAppID, m_nPublishedFileId, m_hLegacyContent,
	// m_unManifestID.
	itemInstalledFields = []field{fieldInt32, fieldUint64, fieldUint64, fieldUint64}
	// DownloadItemResult_t: m_unAppID, m_nPublishedFileId, m_eResult.
	downloadItemResultFields = []field{fieldInt32, fieldUint64, fieldInt32}
	// SteamUGCDetails_t, all 27 fields: m_nPublishedFileId, m_eResult,
	// m_eFileType, m_nCreatorAppID, m_nConsumerAppID, m_rgchTitle[129],
	// m_rgchDescription[8000], m_ulSteamIDOwner, m_rtimeCreated,
	// m_rtimeUpdated, m_rtimeAddedToUserList, m_eVisibility, m_bBanned,
	// m_bAcceptedForUse, m_bTagsTruncated, m_rgchTags[1025], m_hFile,
	// m_hPreviewFile, m_pchFileName[260], m_nFileSize, m_nPreviewFileSize,
	// m_rgchURL[256], m_unVotesUp, m_unVotesDown, m_flScore,
	// m_unNumChildren, m_ulTotalFilesSize.
	ugcDetailsFields = []field{
		fieldUint64, fieldInt32, fieldInt32, fieldInt32, fieldInt32, fieldChars(129), fieldChars(8000),
		fieldUint64, fieldInt32, fieldInt32, fieldInt32, fieldInt32, fieldBool, fieldBool, fieldBool,
		fieldChars(1025), fieldUint64, fieldUint64, fieldChars(260), fieldInt32, fieldInt32,
		fieldChars(256), fieldInt32, fieldInt32, fieldInt32, fieldInt32, fieldUint64,
	}
)

func u32(b []byte, offset int) uint32 { return binary.LittleEndian.Uint32(b[offset:]) }
func i32(b []byte, offset int) int32  { return int32(binary.LittleEndian.Uint32(b[offset:])) }
func u64(b []byte, offset int) uint64 { return binary.LittleEndian.Uint64(b[offset:]) }
func flag(b []byte, offset int) bool  { return b[offset] != 0 }

// cString reads the text of a char array of n bytes at offset: up to its
// end mark, never past the array.
func cString(b []byte, offset, n int) string {
	text := b[offset : offset+n]
	if end := bytes.IndexByte(text, 0); end >= 0 {
		text = text[:end]
	}
	return string(text)
}

// readable says whether b holds the whole struct of l.
func readable(b []byte, l layout) bool { return len(b) >= l.size }

// CreateItemResult is the result of CreateItem.
type CreateItemResult struct {
	Result          EResult
	PublishedFileID PublishedFileId_t
	// NeedsLegalAgreement: the player must accept the Workshop's legal
	// agreement before the item can be seen (ShowWorkshopEULA).
	NeedsLegalAgreement bool
}

func decodeCreateItemResult(b []byte, pack int) (CreateItemResult, bool) {
	l := layoutOf(createItemResultFields, pack)
	if !readable(b, l) {
		return CreateItemResult{}, false
	}
	return CreateItemResult{EResult(i32(b, l.offsets[0])), PublishedFileId_t(u64(b, l.offsets[1])), flag(b, l.offsets[2])}, true
}

// SubmitItemUpdateResult is the result of SubmitItemUpdate.
type SubmitItemUpdateResult struct {
	Result              EResult
	NeedsLegalAgreement bool
	PublishedFileID     PublishedFileId_t
}

func decodeSubmitItemUpdateResult(b []byte, pack int) (SubmitItemUpdateResult, bool) {
	l := layoutOf(submitItemUpdateResultFields, pack)
	if !readable(b, l) {
		return SubmitItemUpdateResult{}, false
	}
	return SubmitItemUpdateResult{EResult(i32(b, l.offsets[0])), flag(b, l.offsets[1]), PublishedFileId_t(u64(b, l.offsets[2]))}, true
}

// SubscribeResult is the result of SubscribeItem or UnsubscribeItem.
type SubscribeResult struct {
	Result          EResult
	PublishedFileID PublishedFileId_t
}

func decodeSubscribeResult(b []byte, pack int) (SubscribeResult, bool) {
	l := layoutOf(subscribeResultFields, pack)
	if !readable(b, l) {
		return SubscribeResult{}, false
	}
	return SubscribeResult{EResult(i32(b, l.offsets[0])), PublishedFileId_t(u64(b, l.offsets[1]))}, true
}

// QueryCompleted is the result of SendQueryUGCRequest: read the items with
// GetQueryUGCResult, then ReleaseQueryUGCRequest.
type QueryCompleted struct {
	Handle   UGCQueryHandle_t
	Result   EResult
	Returned uint32
	Total    uint32
}

func decodeQueryCompleted(b []byte, pack int) (QueryCompleted, bool) {
	l := layoutOf(queryCompletedFields, pack)
	if !readable(b, l) {
		return QueryCompleted{}, false
	}
	return QueryCompleted{UGCQueryHandle_t(u64(b, l.offsets[0])), EResult(i32(b, l.offsets[1])), u32(b, l.offsets[2]), u32(b, l.offsets[3])}, true
}

// WorkshopEULAStatus is the result of GetWorkshopEULAStatus.
type WorkshopEULAStatus struct {
	Result      EResult
	AppID       AppId_t
	Version     uint32
	Accepted    bool
	NeedsAction bool
}

func decodeEULAStatus(b []byte, pack int) (WorkshopEULAStatus, bool) {
	l := layoutOf(eulaStatusFields, pack)
	if !readable(b, l) {
		return WorkshopEULAStatus{}, false
	}
	return WorkshopEULAStatus{EResult(i32(b, l.offsets[0])), AppId_t(u32(b, l.offsets[1])), u32(b, l.offsets[2]), flag(b, l.offsets[4]), flag(b, l.offsets[5])}, true
}

// ItemInstalled is the callback of an item that was installed or updated.
type ItemInstalled struct {
	AppID           AppId_t
	PublishedFileID PublishedFileId_t
}

func decodeItemInstalled(b []byte, pack int) (ItemInstalled, bool) {
	l := layoutOf(itemInstalledFields, pack)
	if !readable(b, l) {
		return ItemInstalled{}, false
	}
	return ItemInstalled{AppId_t(u32(b, l.offsets[0])), PublishedFileId_t(u64(b, l.offsets[1]))}, true
}

// DownloadItemResult is the callback of a download that ended.
type DownloadItemResult struct {
	AppID           AppId_t
	PublishedFileID PublishedFileId_t
	Result          EResult
}

func decodeDownloadItemResult(b []byte, pack int) (DownloadItemResult, bool) {
	l := layoutOf(downloadItemResultFields, pack)
	if !readable(b, l) {
		return DownloadItemResult{}, false
	}
	return DownloadItemResult{AppId_t(u32(b, l.offsets[0])), PublishedFileId_t(u64(b, l.offsets[1])), EResult(i32(b, l.offsets[2]))}, true
}

// UGCDetails is what the game needs of one item's details.
type UGCDetails struct {
	PublishedFileID PublishedFileId_t
	Result          EResult
	Title           string
	Description     string
	// Owner is the SteamID64 of the item's owner.
	Owner          CSteamID
	TimeUpdated    uint32
	Visibility     ERemoteStoragePublishedFileVisibility
	Banned         bool
	TotalFilesSize uint64
}

func decodeUGCDetails(b []byte, pack int) (UGCDetails, bool) {
	l := layoutOf(ugcDetailsFields, pack)
	if !readable(b, l) {
		return UGCDetails{}, false
	}
	o := l.offsets
	return UGCDetails{
		PublishedFileID: PublishedFileId_t(u64(b, o[0])),
		Result:          EResult(i32(b, o[1])),
		Title:           cString(b, o[5], ugcDetailsFields[5].size),
		Description:     cString(b, o[6], ugcDetailsFields[6].size),
		Owner:           CSteamID(u64(b, o[7])),
		TimeUpdated:     u32(b, o[9]),
		Visibility:      ERemoteStoragePublishedFileVisibility(i32(b, o[11])),
		Banned:          flag(b, o[12]),
		TotalFilesSize:  u64(b, o[26]),
	}, true
}
