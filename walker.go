package walker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/WeiJiadong/walker/internal/base"
	"github.com/WeiJiadong/walker/internal/constant"
	"github.com/WeiJiadong/walker/internal/proto"
)

// walker ...
type walker struct {
	opts walkerOpts
}

type walkerOpts struct {
	uid    string
	passwd string
	step   string
	plat   string
}

// WalkerOpt walkerOpts helper
type WalkerOpt func(opts *walkerOpts)

// WithUid 设置uid
func WithUid(uid string) WalkerOpt {
	return func(opts *walkerOpts) {
		opts.uid = uid
		opts.plat = "email"
		if !strings.Contains(uid, "@") {
			opts.uid = "+86" + uid
			opts.plat = "huami_phone"
		}
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
	access, err := w.getAccess()
	if err != nil {
		log.Fatalln("getAccess error, err:", err)
		return err
	}

	token, err := w.getToken(access)
	if err != nil {
		log.Fatalln("getToken error, err:", err)
		return err
	}

	if err := w.setStep(token); err != nil {
		log.Fatalln("setStep error, err:", err)
		return err
	}

	return nil
}

func (w *walker) getAccess() (string, error) {
	u := url.Values{
		"client_id":    {"HuaMi"},
		"password":     {w.opts.passwd},
		"redirect_uri": {"https://s3-us-west-2.amazonaws.com/hm-registration/successsignin.html"},
		"token":        {"access"},
	}
	req, err := http.NewRequest("POST", base.GenAccessUrl(w.opts.uid), strings.NewReader(u.Encode()))
	if err != nil {
		log.Fatalln("NewRequest error, err:", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Set("User-Agent", "MiFit/4.6.0 (iPhone; iOS 14.0.1; Scale/2.00)")

	cli := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse /* 不进入重定向 */
		}}
	rsp, err := cli.Do(req)
	if err != nil {
		log.Fatalln("Do error, err:", err)
		return "", err
	}
	local, err := rsp.Location()
	if err != nil {
		log.Fatalln("Location error, err:", err)
		return "", err
	}
	return local.Query().Get("access"), nil
}

func (w *walker) getToken(access string) (*proto.Token, error) {
	u := url.Values{
		"app_name":     {"com.xiaomi.hm.health"},
		"app_version":  {"4.6.0"},
		"code":         {access},
		"country_code": {"CN"},
		"device_id":    {"2C8B4939-0CCD-4E94-8CBA-CB8EA6E613A1"},
		"device_model": {"phone"},
		"grant_type":   {"access_token"},
		"third_name":   {w.opts.plat},
	}
	req, err := http.NewRequest("POST", base.GenLoginUrl(), strings.NewReader(u.Encode()))
	if err != nil {
		log.Fatalln("NewRequest error, err:", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Set("User-Agent", "MiFit/4.6.0 (iPhone; iOS 14.0.1; Scale/2.00)")
	cli := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse /* 不进入重定向 */
		}}
	rsp, err := cli.Do(req)
	if err != nil {
		log.Fatalln("http client Do error, err:", err)
		return nil, err
	}

	result, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		log.Fatalln("ReadAll error, err:", err)
		return nil, err
	}

	token := proto.Token{}
	if err = json.Unmarshal(result, &token); err != nil {
		log.Fatalln("Unmarshal error, err:", err)
		return nil, err
	}

	return &token, nil
}

func (w *walker) setStep(token *proto.Token) (err error) {
	data := constant.ReqData
	data = strings.Replace(data, "__date__", time.Now().Format("2006-01-02"), -1)
	data = strings.Replace(data, "__ttl__", w.opts.step, -1)

	enEscapeUrl, _ := url.QueryUnescape(data)
	u := url.Values{
		"userid":              {token.TokenInfo.UserId},
		"last_sync_data_time": {"1597306380"},
		"device_type":         {"0"},
		"last_deviceid":       {"DA932FFFFE8888E8"},
		"data_json":           {enEscapeUrl},
	}
	req, err := http.NewRequest("POST", base.GenSetStepUrl(), strings.NewReader(u.Encode()))
	if err != nil {
		log.Fatalln("NewRequest error, err:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Set("User-Agent", "MiFit/4.6.0 (iPhone; iOS 14.0.1; Scale/2.00)")
	req.Header.Set("apptoken", token.GetTokenInfo().GetAppToken())

	cli := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse /* 不进入重定向 */
		}}
	rsp, err := cli.Do(req)
	if err != nil {
		log.Fatalln("http client Do error, err:", err)
		return err
	}

	result, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		log.Fatalln("ReadAll error, err:", err)
		return err
	}

	ssRsp := proto.SetStepRsp{}
	err = json.Unmarshal(result, &ssRsp)
	if err != nil {
		log.Fatalln("Unmarshal error, err:", err)
		return err
	}
	if ssRsp.GetCode() != 1 {
		log.Fatalln("GetCode error, err:", ssRsp)
		return fmt.Errorf("%+v", ssRsp)
	}
	return nil
}
