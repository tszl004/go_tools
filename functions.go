package tools

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/sftp"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"golang.org/x/crypto/ssh"
)

// MD5 md5加密
func MD5(params string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(params))
	return hex.EncodeToString(md5Ctx.Sum(nil))
}

// Base64Md5 先base64，然后MD5
func Base64Md5(params string) string {
	return MD5(base64.StdEncoding.EncodeToString([]byte(params)))
}

func GetFilesMimeByFileName(filepath string) string {
	f, err := os.Open(filepath)
	if err != nil {
		return ""
	}
	defer f.Close()

	// 只需要前 32 个字节就可以了
	buffer := make([]byte, 32)
	if _, err := f.Read(buffer); err != nil {
		return ""
	}

	return http.DetectContentType(buffer)
}

// GetFilesMimeByFp 通过文件指针获取文件mime信息
func GetFilesMimeByFp(fp multipart.File) string {
	buffer := make([]byte, 32)
	if _, err := fp.Read(buffer); err != nil {
		return ""
	}

	return http.DetectContentType(buffer)
}

func IntToDate(date int, loc *time.Location) time.Time {
	year := date / 10000
	month := date % 10000 / 100
	day := date % 100
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc)
}

func LogErr(err error) {
	if err != nil {
		fmt.Println("出错错误", err)
		log.Println("出错错误", err)
		log.Println(string(debug.Stack()))
	}
}

func Md5(str string) string {
	h := md5.New()
	_, _ = io.WriteString(h, str)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func DecodeB64(message string) string {
	base64Text := make([]byte, base64.StdEncoding.DecodedLen(len(message)))
	_, _ = base64.StdEncoding.Decode(base64Text, []byte(message))
	return string(base64Text)
}

/// 飞书通知渠道
type notifyChannelFeiShu struct {
	// 渠道接口地址
	Webhook string
}
type FeishuContentPiece struct {
	tag  string
	text string
}

func (n notifyChannelFeiShu) Send(title, message string, content map[string]string) (res bool, err error) {
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
	err = PostJson(n.Webhook, &resp, params, map[string]string{}, "application/json")
	if err != nil {
		return
	}
	if code, ok := resp["StatusCode"]; !ok || code.(float64) != 0 {
		return false, fmt.Errorf("飞书发送失败 响应：%v", resp)
	}
	return true, nil
}
func (n *notifyChannelFeiShu) feishuToken() (string, error) {
	var target = new(struct {
		Code              int
		Msg               string
		TenantAccessToken string
		Expire            int
	})
	params := url.Values{}
	params.Set("app_id", "")
	params.Set("app_secret", "")
	err := GetJson("https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal/", target, params, map[string]string{})
	if err != nil {
		return "", err
	}
	return target.TenantAccessToken, nil
}
func (n notifyChannelFeiShu) SendHTML(title, message, content string) (res bool, err error) {
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
	err = PostJson(n.Webhook, &resp, params, map[string]string{}, "application/json; charset=utf-8")
	if err != nil {
		return
	}
	if code, ok := resp["StatusCode"]; !ok || code.(float64) != 0 {
		return false, fmt.Errorf("飞书发送失败 响应：%v", resp)
	}
	return true, nil
}

/// 钉钉通知渠道
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
	err = PostJson(n.api+query.Encode(), &resp, params, map[string]string{}, "application/json")
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

func NewFeiShuNotify(ApiUrl string) NotifyChan {
	return notifyChannelFeiShu{ApiUrl}
}
func NewDingTalkNotify(webhook, secret string) NotifyChan {
	accessToken := strings.Split(webhook, "access_token=")[1]
	return notifyChannelDingTalk{webhook, secret, accessToken, "https://oapi.dingtalk.com/robot/send?"}
}

type NotifyChan interface {
	Send(title, message string, content map[string]string) (res bool, err error)
	SendHTML(title, message, content string) (res bool, err error)
}

func SliceColumn(structSlice []interface{}, key string) []interface{} {
	rt := reflect.TypeOf(structSlice)
	rv := reflect.ValueOf(structSlice)
	if rt.Kind() == reflect.Slice { //切片类型
		var sliceColumn []interface{}
		elemt := rt.Elem() //获取切片元素类型
		for i := 0; i < rv.Len(); i++ {
			inxv := rv.Index(i)
			if elemt.Kind() == reflect.Struct {
				for i := 0; i < elemt.NumField(); i++ {
					if elemt.Field(i).Name == key {
						strf := inxv.Field(i)
						switch strf.Kind() {
						case reflect.String:
							sliceColumn = append(sliceColumn, strf.String())
						case reflect.Float64:
							sliceColumn = append(sliceColumn, strf.Float())
						case reflect.Int, reflect.Int64:
							sliceColumn = append(sliceColumn, strf.Int())
						default:
							//do nothing
						}
					}
				}
			}
		}
		return sliceColumn
	}
	return nil
}

func UploadS3(filename, key, contentType string) (string, error) {
	accessKey := "AKIA43GWRT4J5RNRHPKN"
	secretKey := "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"
	//endPoint := "http://ap-southeast-1.amazonaws.com" //endpoint设置，不要动

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:    nil, //aws.String(endPoint),
		Region:      aws.String("ap-southeast-1"),
		//DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(false), //virtual-host style方式，不要修改
	})

	bucket := "okrdslog-bucket"

	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("Unable to open file %s, %+v", filename, err)
	}

	defer func() { _ = file.Close() }()

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		Body:        file,
	})
	return key, err
}

