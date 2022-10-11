package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 一致性 hash 算法
// https://geektutu.com/post/geecache-day4.html

/*

   1、定义了函数类型 Hash，采取依赖注入的方式，允许用于替换成自定义的 Hash 函数，也方便测试时替换，默认为 crc32.ChecksumIEEE 算法。
   2、Map 是一致性哈希算法的主数据结构，
   	包含 4 个成员变量：
	Hash 函数 hash；
	虚拟节点倍数 replicas；
	哈希环 keys；
	虚拟节点与真实节点的映射表 hashMap，键是虚拟节点的哈希值，值是真实节点的名称。
   3、构造函数 New() 允许自定义虚拟节点倍数和 Hash 函数。

*/

// Hash maps bytes to uint32
type Hash func(data []byte) uint32

// Map 包含所有散列键
type Map struct {
	hash     Hash           // hash 函数
	replicas int            // 虚拟节点个数（保证key分布均匀）
	keys     []int          // sorted 环
	hashMap  map[int]string // key 与 机器的映射
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

/*

   1、Add 函数允许传入 0 或 多个真实节点的名称。
   2、对每一个真实节点 key，对应创建 m.replicas 个虚拟节点，虚拟节点的名称是：strconv.Itoa(i) + key，即通过添加编号的方式区分不同虚拟节点。
   3、使用 m.hash() 计算虚拟节点的哈希值，使用 append(m.keys, hash) 添加到环上。
   4、在 hashMap 中增加虚拟节点和真实节点的映射关系。
   5、最后一步，环上的哈希值排序。

*/
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys) // 排序
}

/*

   1、选择节点就非常简单了，第一步，计算 key 的哈希值。
   2、第二步，顺时针找到第一个匹配的虚拟节点的下标 idx，从 m.keys 中获取到对应的哈希值。
   	如果 idx == len(m.keys)，说明应选择 m.keys[0]，因为 m.keys 是一个环状结构，所以用取余数的方式来处理这种情况。
   3、第三步，通过 hashMap 映射得到真实的节点。

*/
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// 二分查找合适的副本
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
