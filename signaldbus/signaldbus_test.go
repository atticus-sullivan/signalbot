package signaldbus_test

import (
	"io"
	"reflect"
	"signalbot_go/signaldbus"
	"signalbot_go/signaldbus/mock"
	"testing"

	"golang.org/x/exp/slog"
)

var log *slog.Logger = slog.New(slog.HandlerOptions{AddSource: true, Level: slog.LevelError}.NewTextHandler(io.Discard))

func TestGetContactName(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Numbers:     []string{"+94453", "+351343"},
		Name:        "noname",
	}
	conn, err := mock.Setup_dbus(&d, []string{"GetContactName"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	name, err := acc.GetContactName(
		d.Numbers[0],
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 1 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 1)
	}
	v, ok := mock.Args[0].(string)
	if !ok || !reflect.DeepEqual(v, d.Numbers[0]) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Numbers[0])
	}
	// check returned value
	if !reflect.DeepEqual(name, d.Name) {
		t.Fatalf("name do not match. Is %v (should: %v)", name, d.Name)
	}
}

func TestGetContactNumber(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Name:        "testing",
		Numbers:     []string{"+49424", "+98424"},
	}
	conn, err := mock.Setup_dbus(&d, []string{"GetContactNumber"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	numbers, err := acc.GetContactNumber(
		d.Name,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 1 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 1)
	}
	v, ok := mock.Args[0].(string)
	if !ok || !reflect.DeepEqual(v, d.Name) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Name)
	}
	// check returned value
	if !reflect.DeepEqual(numbers, d.Numbers) {
		t.Fatalf("numbers do not match. Is %v (should: %v)", numbers, d.Numbers)
	}
}

func TestGetSelfNumber(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+494242",
	}
	conn, err := mock.Setup_dbus(&d, []string{}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	number, err := acc.GetSelfNumber()
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 0 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 0)
	}
	// check returned value
	if !reflect.DeepEqual(number, d.Number_self) {
		t.Fatalf("number do not match. Is %v (should: %v)", number, d.Number_self)
	}
}

func TestIsContactBlocked(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Numbers:     []string{"+9532", "+135364", "+2423"},
		Blocked:     true,
	}
	conn, err := mock.Setup_dbus(&d, []string{"IsContactBlocked"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	blocked, err := acc.IsContactBlocked(
		d.Numbers[0],
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 1 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 1)
	}
	v, ok := mock.Args[0].(string)
	if !ok || !reflect.DeepEqual(v, d.Numbers[0]) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Numbers[0])
	}
	// check returned value
	if !reflect.DeepEqual(blocked, d.Blocked) {
		t.Fatalf("blocked do not match. Is %v (should: %v)", blocked, d.Blocked)
	}
}

func TestIsRegistered(t *testing.T) {
	d := mock.Dbus_api{
		Result: true,
	}
	conn, err := mock.Setup_dbus(&d, []string{"IsRegistered"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	result, err := acc.IsRegistered()
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 0 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 0)
	}
	// check returned value
	if !reflect.DeepEqual(result, d.Result) {
		t.Fatalf("result do not match. Is %v (should: %v)", result, d.Result)
	}
}

func TestIsRegistered_num(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Numbers:     []string{"+2324", "+324"},
		Result:      true,
	}
	conn, err := mock.Setup_dbus(&d, []string{"IsRegistered_num"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	result, err := acc.IsRegistered_num(
		d.Numbers[0],
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 1 {
		t.Fatalf("wrong number of arguments passed. is %v (should: %v)", len(mock.Args), 1)
	}

	if v, ok := mock.Args[0].(string); !ok || !reflect.DeepEqual(v, d.Numbers[0]) {
		t.Fatalf("wrong argument passed. is %v (should: %v)", mock.Args[0], d.Numbers[0])
	}
	// check returned value
	if !reflect.DeepEqual(result, d.Result) {
		t.Fatalf("result do not match. is %v (should: %v)", result, d.Result)
	}
}

func TestIsRegistered_nums(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Numbers:     []string{"+1231", "+24", "+4234"},
		Results:     []bool{true, false, false},
	}
	conn, err := mock.Setup_dbus(&d, []string{"IsRegistered_nums"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	results, err := acc.IsRegistered_nums(
		d.Numbers,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 1 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 1)
	}

	if v, ok := mock.Args[0].([]string); !ok || !reflect.DeepEqual(v, d.Numbers) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Numbers)
	}
	// check returned value
	if !reflect.DeepEqual(results, d.Results) {
		t.Fatalf("results do not match. Is %v (should: %v)", results, d.Results)
	}
}

func TestListNumbers(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Numbers:     []string{"+324", "+4353"},
	}
	conn, err := mock.Setup_dbus(&d, []string{"ListNumbers"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	numbers, err := acc.ListNumbers()
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 0 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 0)
	}
	// check returned value
	if !reflect.DeepEqual(numbers, d.Numbers) {
		t.Fatalf("numbers do not match. Is %v (should: %v)", numbers, d.Numbers)
	}
}

func TestRemovePin(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
	}
	conn, err := mock.Setup_dbus(&d, []string{"RemovePin"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	err = acc.RemovePin()
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 0 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 0)
	}
	// // check returned value
	// if !reflect.DeepEqual(, d.) {
	// 	t.Fatalf(" do not match. Is %v (should: %v)", , d.)
	// }
}

