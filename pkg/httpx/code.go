package httpx

const (
	// 成功
	Ok int = 0

	// 参数错误
	InvalidParam = 400

	// 未登录
	Unauthorized = 401

	// 权限不足
	Forbidden = 403

	// 系统异常
	SystemError = 500
)

type codeMeta struct {
	bizCode int
	msg     string
}

var metaMap = map[int]codeMeta{
	Ok: {
		bizCode: Ok,
		msg:     "success",
	},

	InvalidParam: {
		bizCode: InvalidParam,
		msg:     "invalid params",
	},

	Unauthorized: {
		bizCode: Unauthorized,
		msg:     "unauthorized",
	},

	Forbidden: {
		bizCode: Forbidden,
		msg:     "forbidden",
	},

	SystemError: {
		bizCode: SystemError,
		msg:     "system error",
	},
}

func GetMeta(bizCode int) codeMeta {
	m, ok := metaMap[bizCode]
	if !ok {
		return metaMap[SystemError]
	}
	return m
}
