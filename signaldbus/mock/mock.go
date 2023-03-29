package mock

import (
	"github.com/godbus/dbus/v5"
	"testing"
)

// for testing methods
type Dbus_api struct {
	About                string
	Attachments          []string
	Avatar               string
	Block                bool
	Blocked              bool
	Captcha              string
	Challenge            string
	Emoji                string
	Expiration           int32
	FamilyName           string
	GivenName            string
	GroupId              []byte
	GroupName            string
	InviteURI            string
	Members              []string
	Message              string
	Name                 string
	Number_self          string
	Numbers              []string
	Pin                  string
	Recipients           []string
	Result               bool
	Results              []bool
	Remove               bool
	StickerPackPath      string
	Stop                 bool
	TargetAuthor         string
	TargetSentTimestamps []int64
	Timestamp            int64
	Version_             string
	Url                  string
}

var Args []any

func (s *Dbus_api) GetContactName(number string) (string, *dbus.Error) {
	Args = []any{
		number,
	}
	return s.Name, nil
}
func (s *Dbus_api) GetContactNumber(name string) ([]string, *dbus.Error) {
	Args = []any{
		name,
	}
	return s.Numbers, nil
}
func (s *Dbus_api) GetSelfNumber() (string, *dbus.Error) {
	Args = []any{}
	return s.Number_self, nil
}
func (s *Dbus_api) IsContactBlocked(number string) (bool, *dbus.Error) {
	Args = []any{
		number,
	}
	return s.Blocked, nil
}
func (s *Dbus_api) IsRegistered() (bool, *dbus.Error) {
	Args = []any{}
	return s.Result, nil
}

func (s *Dbus_api) IsRegistered_num(number string) (bool, *dbus.Error) {
	Args = []any{
		number,
	}
	return s.Result, nil
}

func (s *Dbus_api) IsRegistered_nums(numbers []string) ([]bool, *dbus.Error) {
	Args = []any{
		numbers,
	}
	return s.Results, nil
}

func (s *Dbus_api) ListNumbers() ([]string, *dbus.Error) {
	Args = []any{}
	return s.Numbers, nil
}
func (s *Dbus_api) RemovePin() *dbus.Error {
	Args = []any{}
	return nil
}
func (s *Dbus_api) SendEndSessionMessage(recipients []string) *dbus.Error {
	Args = []any{
		recipients,
	}
	return nil
}
func (s *Dbus_api) SendMessage(message string, attachments []string, recipient string) (int64, *dbus.Error) {
	Args = []any{
		message,
		attachments,
		recipient,
	}
	return s.Timestamp, nil
}

func (s *Dbus_api) SendMessage_multi(message string, attachments []string, recipients []string) (int64, *dbus.Error) {
	Args = []any{
		message,
		attachments,
		recipients,
	}
	return s.Timestamp, nil
}

func (s *Dbus_api) SendMessageReaction(emoji string, remove bool, targetAuthor string, targetSentTimestamp int64, recipient string) (int64, *dbus.Error) {
	Args = []any{
		emoji,
		remove,
		targetAuthor,
		targetSentTimestamp,
		recipient,
	}
	return s.Timestamp, nil
}

func (s *Dbus_api) SendMessageReaction_multi(emoji string, remove bool, targetAuthor string, targetSentTimestamp int64, recipients []string) (int64, *dbus.Error) {
	Args = []any{
		emoji,
		remove,
		targetAuthor,
		targetSentTimestamp,
		recipients,
	}
	return s.Timestamp, nil
}

// func (s *dbus_api) SendPaymentNotification( receipt []byte, note string, recipient string,) (int64, *dbus.Error) {
// 	args = []any{
// 		receipt,
// 		note,
// 		recipient,
// 	}
// 	return s.timestamp, nil
// }

