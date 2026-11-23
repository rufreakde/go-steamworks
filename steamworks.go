// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021 The go-steamworks Authors

//go:generate go run gen.go

package steamworks

import (
	"structs"
	"unsafe"
)

type AppId_t uint32
type CSteamID uint64
type HAuthTicket uint32
type InputHandle_t uint64
type HSteamPipe int32
type HSteamUser int32

const HAuthTicketInvalid HAuthTicket = 0

type EResult int32

const (
	EResultNone                  EResult = 0
	EResultOK                    EResult = 1
	EResultFail                  EResult = 2
	EResultNoConnection          EResult = 3
	EResultInvalidPassword       EResult = 5
	EResultInvalidProtocolVer    EResult = 7
	EResultInvalidParam          EResult = 8
	EResultBusy                  EResult = 10
	EResultInvalidState          EResult = 11
	EResultTimeout               EResult = 16
	EResultNotLoggedOn           EResult = 21
	EResultPending               EResult = 22
	EResultInsufficientPrivilege EResult = 24
)

type ESteamAPIInitResult int32

const (
	ESteamAPIInitResult_OK              ESteamAPIInitResult = 0
	ESteamAPIInitResult_FailedGeneric   ESteamAPIInitResult = 1
	ESteamAPIInitResult_NoSteamClient   ESteamAPIInitResult = 2
	ESteamAPIInitResult_VersionMismatch ESteamAPIInitResult = 3
)

type ESteamInputType int32

const (
	ESteamInputType_Unknown                  ESteamInputType = 0
	ESteamInputType_SteamController          ESteamInputType = 1
	ESteamInputType_XBox360Controller        ESteamInputType = 2
	ESteamInputType_XBoxOneController        ESteamInputType = 3
	ESteamInputType_GenericGamepad           ESteamInputType = 4
	ESteamInputType_PS4Controller            ESteamInputType = 5
	ESteamInputType_AppleMFiController       ESteamInputType = 6 // Unused
	ESteamInputType_AndroidController        ESteamInputType = 7 // Unused
	ESteamInputType_SwitchJoyConPair         ESteamInputType = 8 // Unused
	ESteamInputType_SwitchJoyConSingle       ESteamInputType = 9 // Unused
	ESteamInputType_SwitchProController      ESteamInputType = 10
	ESteamInputType_MobileTouch              ESteamInputType = 11
	ESteamInputType_PS3Controller            ESteamInputType = 12
	ESteamInputType_PS5Controller            ESteamInputType = 13
	ESteamInputType_SteamDeckController      ESteamInputType = 14
	ESteamInputType_SteamOSHandheld          ESteamInputType = 15
	ESteamInputType_Switch2ProController     ESteamInputType = 16
	ESteamInputType_SteamController2026      ESteamInputType = 17
	ESteamInputType_SteamFrameControllerPair ESteamInputType = 18
	ESteamInputType_Count                    ESteamInputType = 19
	ESteamInputType_MaximumPossibleValue     ESteamInputType = 255
)

// Deprecated: Use ESteamInputType_GenericGamepad.
const ESteamInputType_GenericXInput = ESteamInputType_GenericGamepad

const (
	_STEAM_INPUT_MAX_COUNT = 16
)

type EFloatingGamepadTextInputMode int32

const (
	EFloatingGamepadTextInputMode_ModeSingleLine    EFloatingGamepadTextInputMode = 0
	EFloatingGamepadTextInputMode_ModeMultipleLines EFloatingGamepadTextInputMode = 1
	EFloatingGamepadTextInputMode_ModeEmail         EFloatingGamepadTextInputMode = 2
	EFloatingGamepadTextInputMode_ModeNumeric       EFloatingGamepadTextInputMode = 3
)

type ISteamApps interface {
	BGetDLCDataByIndex(iDLC int) (appID AppId_t, available bool, pchName string, success bool)
	BIsDlcInstalled(appID AppId_t) bool
	GetAppInstallDir(appID AppId_t) string
	GetCurrentGameLanguage() string
	GetDLCCount() int32
	// GetLaunchCommandLine reads the command line Steam appended when it
	// launched this process via a Steam URL — the real, documented
	// delivery mechanism for a Rich Presence "connect" value when a
	// friend's Join Game click launches a not-already-running game
	// (Valve's own ISteamApps doc, partner.steamgames.com/doc/api/
	// ISteamApps). Real, synchronous, flat-API call — same shape as
	// GetAppInstallDir above, added for modwars-workspace issue #178.
	GetLaunchCommandLine() string
}

