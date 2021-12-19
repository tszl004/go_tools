package notify

import (
	"encoding/json"
	"fmt"
	"github.com/tszl004/go_tools"
	"net/url"
	"reflect"
)

// 飞书通知渠道
type feiShu struct {
	// 渠道接口地址
	Webhook string
}

type feiShuContentPiece struct {
	tag  string
	text string
}

func (n feiShu) Send(title, message string, content map[string]string) (err error) {
	resp := make(map[string]interface{})
	var con [][]map[string]string
	con = append(con, []map[string]string{{"tag": "text", "text": message}})
	for _, k := range reflect.ValueOf(content).MapKeys() {
		con = append(con, []map[string]string{{"tag": "text", "text": k.String() + content[k.String()]}})
	}
	tmp := map[string]interface{}{
		"msg_type": "post",
		"content": map[string]interface{}{
			"post": map[string]interface{}{
				"zh_cn": map[string]interface{}{
					"title":   title,
					"content": con,
				},
			},
		},
	}
	params, _ := json.Marshal(tmp)
	err = tools.PostJson(n.Webhook, &resp, params, map[string]string{}, "application/json")
	if err != nil {
		return
	}
	if code, ok := resp["StatusCode"]; !ok || code.(float64) != 0 {
		return fmt.Errorf("飞书发送失败 响应：%v", resp)
	}
	return nil
}

func (n *feiShu) token() (string, error) {
	var target = new(struct {
		Code              int
		Msg               string
		TenantAccessToken string
		Expire            int
	})
	params := url.Values{}
	params.Set("app_id", "")
	params.Set("app_secret", "")
	_, err := tools.GetJson("https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal/", target, params, map[string]string{})
	if err != nil {
		return "", err
	}
	return target.TenantAccessToken, nil
}

func (n feiShu) SendHTML(title, message, content string) (err error) {
	resp := make(map[string]interface{})
	var con [][]map[string]string
	con = append(con, []map[string]string{{"tag": "text", "text": message}, {"tag": "code", "text": content}})
	tmp := map[string]interface{}{
		"msg_type": "post",
		"content": map[string]interface{}{
			"post": map[string]interface{}{
				"zh_cn": map[string]interface{}{
					"title":   title,
					"content": con,
				},
			},
		},
	}
	params, _ := json.Marshal(tmp)
	err = tools.PostJson(n.Webhook, &resp, params, map[string]string{}, "application/json; charset=utf-8")
	if err != nil {
		return
	}
	if code, ok := resp["StatusCode"]; !ok || code.(float64) != 0 {
		return fmt.Errorf("飞书发送失败 响应：%v", resp)
	}
	return nil
}