func TestSendEndSessionMessage(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Recipients:  []string{"+424", "+3242", "+344"},
	}
	conn, err := mock.Setup_dbus(&d, []string{"SendEndSessionMessage"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	err = acc.SendEndSessionMessage(
		d.Recipients,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 1 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 1)
	}
	v, ok := mock.Args[0].([]string)
	if !ok || !reflect.DeepEqual(v, d.Recipients) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Recipients)
	}
	// // check returned value
	// if !reflect.DeepEqual(, d.) {
	// 	t.Fatalf(" do not match. Is %v (should: %v)", , d.)
	// }
}

func TestSendMessage(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Message:     "hello test",
		Attachments: []string{"+314", "+3242", "+342"},
		Recipients:  []string{"+42352", "+2435"},
		Timestamp:   1245,
	}
	conn, err := mock.Setup_dbus(&d, []string{"SendMessage"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	timestamp, err := acc.SendMessage(
		d.Message,
		d.Attachments,
		d.Recipients[0],
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 3 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 3)
	}
	{
		if v, ok := mock.Args[0].(string); !ok || !reflect.DeepEqual(v, d.Message) {
			t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Message)
		}
	}
	if v, ok := mock.Args[1].([]string); !ok || !reflect.DeepEqual(v, d.Attachments) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.Attachments)
	}
	if v, ok := mock.Args[2].(string); !ok || !reflect.DeepEqual(v, d.Recipients[0]) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[2], d.Recipients[0])
	}
	// check returned value
	if !reflect.DeepEqual(timestamp, d.Timestamp) {
		t.Fatalf("timestamp do not match. Is %v (should: %v)", timestamp, d.Timestamp)
	}
}

func TestSendMessage_multi(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Message:     "hello test2",
		Attachments: []string{"attach"},
		Recipients:  []string{"+3242", "3576"},
		Timestamp:   7567,
	}
	conn, err := mock.Setup_dbus(&d, []string{"SendMessage_multi"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	timestamp, err := acc.SendMessage_multi(
		d.Message,
		d.Attachments,
		d.Recipients,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 3 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 3)
	}
	v, ok := mock.Args[0].(string)
	if !ok || !reflect.DeepEqual(v, d.Message) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Message)
	}
	if v, ok := mock.Args[1].([]string); !ok || !reflect.DeepEqual(v, d.Attachments) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.Attachments)
	}
	if v, ok := mock.Args[2].([]string); !ok || !reflect.DeepEqual(v, d.Recipients) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[2], d.Recipients)
	}
	// check returned value
	if !reflect.DeepEqual(timestamp, d.Timestamp) {
		t.Fatalf("timestamp do not match. Is %v (should: %v)", timestamp, d.Timestamp)
	}
}

func TestSendMessageReaction(t *testing.T) {
	d := mock.Dbus_api{
		Number_self:          "+49",
		Emoji:                "em",
		Remove:               false,
		TargetAuthor:         "+1242",
		TargetSentTimestamps: []int64{34546},
		Recipients:           []string{"+7567"},
		Timestamp:            9870,
	}
	conn, err := mock.Setup_dbus(&d, []string{"SendMessageReaction"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	timestamp, err := acc.SendMessageReaction(
		d.Emoji,
		d.Remove,
		d.TargetAuthor,
		d.TargetSentTimestamps[0],
		d.Recipients[0],
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 5 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 5)
	}
	v, ok := mock.Args[0].(string)
	if !ok || !reflect.DeepEqual(v, d.Emoji) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Emoji)
	}
	if v, ok := mock.Args[1].(bool); !ok || !reflect.DeepEqual(v, d.Remove) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.Remove)
	}
	if v, ok := mock.Args[2].(string); !ok || !reflect.DeepEqual(v, d.TargetAuthor) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[2], d.TargetAuthor)
	}
	if v, ok := mock.Args[3].(int64); !ok || !reflect.DeepEqual(v, d.TargetSentTimestamps[0]) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[3], d.TargetSentTimestamps[0])
	}
	if v, ok := mock.Args[4].(string); !ok || !reflect.DeepEqual(v, d.Recipients[0]) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[4], d.Recipients[0])
	}
	// check returned value
	if !reflect.DeepEqual(timestamp, d.Timestamp) {
		t.Fatalf("timestamp do not match. Is %v (should: %v)", timestamp, d.Timestamp)
	}
}

