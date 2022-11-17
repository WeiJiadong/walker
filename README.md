# walker
golang实现的刷步数的库，基于Zapp Life(原小米运动)，支持手机号和邮箱两种登陆方式。

使用方法：
```go

func TestNewWalker(t *testing.T) {
	w := NewWalker(WithUid("手机号或者邮箱"), WithPasswd("密码"), WithStep("步数"))
	err := w.Do()
	if err != nil {
		log.Fatalln(err)
	}
}
```

# 刷步原理
![image](https://user-images.githubusercontent.com/10074838/131337300-3adf9626-5786-4ba3-9a26-1688d0ba8fa1.png)

