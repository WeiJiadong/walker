# walker
golang实现的刷步数的库

使用方法：
```go

func TestNewWalker(t *testing.T) {
	w := NewWalker(WithUid("手机号"), WithPasswd("密码"), WithStep("步数"))
	err := w.Do()
	if err != nil {
		log.Fatalln(err)
	}
}
```