func TestSendMessageReaction_multi(t *testing.T) {
	d := mock.Dbus_api{
		Number_self:          "+49",
		Emoji:                "emo",
		Remove:               true,
		TargetAuthor:         "+65732",
		TargetSentTimestamps: []int64{56487},
		Recipients:           []string{"+4253", "+65734"},
		Timestamp:            42820,
	}
	conn, err := mock.Setup_dbus(&d, []string{"SendMessageReaction_multi"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	timestamp, err := acc.SendMessageReaction_multi(
		d.Emoji,
		d.Remove,
		d.TargetAuthor,
		d.TargetSentTimestamps[0],
		d.Recipients,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 5 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 5)
	}
	v, ok := mock.Args[0].(string)
	if !ok || !reflect.DeepEqual(v, d.Emoji) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Emoji)
	}

	if v, ok := mock.Args[1].(bool); !ok || !reflect.DeepEqual(v, d.Remove) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.Remove)
	}

	if v, ok := mock.Args[2].(string); !ok || !reflect.DeepEqual(v, d.TargetAuthor) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[2], d.TargetAuthor)
	}

	if v, ok := mock.Args[3].(int64); !ok || !reflect.DeepEqual(v, d.TargetSentTimestamps[0]) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[3], d.TargetSentTimestamps[0])
	}

	if v, ok := mock.Args[4].([]string); !ok || !reflect.DeepEqual(v, d.Recipients) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[4], d.Recipients)
	}
	// check returned value
	if !reflect.DeepEqual(timestamp, d.Timestamp) {
		t.Fatalf("timestamp do not match. Is %v (should: %v)", timestamp, d.Timestamp)
	}
}

// func TestSendPaymentNotification(t *testing.T) {
// 	d := dbus_api{
// 		number_self: "+49",
// 		receipt:     "+34535",
// 		note:        "",
// 		recipient:   "",
// 		timestamp:   "",
// 	}
// 	conn, err := mock.Setup_dbus(&d, []string{"SendPaymentNotification"}, t)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer conn.Close()
//
// 	acc, err := signaldbus.NewAccount(signaldbus.SessionBus, false)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer acc.Close()
//
// 	timestamp, err := acc.SendPaymentNotification(
// 		d.Receipt,
// 		d.Note,
// 		d.Recipient,
// 	)
// 	if err != nil {
// 		panic(err)
// 	}
// 	// check arguments
// 	if len(args) != 0 {
// 		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(args), 0)
// 	}
// 	v, ok := args[0].([]byte)
// 	if !ok || !reflect.DeepEqual(v, d.Receipt) {
// 		t.Fatalf("Wrong argument passed. Is %v (should: %v)", args[0], d.Receipt)
// 	}
// 	v, ok := args[1].(string)
// 	if !ok || !reflect.DeepEqual(v, d.Note) {
// 		t.Fatalf("Wrong argument passed. Is %v (should: %v)", args[1], d.Note)
// 	}
// 	v, ok := args[2].(string)
// 	if !ok || !reflect.DeepEqual(v, d.Recipient) {
// 		t.Fatalf("Wrong argument passed. Is %v (should: %v)", args[2], d.Recipient)
// 	}
// 	// check returned value
// 	if !reflect.DeepEqual(timestamp, d.Timestamp) {
// 		t.Fatalf("timestamp do not match. Is %v (should: %v)", timestamp, d.Timestamp)
// 	}
// }

func TestSendNoteToSelfMessage(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Message:     "hello testing self",
		Attachments: []string{"attachment1", "att2"},
		Timestamp:   234345643,
	}
	conn, err := mock.Setup_dbus(&d, []string{"SendNoteToSelfMessage"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	timestamp, err := acc.SendNoteToSelfMessage(
		d.Message,
		d.Attachments,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 2 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 2)
	}

	if v, ok := mock.Args[0].(string); !ok || !reflect.DeepEqual(v, d.Message) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Message)
	}

	if v, ok := mock.Args[1].([]string); !ok || !reflect.DeepEqual(v, d.Attachments) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.Attachments)
	}
	// check returned value
	if !reflect.DeepEqual(timestamp, d.Timestamp) {
		t.Fatalf("timestamp do not match. Is %v (should: %v)", timestamp, d.Timestamp)
	}
}

func TestSendReadReceipt(t *testing.T) {
	d := mock.Dbus_api{
		Number_self:          "+49",
		Recipients:           []string{"+6474"},
		TargetSentTimestamps: []int64{42534643, 23435},
	}
	conn, err := mock.Setup_dbus(&d, []string{"SendReadReceipt"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	err = acc.SendReadReceipt(
		d.Recipients[0],
		d.TargetSentTimestamps,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 2 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 2)
	}

	if v, ok := mock.Args[0].(string); !ok || !reflect.DeepEqual(v, d.Recipients[0]) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Recipients[0])
	}

	if v, ok := mock.Args[1].([]int64); !ok || !reflect.DeepEqual(v, d.TargetSentTimestamps) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.TargetSentTimestamps)
	}
	// // check returned value
	// if !reflect.DeepEqual(, d.) {
	// 	t.Fatalf(" do not match. Is %v (should: %v)", , d.)
	// }
}

func TestSendViewedReceipt(t *testing.T) {
	d := mock.Dbus_api{
		Number_self:          "+49",
		Recipients:           []string{"+536", "+3543"},
		TargetSentTimestamps: []int64{53232},
	}
	conn, err := mock.Setup_dbus(&d, []string{"SendViewedReceipt"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	err = acc.SendViewedReceipt(
		d.Recipients[0],
		d.TargetSentTimestamps,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 2 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 2)
	}

	if v, ok := mock.Args[0].(string); !ok || !reflect.DeepEqual(v, d.Recipients[0]) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Recipients[0])
	}

	if v, ok := mock.Args[1].([]int64); !ok || !reflect.DeepEqual(v, d.TargetSentTimestamps) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.TargetSentTimestamps)
	}
	// // check returned value
	// if !reflect.DeepEqual(, d.) {
	// 	t.Fatalf(" do not match. Is %v (should: %v)", , d.)
	// }
}

