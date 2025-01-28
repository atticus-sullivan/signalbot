package spotify

import (
	"log/slog"
	"testing"

	"github.com/neilotoole/slogt"
)

func TestAuth(t *testing.T) {
	var testLog *slog.Logger = slogt.New(t)
	fetcher := NewFetcher(testLog, "881f7e3174b346abbf82eab14c4898c1", "fd8154a2c730477c860bfb910e04efff")
	token, _ := fetcher.auth()
	println(token)
	t.Fail()
}
