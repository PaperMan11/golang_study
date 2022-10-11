package geecache

import (
	"GeeCache/geecache/singleflight"
	"fmt"
	"log"
	"sync"
)

/*

   定义接口 Getter 和 回调函数 Get(key string)([]byte, error)，参数是 key，返回值是 []byte。
   定义函数类型 GetterFunc，并实现 Getter 接口的 Get 方法。
   函数类型实现某一个接口，称之为接口型函数，方便使用者在调用时既能够传入函数作为参数，也能够传入实现了该接口的结构体作为参数。

*/

// Getter为键加载数据
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc通过函数实现Getter
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

/*

   一个 Group 可以认为是一个缓存的命名空间，每个 Group 拥有一个唯一的名称 name。比如可以创建三个 Group，缓存学生的成绩命名为 scores，缓存学生信息的命名为 info，缓存学生课程的命名为 courses。
   第二个属性是 getter Getter，即缓存未命中时获取源数据的回调(callback)。
   第三个属性是 mainCache cache，即一开始实现的并发缓存。
   构建函数 NewGroup 用来实例化 Group，并且将 group 存储在全局变量 groups 中。
   GetGroup 用来特定名称的 Group，这里使用了只读锁 RLock()，因为不涉及任何冲突变量的写操作。

*/

// Group 是一个缓存命名空间，并将加载的相关数据分散开来
type Group struct {
	name      string
	getter    Getter // 回调函数
	mainCache cache
	peers     PeerPicker
	loader    *singleflight.Group // 确保相同的 key 只调用一次
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup 创建 Group 的新实例
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// GetGroup 返回先前用 NewGroup 创建的命名组 or nil
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

/*
                            是
接收 key --> 检查是否被缓存 -----> 返回缓存值 ⑴
                |  否                         是
                |-----> 是否应当从远程节点获取 -----> 与远程节点交互 --> 返回缓存值 ⑵
                            |  否
                            |-----> 调用`回调函数`，获取值并添加到缓存 --> 返回缓存值 ⑶
*/
/*

   Get 方法实现了上述所说的流程 ⑴ 和 ⑶。
   流程 ⑴ ：从 mainCache 中查找缓存，如果存在则返回缓存值。
   流程 ⑶ ：缓存不存在，则调用 load 方法，load 调用 getLocally（分布式场景下会调用 getFromPeer 从其他节点获取），
   			getLocally 调用用户回调函数 g.getter.Get() 获取源数据，并且将源数据添加到缓存 mainCache 中（通过 populateCache 方法）

*/
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	return g.load(key)
}

// RegisterPeers 注册一个PeerPicker (type HTTPPool)
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error) {
	// 在高并发时，确保每个键只获得一次（远程、本地）
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

// 从远端机器获取 key
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key) // 发送 http 请求获取缓存
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

// 从本地数据库获取 key
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err

	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
