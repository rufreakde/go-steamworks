// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

package steamworks

import "unsafe"

// This file adds the Session Ticket for the Steamworks Web API: what a
// game's own backend server needs to learn, from Steam and not from the
// client, which Steam account a connection belongs to
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
//
// Both flat functions (SteamAPI_ISteamUser_GetAuthTicketForWebApi,
// SteamAPI_ISteamUser_CancelAuthTicket) were confirmed present via `nm`
// against this package's own vendored libsteam_api.dylib and
// libsteam_api64.so.

// HAuthTicket is the handle of a ticket this process asked for.
type HAuthTicket uint32

// HAuthTicketInvalid is what GetAuthTicketForWebApi returns when no ticket
// could be asked for (k_HAuthTicketInvalid).
const HAuthTicketInvalid HAuthTicket = 0

// EResult is Steam's general result code. Only EResultOK is named here.
type EResult int32

// EResultOK is k_EResultOK.
const EResultOK EResult = 1

// k_iCallbackGetTicketForWebApiResponse is
// GetTicketForWebApiResponse_t's own k_iCallback
// (k_iSteamUserCallbacks (100) + 68).
const k_iCallbackGetTicketForWebApiResponse = 168

// maxTicketForWebAPILength mirrors the fixed size of
// GetTicketForWebApiResponse_t.m_rgubTicket (m_cubTicketMaxLength).
const maxTicketForWebAPILength = 2560

// ticketForWebAPIResponseSize is the size of GetTicketForWebApiResponse_t:
//
//	struct GetTicketForWebApiResponse_t {
//	    HAuthTicket m_hAuthTicket;   // offset 0
//	    EResult     m_eResult;       // offset 4
//	    int         m_cubTicket;     // offset 8
//	    uint8       m_rgubTicket[2560]; // offset 12
//	};
const ticketForWebAPIResponseSize = 12 + maxTicketForWebAPILength

// TicketForWebAPIResponse is one GetTicketForWebApiResponse_t: the answer
// to one GetAuthTicketForWebApi call. Ticket is empty unless Result is
// EResultOK.
type TicketForWebAPIResponse struct {
	Handle HAuthTicket
	Result EResult
	Ticket []byte
}

// decodeTicketForWebAPIResponse reinterprets a callbackMsg's pubParam as
// GetTicketForWebApiResponse_t and copies the ticket out of Steam's own
// memory, which is freed right after the poll.
func decodeTicketForWebAPIResponse(pubParam uintptr) TicketForWebAPIResponse {
	raw := unsafe.Slice((*byte)(unsafe.Pointer(pubParam)), ticketForWebAPIResponseSize)
	response := TicketForWebAPIResponse{
		Handle: HAuthTicket(*(*uint32)(unsafe.Pointer(&raw[0]))),
		Result: EResult(*(*int32)(unsafe.Pointer(&raw[4]))),
	}
	length := int(*(*int32)(unsafe.Pointer(&raw[8])))
	if length < 0 {
		length = 0
	}
	if length > maxTicketForWebAPILength {
		length = maxTicketForWebAPILength
	}
	if response.Result == EResultOK && length > 0 {
		response.Ticket = append([]byte(nil), raw[12:12+length]...)
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
}

// PollCallbacks pumps ManualDispatch's own dispatch frame once, drains
// every callback currently queued, and returns the ones this binding
// knows. Every other callback is freed and discarded unread.
//
// A process that needs both a Join Game request and a Web API ticket must
// call this and never PollGameRichPresenceJoinRequested: that one discards
// a ticket it finds, and the ticket is then lost.
//
// Must only be called after EnableManualDispatch.
func PollCallbacks(pipe HSteamPipe) Callbacks {
	ptrAPI_ManualDispatch_RunFrame(pipe)

	var found Callbacks
	for {
		var msg callbackMsg
		if !ptrAPI_ManualDispatch_GetNextCallback(pipe, uintptr(unsafe.Pointer(&msg))) {
			break
		}
		if msg.pubParam != 0 {
			switch msg.callback {
			case k_iCallbackGameRichPresenceJoinRequested:
				request := decodeGameRichPresenceJoinRequested(msg.pubParam)
				found.JoinRequested = &request
			case k_iCallbackGetTicketForWebApiResponse:
				if int(msg.cubParam) >= ticketForWebAPIResponseSize {
					found.WebAPITickets = append(found.WebAPITickets, decodeTicketForWebAPIResponse(msg.pubParam))
				}
			}
		}
		ptrAPI_ManualDispatch_FreeLastCallback(pipe)
	}
	return found
}
