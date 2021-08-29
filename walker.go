package walker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"walker/internal/constant"
	"walker/internal/proto"
)

type http_header map[string]string
type http_method string
type http_result struct {
	Result   []byte
	Location *url.URL
}

// walker ...
type walker struct {
	opts walkerOpts
}

type walkerOpts struct {
	uid    string
	passwd string
	step   string
}

// WalkerOpt walkerOpts helper
type WalkerOpt func(opts *walkerOpts)

// WithUid 设置uid
func WithUid(uid string) WalkerOpt {
	return func(opts *walkerOpts) {
		opts.uid = uid
	}
}

// WithPasswd 设置密码
func WithPasswd(passwd string) WalkerOpt {
	return func(opts *walkerOpts) {
		opts.passwd = passwd
	}
}

// WithStep 设置步数
func WithStep(step string) WalkerOpt {
	return func(opts *walkerOpts) {
		opts.step = step
	}
}

// NewWalker 构造函数
func NewWalker(opt ...WalkerOpt) *walker {
	opts := walkerOpts{}
	for i := range opt {
		opt[i](&opts)
	}

	return &walker{opts: opts}
}

// Do 功能执行入口
func (w *walker) Do() error {
	token, err := w.getToken()
	if err != nil {
		log.Fatalln("getToken error:", err)
		return err
	}

	if err := w.setStep(token); err != nil {
		return err
	}

	return nil
}

func (w walker) getToken() (*proto.Token, error) {
	return login(w.opts.uid, w.opts.passwd)
}

func (w walker) setStep(token *proto.Token) (err error) {
	return set(*token, w.opts.step)
}

func login(user, password string) (*proto.Token, error) {
	token := proto.Token{}
	_url := "https://api-user.huami.com/registrations/+86" + user + "/tokens"

	data := url.Values{
		"client_id":    {"HuaMi"},
		"password":     {password},
		"redirect_uri": {"https://s3-us-west-2.amazonaws.com/hm-registration/successsignin.html"},
		"token":        {"access"},
	}

	header := http_header{
		"Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
		"User-Agent":   "MiFit/4.6.0 (iPhone; iOS 14.0.1; Scale/2.00)",
	}

	var method http_method = "POST"

	res, err := Curl(_url, strings.NewReader(data.Encode()), method, header)
	if err != nil {
		return nil, errors.New("获取token失败")
	}
	access := res.Location.Query().Get("access")

	_url = "https://account.huami.com/v2/client/login"
	//登录
	data = url.Values{
		"app_name":     {"com.xiaomi.hm.health"},
		"app_version":  {"4.6.0"},
		"code":         {access},
		"country_code": {"CN"},
		"device_id":    {"2C8B4939-0CCD-4E94-8CBA-CB8EA6E613A1"},
		"device_model": {"phone"},
		"grant_type":   {"access_token"},
		"third_name":   {"huami_phone"},
	}
	res, err = Curl(_url, strings.NewReader(data.Encode()), method, header)
	if err != nil {
		return nil, errors.New("获取登录信息失败")
	}

	err = json.Unmarshal(res.Result, &token)
	if err != nil {
		return nil, errors.New("登录信息转换map失败")
	}
	return &token, nil
}

//设置步数
func set(token proto.Token, step string) error {
	data := constant.ReqData
	data = strings.Replace(data, "__date__", time.Now().Format("2006-01-02"), -1)
	data = strings.Replace(data, "__ttl__", step, -1)
	_url := "https://api-mifit-cn.huami.com/v1/data/band_data.json?&t=" + strconv.Itoa(int(time.Now().Unix()))
	enEscapeUrl, _ := url.QueryUnescape(data)
	header := http_header{
		"Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
		"User-Agent":   "MiFit/4.6.0 (iPhone; iOS 14.0.1; Scale/2.00)",
		"apptoken":     token.TokenInfo.AppToken,
	}

	u := url.Values{
		"userid":              {token.TokenInfo.UserId},
		"last_sync_data_time": {"1597306380"},
		"device_type":         {"0"},
		"last_deviceid":       {"DA932FFFFE8888E8"},
		"data_json":           {enEscapeUrl},
	}

	res, err := Curl(_url, strings.NewReader(u.Encode()), header)
	fmt.Println(string(res.Result))
	if err != nil {
		return err
	}
	m := map[string]interface{}{}
	err = json.Unmarshal(res.Result, &m)
	if err != nil {
		return err
	}
	if m["code"].(float64) != 1 || m["message"].(string) != "success" {
		return errors.New("响应包解析失败")
	}
	return nil
}

//封装curl
func Curl(_url string, data io.Reader, options ...interface{}) (http_result, error) {
	//options -》 method string,data string,hearder map[string]string
	result := http_result{}
	//获取访问方法
	var method http_method = "POST"
	//获取头
	header := http_header{}
	for _, value := range options {
		switch value.(type) {
		case http_header:
			header = value.(http_header)
		case http_method:
			method = value.(http_method)
		default:
			break
		}

	}

	req, _ := http.NewRequest(string(method), _url, data)
	//设置请求头
	for key, value := range header {
		req.Header.Set(key, value)
	}
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse /* 不进入重定向 */
		},
	}

	resp, err := (client).Do(req)

	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	result.Result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	result.Location, _ = resp.Location()
	return result, nil
}
