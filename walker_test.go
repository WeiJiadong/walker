package walker

import (
	"log"
	"testing"
)

func TestNewWalker(t *testing.T) {
	w := NewWalker(
		WithUid("15338729859"),
		WithPasswd("u7758258"),
		WithStep("12851"))
	err := w.Do()
	if err != nil {
		log.Fatalln(err)
	}
}