func TestSendRemoteDeleteMessage(t *testing.T) {
	d := mock.Dbus_api{
		Number_self:          "+49",
		TargetSentTimestamps: []int64{1342342},
		Recipients:           []string{"+43253"},
		Timestamp:            4363,
	}
	conn, err := mock.Setup_dbus(&d, []string{"SendRemoteDeleteMessage"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	timestamp, err := acc.SendRemoteDeleteMessage(
		d.TargetSentTimestamps[0],
		d.Recipients[0],
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 2 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 2)
	}

	if v, ok := mock.Args[0].(int64); !ok || !reflect.DeepEqual(v, d.TargetSentTimestamps[0]) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.TargetSentTimestamps[0])
	}

	if v, ok := mock.Args[1].(string); !ok || !reflect.DeepEqual(v, d.Recipients[0]) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.Recipients[0])
	}
	// check returned value
	if !reflect.DeepEqual(timestamp, d.Timestamp) {
		t.Fatalf("timestamp do not match. Is %v (should: %v)", timestamp, d.Timestamp)
	}
}

func TestSendRemoteDeleteMessage_multi(t *testing.T) {
	d := mock.Dbus_api{
		Number_self:          "+49",
		TargetSentTimestamps: []int64{3142},
		Recipients:           []string{"+42352", "+3235"},
		Timestamp:            4253,
	}
	conn, err := mock.Setup_dbus(&d, []string{"SendRemoteDeleteMessage_multi"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	timestamp, err := acc.SendRemoteDeleteMessage_multi(
		d.TargetSentTimestamps[0],
		d.Recipients,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 2 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 2)
	}

	if v, ok := mock.Args[0].(int64); !ok || !reflect.DeepEqual(v, d.TargetSentTimestamps[0]) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.TargetSentTimestamps[0])
	}

	if v, ok := mock.Args[1].([]string); !ok || !reflect.DeepEqual(v, d.Recipients) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.Recipients)
	}
	// check returned value
	if !reflect.DeepEqual(timestamp, d.Timestamp) {
		t.Fatalf("timestamp do not match. Is %v (should: %v)", timestamp, d.Timestamp)
	}
}

func TestSendTyping(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Recipients:  []string{"+2423", "+42342"},
		Stop:        true,
	}
	conn, err := mock.Setup_dbus(&d, []string{"SendTyping"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	err = acc.SendTyping(
		d.Recipients[0],
		d.Stop,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 2 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 2)
	}

	if v, ok := mock.Args[0].(string); !ok || !reflect.DeepEqual(v, d.Recipients[0]) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Recipients[0])
	}

	if v, ok := mock.Args[1].(bool); !ok || !reflect.DeepEqual(v, d.Stop) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.Stop)
	}
	// // check returned value
	// if !reflect.DeepEqual(, d.) {
	// 	t.Fatalf(" do not match. Is %v (should: %v)", , d.)
	// }
}

func TestSetContactBlocked(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Numbers:     []string{"+23423"},
		Block:       false,
	}
	conn, err := mock.Setup_dbus(&d, []string{"SetContactBlocked"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	err = acc.SetContactBlocked(
		d.Numbers[0],
		d.Block,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 2 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 2)
	}

	if v, ok := mock.Args[0].(string); !ok || !reflect.DeepEqual(v, d.Numbers[0]) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Numbers[0])
	}

	if v, ok := mock.Args[1].(bool); !ok || !reflect.DeepEqual(v, d.Block) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.Block)
	}
	// // check returned value
	// if !reflect.DeepEqual(, d.) {
	// 	t.Fatalf(" do not match. Is %v (should: %v)", , d.)
	// }
}

func TestSetContactName(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Numbers:     []string{"+4234"},
		Name:        "xyza",
	}
	conn, err := mock.Setup_dbus(&d, []string{"SetContactName"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	err = acc.SetContactName(
		d.Numbers[0],
		d.Name,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 2 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 2)
	}

	if v, ok := mock.Args[0].(string); !ok || !reflect.DeepEqual(v, d.Numbers[0]) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Numbers[0])
	}

	if v, ok := mock.Args[1].(string); !ok || !reflect.DeepEqual(v, d.Name) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.Name)
	}
	// // check returned value
	// if !reflect.DeepEqual(, d.) {
	// 	t.Fatalf(" do not match. Is %v (should: %v)", , d.)
	// }
}

func TestDeleteContact(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Numbers:     []string{"+4234", "+5353"},
	}
	conn, err := mock.Setup_dbus(&d, []string{"DeleteContact"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	err = acc.DeleteContact(
		d.Numbers[0],
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 1 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 1)
	}

	if v, ok := mock.Args[0].(string); !ok || !reflect.DeepEqual(v, d.Numbers[0]) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Numbers[0])
	}
	// // check returned value
	// if !reflect.DeepEqual(, d.) {
	// 	t.Fatalf(" do not match. Is %v (should: %v)", , d.)
	// }
}

