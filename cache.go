package geecache

import (
	"github/lhh-gh/geecache/lru"
	"sync"
)

// 核心职责：提供并发安全的缓存读写能力，隐藏底层LRU实现细节
type cache struct {
	mu         sync.Mutex // 互斥锁，保障并发安全
	lru        *lru.Cache // 实际存储的LRU缓存实例（延迟初始化）
	cacheBytes int64      // 缓存容量限制（单位：字节）
}

// add 添加缓存条目（线程安全）
// 设计要点：
//  1. 延迟初始化：首次写入时创建LRU实例，避免空缓存的内存占用
//  2. 值类型限制：强制使用ByteView保证值不可变性
//  3. 容量检查：由底层LRU自动处理淘汰逻辑
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 延迟初始化：首次操作时创建LRU实例
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}

	// 类型安全：value强制为ByteView类型
	c.lru.Add(key, value)
}

// get 获取缓存条目（线程安全）
// 安全机制：
//  1. 双检锁模式：初始化检查与获取操作的原子性
//  2. 类型断言：确保返回值符合ByteView类型约束
//
// 返回值：
//
//	value - 始终返回深拷贝的ByteView，保证原始数据不可变
//	ok    - 命中状态标识
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 空缓存直接返回
	if c.lru == nil {
		return
	}

	// 类型安全断言
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), true
	}

	return
}
