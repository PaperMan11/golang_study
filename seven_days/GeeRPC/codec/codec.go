package codec

import "io"

type Header struct {
	ServiceMethod string // format "Service.Method"
	Seq           uint64 // 序号，某个请求的ID，用于区分不同的请求
	Error         string // 错误信息
}

type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

type NewCodecFunc func(io.ReadWriteCloser) Codec
type Type string

// 我们定义了 2 种 Codec，Gob 和 Json，
// 但是实际代码中只实现了 Gob 一种，事实上，2 者的实现非常接近，
// 甚至只需要把 gob 换成 json 即可。
const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json" // 未实现
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
