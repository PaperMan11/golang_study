package lru

import "container/list"

/*

   1、在这里我们直接使用 Go 语言标准库实现的双向链表list.List。
   2、字典的定义是 map[string]*list.Element，键是字符串，值是双向链表中对应节点的指针。
   3、maxBytes 是允许使用的最大内存，nbytes 是当前已使用的内存，OnEvicted 是某条记录被移除时的回调函数，可以为 nil。
   4、键值对 entry 是双向链表节点的数据类型，在链表中仍保存每个值对应的 key 的好处在于，淘汰队首节点时，需要用 key 从字典中删除对应的映射。
   5、为了通用性，我们允许值是实现了 Value 接口的任意类型，该接口只包含了一个方法 Len() int，用于返回值所占用的内存大小。

*/

// LRU Cache（最近最少使用）
type Cache struct {
	maxBytes  int64
	nbytes    int64
	ll        *list.List // 双向链表
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value) // hook 在清除条目时执行。
}

type entry struct {
	key   string
	value Value
}

// Value 用来计算字节数
type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

/*

   1、如果键对应的链表节点存在，则将对应节点移动到队尾，并返回查找到的值。
   2、c.ll.MoveToFront(ele)，即将链表中的节点 ele 移动到队尾（双向链表作为队列，队首队尾是相对的，在这里约定 front 为队尾）

*/
// Get 查找键的值
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

/*

   1、c.ll.Back() 取到队首节点，从链表中删除。
   2、delete(c.cache, kv.key)，从字典中 c.cache 删除该节点的映射关系。
   3、更新当前所用的内存 c.nbytes。
   4、如果回调函数 OnEvicted 不为 nil，则调用回调函数。

*/
// RemoveOldest 缓存淘汰
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

/*

   1、如果键存在，则更新对应节点的值，并将该节点移到队尾。
   2、不存在则是新增场景，首先队尾添加新节点 &entry{key, value}, 并字典中添加 key 和节点的映射关系。
   3、更新 c.nbytes，如果超过了设定的最大值 c.maxBytes，则移除最少访问的节点。

*/
// Add 新增缓存
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len()) // value 替换了
		kv.value = value
	} else {
		// 新增节点
		ele := c.ll.PushFront(&entry{key: key, value: value})
		c.cache[key] = ele
		c.nbytes += int64(value.Len()) + int64(len(key))
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes { // 缓存溢出
		c.RemoveOldest()
	}
}

// Len 缓存数
func (c *Cache) Len() int {
	return c.ll.Len()
}
