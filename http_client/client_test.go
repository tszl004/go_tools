package http_client

import (
	"fmt"
	"net/url"
	"testing"
)

func TestParseUrl(t *testing.T) {
	reqUrl , err:= ParseUrl("https://www.baidu.com/index.php?abc=1&bbc[0]=2", url.Values{
		"bbc":{"3"},
		"cbc":{"0"},
	})
	if err != nil {
		t.Fatalf("%+v", err)
	}
	fmt.Println(reqUrl)
}