func (s *Dbus_api) SendNoteToSelfMessage(message string, attachments []string) (int64, *dbus.Error) {
	Args = []any{
		message,
		attachments,
	}
	return s.Timestamp, nil
}
func (s *Dbus_api) SendReadReceipt(recipient string, targetSentTimestamps []int64) *dbus.Error {
	Args = []any{
		recipient,
		targetSentTimestamps,
	}
	return nil
}
func (s *Dbus_api) SendViewedReceipt(recipient string, targetSentTimestamp []int64) *dbus.Error {
	Args = []any{
		recipient,
		targetSentTimestamp,
	}
	return nil
}
func (s *Dbus_api) SendRemoteDeleteMessage(targetSentTimestamp int64, recipient string) (int64, *dbus.Error) {
	Args = []any{
		targetSentTimestamp,
		recipient,
	}
	return s.Timestamp, nil
}

func (s *Dbus_api) SendRemoteDeleteMessage_multi(targetSentTimestamp int64, recipients []string) (int64, *dbus.Error) {
	Args = []any{
		targetSentTimestamp,
		recipients,
	}
	return s.Timestamp, nil
}

func (s *Dbus_api) SendTyping(recipient string, stop bool) *dbus.Error {
	Args = []any{
		recipient,
		stop,
	}
	return nil
}
func (s *Dbus_api) SetContactBlocked(number string, block bool) *dbus.Error {
	Args = []any{
		number,
		block,
	}
	return nil
}
func (s *Dbus_api) SetContactName(number string, name string) *dbus.Error {
	Args = []any{
		number,
		name,
	}
	return nil
}
func (s *Dbus_api) DeleteContact(number string) *dbus.Error {
	Args = []any{
		number,
	}
	return nil
}
func (s *Dbus_api) DeleteRecipient(number string) *dbus.Error {
	Args = []any{
		number,
	}
	return nil
}
func (s *Dbus_api) SetExpirationTimer(number string, expiration int32) *dbus.Error {
	Args = []any{
		number,
		expiration,
	}
	return nil
}
func (s *Dbus_api) SetPin(pin string) *dbus.Error {
	Args = []any{
		pin,
	}
	return nil
}
func (s *Dbus_api) SubmitRateLimitChallenge(challenge string, captcha string) *dbus.Error {
	Args = []any{
		challenge,
		captcha,
	}
	return nil
}
func (s *Dbus_api) UpdateProfile(name string, about string, aboutEmoji string, avatar string, remove bool) *dbus.Error {
	Args = []any{
		name,
		about,
		aboutEmoji,
		avatar,
		remove,
	}
	return nil
}

func (s *Dbus_api) UpdateProfile_firstLastName(givenName string, familyName string, about string, aboutEmoji string, avatar string, remove bool) *dbus.Error {
	Args = []any{
		givenName,
		familyName,
		about,
		aboutEmoji,
		avatar,
		remove,
	}
	return nil
}

func (s *Dbus_api) UploadStickerPack(stickerPackPath string) (string, *dbus.Error) {
	Args = []any{
		stickerPackPath,
	}
	return s.Url, nil
}
func (s *Dbus_api) Version() (string, *dbus.Error) {
	Args = []any{}
	return s.Version_, nil
}
func (s *Dbus_api) CreateGroup(groupName string, members []string, avatar string) ([]byte, *dbus.Error) {
	Args = []any{
		groupName,
		members,
		avatar,
	}
	return s.GroupId, nil
}

// func (s *dbus_api) GetGroup( groupId []byte,) (, *dbus.Error) {
//		args = []any{
//		groupId,
//		}
//		return s.objectPath, nil
//	}

