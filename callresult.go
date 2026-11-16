// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

package steamworks

import "unsafe"

// Call results under ManualDispatch (steam_api.h): a call that answers later
// returns a SteamAPICall_t. When it is done, Steam queues a
// SteamAPICallCompleted_t that names the call, the kind of its result and
// the result's size; the result itself is fetched with
// SteamAPI_ManualDispatch_GetAPICallResult. Added for modwars-workspace
// issue #802 (the Workshop).

// The callback ids this file knows (steam_api.json of SDK 1.65).
const (
	steamAPICallCompletedCallbackID  = 703
	subscribeResultCallbackID        = 1313
	unsubscribeResultCallbackID      = 1315
	queryCompletedCallbackID         = 3401
	createItemResultCallbackID       = 3403
	submitItemUpdateResultCallbackID = 3404
	itemInstalledCallbackID          = 3405
	downloadItemResultCallbackID     = 3406
	eulaStatusCallbackID             = 3420
)

const (
	// callCompletedSize is SteamAPICallCompleted_t: m_hAsyncCall (8),
	// m_iCallback (4), m_cubParam (4). isteamutils.h sets no packing for
	// it, so it is the same on every system.
	callCompletedSize = 16
	// callResultMaxBytes is the largest result this binding fetches. The
	// largest it knows (SteamUGCQueryCompleted_t) has 280 bytes.
	callResultMaxBytes = 64 << 10
)

// CallResult is the result of one call: the handle the call returned, and
// what Steam answered.
type CallResult[T any] struct {
	Call  SteamAPICall_t
	Value T
}

// getAPICallResult fetches the result of call into buffer. A variable, so
// that a test needs no Steam.
var getAPICallResult = func(pipe HSteamPipe, call SteamAPICall_t, buffer []byte, expected int32) (ok, failed bool) {
	ok = ptrAPI_ManualDispatch_GetAPICallResult(pipe, call, &buffer[0], int32(len(buffer)), expected, &failed)
	return ok, failed
}

// paramBytes copies a callback's payload out of Steam's memory, which is
// freed right after the callback is read.
func paramBytes(msg CallbackMsg) []byte {
	if msg.Param == nil || msg.ParamSize <= 0 {
		return nil
	}
	return append([]byte(nil), unsafe.Slice((*byte)(msg.Param), msg.ParamSize)...)
}

// takeWorkshop reads the Workshop's callbacks and call results into c. It
// reports whether msg was one of them.
func (c *Callbacks) takeWorkshop(msg CallbackMsg) bool {
	switch msg.CallbackID {
	case steamAPICallCompletedCallbackID:
		b := paramBytes(msg)
		if len(b) >= callCompletedSize {
			c.takeCallResult(SteamAPICall_t(u64(b, 0)), i32(b, 8), u32(b, 12))
		}
	case itemInstalledCallbackID:
		if installed, ok := decodeItemInstalled(paramBytes(msg), callbackPack); ok {
			c.ItemsInstalled = append(c.ItemsInstalled, installed)
		}
	case downloadItemResultCallbackID:
		if download, ok := decodeDownloadItemResult(paramBytes(msg), callbackPack); ok {
			c.ItemDownloads = append(c.ItemDownloads, download)
		}
	default:
		return false
	}
	return true
}

// takeCallResult fetches the result of a call that is done and reads it
// when this binding knows its kind. A call Steam says failed, or whose
// result cannot be fetched, goes to FailedCalls.
func (c *Callbacks) takeCallResult(call SteamAPICall_t, kind int32, size uint32) {
	if !knownCallResult(kind) {
		return
	}
	if size == 0 || size > callResultMaxBytes {
		c.FailedCalls = append(c.FailedCalls, call)
		return
	}
	buffer := make([]byte, size)
	if ok, failed := getAPICallResult(c.pipe, call, buffer, kind); !ok || failed {
		c.FailedCalls = append(c.FailedCalls, call)
		return
	}
	read := false
	switch kind {
	case createItemResultCallbackID:
		var v CreateItemResult
		if v, read = decodeCreateItemResult(buffer, callbackPack); read {
			c.CreatedItems = append(c.CreatedItems, CallResult[CreateItemResult]{call, v})
		}
	case submitItemUpdateResultCallbackID:
		var v SubmitItemUpdateResult
		if v, read = decodeSubmitItemUpdateResult(buffer, callbackPack); read {
			c.SubmittedUpdates = append(c.SubmittedUpdates, CallResult[SubmitItemUpdateResult]{call, v})
		}
	case subscribeResultCallbackID, unsubscribeResultCallbackID:
		var v SubscribeResult
		if v, read = decodeSubscribeResult(buffer, callbackPack); read {
			result := CallResult[SubscribeResult]{call, v}
			if kind == subscribeResultCallbackID {
				c.Subscribed = append(c.Subscribed, result)
			} else {
				c.Unsubscribed = append(c.Unsubscribed, result)
			}
		}
	case queryCompletedCallbackID:
		var v QueryCompleted
		if v, read = decodeQueryCompleted(buffer, callbackPack); read {
			c.QueriesCompleted = append(c.QueriesCompleted, CallResult[QueryCompleted]{call, v})
		}
	case eulaStatusCallbackID:
		var v WorkshopEULAStatus
		if v, read = decodeEULAStatus(buffer, callbackPack); read {
			c.EULAStatuses = append(c.EULAStatuses, CallResult[WorkshopEULAStatus]{call, v})
		}
	}
	if !read {
		c.FailedCalls = append(c.FailedCalls, call)
	}
}

// knownCallResult says whether this binding reads results of kind.
func knownCallResult(kind int32) bool {
	switch kind {
	case createItemResultCallbackID, submitItemUpdateResultCallbackID, subscribeResultCallbackID,
		unsubscribeResultCallbackID, queryCompletedCallbackID, eulaStatusCallbackID:
		return true
	}
	return false
}
