package tools

import (
	"net/http"
	"net/url"
	"strings"
)

func GetJson(reqUrl string, target interface{}, params url.Values, headers map[string]string) (*http.Response, error) {
	return httpCli.GetJson(reqUrl, target, params, headers)
}

func Get(reqUrl string, params url.Values, headers map[string]string) (bodyBytes []byte, req *http.Request, resp *http.Response, err error) {
	return httpCli.Get(reqUrl, params, headers)
}

func Post(reqUrl string, params interface{}, headers map[string]string, contentType string) (bodyBytes []byte, req *http.Request, resp *http.Response, err error) {
	return httpCli.Post(reqUrl, params, headers, contentType)
}

func PostJson(reqUrl string, target interface{}, params interface{}, headers map[string]string, contentType string) error {
	return httpCli.PostJson(reqUrl, target, params, headers, contentType)
}

func GetCookieByRespHeader(header http.Header) string {
	cookie := ""
	for _, item := range header.Values("Set-Cookie") {
		if strings.Index(item, "=deleted;") > 0 {
			continue
		}
		cookie += strings.Split(item, ";")[0] + ";"
	}
	return cookie
}