func UploadQiniu(filename string, key string) (string, error) {
	bucket := "images"
	putPolicy := storage.PutPolicy{
		Scope: bucket + ":" + key,
	}
	accessKey := "nzLnJ7bLZV2NzisSuPh4Kut0AjiIffBwlg8OlrLB"
	secretKey := "U9Mhfg39pyiozfDwhKf0D95ozARq9IJtlf-sb56u"

	mac := qbox.NewMac(accessKey, secretKey)
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	// 空间对应的机房
	cfg.Zone = &storage.ZoneHuadong
	// 是否使用https域名
	cfg.UseHTTPS = true
	// 上传是否使用CDN上传加速
	cfg.UseCdnDomains = false
	// 构建表单上传的对象
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	// 可选配置
	putExtra := storage.PutExtra{
		Params: map[string]string{},
	}
	err := formUploader.PutFile(context.Background(), &ret, upToken, key, filename, &putExtra)
	if err != nil {
		return "", err
	}
	return key, nil
}

func CreateSSHConnByPass(host, user, pass string, port ...int) *ssh.Client {
	hostPort := 22
	if len(port) > 0 {
		hostPort = port[0]
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// connect
	sshConn, err := ssh.Dial("tcp", fmt.Sprintf("%v:%v", host, hostPort), config)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return sshConn
}

func CreateSSHConnByPrivateKey(host, user, keyPath, keyPass string, port ...int) *ssh.Client {
	hostPort := 22
	if len(port) > 0 {
		hostPort = port[0]
	}
	pemBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Fatalf("unable to read private pemBytes: %v", err)
	}

	// Create the Signer for this private pemBytes.
	var signer ssh.Signer
	if keyPass != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(keyPass))
	} else {
		signer, err = ssh.ParsePrivateKey(pemBytes)
	}

	if err != nil {
		log.Fatalf("unable to parse private pemBytes: %v", err)
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// connect
	sshConn, err := ssh.Dial("tcp", fmt.Sprintf("%v:%v", host, hostPort), config)
	if err != nil {
		log.Fatalf("dial fail:%+v", err)
		return nil
	}
	return sshConn
}

func SSHDownload(path, dest string, conn *ssh.Client) error {

	// create new SFTP client
	client, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	// open remote file
	srcFile, err := client.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()
	log.Println("打开远程文件成功", path)

	// create dst file
	dstFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	// copy source file to destination file
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	log.Printf("复制文件内容成功 %s -> %s\n", path, dest)

	// flush in-memory copy
	err = dstFile.Sync()
	if err != nil {
		return err
	}
	return nil
}

func ParseToUrlValues(i interface{}) (values url.Values) {
	values = url.Values{}
	iVal := reflect.ValueOf(i).Elem()
	typ := iVal.Type()
	for i := 0; i < iVal.NumField(); i++ {
		f := iVal.Field(i)
		// You ca use tags here...
		// tag := typ.Field(i).Tag.Get("tagname")
		// Convert each type into a string for the url.Values string map
		var v string
		switch f.Interface().(type) {
		case int, int8, int16, int32, int64:
			v = strconv.FormatInt(f.Int(), 10)
		case uint, uint8, uint16, uint32, uint64:
			v = strconv.FormatUint(f.Uint(), 10)
		case float32:
			v = strconv.FormatFloat(f.Float(), 'f', 4, 32)
		case float64:
			v = strconv.FormatFloat(f.Float(), 'f', 4, 64)
		case []byte:
			v = string(f.Bytes())
		case string:
			v = f.String()
		}
		values.Set(typ.Field(i).Name, v)
	}
	return
}
