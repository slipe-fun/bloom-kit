package main

/*
#include <stdlib.h>

typedef void (*chats_callback_t)(const char* json_data);
typedef void (*messages_callback_t)(const char* json_data);

static inline void call_chats_callback(chats_callback_t cb, const char* json_data) {
    if (cb != NULL) {
        cb(json_data);
    }
}

static inline void call_messages_callback(messages_callback_t cb, const char* json_data) {
    if (cb != NULL) {
        cb(json_data);
    }
}
*/

import "C"
import (
	"encoding/json"
	"unsafe"

	"github.com/slipe-fun/bloom-kit/client"
)

var globalClient *client.BloomClient

type ErrorResponse struct {
	Error string `json:"error"`
}

func makeErrorJSON(err error) *C.char {
	resp := ErrorResponse{Error: err.Error()}
	bytes, _ := json.Marshal(resp)
	return C.CString(string(bytes))
}

func makeValueJSON(key, value string) *C.char {
	resp := map[string]string{key: value}
	bytes, _ := json.Marshal(resp)
	return C.CString(string(bytes))
}

//export InitClient
func InitClient(baseURL, wsURL, storagePath *C.char, encKey *C.uchar, keyLen C.int) *C.char {
	goBaseURL := C.GoString(baseURL)
	goWsURL := C.GoString(wsURL)
	goStorage := C.GoString(storagePath)
	goKey := C.GoBytes(unsafe.Pointer(encKey), keyLen)

	globalClient = client.NewClient(goBaseURL, goWsURL, goStorage, goKey)
	if err := globalClient.Initialize(); err != nil {
		return makeErrorJSON(err)
	}
	return nil
}

//export FreeString
func FreeString(p *C.char) {
	if p != nil {
		C.free(unsafe.Pointer(p))
	}
}

//export Register
func Register() *C.char {
	if globalClient == nil {
		return C.CString(`{"error": "client not initialized"}`)
	}
	res, err := globalClient.Register()
	if err != nil {
		return makeErrorJSON(err)
	}
	bytes, err := json.Marshal(res)
	if err != nil {
		return makeErrorJSON(err)
	}
	return C.CString(string(bytes))
}

//export Login
func Login(recoveryKey *C.char) *C.char {
	if globalClient == nil {
		return C.CString(`{"error": "client not initialized"}`)
	}
	goRecoveryKey := C.GoString(recoveryKey)
	res, err := globalClient.Login(goRecoveryKey)
	if err != nil {
		return makeErrorJSON(err)
	}
	bytes, err := json.Marshal(res)
	if err != nil {
		return makeErrorJSON(err)
	}
	return C.CString(string(bytes))
}

//export RestoreSession
func RestoreSession() *C.char {
	if globalClient == nil {
		return C.CString(`{"error": "client not initialized"}`)
	}
	res, err := globalClient.RestoreSession()
	if err != nil {
		return makeErrorJSON(err)
	}
	bytes, err := json.Marshal(res)
	if err != nil {
		return makeErrorJSON(err)
	}
	return C.CString(string(bytes))
}

//export ClearCredentials
func ClearCredentials() {
	if globalClient != nil {
		globalClient.ClearCredentials()
	}
}

//export GetMe
func GetMe() *C.char {
	if globalClient == nil {
		return C.CString(`{"error": "client not initialized"}`)
	}
	res, err := globalClient.GetMe()
	if err != nil {
		return makeErrorJSON(err)
	}
	return C.CString(string(res))
}

//export SearchUsers
func SearchUsers(query *C.char) *C.char {
	if globalClient == nil {
		return C.CString(`{"error": "client not initialized"}`)
	}
	goQuery := C.GoString(query)
	res, err := globalClient.SearchUsers(goQuery)
	if err != nil {
		return makeErrorJSON(err)
	}
	return C.CString(string(res))
}

//export EditUser
func EditUser(username, displayName, description *C.char, hasUsername, hasDisplayName, hasDescription C.int) *C.char {
	if globalClient == nil {
		return C.CString(`{"error": "client not initialized"}`)
	}
	req := &client.EditRequest{
		Username:       C.GoString(username),
		HasUsername:    hasUsername != 0,
		DisplayName:    C.GoString(displayName),
		HasDisplayName: hasDisplayName != 0,
		Description:    C.GoString(description),
		HasDescription: hasDescription != 0,
	}
	res, err := globalClient.EditUser(req)
	if err != nil {
		return makeErrorJSON(err)
	}
	return C.CString(string(res))
}

//export GetUser
func GetUser(userID *C.char) *C.char {
	if globalClient == nil {
		return C.CString(`{"error": "client not initialized"}`)
	}
	goUserID := C.GoString(userID)
	res, err := globalClient.GetUser(goUserID)
	if err != nil {
		return makeErrorJSON(err)
	}
	return C.CString(string(res))
}

