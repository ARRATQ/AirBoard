package utils

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

// signalWriter transmet chaque ligne de journal sur un canal.
type signalWriter chan string

func (w signalWriter) Write(p []byte) (int, error) {
	w <- string(p)
	return len(p), nil
}

func TestRecoverAndLogSwallowsPanic(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	func() {
		defer RecoverAndLog("sync")
		panic("boom")
	}()

	if !strings.Contains(buf.String(), `goroutine "sync" : boom`) {
		t.Fatalf("panique non journalisée : %q", buf.String())
	}
}

func TestRecoverAndLogWithoutPanic(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	func() {
		defer RecoverAndLog("calme")
	}()

	if buf.Len() != 0 {
		t.Fatalf("journal inattendu : %q", buf.String())
	}
}

func TestSafeGoRecoversPanic(t *testing.T) {
	lines := make(signalWriter, 1)
	log.SetOutput(lines)
	defer log.SetOutput(os.Stderr)

	SafeGo("notifier", func() { panic("boom") })

	select {
	case line := <-lines:
		if !strings.Contains(line, `goroutine "notifier" : boom`) {
			t.Fatalf("journal inattendu : %q", line)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("la panique n'a pas été journalisée")
	}
}

func TestSafeGoRunsFunction(t *testing.T) {
	done := make(chan int, 1)
	SafeGo("ok", func() { done <- 42 })
	if got := <-done; got != 42 {
		t.Fatalf("got %d", got)
	}
}
