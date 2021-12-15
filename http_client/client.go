package http_client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type client struct {
	proxyUrl string
	cli *http.Client
}

func (c client) GetJson(reqUrl string, target interface{}, params url.Values, headers map[string]string) (*http.Response, error) {
	bodyBytes, _, resp, err := c.Get(reqUrl, params, headers)
	bodyStr := string(bodyBytes)
	if bodyStr == "" {
		return resp, errors.New("响应体为空")
	}
	err = json.Unmarshal([]byte(bodyStr), target)
	if err != nil {
		return resp, errors.New(err.Error() + ", JSON:" + bodyStr)
	}
	// 会出现 EOF错误暂时不知道是什么原因引起的
	// err = json.NewDecoder(resp.Body).Decode(target)
	return resp, nil
}

func (c client) Get(reqUrl string, params url.Values, headers map[string]string) (bodyBytes []byte, req *http.Request, resp *http.Response, err error) {
	reqHeader := http.Header{}
	for k := range headers {
		reqHeader.Set(k, headers[k])
	}
	reqUrl += "?" + params.Encode()

	return c.httpClient(http.MethodGet, reqUrl, nil, reqHeader)
}

func (c client) getClient() *http.Client{
	if c.cli != nil {
		return c.cli
	}
	if c.proxyUrl != ""{
		proxy, _ := url.Parse(c.proxyUrl)
		tr := &http.Transport{
			Proxy:           http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		c.cli = &http.Client{
			Transport: tr,
			Timeout:   time.Second * 5, //超时时间
		}
	} else {
		c.cli = &http.Client{}
	}

	return c.cli
}

func (c client) httpClient(reqMethod string, reqUrl string, reqBody io.Reader, headers http.Header) (bodyBytes []byte, req *http.Request, resp *http.Response, err error) {
	req, err = http.NewRequest(reqMethod, reqUrl, reqBody)
	if err != nil {
		return nil, req, nil, err
	}
	req.Header = headers

	resp, err = c.getClient().Do(req)
	if err != nil {
		return nil, req, nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, req, resp, err
	}

	return bodyBytes, req, resp, err
}

func (c client) Post(reqUrl string, params interface{}, headers map[string]string, contentType string) (bodyBytes []byte, req *http.Request, resp *http.Response, err error) {
	var body io.Reader
	switch params.(type) {
	case url.Values:
		body = strings.NewReader(params.(url.Values).Encode())
	case string:
		body = strings.NewReader(params.(string))
	case []byte:
		body = bytes.NewBuffer(params.([]byte))
	default:
		return nil, req, resp, errors.New("invalid params must be url.Values,[]byte or string")
	}
	reqHeader := http.Header{}
	if contentType == "" {
		contentType = "application/x-www-form-urlencoded"
	}
	reqHeader.Set("Content-Type", contentType)

	for k := range headers {
		reqHeader.Set(k, headers[k])
	}
	return c.httpClient(http.MethodPost, reqUrl, body, reqHeader)
}

func (c client) PostJson(reqUrl string, target interface{}, params interface{}, headers map[string]string, contentType string) error {
	bodyBytes, req, _, err := c.Post(reqUrl, params, headers, contentType)
	bodyStr := string(bodyBytes)
	if bodyStr == "" {
		return errors.New("响应体为空")
	}
	err = json.Unmarshal([]byte(bodyStr), target)
	if err != nil {
		return errors.New("JSON 解析失败 error: " + err.Error() + "api:" + req.URL.String() + " json:" + bodyStr)
	}

	// 会出现 EOF错误暂时不知道是什么原因引起的
	// err = json.NewDecoder(resp.Body).Decode(target)
	return nil
}

func NewClient(proxy ...string) *client {
	var proxyUrl string
	if len(proxy) > 0 {
		proxyUrl = proxy[0]
	}
	return &client{proxyUrl: proxyUrl}
}