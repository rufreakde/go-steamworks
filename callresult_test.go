// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

package steamworks

import (
	"encoding/binary"
	"testing"
	"unsafe"
)

// callCompleted is a SteamAPICallCompleted_t callback for call, naming the
// kind and the size of its result.
func callCompleted(call SteamAPICall_t, kind int32, size uint32) CallbackMsg {
	b := make([]byte, callCompletedSize)
	binary.LittleEndian.PutUint64(b[0:], uint64(call))
	binary.LittleEndian.PutUint32(b[8:], uint32(kind))
	binary.LittleEndian.PutUint32(b[12:], size)
	return CallbackMsg{CallbackID: steamAPICallCompletedCallbackID, Param: unsafe.Pointer(&b[0]), ParamSize: int32(len(b))}
}

// steamAnswers makes getAPICallResult hand out results[call] until the test
// ends; a call with no result fails. It records what it was asked.
func steamAnswers(t *testing.T, results map[SteamAPICall_t][]byte) *[]string {
	t.Helper()
	saved := getAPICallResult
	t.Cleanup(func() { getAPICallResult = saved })
	var asked []string
	getAPICallResult = func(pipe HSteamPipe, call SteamAPICall_t, buffer []byte, expected int32) (bool, bool) {
		asked = append(asked, string(rune('0'+pipe)))
		result, ok := results[call]
		if !ok {
			return true, true
		}
		copy(buffer, result)
		return true, false
	}
	return &asked
}

func structOf(fields []field, values ...any) []byte {
	l := layoutOf(fields, callbackPack)
	b := make([]byte, l.size)
	for i, v := range values {
		if v != nil {
			put(b, l.offsets[i], v)
		}
	}
	return b
}

// Happy path: every kind of Workshop result is fetched from Steam and read,
// under the call that asked for it.
func TestPollCallbacks_ReadsTheWorkshopCallResults(t *testing.T) {
	created := structOf(createItemResultFields, int32(EResultOK), uint64(1001), true)
	submitted := structOf(submitItemUpdateResultFields, int32(EResultOK), nil, uint64(1001))
	subscribed := structOf(subscribeResultFields, int32(EResultOK), uint64(2002))
	unsubscribed := structOf(subscribeResultFields, int32(EResultOK), uint64(2003))
	query := structOf(queryCompletedFields, uint64(55), int32(EResultOK), uint32(1), uint32(1))
	eula := structOf(eulaStatusFields, int32(EResultOK), uint32(5534970), uint32(2), nil, true, false)
	asked := steamAnswers(t, map[SteamAPICall_t][]byte{1: created, 2: submitted, 3: subscribed, 4: unsubscribed, 5: query, 6: eula})

	found := Callbacks{pipe: 7}
	for _, msg := range []CallbackMsg{
		callCompleted(1, createItemResultCallbackID, uint32(len(created))),
		callCompleted(2, submitItemUpdateResultCallbackID, uint32(len(submitted))),
		callCompleted(3, subscribeResultCallbackID, uint32(len(subscribed))),
		callCompleted(4, unsubscribeResultCallbackID, uint32(len(unsubscribed))),
		callCompleted(5, queryCompletedCallbackID, uint32(len(query))),
		callCompleted(6, eulaStatusCallbackID, uint32(len(eula))),
	} {
		found.take(msg)
	}

	if len(found.CreatedItems) != 1 || found.CreatedItems[0] != (CallResult[CreateItemResult]{1, CreateItemResult{EResultOK, 1001, true}}) {
		t.Errorf("CreatedItems %+v", found.CreatedItems)
	}
	if len(found.SubmittedUpdates) != 1 || found.SubmittedUpdates[0].Value.PublishedFileID != 1001 || found.SubmittedUpdates[0].Call != 2 {
		t.Errorf("SubmittedUpdates %+v", found.SubmittedUpdates)
	}
	if len(found.Subscribed) != 1 || found.Subscribed[0].Value.PublishedFileID != 2002 {
		t.Errorf("Subscribed %+v", found.Subscribed)
	}
	if len(found.Unsubscribed) != 1 || found.Unsubscribed[0].Value.PublishedFileID != 2003 {
		t.Errorf("Unsubscribed %+v", found.Unsubscribed)
	}
	if len(found.QueriesCompleted) != 1 || found.QueriesCompleted[0].Value.Handle != 55 {
		t.Errorf("QueriesCompleted %+v", found.QueriesCompleted)
	}
	if len(found.EULAStatuses) != 1 || !found.EULAStatuses[0].Value.Accepted || found.EULAStatuses[0].Value.AppID != 5534970 {
		t.Errorf("EULAStatuses %+v", found.EULAStatuses)
	}
	if len(found.FailedCalls) != 0 {
		t.Errorf("FailedCalls %v", found.FailedCalls)
	}
	// Every result was fetched on the pipe that was polled.
	for _, pipe := range *asked {
		if pipe != "7" {
			t.Errorf("fetched on pipe %q", pipe)
		}
	}
}

