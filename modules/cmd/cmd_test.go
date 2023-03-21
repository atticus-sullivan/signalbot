package cmd_test

import (
	"fmt"
	"signalbot_go/modules/cmd"
	"signalbot_go/signaldbus"

	"golang.org/x/exp/slog"
)

type mock struct {
}
func (mock *mock) Respond(message string, attachments []string, m *signaldbus.Message) (timestamp int64, err error){
	fmt.Println(message)
	return 0, nil
}
func (mock *mock) SendGeneric(message string, attachments []string, recipient string, groupID []byte) (timestamp int64, err error) {
	return 0, nil
}

// func main(){
// 	cmd := cmd.Cmd{
// 		Log: slog.Default(),
// 	}
// 	cmd.Handle(&signaldbus.Message{
// 		Message: "space",
// 	}, &mock{})
// }
