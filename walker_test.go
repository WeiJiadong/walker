package walker

import (
	"log"
	"testing"
)

func TestNewWalker(t *testing.T) {
	w := NewWalker(
		WithUid("手机号/邮箱"),
		WithPasswd("密码"),
		WithStep("步数"))
	err := w.Do()
	if err != nil {
		log.Fatalln(err)
	}
}