// A call that failed, a result whose size is absurd and a result too small
// for its kind are each reported as a failed call; a result of a kind this
// binding does not read is left alone.
func TestPollCallbacks_AFailedCallIsReported(t *testing.T) {
	asked := steamAnswers(t, map[SteamAPICall_t][]byte{3: {1, 2}})
	var found Callbacks
	found.take(callCompleted(1, createItemResultCallbackID, 16))          // Steam says it failed
	found.take(callCompleted(2, submitItemUpdateResultCallbackID, 0))     // no result
	found.take(callCompleted(4, submitItemUpdateResultCallbackID, 1<<20)) // far too large
	found.take(callCompleted(3, subscribeResultCallbackID, 2))            // too small for its kind
	found.take(callCompleted(9, 1234, 8))                                 // a kind this binding does not read
	short := callCompleted(5, createItemResultCallbackID, 16)             // a SteamAPICallCompleted_t cut short
	short.ParamSize = 8
	found.take(short)

	if want := []SteamAPICall_t{1, 2, 4, 3}; len(found.FailedCalls) != len(want) {
		t.Errorf("FailedCalls %v, want %v", found.FailedCalls, want)
	} else {
		for i := range want {
			if found.FailedCalls[i] != want[i] {
				t.Errorf("FailedCalls %v, want %v", found.FailedCalls, want)
			}
		}
	}
	if len(*asked) != 2 {
		t.Errorf("Steam was asked %d times, want 2 (a call with no or a too large result is not fetched)", len(*asked))
	}
}

func TestPollCallbacks_ReadsTheWorkshopCallbacks(t *testing.T) {
	installed := structOf(itemInstalledFields, uint32(5534970), uint64(77))
	download := structOf(downloadItemResultFields, uint32(5534970), uint64(78), int32(EResultOK))
	var found Callbacks
	found.take(message(itemInstalledCallbackID, installed))
	found.take(message(downloadItemResultCallbackID, download))
	found.take(message(itemInstalledCallbackID, []byte{1}))      // too small: left out
	found.take(message(downloadItemResultCallbackID, []byte{1})) // too small: left out
	if len(found.ItemsInstalled) != 1 || found.ItemsInstalled[0] != (ItemInstalled{5534970, 77}) {
		t.Errorf("ItemsInstalled %+v", found.ItemsInstalled)
	}
	if len(found.ItemDownloads) != 1 || found.ItemDownloads[0] != (DownloadItemResult{5534970, 78, EResultOK}) {
		t.Errorf("ItemDownloads %+v", found.ItemDownloads)
	}
}

func TestParamBytes_NoPayloadIsNothing(t *testing.T) {
	if b := paramBytes(CallbackMsg{}); b != nil {
		t.Errorf("paramBytes = %v", b)
	}
	b := []byte{9}
	if got := paramBytes(CallbackMsg{Param: unsafe.Pointer(&b[0]), ParamSize: -1}); got != nil {
		t.Errorf("a negative size = %v", got)
	}
}

// The real fetch hands Steam the pipe, the call, the buffer with its size
// and the kind it expects, and gives back what Steam answered.
func TestGetAPICallResult_PassesItsArgumentsThrough(t *testing.T) {
	saved := ptrAPI_ManualDispatch_GetAPICallResult
	t.Cleanup(func() { ptrAPI_ManualDispatch_GetAPICallResult = saved })
	var gotPipe HSteamPipe
	var gotCall SteamAPICall_t
	var gotSize, gotKind int32
	ptrAPI_ManualDispatch_GetAPICallResult = func(pipe HSteamPipe, call SteamAPICall_t, buffer *byte, size, kind int32, failed *bool) bool {
		gotPipe, gotCall, gotSize, gotKind = pipe, call, size, kind
		*buffer = 42
		*failed = true
		return true
	}
	buffer := make([]byte, 16)
	ok, failed := getAPICallResult(3, 99, buffer, createItemResultCallbackID)
	if !ok || !failed || gotPipe != 3 || gotCall != 99 || gotSize != 16 || gotKind != createItemResultCallbackID || buffer[0] != 42 {
		t.Errorf("ok %v failed %v pipe %d call %d size %d kind %d buffer %d", ok, failed, gotPipe, gotCall, gotSize, gotKind, buffer[0])
	}
}
