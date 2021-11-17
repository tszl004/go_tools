package core_const

// 这里定义的常量，一般是具有错误代码+错误说明组成，一般用于接口返回
const (
	ProcessKilled string = "收到信号，进程被结束" // 进程被结束

	CodeErr int = 400000 // 公用错误

	ValidatorPrefix      string = "Form_Validator_" // 表单验证器前缀
	CodeValidatorFail    int    = 400001
	CodeValidatorInvalid int    = 400002
	MsgValidatorFail     string = "参数校验失败"
	MsgValidatorInvalid  string = "验证器缺失"

	CodeServerOccurred int    = 500 // 服务器代码发生错误
	MsgServerOccurred  string = "服务器内部发生代码执行错误, "

	CodeTokenOk        int = 200100 // token有效
	CodeTokenInvalid   int = 400100 // 无效的token
	CodeTokenExpired   int = 400101 // 过期的token
	CodeTokenFormatErr int = 400102 // 提交的 token 格式错误

	CodeOk int    = 200 // CURD 常用业务状态码
	MsgOk  string = "Success"

	CodeCurdCreatFail    int = 400200
	CodeCurdUpdateFail   int = 400201
	CodeCurdDeleteFail   int = 400202
	CodeCurdSelectFail   int = 400203
	CodeRefreshTokenFail int = 400206

	CodeUploadFail        int    = 400250 // 文件上传
	MsgUploadFail         string = "文件上传失败, 获取上传文件发生错误!"
	CodeUploadSizeErr     int    = 400251
	MsgUploadSizeErr      string = "长传文件超过系统设定的最大值,系统允许的最大值（M）："
	CodeUploadMimeTypeErr int    = 400252
	MsgUploadMimeTypeErr  string = "文件mime类型不允许"

	CodeWsServerNotStart int    = 400300 // websocket
	MsgWsServerNotStart  string = "websocket 服务没有开启，请在配置文件开启，相关路径：config/config.yml"
	CodeWsOpenFail       int    = 400301
	MsgWsOpenFail        string = "websocket open阶段初始化基本参数失败"

	MsgCaptchaGetParamsInvalid  string = "获取验证码：提交的验证码参数无效,请检查验证码ID以及文件名后缀是否完整" // 验证码
	CodeCaptchaGetParamsInvalid int    = 400350
	MsgCaptchaCheckInvalid      string = "校验验证码：提交的参数无效，请检查 【验证码ID、验证码值】 提交时的键名是否与配置项一致"
	CodeCaptchaCheckInvalid     int    = 400351
	MsgCaptchaCheckOk           string = "验证码校验通过"
	CodeCaptchaCheckOk          int    = 200
	CodeCaptchaCheckFail        int    = 400355
	MsgCaptchaCheckFail         string = "验证码校验失败"

	StartTimeStamp = int64(1483228800000) // 开始时间截 (2017-01-01)
	MachineIdBits  = uint(10)             // 机器id所占的位数
	SequenceBits   = uint(12)             // 序列所占的位数

	MachineIdMax   = int64(-1 ^ (-1 << MachineIdBits)) // 支持的最大机器id数量
	SequenceMask   = int64(-1 ^ (-1 << SequenceBits))  //
	MachineIdShift = SequenceBits                      // 机器id左移位数
	TimestampShift = SequenceBits + MachineIdBits      // 时间戳左移位数
)