func (s *Dbus_api) GetGroupName(groupId []byte) (string, *dbus.Error) {
	Args = []any{
		groupId,
	}
	return s.Name, nil
}
func (s *Dbus_api) GetGroupMembers(groupId []byte) ([]string, *dbus.Error) {
	Args = []any{
		groupId,
	}
	return s.Members, nil
}
func (s *Dbus_api) JoinGroup(inviteURI string) *dbus.Error {
	Args = []any{
		inviteURI,
	}
	return nil
}
func (s *Dbus_api) SendGroupMessage(message string, attachments []string, groupId []byte) (int64, *dbus.Error) {
	Args = []any{
		message,
		attachments,
		groupId,
	}
	return s.Timestamp, nil
}
func (s *Dbus_api) SendGroupTyping(groupId []byte, stop bool) *dbus.Error {
	Args = []any{
		groupId,
		stop,
	}
	return nil
}
func (s *Dbus_api) SendGroupMessageReaction(emoji string, remove bool, targetAuthor string, targetSentTimestamp int64, groupId []byte) (int64, *dbus.Error) {
	Args = []any{
		emoji,
		remove,
		targetAuthor,
		targetSentTimestamp,
		groupId,
	}
	return s.Timestamp, nil
}
func (s *Dbus_api) SendGroupRemoteDeleteMessage(targetSentTimestamp int64, groupId []byte) (int64, *dbus.Error) {
	Args = []any{
		targetSentTimestamp,
		groupId,
	}
	return s.Timestamp, nil
}

var func_map = map[string]string{
	"GetContactName":            "getContactName",
	"GetContactNumber":          "getContactNumber",
	"GetSelfNumber":             "getSelfNumber",
	"IsContactBlocked":          "isContactBlocked",
	"IsRegistered":              "isRegistered",
	"IsRegistered_num":          "isRegistered",
	"IsRegistered_nums":         "isRegistered",
	"ListNumbers":               "listNumbers",
	"RemovePin":                 "removePin",
	"SendEndSessionMessage":     "sendEndSessionMessage",
	"SendMessage":               "sendMessage",
	"SendMessage_multi":         "sendMessage",
	"SendMessageReaction":       "sendMessageReaction",
	"SendMessageReaction_multi": "sendMessageReaction",
	// "SendPaymentNotification":       "sendPaymentNotification",
	"SendNoteToSelfMessage":         "sendNoteToSelfMessage",
	"SendReadReceipt":               "sendReadReceipt",
	"SendViewedReceipt":             "sendViewedReceipt",
	"SendRemoteDeleteMessage":       "sendRemoteDeleteMessage",
	"SendRemoteDeleteMessage_multi": "sendRemoteDeleteMessage",
	"SendTyping":                    "sendTyping",
	"SetContactBlocked":             "setContactBlocked",
	"SetContactName":                "setContactName",
	"DeleteContact":                 "deleteContact",
	"DeleteRecipient":               "deleteRecipient",
	"SetExpirationTimer":            "setExpirationTimer",
	"SetPin":                        "setPin",
	"SubmitRateLimitChallenge":      "submitRateLimitChallenge",
	"UpdateProfile":                 "updateProfile",
	"UpdateProfile_firstLastName":   "updateProfile",
	"UploadStickerPack":             "uploadStickerPack",
	"Version":                       "version",
	"CreateGroup":                   "createGroup",
	// "GetGroup":                      "getGroup",
	"GetGroupMembers":              "getGroupMembers",
	"GetGroupName":                 "getGroupName",
	"JoinGroup":                    "joinGroup",
	"SendGroupMessage":             "sendGroupMessage",
	"SendGroupTyping":              "sendGroupTyping",
	"SendGroupMessageReaction":     "sendGroupMessageReaction",
	"SendGroupRemoteDeleteMessage": "sendGroupRemoteDeleteMessage",
}

func Setup_dbus(s *Dbus_api, funcs []string, t *testing.T) (*dbus.Conn, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		t.Error(err)
	}

	m := make(map[string]string)
	for _, k := range funcs {
		if v, ok := func_map[k]; !ok {
			t.Error(k, "no valid function")
		} else {
			m[k] = v
		}
	}
	m["GetSelfNumber"] = func_map["GetSelfNumber"]

	err = conn.ExportWithMap(s, m, "/org/asamk/Signal", "org.asamk.Signal")
	if err != nil {
		panic(err)
	}

	reply, err := conn.RequestName("org.asamk.Signal", dbus.NameFlagDoNotQueue)
	if err != nil {
		panic(err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		panic("name already taken")
	}

	// fmt.Println("Listening on org.asamk.Signal / /org/asamk/Signal ...")

	return conn, nil
}
