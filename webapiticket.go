// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

package steamworks

import "unsafe"

// This file adds the reading of the Session Ticket for the Steamworks Web
// API: what a game's own backend server needs to learn, from Steam and not
// from the client, which Steam account a connection belongs to
// (https://partner.steamgames.com/doc/features/auth, "Backend Server").
//
//  1. The client calls ISteamUser.GetAuthTicketForWebApi(identity).
//  2. It waits for the GetTicketForWebApiResponse_t callback (PollCallbacks).
//  3. It sends the ticket bytes to the backend, which asks
//     ISteamUserAuth/AuthenticateUserTicket with its publisher key.
//  4. It calls ISteamUser.CancelAuthTicket when it needs the ticket no more.
//
// GetAuthSessionTicket, the older call, is the wrong one for this: Valve's
// own documentation says its tickets cannot be used with the Web API.

// TicketForWebAPIResponse is one GetTicketForWebApiResponse_t: the answer
// to one GetAuthTicketForWebApi call, copied out of Steam's memory. Ticket
// is empty unless Result is EResultOK.
type TicketForWebAPIResponse struct {
	Handle HAuthTicket
	Result EResult
	Ticket []byte
}

// decodeTicketForWebAPIResponse reads a callback's Param as
// GetTicketForWebApiResponse_t and copies the ticket out of Steam's own
// memory, which is freed right after the poll.
func decodeTicketForWebAPIResponse(param unsafe.Pointer) TicketForWebAPIResponse {
	native := (*GetTicketForWebApiResponse)(param)
	response := TicketForWebAPIResponse{
		Handle: native.Handle,
		Result: native.Result,
	}
	length := int(native.TicketLen)
	if length < 0 {
		length = 0
	}
	if length > len(native.Ticket) {
		length = len(native.Ticket)
	}
	if response.Result == EResultOK && length > 0 {
		response.Ticket = append([]byte(nil), native.Ticket[:length]...)
	}
	return response
}

// Callbacks is what one PollCallbacks found.
type Callbacks struct {
	// JoinRequested is the most recent GameRichPresenceJoinRequested_t,
	// nil when there was none.
	JoinRequested *GameRichPresenceJoinRequested
	// WebAPITickets are the GetTicketForWebApiResponse_t callbacks, in
	// the order Steam gave them.
	WebAPITickets []TicketForWebAPIResponse

	// The Workshop (callresult.go): the results of its calls, in the order
	// Steam gave them, and its two callbacks.
	CreatedItems     []CallResult[CreateItemResult]
	SubmittedUpdates []CallResult[SubmitItemUpdateResult]
	Subscribed       []CallResult[SubscribeResult]
	Unsubscribed     []CallResult[SubscribeResult]
	QueriesCompleted []CallResult[QueryCompleted]
	EULAStatuses     []CallResult[WorkshopEULAStatus]
	// FailedCalls are calls of those kinds that Steam says failed, or whose
	// result could not be read.
	FailedCalls    []SteamAPICall_t
	ItemsInstalled []ItemInstalled
	ItemDownloads  []DownloadItemResult

	// pipe is the pipe PollCallbacks reads, for the call results.
	pipe HSteamPipe
}

// PollCallbacks pumps ManualDispatch's own dispatch frame once, drains
// every callback currently queued, and returns the ones this binding
// knows. Every other callback is freed and discarded unread.
//
// Must only be called after EnableManualDispatch.
func PollCallbacks(pipe HSteamPipe) Callbacks {
	ManualDispatch_RunFrame(pipe)

	found := Callbacks{pipe: pipe}
	for {
		var msg CallbackMsg
		if !ManualDispatch_GetNextCallback(pipe, &msg) {
			break
		}
		found.take(msg)
		ManualDispatch_FreeLastCallback(pipe)
	}
	return found
}

// take reads one callback into c when this binding knows its kind and its
// payload is as large as that kind needs. msg.Param is Steam's memory: it
// is valid only until the callback is freed, so take copies what it keeps.
func (c *Callbacks) take(msg CallbackMsg) {
	if msg.Param == nil || c.takeWorkshop(msg) {
		return
	}
	switch msg.CallbackID {
	case gameRichPresenceJoinRequestedCallbackID:
		if request, ok := decodeGameRichPresenceJoinRequested(msg.Param, int(msg.ParamSize)); ok {
			c.JoinRequested = &request
		}
	case GetTicketForWebApiResponseCallbackID:
		if uintptr(msg.ParamSize) < unsafe.Sizeof(GetTicketForWebApiResponse{}) {
			return
		}
		c.WebAPITickets = append(c.WebAPITickets, decodeTicketForWebAPIResponse(msg.Param))
	}
}
