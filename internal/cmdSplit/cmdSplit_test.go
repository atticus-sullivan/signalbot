package cmdsplit_test

import (
	cmdsplit "signalbot_go/internal/cmdSplit"
	"testing"
)

func TestSplitSucc(t *testing.T) {
	strL, err := cmdsplit.Split(`hello world "arg with\ space" space\ arg backslash\\ 'arg with\ space'`)
	if err != nil {
		t.Fatalf("Err: %v", err)
	}

	if len(strL) != 6 {
		t.Fatalf("Wrong amount of elements. Is %d but should be 6", len(strL))
	}
	if strL[0] != "hello" {
		t.Fatalf("Was: %v but should be %v", strL[0], "hello")
	}
	if strL[1] != "world" {
		t.Fatalf("Was: %v but should be %v", strL[1], "world")
	}
	if strL[2] != "arg with space" {
		t.Fatalf("Was: %v but should be %v", strL[2], "arg with space")
	}
	if strL[3] != "space arg" {
		t.Fatalf("Was: %v but should be %v", strL[3], "space arg")
	}
	if strL[4] != "backslash\\" {
		t.Fatalf("Was: %v but should be %v", strL[4], "backslash\\")
	}
	if strL[5] != "arg with space" {
		t.Fatalf("Was: %v but should be %v", strL[5], "arg with space")
	}
}

func TestSplitFail(t *testing.T) {
	_, err := cmdsplit.Split(`hello world "arg with`)
	if err != cmdsplit.ErrTrailingQuoteEscape {
		t.Fatalf("Should have returned %v", cmdsplit.ErrTrailingQuoteEscape)
	}
}

func TestUnescapeSucc(t *testing.T) {
	str, err := cmdsplit.Unescape(`hello\ world\\slash "nothing useful"`)
	if err != nil {
		t.Fatalf("Err: %v", err)
	}

	if str != `hello world\slash "nothing useful"` {
		t.Fatalf("Was: %v but should be %v", str, `hello world\slash "nothing useful"`)
	}

	str, err = cmdsplit.Unescape(`hello\ world\\slash "nothing useful`)
	if err != nil {
		t.Fatalf("Err: %v", err)
	}

	if str != `hello world\slash "nothing useful` {
		t.Fatalf("Was: %v but should be %v", str, `hello world\slash "nothing useful`)
	}
}

func TestUnescapeFail(t *testing.T) {
	_, err := cmdsplit.Unescape(`hello\ world\`)
	if err != cmdsplit.ErrTrailingEscape {
		t.Fatalf("Should have returned %v", cmdsplit.ErrTrailingEscape)
	}
}
