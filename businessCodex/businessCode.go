package businessCodex

// Response 统一返回结构. 错误时返回code和msg, 正确时返回data.
type Response struct {
	Code int         `json:"code"` // 业务码
	Msg  string      `json:"msg"`  // 描述信息
	Data interface{} `json:"data"` // 返回值
}

// Failure 错误时返回结构 (保留type, 增加兼容性).
type Failure Response

const (
	// ZhCN 简体中文 - 中国
	ZhCN = "zh-cn"

	// EnUS 英文 - 美国
	EnUS = "en-us"
)

var (
	succeedCode = 1000000000
)

func SetSucceedCode(code int) {
	succeedCode = code
}

func GetSucceedCode() (code int) {
	code = succeedCode
	return
}

var (
	// ServerError 110001 程序发生未知 panic 被框架捕捉到的时候.
	serverErrorCode = succeedCode + 1
	// TooManyRequests 110002 开启了框架的限速器被限速了
	tooManyRequestsCode = serverErrorCode + 1
	// ParamBindError 110003 使用框架序列化参数时报错, 一般是前端参数传错了.
	paramBindErrorCode = tooManyRequestsCode + 1
	// MySQLExecError 110004 数据库执行时报错, 一般用作未知的SQL执行异常.
	mySQLExecErrorCode = paramBindErrorCode + 1
)

func SetServerErrorCode(code int) {
	serverErrorCode = code
}

func GetServerErrorCode() (code int) {
	code = serverErrorCode
	return
}

func SetTooManyRequestsCode(code int) {
	tooManyRequestsCode = code
}

func GetTooManyRequestsCode() (code int) {
	code = tooManyRequestsCode
	return
}

func SetParamBindErrorCode(code int) {
	paramBindErrorCode = code
}

func GetParamBindErrorCode() (code int) {
	code = paramBindErrorCode
	return
}

func SetMySQLExecErrorCode(code int) {
	mySQLExecErrorCode = code
}

func GetMySQLExecErrorCode() (code int) {
	code = mySQLExecErrorCode
	return
}

var lang string

func SetLang(l string) {
	lang = l
}

var (
	zhCnTextMap  = make(map[int]string)
	enUsTextMap  = make(map[int]string)
	return401Map = make(map[int]struct{})
)

func Init(enUS bool) {
	if enUS {
		SetLang(EnUS)
	} else {
		SetLang(ZhCN)
	}
	SetZhCNText(zhCNText())
	SetEnUSText(enUSText())

}

func SetZhCNText(textMap map[int]string) {
	for k, v := range textMap {
		zhCnTextMap[k] = v
	}
}
func GetZhCNText() map[int]string {
	return zhCnTextMap
}

func SetEnUSText(textMap map[int]string) {
	for k, v := range textMap {
		enUsTextMap[k] = v
	}
}

func GetEnUSText() map[int]string {
	return enUsTextMap
}

func SetReturn401Map(intMap map[int]struct{}) {
	for k, v := range intMap {
		return401Map[k] = v
	}
}

func GetReturn401Map() map[int]struct{} {
	return return401Map
}

func Text(code int) string {
	if lang == ZhCN {
		return GetZhCNText()[code]
	}

	if lang == EnUS {
		return GetEnUSText()[code]
	}

	return GetZhCNText()[code]
}
