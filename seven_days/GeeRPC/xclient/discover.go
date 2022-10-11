package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

const (
	RandomSelect     SelectMode = iota // 随机
	RoundRobinSelect                   // 轮询
)

type Discover interface {
	Refresh() error // 从远程注册中心刷新
	Update(servers []string) error
	Get(mode SelectMode) (string, error)
	GetAll() ([]string, error)
}

// MultiServersDiscovery 是对没有注册中心的多服务器的发现
// 用户显式地提供服务器地址
type MultiServersDiscovery struct {
	r       *rand.Rand // 生成随机数
	mu      sync.RWMutex
	servers []string
	index   int // 记录所选位置
}

var _ Discover = (*MultiServersDiscovery)(nil)

func NewMutiServersDiscovery(servers []string) *MultiServersDiscovery {
	d := &MultiServersDiscovery{
		servers: servers,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	d.index = d.r.Intn(math.MaxInt32 - 1) // 避免每次从0开始
	return d
}

// Refresh 对MultiServersDiscovery没有意义，所以忽略它
func (d *MultiServersDiscovery) Refresh() error {
	return nil
}

// Update 动态发现服务器
func (d *MultiServersDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	return nil
}

func (d *MultiServersDiscovery) Get(mode SelectMode) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	n := len(d.servers)
	if n == 0 {
		return "", errors.New("rpc discovery: no available servers")
	}

	switch mode {
	case RandomSelect:
		return d.servers[d.r.Intn(n)], nil
	case RoundRobinSelect:
		s := d.servers[d.index%n] // servers 可能更新，所以mod
		d.index = (d.index + 1) % n
		return s, nil
	default:
		return "", errors.New("rpc discovery: not supported select mode")
	}
}

// GetAll 返回所有服务器
func (d *MultiServersDiscovery) GetAll() ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	servers := make([]string, len(d.servers), len(d.servers))
	copy(servers, d.servers)
	return servers, nil
}
