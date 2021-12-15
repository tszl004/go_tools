package notify

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/tszl004/go_tools"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// 钉钉通知渠道
type notifyChannelDingTalk struct {
	// 渠道接口地址
	Webhook string
	// 秘钥
	Secret      string
	accessToken string
	api         string
}

func (n notifyChannelDingTalk) getSign(timestamp int64) string {
	str := fmt.Sprintf("%v\n%v", timestamp, n.Secret)
	h := hmac.New(sha256.New, []byte(n.Secret))
	h.Write([]byte(str))
	// 将加密之后的字符串 base64加密 然后url encode加密
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (n notifyChannelDingTalk) Send(title, message string, content map[string]string) (res bool, err error) {
	resp := make(map[string]interface{})

	var con string
	con = "####" + message
	for _, k := range reflect.ValueOf(content).MapKeys() {
		con += fmt.Sprintf("\n%s:\t%s", k.String(), content[k.String()])
	}
	tmp := map[string]interface{}{
		"msgtype": "actionCard",
		"actionCard": map[string]interface{}{
			"title": title,
			"text":  con,
			"btns":  []int{},
		},
	}
	params, _ := json.Marshal(tmp)
	query := url.Values{}
	query.Set("access_token", n.accessToken)
	if n.Secret != "" {
		timestamp := time.Now().UnixNano() / 1e6
		query.Set("timestamp", fmt.Sprintf("%v", timestamp))
		query.Set("sign", n.getSign(timestamp))
	}
	err = tools.PostJson(n.api+query.Encode(), &resp, params, map[string]string{}, "application/json")
	if err != nil {
		return
	}
	if code, ok := resp["errcode"]; !ok || code.(float64) != 0 {
		return false, fmt.Errorf("钉钉消息发送失败 响应：%v", resp)
	}
	return true, nil
}
func (n notifyChannelDingTalk) SendHTML(title, message, content string) (res bool, err error) {

	return true, nil
}

func NewDingTalkNotify(webhook, secret string) Chan {
	accessToken := strings.Split(webhook, "access_token=")[1]
	return notifyChannelDingTalk{webhook, secret, accessToken, "https://oapi.dingtalk.com/robot/send?"}
}
