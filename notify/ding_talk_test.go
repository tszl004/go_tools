package notify

import "testing"

func TestDingTalk_Send(t *testing.T) {
	sender := dingTalk{Webhook: "", Secret: ""}
	ok, err := sender.Send("测试", "测试", nil)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if !ok {
		t.Fatalf("发送失败")
	}
}
