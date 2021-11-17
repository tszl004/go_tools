package tools

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func GetJson(reqUrl string, target interface{}, params url.Values, headers map[string]string) error {
	bodyBytes, _, _, err := Get(reqUrl, params, headers)
	bodyStr := string(bodyBytes)
	if bodyStr == "" {
		return errors.New("响应体为空")
	}
	err = json.Unmarshal([]byte(bodyStr), target)
	if err != nil {
		return errors.New(err.Error() + ", JSON:" + bodyStr)
	}
	// 会出现 EOF错误暂时不知道是什么原因引起的
	//err = json.NewDecoder(resp.Body).Decode(target)
	return nil
}

func Get(reqUrl string, params url.Values, headers map[string]string) (bodyBytes []byte, respHeaders http.Header, req *http.Request, err error) {
	body := strings.NewReader(params.Encode())
	reqHeader := http.Header{}
	for k := range headers {
		reqHeader.Set(k, headers[k])
	}

	return httpClient(http.MethodGet, reqUrl, body, reqHeader)
}

func httpClient(reqMethod string, reqUrl string, reqBody io.Reader, headers http.Header) (bodyBytes []byte, respHeaders http.Header, req *http.Request, err error) {
	req, err = http.NewRequest(reqMethod, reqUrl, reqBody)
	if err != nil {
		return nil, nil, req, err
	}
	req.Header = headers

	cli := http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, nil, req, err
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, req, err
	}
	return bodyBytes, resp.Header, req, err
}

func Post(reqUrl string, params interface{}, headers map[string]string, contentType string) (bodyBytes []byte, respHeaders http.Header, req *http.Request, err error) {
	var body io.Reader
	switch params.(type) {
	case url.Values:
		body = strings.NewReader(params.(url.Values).Encode())
	case string:
		body = strings.NewReader(params.(string))
	case []byte:
		body = bytes.NewBuffer(params.([]byte))
	default:
		return nil, nil, req, errors.New("invalid params must be url.Values,[]byte or string")
	}
	reqHeader := http.Header{}
	if contentType == "" {
		contentType = "application/x-www-form-urlencoded"
	}
	reqHeader.Set("Content-Type", contentType)

	for k := range headers {
		reqHeader.Set(k, headers[k])
	}
	return httpClient(http.MethodPost, reqUrl, body, reqHeader)
}

func PostJson(reqUrl string, target interface{}, params interface{}, headers map[string]string, contentType string) error {
	bodyBytes, _, req, err := Post(reqUrl, params, headers, contentType)
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