func TestDeleteRecipient(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Numbers:     []string{"+42353", "+2455"},
	}
	conn, err := mock.Setup_dbus(&d, []string{"DeleteRecipient"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	err = acc.DeleteRecipient(
		d.Numbers[0],
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 1 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 1)
	}
	v, ok := mock.Args[0].(string)
	if !ok || !reflect.DeepEqual(v, d.Numbers[0]) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Numbers[0])
	}
	// // check returned value
	// if !reflect.DeepEqual(, d.) {
	// 	t.Fatalf(" do not match. Is %v (should: %v)", , d.)
	// }
}

func TestSetExpirationTimer(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Numbers:     []string{"+42353"},
		Expiration:  353,
	}
	conn, err := mock.Setup_dbus(&d, []string{"SetExpirationTimer"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	err = acc.SetExpirationTimer(
		d.Numbers[0],
		d.Expiration,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 2 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 2)
	}

	if v, ok := mock.Args[0].(string); !ok || !reflect.DeepEqual(v, d.Numbers[0]) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Numbers[0])
	}

	if v, ok := mock.Args[1].(int32); !ok || !reflect.DeepEqual(v, d.Expiration) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.Expiration)
	}
	// // check returned value
	// if !reflect.DeepEqual(, d.) {
	// 	t.Fatalf(" do not match. Is %v (should: %v)", , d.)
	// }
}

func TestSetPin(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Pin:         "5363",
	}
	conn, err := mock.Setup_dbus(&d, []string{"SetPin"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	err = acc.SetPin(
		d.Pin,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 1 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 1)
	}

	if v, ok := mock.Args[0].(string); !ok || !reflect.DeepEqual(v, d.Pin) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Pin)
	}
	// // check returned value
	// if !reflect.DeepEqual(, d.) {
	// 	t.Fatalf(" do not match. Is %v (should: %v)", , d.)
	// }
}

func TestSubmitRateLimitChallenge(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Challenge:   "ewfege",
		Captcha:     "ewtgnönliu",
	}
	conn, err := mock.Setup_dbus(&d, []string{"SubmitRateLimitChallenge"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	err = acc.SubmitRateLimitChallenge(
		d.Challenge,
		d.Captcha,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 2 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 2)
	}

	if v, ok := mock.Args[0].(string); !ok || !reflect.DeepEqual(v, d.Challenge) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Challenge)
	}

	if v, ok := mock.Args[1].(string); !ok || !reflect.DeepEqual(v, d.Captcha) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.Captcha)
	}
	// // check returned value
	// if !reflect.DeepEqual(, d.) {
	// 	t.Fatalf(" do not match. Is %v (should: %v)", , d.)
	// }
}

func TestUpdateProfile(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Name:        "noname",
		About:       "about me",
		Emoji:       "abEm",
		Avatar:      "av",
		Remove:      false,
	}
	conn, err := mock.Setup_dbus(&d, []string{"UpdateProfile"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	err = acc.UpdateProfile(
		d.Name,
		d.About,
		d.Emoji,
		d.Avatar,
		d.Remove,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 5 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 5)
	}

	if v, ok := mock.Args[0].(string); !ok || !reflect.DeepEqual(v, d.Name) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Name)
	}

	if v, ok := mock.Args[1].(string); !ok || !reflect.DeepEqual(v, d.About) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.About)
	}

	if v, ok := mock.Args[2].(string); !ok || !reflect.DeepEqual(v, d.Emoji) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[2], d.Emoji)
	}

	if v, ok := mock.Args[3].(string); !ok || !reflect.DeepEqual(v, d.Avatar) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[3], d.Avatar)
	}

	if v, ok := mock.Args[4].(bool); !ok || !reflect.DeepEqual(v, d.Remove) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[4], d.Remove)
	}
	// // check returned value
	// if !reflect.DeepEqual(, d.) {
	// 	t.Fatalf(" do not match. Is %v (should: %v)", , d.)
	// }
}

func TestUpdateProfile_firstLastName(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		GivenName:   "given",
		FamilyName:  "family",
		About:       "about",
		Emoji:       "abEmo",
		Avatar:      "avatar",
		Remove:      false,
	}
	conn, err := mock.Setup_dbus(&d, []string{"UpdateProfile_firstLastName"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	err = acc.UpdateProfile_firstLastName(
		d.GivenName,
		d.FamilyName,
		d.About,
		d.Emoji,
		d.Avatar,
		d.Remove,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 6 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 6)
	}

	if v, ok := mock.Args[0].(string); !ok || !reflect.DeepEqual(v, d.GivenName) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.GivenName)
	}

	if v, ok := mock.Args[1].(string); !ok || !reflect.DeepEqual(v, d.FamilyName) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.FamilyName)
	}

	if v, ok := mock.Args[2].(string); !ok || !reflect.DeepEqual(v, d.About) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[2], d.About)
	}

	if v, ok := mock.Args[3].(string); !ok || !reflect.DeepEqual(v, d.Emoji) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[3], d.Emoji)
	}

	if v, ok := mock.Args[4].(string); !ok || !reflect.DeepEqual(v, d.Avatar) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[4], d.Avatar)
	}

	if v, ok := mock.Args[5].(bool); !ok || !reflect.DeepEqual(v, d.Remove) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[5], d.Remove)
	}
	// // check returned value
	// if !reflect.DeepEqual(, d.) {
	// 	t.Fatalf(" do not match. Is %v (should: %v)", , d.)
	// }
}

