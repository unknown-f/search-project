package errmsg

const (
	SUCCESS = 200
	ERROR   = 500

	// code = 1000.. 用户模块的错误码
	ERROR_USERNAME_USED  = 1001
	ERROR_PASSWORD_WRONG = 1002
	ERROR_USER_NOT_EXIST = 1003
	ERROR_TOKEN_EXIST    = 1004
	ERROR_TOKEN_RUNTIME  = 1005
	ERROR_TOKEN_WRONG    = 1006
	ERROR_TOKEN_TYPE     = 1007

	// code = 2000.. Link 模块的错误码
	ERROR_LINKNAME_USED = 2001

	// code = 3000.. 收藏夹模块的错误码
	ERROR_FAVORITENAME_USED = 3001
)

var codeMsg = map[int]string{
	SUCCESS:              "OK",
	ERROR:                "FAIL",
	ERROR_USERNAME_USED:  "用户名已存在",
	ERROR_PASSWORD_WRONG: "密码错误",
	ERROR_USER_NOT_EXIST: "用户不存在",
	ERROR_TOKEN_EXIST:    "Token不存在",
	ERROR_TOKEN_RUNTIME:  "Token已过期",
	ERROR_TOKEN_WRONG:    "Token不正确",
	ERROR_TOKEN_TYPE:     "Token格式错误",

	ERROR_LINKNAME_USED: "该链接名已使用",

	ERROR_FAVORITENAME_USED: "该收藏夹已存在",
}

func GetErrMsg(code int) string {
	return codeMsg[code]
}
