package notify
// Chan 通知渠道接口
type Chan interface {
	Send(title, message string, content map[string]string) (res bool, err error)
	SendHTML(title, message, content string) (res bool, err error)
}
