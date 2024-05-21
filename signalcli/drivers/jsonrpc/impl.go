package signaljsonrpc

import (
	"context"
	"encoding/base64"
)

type sendResult struct {
	Results []string
	Timestamp int64
}

func (d *SignalCliDriver) SendMessage(message string, attachments []string, recipient string, notifySelf bool) (timestamp int64, err error) {
	ctx := context.Background()
	var result sendResult
	err = d.conn.Call(ctx, "send", map[string]any{"recipient": recipient, "message": message, "attachments": attachments, "notifySelf": notifySelf}).Await(ctx, &result)
	if err != nil {
		return 0, err
	}
	return result.Timestamp, nil
}

func (d *SignalCliDriver) SendGroupMessage(message string, attachments []string, groupId []byte) (timestamp int64, err error) {
	gid := base64.StdEncoding.EncodeToString(groupId)
	ctx := context.Background()
	var result sendResult
	err = d.conn.Call(ctx, "send", map[string]any{"groupId": gid, "message": message, "attachments": attachments}).Await(ctx, &result)
	if err != nil {
		return 0, err
	}
	return result.Timestamp, nil
}

type groupResult struct {
	Name string
	Description string
}

func (d *SignalCliDriver) GetGroupName(groupId []byte) (name string, err error) {
	gid := base64.StdEncoding.EncodeToString(groupId)
	ctx := context.Background()
	var result groupResult
	err = d.conn.Call(ctx, "listGroups", map[string]any{"groupId": gid}).Await(ctx, &result)
	if err != nil {
		return "", err
	}
	return result.Name, nil
}
