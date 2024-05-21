package signaljsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/exp/jsonrpc2"
)

// RawFramer returns a new Framer.
// The messages are sent with no wrapping, and rely on json decode consistency
// to determine message boundaries.
func RawFramerNewline() jsonrpc2.Framer { return rawFramerNewline{} }

type rawFramerNewline struct{}
type rawReader struct{ in *json.Decoder }
type rawWriterNewline struct{ out io.Writer }

func (rawFramerNewline) Reader(rw io.Reader) jsonrpc2.Reader {
	return &rawReader{in: json.NewDecoder(rw)}
}

func (rawFramerNewline) Writer(rw io.Writer) jsonrpc2.Writer {
	return &rawWriterNewline{out: rw}
}

func (r *rawReader) Read(ctx context.Context) (jsonrpc2.Message, int64, error) {
	select {
	case <-ctx.Done():
		return nil, 0, ctx.Err()
	default:
	}
	var raw json.RawMessage
	if err := r.in.Decode(&raw); err != nil {
		return nil, 0, err
	}
	msg, err := jsonrpc2.DecodeMessage(raw)
	return msg, int64(len(raw)), err
}

func (w *rawWriterNewline) Write(ctx context.Context, msg jsonrpc2.Message) (int64, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}
	data, err := jsonrpc2.EncodeMessage(msg)
	if err != nil {
		return 0, fmt.Errorf("marshaling message: %v", err)
	}
	data = append(data, '\n')
	n, err := w.out.Write(data)
	return int64(n), err
}
