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

type Client struct {
	proxyUrl    string
	cli         *http.Client
	notRedirect bool
}

func (c Client) GetJson(reqUrl string, target interface{}, params url.Values, headers map[string]string) (*http.Response, error) {
	bodyBytes, _, resp, err := c.Get(reqUrl, params, headers)
	if len(bodyBytes) == 0 {
		return resp, errors.Wrapf(err, "响应体为空")
	}
	err = json.Unmarshal(bodyBytes, target)
	if err != nil {
		return resp, errors.New(err.Error() + ", JSON:" + string(bodyBytes))
	}
	// 会出现 EOF错误暂时不知道是什么原因引起的
	// err = json.NewDecoder(resp.Body).Decode(target)
	return resp, nil
}

func (c Client) Get(reqUrl string, params url.Values, headers map[string]string) (bodyBytes []byte, req *http.Request, resp *http.Response, err error) {
	reqHeader := http.Header{}
	for k := range headers {
		reqHeader.Set(k, headers[k])
	}
	req, err = http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, req, nil, err
	}
	req.URL, err = ParseUrl(reqUrl, params)
	if err != nil {
		return nil, req, nil, err
	}

	return c.httpClient(req, nil, reqHeader)
}

func ParseUrl(reqUrl string, params url.Values) (*url.URL, error) {
	reqUrlObj, err := url.Parse(reqUrl)
	if err != nil {
		return reqUrlObj, err
	}
	urlParams := reqUrlObj.Query()
	if params != nil {
		for k, v := range params {
			urlParams[k] = v
		}
		reqUrlObj.RawQuery = urlParams.Encode()
	}

	return reqUrlObj, nil
}

func (c Client) getClient() *http.Client {
	if c.cli != nil {
		return c.cli
	}
	if c.proxyUrl != "" {
		proxy, _ := url.Parse(c.proxyUrl)
		tr := &http.Transport{
			Proxy:           http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		c.cli = &http.Client{
			Transport: tr,
			Timeout:   time.Second * 5, // 超时时间
		}
	} else {
		c.cli = &http.Client{}
	}
	if c.notRedirect {
		c.cli.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return NotRedirectErr
		}
	}

	return c.cli
}

func (c Client) httpClient(httpReq *http.Request, reqBody io.Reader, headers http.Header) (bodyBytes []byte, req *http.Request, resp *http.Response, err error) {
	rc, ok := reqBody.(io.ReadCloser)
	if !ok && reqBody != nil {
		rc = io.NopCloser(reqBody)
	}
	httpReq.Body = rc
	httpReq.Header = headers

	resp, err = c.getClient().Do(httpReq)
	// Not follow redirect err Skip
	if err != nil && strings.Index(err.Error(), NotRedirectErr.Error()) < 0 {
		return nil, httpReq, resp, err
	}
	defer func() { _ = resp.Body.Close() }()
	bodyBytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, httpReq, resp, err
	}

	return bodyBytes, httpReq, resp, err
}

func (c Client) Post(reqUrl string, params interface{}, headers map[string]string, contentType string) (bodyBytes []byte, req *http.Request, resp *http.Response, err error) {
	var body io.Reader
	switch params.(type) {
	case url.Values:
		body = strings.NewReader(params.(url.Values).Encode())
	case string:
		body = strings.NewReader(params.(string))
	case []byte:
		body = bytes.NewBuffer(params.([]byte))
	case nil:

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
	req, err = http.NewRequest(http.MethodPost, reqUrl, body)
	if err != nil {
		return nil, nil, nil, err
	}
	return c.httpClient(req, body, reqHeader)
}

func (c Client) PostJson(reqUrl string, target interface{}, params interface{}, headers map[string]string, contentType string) error {
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

func (c Client) FollowRedirect(redirect bool) {
	c.notRedirect = redirect
}

func NewClient(proxy ...string) *Client {
	var proxyUrl string
	if len(proxy) > 0 {
		proxyUrl = proxy[0]
	}
	return &Client{proxyUrl: proxyUrl}
}

func NewNotRedirectClient(proxy ...string) *Client {
	cli := NewClient(proxy...)
	cli.notRedirect = true
	return cli
}
