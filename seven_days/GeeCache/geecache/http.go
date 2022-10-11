package geecache

import (
	"GeeCache/geecache/consistenthash"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache"
	defaultReplicas = 50 // 虚拟节点个数
)

/*

   1、HTTPPool 只有 2 个参数，一个是 self，用来记录自己的地址，包括主机名/IP 和端口。
   2、另一个是 basePath，作为节点间通讯地址的前缀，默认是 /_geecache/，
   	那么 http://example.com/_geecache/ 开头的请求，就用于节点间的访问。
   	因为一个主机上还可能承载其他的服务，加一段 Path 是一个好习惯。
   	比如，大部分网站的 API 接口，一般以 /api 作为前缀。

*/
// HTTPPool 为一个 HTTP 对等体池实现了 PeerPicker。
type HTTPPool struct {
	self        string
	basePath    string
	mu          sync.Mutex
	peers       *consistenthash.Map    // 用来根据具体的 key 选择节点
	httpGetters map[string]*httpGetter // keyed by e.g. "http://10.0.0.2:8008"，所需要的 key 与机器地址的映射
}

// NewHTTPPool initializes an HTTP pool of peers.
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log info with server name
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

/*

   1、ServeHTTP 的实现逻辑是比较简单的，首先判断访问路径的前缀是否是 basePath，不是返回错误。
   2、我们约定访问路径格式为 /<basepath>/<groupname>/<key>，通过 groupname 得到 group 实例，再使用 group.Get(key) 获取缓存数据。
   3、最终使用 w.Write() 将缓存值作为 httpResponse 的 body 返回。

*/
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath)+1:], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

// Set 一致性 hash 实例，将节点加入环中
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer 根据 key 选择一个 peer （节点）
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

// 这是确保接口被实现常用的方式。即利用强制类型转换，确保 struct HTTPPool 实现了接口 PeerPicker。
// 这样 IDE 和编译期间就可以检查，而不是等到使用的时候。
var _ PeerPicker = (*HTTPPool)(nil)

/********************************Clinet*************************************/
type httpGetter struct {
	baseURL string // 例如 http://example.com/_geecache/
}

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v%v",
		h.baseURL,
		url.QueryEscape(group), // QueryEscape函数对s进行转码使之可以安全的用在URL查询里。
		url.QueryEscape(key),
	)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, nil
}

var _ PeerGetter = (*httpGetter)(nil)
