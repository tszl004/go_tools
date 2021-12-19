package tools

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/tszl004/go_tools/core_vars"
	"github.com/tszl004/go_tools/http_client"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"golang.org/x/crypto/ssh"
)

var (
	httpCli = http_client.NewClient()
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

func SliceColumn(structSlice []interface{}, key string) []interface{} {
	rt := reflect.TypeOf(structSlice)
	rv := reflect.ValueOf(structSlice)
	if rt.Kind() == reflect.Slice { // 切片类型
		var sliceColumn []interface{}
		elemt := rt.Elem() // 获取切片元素类型
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
							// do nothing
						}
					}
				}
			}
		}
		return sliceColumn
	}
	return nil
}

// UploadS3
// Deprecated: 这里暂不可使用
func UploadS3(filename, key, contentType string) (string, error) {
	// todo s3 代码有问题暂时不用
	// accessKey := "AKIA43GWRT4J5RNRHPKN"
	// secretKey := "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"
	// endPoint := "http://ap-southeast-1.amazonaws.com" //endpoint设置，不要动

	// sess, err := session.NewSession(&aws.Config{
	// 	Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	// 	Endpoint:    nil, //aws.String(endPoint),
	// 	Region:      aws.String("ap-southeast-1"),
	// 	//DisableSSL:       aws.Bool(true),
	// 	S3ForcePathStyle: aws.Bool(false), //virtual-host style方式，不要修改
	// })

	// bucket := "okrdslog-bucket"

	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("Unable to open file %s, %+v", filename, err)
	}

	defer func() { _ = file.Close() }()

	// uploader := s3manager.NewUploader(sess)
	// _, err = uploader.Upload(&s3manager.UploadInput{
	// 	Bucket:      aws.String(bucket),
	// 	Key:         aws.String(key),
	// 	ContentType: aws.String(contentType),
	// 	Body:        file,
	// })
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

func ParseInt(s string) int {
	s = strings.ReplaceAll(s, ",", "")
	reg, _ := regexp.Compile(`^\d+`)
	numStr := reg.FindString(s)
	num, _ := strconv.Atoi(numStr)
	return num
}

func ParseFloat64(s string) float64 {
	f, _ := strconv.ParseFloat(strings.ReplaceAll(s, ",", ""), 64)
	return f
}

func SplitToSliceInt(s, sep string) []int {
	res := make([]int, 0)
	for _, numStr := range strings.Split(s, sep) {
		res = append(res, ParseInt(numStr))
	}
	return res
}

func RegexpMatchFirstStr(s, expr string) string {
	r, err := regexp.Compile(expr)
	if err != nil {
		return ""
	}
	return r.FindString(s)
}

func RegexMatchAll(s, expr string) []string {
	r, err := regexp.Compile(expr)
	if err != nil {
		return nil
	}
	return r.FindAllString(s, -1)
}

func randSeedReset() {
	rand.Seed(time.Now().UnixNano())
}

func RandIntN(n int) int {
	randSeedReset()
	return rand.Intn(n)
}

func RandRange(min, max int) int {
	return rand.Intn(max-min) + min
}

// GbkToUtf8 GBK 转 UTF-8
func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func NumToNumByteSlice(num int) []int {
	res := make([]int, 0)
	for num > 0 {
		res = append(res, num%10)
		num /= 10
	}
	return ReverseIntSlice(res)
}

func ReverseIntSlice(list []int) []int {
	length := len(list)
	res := make([]int, length)
	for i := 0; i < length; i++ {
		res[length-1-i] = list[i]
	}
	return res
}

func FileGetContents(filePath string) ([]byte, error) {
	filename, _ := filepath.Abs(filePath)
	fileCon, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}
	return fileCon, nil
}

func Tomorrow(locArgs ...*time.Location) time.Time {
	var loc *time.Location
	if len(locArgs) == 0 {
		loc = core_vars.RPCLoc
	} else {
		loc = locArgs[0]
	}
	tTime := time.Now().In(loc).Add(24 * time.Hour).Format(core_vars.DateLayout)
	tomorrow, _ := time.ParseInLocation(core_vars.DateLayout, tTime, loc)
	return tomorrow
}
