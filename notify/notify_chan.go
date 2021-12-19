package notify

import (
	"errors"

	"github.com/mitchellh/mapstructure"
)

// Sender 通知渠道接口
type Sender interface {
	Send(title, message string, content map[string]string) error
	SendHTML(title, message, content string) error
}
type DriverType string

var (
	ErrDriverExists            = errors.New("driver doesn't exists")
	DriverDingTalk  DriverType = "DingTalk"
	DriverFeiShu    DriverType = "FeiShu"
	driverMap                  = map[DriverType]Sender{}
)

func SetSenderCfg(driver DriverType, config map[string]string) (sender Sender, err error) {
	switch driver {
	case DriverFeiShu:
		sender = feiShu{}
	case DriverDingTalk:
		sender = dingTalk{}
	}
	err = mapstructure.Decode(config, &sender)
	if err != nil {
		return nil, err
	}
	driverMap[driver] = sender
	return sender, nil
}

func GetSender(driver DriverType) (Sender, error) {
	sender, ok := driverMap[driver]
	if !ok {
		return nil, ErrDriverExists
	}
	return sender, nil
}
