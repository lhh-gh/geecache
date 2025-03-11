package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 定义哈希函数类型，将字节数据映射为32位无符号整数
// 默认实现为CRC32校验和，平衡性能与分布均匀性
type Hash func(data []byte) uint32

// Map 实现一致性哈希的核心数据结构
// 设计要点：
//   - 虚拟节点机制改善负载均衡
//   - 排序环状结构实现高效查询
//   - 哈希空间复用减少内存占用
type Map struct {
	hash     Hash           // 哈希函数（可自定义）
	replicas int            // 每个真实节点对应的虚拟节点数
	keys     []int          // 排序后的虚拟节点哈希值（构成哈希环）
	hashMap  map[int]string // 虚拟节点哈希到真实节点的映射
}

// New 创建一致性哈希实例
// 参数说明：
//
//	replicas - 虚拟节点倍数（建议>=200）
//	fn       - 自定义哈希函数（可选，默认CRC32）
//
// 设计约束：
//
//	虚拟节点数需>0以保证哈希环有效性
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE // 默认使用工业标准CRC32算法
	}
	return m
}

// Add 将真实节点加入哈希环
// 核心流程：
//  1. 为每个真实节点生成replicas个虚拟节点
//  2. 虚拟节点格式为"i+key"（i为虚拟节点编号）
//  3. 将所有虚拟节点加入哈希环并排序
//
// 技术细节：
//   - 使用字符串拼接保证不同虚拟节点的唯一性
//   - 排序操作时间复杂度O(n log n)
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			// 生成虚拟节点唯一标识
			virtualKey := strconv.Itoa(i) + key
			// 计算虚拟节点哈希值
			hash := int(m.hash([]byte(virtualKey)))
			m.keys = append(m.keys, hash)
			// 建立虚拟节点到真实节点的映射
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys) // 哈希环排序，支持二分查找
}

// Get 根据键查找对应的真实节点
// 执行流程：
//  1. 计算键的哈希值
//  2. 二分查找第一个>=该哈希的虚拟节点
//  3. 环状处理（取模运算）
//
// 时间复杂度：O(log n) 得益于预排序和二分查找
// 边界情况处理：
//   - 空哈希环返回空字符串
//   - 哈希值超过最大值时回绕到环首
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return "" // 防御性编程
	}

	// 计算键的哈希值
	hash := int(m.hash([]byte(key)))

	// 二分查找第一个>=目标哈希的虚拟节点
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	// 环状处理：当查找结果超出范围时取模回绕
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