type ISteamInput interface {
	GetConnectedControllers() []InputHandle_t
	// GetControllerForGamepadIndex is the handle behind an emulated gamepad
	// slot, 0 for none (api.go). Added for modwars-workspace issue #676.
	GetControllerForGamepadIndex(index int) InputHandle_t
	GetInputTypeForHandle(inputHandle InputHandle_t) ESteamInputType
	Init(bExplicitlyCallRunFrame bool) bool
	RunFrame()
}

type ISteamRemoteStorage interface {
	FileWrite(file string, data []byte) bool
	FileRead(file string, data []byte) int32
	FileDelete(file string) bool
	GetFileSize(file string) int32
}

type ISteamUser interface {
	GetSteamID() CSteamID
	// GetAuthTicketForWebApi asks Steam for a ticket the named backend
	// service can check with the Steamworks Web API. The ticket itself
	// arrives later, as a callback (PollCallbacks). See webapiticket.go.
	GetAuthTicketForWebApi(identity string) HAuthTicket
	// CancelAuthTicket gives a ticket back when it is needed no more.
	CancelAuthTicket(ticket HAuthTicket)
}

type GetTicketForWebApiResponse struct {
	_         structs.HostLayout
	Handle    HAuthTicket
	Result    EResult
	TicketLen int32
	Ticket    [2560]byte
}

// GetTicketForWebApiResponseCallbackID is the k_iCallback of
// GetTicketForWebApiResponse_t (k_iSteamUserCallbacks + 68, isteamuser.h).
const GetTicketForWebApiResponseCallbackID = 168

// CallbackMsg mirrors the native CallbackMsg_t returned by
// ManualDispatch_GetNextCallback. Param points to the callback payload and is
// only valid until ManualDispatch_FreeLastCallback is called.
type CallbackMsg struct {
	_          structs.HostLayout
	SteamUser  HSteamUser
	CallbackID int32
	Param      unsafe.Pointer
	ParamSize  int32
}

type ISteamUserStats interface {
	GetAchievement(name string) (achieved, success bool)
	SetAchievement(name string) bool
	ClearAchievement(name string) bool
	StoreStats() bool
}

type ISteamUtils interface {
	IsOverlayEnabled() bool
	// GetAppID is the App ID the game runs under. Added for
	// modwars-workspace issue #798 (a Workshop item is made for it).
	GetAppID() AppId_t
	IsRunningOnSteamHardware() ESteamHardwareType
	GetSteamHardwareDefaultConfig() ESteamHardwareDefaultConfig
	IsRunningUnderProton() bool
	// Deprecated: Use IsRunningOnSteamHardware.
	IsSteamRunningOnSteamDeck() bool
	ShowFloatingGamepadTextInput(keyboardMode EFloatingGamepadTextInputMode, textFieldXPosition, textFieldYPosition, textFieldWidth, textFieldHeight int32) bool
}

type ESteamHardwareType int32

const (
	ESteamHardwareType_None         ESteamHardwareType = 0
	ESteamHardwareType_SteamDeck    ESteamHardwareType = 1
	ESteamHardwareType_SteamMachine ESteamHardwareType = 2
	ESteamHardwareType_SteamFrame   ESteamHardwareType = 3
)

type ESteamHardwareDefaultConfig int32

const (
	ESteamHardwareDefaultConfig_None         ESteamHardwareDefaultConfig = 0
	ESteamHardwareDefaultConfig_Low          ESteamHardwareDefaultConfig = 1
	ESteamHardwareDefaultConfig_Medium       ESteamHardwareDefaultConfig = 2
	ESteamHardwareDefaultConfig_High         ESteamHardwareDefaultConfig = 3
	ESteamHardwareDefaultConfig_Max          ESteamHardwareDefaultConfig = 4
	ESteamHardwareDefaultConfig_SteamDeck    ESteamHardwareDefaultConfig = 5
	ESteamHardwareDefaultConfig_SteamMachine ESteamHardwareDefaultConfig = 6
	ESteamHardwareDefaultConfig_SteamFrame   ESteamHardwareDefaultConfig = 7
)

type ISteamFriends interface {
	GetPersonaName() string
	SetRichPresence(string, string) bool
	// ActivateGameOverlayInviteDialogConnectString opens the invite dialog
	// of the Steam overlay. Each friend the player picks there gets an
	// invite that carries connect, the same string a Join Game request
	// gives (Rich Presence "connect"). It does nothing when the overlay is
	// not there. Added for modwars-workspace issue #794.
	ActivateGameOverlayInviteDialogConnectString(connect string)
	// ActivateGameOverlayToWebPage opens url in the Steam overlay's web
	// browser; a steam:// address opens Steam's own page, for example a
	// Workshop item's (steam://url/CommunityFilePage/<item>). Added for
	// modwars-workspace issue #798.
	ActivateGameOverlayToWebPage(url string, mode EActivateGameOverlayToWebPageMode)
}