func TestUploadStickerPack(t *testing.T) {
	d := mock.Dbus_api{
		Number_self:     "+49",
		StickerPackPath: "dfj",
		Url:             "http://fwfjn",
	}
	conn, err := mock.Setup_dbus(&d, []string{"UploadStickerPack"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	url, err := acc.UploadStickerPack(
		d.StickerPackPath,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 1 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 1)
	}

	if v, ok := mock.Args[0].(string); !ok || !reflect.DeepEqual(v, d.StickerPackPath) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.StickerPackPath)
	}
	// check returned value
	if !reflect.DeepEqual(url, d.Url) {
		t.Fatalf("url do not match. Is %v (should: %v)", url, d.Url)
	}
}

func TestVersion(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Version_:    "1.45",
	}
	conn, err := mock.Setup_dbus(&d, []string{"Version"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	version, err := acc.Version()
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 0 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 0)
	}
	// check returned value
	if !reflect.DeepEqual(version, d.Version_) {
		t.Fatalf("version do not match. Is %v (should: %v)", version, d.Version_)
	}
}

func TestCreateGroup(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		GroupName:   "group",
		Members:     []string{"rr", "fewfw"},
		Avatar:      "dfwg/dwf",
		GroupId:     []byte{242, 45, 42},
	}
	conn, err := mock.Setup_dbus(&d, []string{"CreateGroup"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	groupId, err := acc.CreateGroup(
		d.GroupName,
		d.Members,
		d.Avatar,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 3 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 3)
	}

	if v, ok := mock.Args[0].(string); !ok || !reflect.DeepEqual(v, d.GroupName) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.GroupName)
	}

	if v, ok := mock.Args[1].([]string); !ok || !reflect.DeepEqual(v, d.Members) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.Members)
	}

	if v, ok := mock.Args[2].(string); !ok || !reflect.DeepEqual(v, d.Avatar) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[2], d.Avatar)
	}
	// check returned value
	if !reflect.DeepEqual(groupId, d.GroupId) {
		t.Fatalf("groupId do not match. Is %v (should: %v)", groupId, d.GroupId)
	}
}

// func TestGetGroup(t *testing.T) {
// 	d := dbus_api{
// 		number_self: "+49",
// 		groupId: "",
// 		objectPath: "",
// 	}
// 	conn, err := mock.Setup_dbus(&d, []string{"GetGroup"}, t)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer conn.Close()
//
// 	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer acc.Close()
//
// 	objectPath, err := acc.GetGroup(
// 	d.GroupId,
// 	)
// 	if err != nil {
// 		panic(err)
// 	}
// 	// check arguments
// 	if len(args) != 0 {
// 		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(args), 0)
// 	}
// 	v, ok := args[0].([]byte)
// 	if !ok || !reflect.DeepEqual(v, d.GroupId) {
// 		t.Fatalf("Wrong argument passed. Is %v (should: %v)", args[0], d.GroupId)
// 	}
// 	// check returned value
// 	if !reflect.DeepEqual(objectPath, d.ObjectPath) {
// 		t.Fatalf("objectPath do not match. Is %v (should: %v)", objectPath, d.ObjectPath)
// 	}
// }

func TestGetGroupName(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		GroupId:     []byte{44, 31, 213},
		Name:        "the group",
	}
	conn, err := mock.Setup_dbus(&d, []string{"GetGroupName"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	name, err := acc.GetGroupName(
		d.GroupId,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 1 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 1)
	}

	if v, ok := mock.Args[0].([]byte); !ok || !reflect.DeepEqual(v, d.GroupId) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.GroupId)
	}
	// check returned value
	if !reflect.DeepEqual(name, d.Name) {
		t.Fatalf("members do not match. Is %v (should: %v)", name, d.Name)
	}
}

func TestGetGroupMembers(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		GroupId:     []byte{44, 31, 213},
		Members:     []string{"+42342", "+342"},
	}
	conn, err := mock.Setup_dbus(&d, []string{"GetGroupMembers"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	members, err := acc.GetGroupMembers(
		d.GroupId,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 1 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 1)
	}

	if v, ok := mock.Args[0].([]byte); !ok || !reflect.DeepEqual(v, d.GroupId) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.GroupId)
	}
	// check returned value
	if !reflect.DeepEqual(members, d.Members) {
		t.Fatalf("members do not match. Is %v (should: %v)", members, d.Members)
	}
}

func TestJoinGroup(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		InviteURI:   "http://invite.url",
	}
	conn, err := mock.Setup_dbus(&d, []string{"JoinGroup"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	err = acc.JoinGroup(
		d.InviteURI,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 1 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 1)
	}

	if v, ok := mock.Args[0].(string); !ok || !reflect.DeepEqual(v, d.InviteURI) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.InviteURI)
	}
	// // check returned value
	// if !reflect.DeepEqual(, d.) {
	// 	t.Fatalf(" do not match. Is %v (should: %v)", , d.)
	// }
}

