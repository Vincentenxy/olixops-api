package httpx

const (
	CodeFail int = -1

	// 成功
	CodeSuccess = 0

	// 参数错误
	CodeInvalidParam = 400

	// 未登录
	CodeUnauthorized = 401

	// 权限不足
	CodeForbidden = 403

	// 系统异常
	CodeSystemError = 500
)

type codeMeta struct {
	bizCode int
	msg     string
}

var metaMap = map[int]codeMeta{
	CodeSuccess: {
		bizCode: CodeSuccess,
		msg:     "success",
	},

	CodeFail: {
		bizCode: CodeFail,
		msg:     "fail",
	},

	CodeInvalidParam: {
		bizCode: CodeInvalidParam,
		msg:     "invalid params",
	},

	CodeUnauthorized: {
		bizCode: CodeUnauthorized,
		msg:     "unauthorized",
	},

	CodeForbidden: {
		bizCode: CodeForbidden,
		msg:     "forbidden",
	},

	CodeSystemError: {
		bizCode: CodeSystemError,
		msg:     "system error",
	},
}

func GetMeta(bizCode int) codeMeta {
	m, ok := metaMap[bizCode]
	if !ok {
		return metaMap[CodeSystemError]
	}
	return m
}