const (
	flatAPI_RestartAppIfNecessary           = "SteamAPI_RestartAppIfNecessary"
	flatAPI_InitFlat                        = "SteamAPI_InitFlat"
	flatAPI_RunCallbacks                    = "SteamAPI_RunCallbacks"
	flatAPI_GetHSteamPipe                   = "SteamAPI_GetHSteamPipe"
	flatAPI_ManualDispatch_Init             = "SteamAPI_ManualDispatch_Init"
	flatAPI_ManualDispatch_RunFrame         = "SteamAPI_ManualDispatch_RunFrame"
	flatAPI_ManualDispatch_GetNextCallback  = "SteamAPI_ManualDispatch_GetNextCallback"
	flatAPI_ManualDispatch_FreeLastCallback = "SteamAPI_ManualDispatch_FreeLastCallback"
	flatAPI_ManualDispatch_GetAPICallResult = "SteamAPI_ManualDispatch_GetAPICallResult"

	flatAPI_SteamApps                         = "SteamAPI_SteamApps_v009"
	flatAPI_ISteamApps_BGetDLCDataByIndex     = "SteamAPI_ISteamApps_BGetDLCDataByIndex"
	flatAPI_ISteamApps_BIsDlcInstalled        = "SteamAPI_ISteamApps_BIsDlcInstalled"
	flatAPI_ISteamApps_GetAppInstallDir       = "SteamAPI_ISteamApps_GetAppInstallDir"
	flatAPI_ISteamApps_GetCurrentGameLanguage = "SteamAPI_ISteamApps_GetCurrentGameLanguage"
	flatAPI_ISteamApps_GetDLCCount            = "SteamAPI_ISteamApps_GetDLCCount"
	flatAPI_ISteamApps_GetLaunchCommandLine   = "SteamAPI_ISteamApps_GetLaunchCommandLine"

	flatAPI_SteamFriends                                               = "SteamAPI_SteamFriends_v018"
	flatAPI_ISteamFriends_GetPersonaName                               = "SteamAPI_ISteamFriends_GetPersonaName"
	flatAPI_ISteamFriends_SetRichPresence                              = "SteamAPI_ISteamFriends_SetRichPresence"
	flatAPI_ISteamFriends_ActivateGameOverlayInviteDialogConnectString = "SteamAPI_ISteamFriends_ActivateGameOverlayInviteDialogConnectString"
	flatAPI_ISteamFriends_ActivateGameOverlayToWebPage                 = "SteamAPI_ISteamFriends_ActivateGameOverlayToWebPage"

	flatAPI_SteamInput                               = "SteamAPI_SteamInput_v007"
	flatAPI_ISteamInput_GetConnectedControllers      = "SteamAPI_ISteamInput_GetConnectedControllers"
	flatAPI_ISteamInput_GetControllerForGamepadIndex = "SteamAPI_ISteamInput_GetControllerForGamepadIndex"
	flatAPI_ISteamInput_GetInputTypeForHandle        = "SteamAPI_ISteamInput_GetInputTypeForHandle"
	flatAPI_ISteamInput_Init                         = "SteamAPI_ISteamInput_Init"
	flatAPI_ISteamInput_RunFrame                     = "SteamAPI_ISteamInput_RunFrame"

	flatAPI_SteamRemoteStorage              = "SteamAPI_SteamRemoteStorage_v016"
	flatAPI_ISteamRemoteStorage_FileWrite   = "SteamAPI_ISteamRemoteStorage_FileWrite"
	flatAPI_ISteamRemoteStorage_FileRead    = "SteamAPI_ISteamRemoteStorage_FileRead"
	flatAPI_ISteamRemoteStorage_FileDelete  = "SteamAPI_ISteamRemoteStorage_FileDelete"
	flatAPI_ISteamRemoteStorage_GetFileSize = "SteamAPI_ISteamRemoteStorage_GetFileSize"

	flatAPI_SteamUser                         = "SteamAPI_SteamUser_v023"
	flatAPI_ISteamUser_GetSteamID             = "SteamAPI_ISteamUser_GetSteamID"
	flatAPI_ISteamUser_GetAuthTicketForWebApi = "SteamAPI_ISteamUser_GetAuthTicketForWebApi"
	flatAPI_ISteamUser_CancelAuthTicket       = "SteamAPI_ISteamUser_CancelAuthTicket"

	flatAPI_SteamUserStats                   = "SteamAPI_SteamUserStats_v013"
	flatAPI_ISteamUserStats_GetAchievement   = "SteamAPI_ISteamUserStats_GetAchievement"
	flatAPI_ISteamUserStats_SetAchievement   = "SteamAPI_ISteamUserStats_SetAchievement"
	flatAPI_ISteamUserStats_ClearAchievement = "SteamAPI_ISteamUserStats_ClearAchievement"
	flatAPI_ISteamUserStats_StoreStats       = "SteamAPI_ISteamUserStats_StoreStats"

	flatAPI_SteamUtils                                = "SteamAPI_SteamUtils_v011"
	flatAPI_ISteamUtils_IsOverlayEnabled              = "SteamAPI_ISteamUtils_IsOverlayEnabled"
	flatAPI_ISteamUtils_GetAppID                      = "SteamAPI_ISteamUtils_GetAppID"
	flatAPI_ISteamUtils_IsRunningOnSteamHardware      = "SteamAPI_ISteamUtils_IsRunningOnSteamHardware"
	flatAPI_ISteamUtils_GetSteamHardwareDefaultConfig = "SteamAPI_ISteamUtils_GetSteamHardwareDefaultConfig"
	flatAPI_ISteamUtils_IsRunningUnderProton          = "SteamAPI_ISteamUtils_IsRunningUnderProton"
	flatAPI_ISteamUtils_ShowFloatingGamepadTextInput  = "SteamAPI_ISteamUtils_ShowFloatingGamepadTextInput"

	flatAPI_SteamUGC                               = "SteamAPI_SteamUGC_v021"
	flatAPI_ISteamUGC_CreateItem                   = "SteamAPI_ISteamUGC_CreateItem"
	flatAPI_ISteamUGC_StartItemUpdate              = "SteamAPI_ISteamUGC_StartItemUpdate"
	flatAPI_ISteamUGC_SetItemTitle                 = "SteamAPI_ISteamUGC_SetItemTitle"
	flatAPI_ISteamUGC_SetItemDescription           = "SteamAPI_ISteamUGC_SetItemDescription"
	flatAPI_ISteamUGC_SetItemVisibility            = "SteamAPI_ISteamUGC_SetItemVisibility"
	flatAPI_ISteamUGC_SetItemContent               = "SteamAPI_ISteamUGC_SetItemContent"
	flatAPI_ISteamUGC_SetItemPreview               = "SteamAPI_ISteamUGC_SetItemPreview"
	flatAPI_ISteamUGC_SetItemTags                  = "SteamAPI_ISteamUGC_SetItemTags"
	flatAPI_ISteamUGC_SubmitItemUpdate             = "SteamAPI_ISteamUGC_SubmitItemUpdate"
	flatAPI_ISteamUGC_GetItemUpdateProgress        = "SteamAPI_ISteamUGC_GetItemUpdateProgress"
	flatAPI_ISteamUGC_SubscribeItem                = "SteamAPI_ISteamUGC_SubscribeItem"
	flatAPI_ISteamUGC_UnsubscribeItem              = "SteamAPI_ISteamUGC_UnsubscribeItem"
	flatAPI_ISteamUGC_GetNumSubscribedItems        = "SteamAPI_ISteamUGC_GetNumSubscribedItems"
	flatAPI_ISteamUGC_GetSubscribedItems           = "SteamAPI_ISteamUGC_GetSubscribedItems"
	flatAPI_ISteamUGC_GetItemState                 = "SteamAPI_ISteamUGC_GetItemState"
	flatAPI_ISteamUGC_GetItemInstallInfo           = "SteamAPI_ISteamUGC_GetItemInstallInfo"
	flatAPI_ISteamUGC_GetItemDownloadInfo          = "SteamAPI_ISteamUGC_GetItemDownloadInfo"
	flatAPI_ISteamUGC_DownloadItem                 = "SteamAPI_ISteamUGC_DownloadItem"
	flatAPI_ISteamUGC_CreateQueryUGCDetailsRequest = "SteamAPI_ISteamUGC_CreateQueryUGCDetailsRequest"
	flatAPI_ISteamUGC_SetReturnLongDescription     = "SteamAPI_ISteamUGC_SetReturnLongDescription"
	flatAPI_ISteamUGC_SendQueryUGCRequest          = "SteamAPI_ISteamUGC_SendQueryUGCRequest"
	flatAPI_ISteamUGC_GetQueryUGCResult            = "SteamAPI_ISteamUGC_GetQueryUGCResult"
	flatAPI_ISteamUGC_ReleaseQueryUGCRequest       = "SteamAPI_ISteamUGC_ReleaseQueryUGCRequest"
	flatAPI_ISteamUGC_GetWorkshopEULAStatus        = "SteamAPI_ISteamUGC_GetWorkshopEULAStatus"
	flatAPI_ISteamUGC_ShowWorkshopEULA             = "SteamAPI_ISteamUGC_ShowWorkshopEULA"
)

type steamErrMsg [1024]byte

func (s *steamErrMsg) String() string {
	for i, b := range s {
		if b == 0 {
			return string(s[:i])
		}
	}
	return ""
}