//export CreateChat
func CreateChat(userID, mlKem768Key, x448Key, ed448Key *C.char) *C.char {
	if globalClient == nil {
		return C.CString(`{"error": "client not initialized"}`)
	}
	req := &client.CreateChatRequest{
		UserID:            C.GoString(userID),
		MlKem768PublicKey: C.GoString(mlKem768Key),
		X448PublicKey:     C.GoString(x448Key),
		Ed448PublicKey:    C.GoString(ed448Key),
	}
	res, err := globalClient.CreateChat(req)
	if err != nil {
		return makeErrorJSON(err)
	}
	return C.CString(string(res))
}

//export GetChats
func GetChats() *C.char {
	if globalClient == nil {
		return C.CString(`{"error": "client not initialized"}`)
	}
	res, err := globalClient.GetChats()
	if err != nil {
		return makeErrorJSON(err)
	}
	return C.CString(string(res))
}

//export GetLocalChats
func GetLocalChats() *C.char {
	if globalClient == nil {
		return C.CString(`{"error": "client not initialized"}`)
	}
	res, err := globalClient.GetLocalChats()
	if err != nil {
		return makeErrorJSON(err)
	}
	return C.CString(string(res))
}

//export SendMessage
func SendMessage(chatID C.int, replyToID C.int, content *C.char) *C.char {
	if globalClient == nil {
		return C.CString(`{"error": "client not initialized"}`)
	}
	goContent := C.GoString(content)

	var replyPtr *int
	if replyToID >= 0 {
		rID := int(replyToID)
		replyPtr = &rID
	}

	res, err := globalClient.SendMessage(int(chatID), replyPtr, goContent)
	if err != nil {
		return makeErrorJSON(err)
	}
	return C.CString(string(res))
}

//export LoadMessages
func LoadMessages(chatID C.int, beforeID C.int) *C.char {
	if globalClient == nil {
		return C.CString(`{"error": "client not initialized"}`)
	}
	res, err := globalClient.LoadMessages(int(chatID), int(beforeID))
	if err != nil {
		return makeErrorJSON(err)
	}
	return C.CString(string(res))
}

//export StartExchangeSession
func StartExchangeSession(exchangeType *C.char) *C.char {
	if globalClient == nil {
		return C.CString(`{"error": "client not initialized"}`)
	}
	goExchangeType := C.GoString(exchangeType)
	roomID, fingerprint, err := globalClient.StartExchangeSession(goExchangeType)
	if err != nil {
		return makeErrorJSON(err)
	}

	resp := map[string]string{
		"room_id":     roomID,
		"fingerprint": fingerprint,
	}
	bytes, _ := json.Marshal(resp)
	return C.CString(string(bytes))
}

//export Exchange
func Exchange(exchangeType, roomID, fingerprint *C.char) *C.char {
	if globalClient == nil {
		return C.CString(`{"error": "client not initialized"}`)
	}
	goExchangeType := C.GoString(exchangeType)
	goRoomID := C.GoString(roomID)
	goFingerprint := C.GoString(fingerprint)

	recoveryKey, err := globalClient.Exchange(goExchangeType, goRoomID, goFingerprint)
	if err != nil {
		return makeErrorJSON(err)
	}
	return makeValueJSON("recovery_key", recoveryKey)
}

//export CancelExchange
func CancelExchange() {
	if globalClient != nil {
		globalClient.CancelExchange()
	}
}

type cgoChatsListener struct {
	callback C.chats_callback_t
}

func (l *cgoChatsListener) OnChatsUpdated(chatsJSON []byte) {
	cStr := C.CString(string(chatsJSON))
	defer C.free(unsafe.Pointer(cStr))
	C.call_chats_callback(l.callback, cStr)
}

//export RegisterChatsCallback
func RegisterChatsCallback(cb C.chats_callback_t) {
	if globalClient != nil {
		globalClient.RegisterChatsListener(&cgoChatsListener{callback: cb})
	}
}

//export UnregisterChatsCallback
func UnregisterChatsCallback() {
	if globalClient != nil {
		globalClient.UnregisterChatsListener()
	}
}

type cgoMessagesListener struct {
	callback C.messages_callback_t
}

func (l *cgoMessagesListener) OnNewMessage(messageJSON []byte) {
	cStr := C.CString(string(messageJSON))
	defer C.free(unsafe.Pointer(cStr))
	C.call_messages_callback(l.callback, cStr)
}

//export RegisterMessagesCallback
func RegisterMessagesCallback(cb C.messages_callback_t) {
	if globalClient != nil {
		globalClient.RegisterMessagesListener(&cgoMessagesListener{callback: cb})
	}
}

//export UnregisterMessagesCallback
func UnregisterMessagesCallback() {
	if globalClient != nil {
		globalClient.UnregisterMessagesListener()
	}
}

func main() {}
