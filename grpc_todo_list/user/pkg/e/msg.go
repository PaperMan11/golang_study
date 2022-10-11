package e

var MsgFlags = map[uint]string{
	Success:       "ok",
	Error:         "fail",
	InvalidParams: "请求参数错误",
}

// 获取状态码对应的信息
func GetMsg(code uint) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	return MsgFlags[Error]
}