func TestSendGroupMessage(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		Message:     "hello group",
		Attachments: []string{"attacch/me"},
		GroupId:     []byte{31, 32, 231},
		Timestamp:   42543,
	}
	conn, err := mock.Setup_dbus(&d, []string{"SendGroupMessage"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	timestamp, err := acc.SendGroupMessage(
		d.Message,
		d.Attachments,
		d.GroupId,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 3 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 3)
	}

	if v, ok := mock.Args[0].(string); !ok || !reflect.DeepEqual(v, d.Message) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Message)
	}

	if v, ok := mock.Args[1].([]string); !ok || !reflect.DeepEqual(v, d.Attachments) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.Attachments)
	}

	if v, ok := mock.Args[2].([]byte); !ok || !reflect.DeepEqual(v, d.GroupId) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[2], d.GroupId)
	}
	// check returned value
	if !reflect.DeepEqual(timestamp, d.Timestamp) {
		t.Fatalf("timestamp do not match. Is %v (should: %v)", timestamp, d.Timestamp)
	}
}

func TestSendGroupTyping(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
		GroupId:     []byte{32, 32, 13},
		Stop:        false,
	}
	conn, err := mock.Setup_dbus(&d, []string{"SendGroupTyping"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	err = acc.SendGroupTyping(
		d.GroupId,
		d.Stop,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 2 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 2)
	}

	if v, ok := mock.Args[0].([]byte); !ok || !reflect.DeepEqual(v, d.GroupId) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.GroupId)
	}

	if v, ok := mock.Args[1].(bool); !ok || !reflect.DeepEqual(v, d.Stop) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.Stop)
	}
	// // check returned value
	// if !reflect.DeepEqual(, d.) {
	// 	t.Fatalf(" do not match. Is %v (should: %v)", , d.)
	// }
}

func TestSendGroupMessageReaction(t *testing.T) {
	d := mock.Dbus_api{
		Number_self:          "+49",
		Emoji:                "emoj",
		Remove:               false,
		TargetAuthor:         "+424",
		TargetSentTimestamps: []int64{3234},
		GroupId:              []byte{34, 45, 31},
		Timestamp:            424,
	}
	conn, err := mock.Setup_dbus(&d, []string{"SendGroupMessageReaction"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	timestamp, err := acc.SendGroupMessageReaction(
		d.Emoji,
		d.Remove,
		d.TargetAuthor,
		d.TargetSentTimestamps[0],
		d.GroupId,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 5 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 5)
	}

	if v, ok := mock.Args[0].(string); !ok || !reflect.DeepEqual(v, d.Emoji) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.Emoji)
	}

	if v, ok := mock.Args[1].(bool); !ok || !reflect.DeepEqual(v, d.Remove) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.Remove)
	}

	if v, ok := mock.Args[2].(string); !ok || !reflect.DeepEqual(v, d.TargetAuthor) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[2], d.TargetAuthor)
	}

	if v, ok := mock.Args[3].(int64); !ok || !reflect.DeepEqual(v, d.TargetSentTimestamps[0]) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[3], d.TargetSentTimestamps[0])
	}

	if v, ok := mock.Args[4].([]byte); !ok || !reflect.DeepEqual(v, d.GroupId) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[4], d.GroupId)
	}
	// check returned value
	if !reflect.DeepEqual(timestamp, d.Timestamp) {
		t.Fatalf("timestamp do not match. Is %v (should: %v)", timestamp, d.Timestamp)
	}
}

func TestSendGroupRemoteDeleteMessage(t *testing.T) {
	d := mock.Dbus_api{
		Number_self:          "+49",
		TargetSentTimestamps: []int64{2343, 342},
		GroupId:              []byte{31, 32, 211},
		Timestamp:            342,
	}
	conn, err := mock.Setup_dbus(&d, []string{"SendGroupRemoteDeleteMessage"}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	timestamp, err := acc.SendGroupRemoteDeleteMessage(
		d.TargetSentTimestamps[0],
		d.GroupId,
	)
	if err != nil {
		panic(err)
	}
	// check arguments
	if len(mock.Args) != 2 {
		t.Fatalf("Wrong number of arguments passed. Is %v (should: %v)", len(mock.Args), 2)
	}

	if v, ok := mock.Args[0].(int64); !ok || !reflect.DeepEqual(v, d.TargetSentTimestamps[0]) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[0], d.TargetSentTimestamps[0])
	}

	if v, ok := mock.Args[1].([]byte); !ok || !reflect.DeepEqual(v, d.GroupId) {
		t.Fatalf("Wrong argument passed. Is %v (should: %v)", mock.Args[1], d.GroupId)
	}
	// check returned value
	if !reflect.DeepEqual(timestamp, d.Timestamp) {
		t.Fatalf("timestamp do not match. Is %v (should: %v)", timestamp, d.Timestamp)
	}
}

