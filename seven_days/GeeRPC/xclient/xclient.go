package xclient

import (
	. "GeeRPC" // 导入包在该作用域
	"context"
	"io"
	"reflect"
	"sync"
)

type XClient struct {
	d       Discover
	mode    SelectMode // 负载均衡模式
	opt     *Option
	mu      sync.Mutex
	clients map[string]*Client
}

var _ io.Closer = (*XClient)(nil)

func NewXClient(d Discover, mode SelectMode, opt *Option) *XClient {
	return &XClient{
		d:       d,
		mode:    mode,
		opt:     opt,
		clients: make(map[string]*Client),
	}
}

func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	for key, client := range xc.clients {
		client.Close()
		delete(xc.clients, key)
	}
	return nil
}

func (xc *XClient) dial(rpcaddr string) (*Client, error) {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	client, ok := xc.clients[rpcaddr] // 检查是否有缓存的服务器
	if ok && !client.IsAvailable() {
		client.Close()
		delete(xc.clients, rpcaddr)
		client = nil
	}
	if client == nil {
		var err error
		client, err = XDial(rpcaddr, xc.opt)
		if err != nil {
			return nil, err
		}
		xc.clients[rpcaddr] = client
	}
	return client, nil
}

func (xc *XClient) call(rpcAddr string, ctx context.Context, serviceMethod string, args, reply interface{}) error {
	client, err := xc.dial(rpcAddr)
	if err != nil {
		return err
	}
	return client.Call(ctx, serviceMethod, args, reply)
}

func (xc *XClient) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	// Xc将选择一个合适的服务器
	rpcAddr, err := xc.d.Get(xc.mode)
	if err != nil {
		return err
	}
	return xc.call(rpcAddr, ctx, serviceMethod, args, reply)
}

// Broadcast 将请求广播到所有的服务实例，
// 如果任意一个实例发生错误，则返回其中一个错误；
// 如果调用成功，则返回其中一个的结果。
func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	servers, err := xc.d.GetAll()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var e error
	replyDone := reply == nil // 如果reply为nil，不需要设置值。
	ctx, cancel := context.WithCancel(ctx)
	for _, rpcAddr := range servers {
		wg.Add(1)
		go func(rpcAddr string) {
			defer wg.Done()
			var clonedReply interface{}
			if reply != nil {
				clonedReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			err := xc.call(rpcAddr, ctx, serviceMethod, args, clonedReply) // 调用
			mu.Lock()
			if err != nil && e == nil {
				e = err
				cancel() // 出错，取消所有 call
			}
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(clonedReply).Elem())
				replyDone = true
			}
			mu.Unlock()
		}(rpcAddr)
	}
	wg.Wait()
	return e
}