func TestMessageReceived(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
	}
	conn, err := mock.Setup_dbus(&d, []string{}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	sync := make(chan int)
	testChan := make(chan *signaldbus.Message)
	acc.AddMessageHandlerFunc(func(m *signaldbus.Message) {
		testChan <- m
	})
	go acc.ListenForSignalsWithSync(sync)
	<-sync

	message_should := signaldbus.Message{
		Timestamp:   1234,
		Sender:      "+494242",
		GroupId:     []byte{23, 23, 55},
		Message:     "hello world",
		Attachments: []string{"attach1"},
	}
	err = conn.Emit("/org/asamk/Signal", "org.asamk.Signal.MessageReceived", message_should.Timestamp, message_should.Sender, message_should.GroupId, message_should.Message, message_should.Attachments)
	if err != nil {
		panic(err)
	}

	message := <-testChan

	t.Log(message, "message")
	t.Log(message_should, "message_should")

	if !reflect.DeepEqual(message.Timestamp, message_should.Timestamp) {
		t.Fatalf("Timestamp do not match. Is '%v' (should: '%v')", message.Timestamp, message_should.Timestamp)
	}
	if !reflect.DeepEqual(message.Sender, message_should.Sender) {
		t.Fatalf("Sender do not match. Is %v (should: %v)", message.Sender, message_should.Sender)
	}
	if !reflect.DeepEqual(message.GroupId, message_should.GroupId) {
		t.Fatalf("GroupId do not match. Is %v (should: %v)", message.GroupId, message_should.GroupId)
	}
	if !reflect.DeepEqual(message.Message, message_should.Message) {
		t.Fatalf("Message do not match. Is %v (should: %v)", message.Message, message_should.Message)
	}
	if !reflect.DeepEqual(message.Attachments, message_should.Attachments) {
		t.Fatalf("Attachments do not match. Is %v (should: %v)", message.Attachments, message_should.Attachments)
	}
}

func TestSyncMessageReceived(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
	}
	conn, err := mock.Setup_dbus(&d, []string{}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	sync := make(chan int)
	testChan := make(chan *signaldbus.SyncMessage)
	acc.AddSyncMessageHandlerFunc(func(m *signaldbus.SyncMessage) {
		testChan <- m
	})
	go acc.ListenForSignalsWithSync(sync)
	<-sync

	message_should := signaldbus.SyncMessage{
		Message: signaldbus.Message{
			Timestamp:   1234,
			Sender:      "+494242",
			GroupId:     []byte{23, 23, 55},
			Message:     "hello world",
			Attachments: []string{"attach1"},
		},
		Destination: "dst",
	}
	err = conn.Emit("/org/asamk/Signal", "org.asamk.Signal.SyncMessageReceived", message_should.Timestamp, message_should.Sender, message_should.Destination, message_should.GroupId, message_should.Message.Message, message_should.Attachments)
	if err != nil {
		panic(err)
	}

	message := <-testChan

	t.Log(message, "message")
	t.Log(message_should, "message_should")

	if !reflect.DeepEqual(message.Timestamp, message_should.Timestamp) {
		t.Fatalf("Timestamp do not match. Is '%v' (should: '%v')", message.Timestamp, message_should.Timestamp)
	}
	if !reflect.DeepEqual(message.Sender, message_should.Sender) {
		t.Fatalf("Sender do not match. Is %v (should: %v)", message.Sender, message_should.Sender)
	}
	if !reflect.DeepEqual(message.GroupId, message_should.GroupId) {
		t.Fatalf("GroupId do not match. Is %v (should: %v)", message.GroupId, message_should.GroupId)
	}
	if !reflect.DeepEqual(message.Message, message_should.Message) {
		t.Fatalf("Message do not match. Is %v (should: %v)", message.Message, message_should.Message)
	}
	if !reflect.DeepEqual(message.Attachments, message_should.Attachments) {
		t.Fatalf("Attachments do not match. Is %v (should: %v)", message.Attachments, message_should.Attachments)
	}
	if !reflect.DeepEqual(message.Destination, message_should.Destination) {
		t.Fatalf("Destination do not match. Is '%v' (should: '%v')", message.Destination, message_should.Destination)
	}
}

func TestReceiptReceived(t *testing.T) {
	d := mock.Dbus_api{
		Number_self: "+49",
	}
	conn, err := mock.Setup_dbus(&d, []string{}, t)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	acc, err := signaldbus.NewAccount(log, signaldbus.SessionBus, false)
	if err != nil {
		panic(err)
	}
	defer acc.Close()

	sync := make(chan int)
	testChan := make(chan *signaldbus.Receipt)
	acc.AddReceiptHandlerFunc(func(m *signaldbus.Receipt) {
		testChan <- m
	})
	go acc.ListenForSignalsWithSync(sync)
	<-sync

	message_should := signaldbus.Receipt{
		Timestamp: 1234,
		Sender:    "+494242",
	}
	err = conn.Emit("/org/asamk/Signal", "org.asamk.Signal.ReceiptReceived", message_should.Timestamp, message_should.Sender)
	if err != nil {
		panic(err)
	}

	message := <-testChan

	t.Log(message, "message")
	t.Log(message_should, "message_should")

	if !reflect.DeepEqual(message.Timestamp, message_should.Timestamp) {
		t.Fatalf("Timestamp do not match. Is '%v' (should: '%v')", message.Timestamp, message_should.Timestamp)
	}
	if !reflect.DeepEqual(message.Sender, message_should.Sender) {
		t.Fatalf("Sender do not match. Is %v (should: %v)", message.Sender, message_should.Sender)
	}
}
