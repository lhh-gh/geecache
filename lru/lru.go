package lru

import "container/list"

// Cache 实现基于LRU（最近最少使用）淘汰策略的缓存结构
// 注意：该实现非并发安全，需在外层通过锁机制保证并发场景下的正确性
type Cache struct {
	maxBytes  int64                         // 允许使用的最大内存（字节）
	nbytes    int64                         // 当前已使用的内存（字节）
	ll        *list.List                    // 双向链表，用于维护访问顺序（链表头为最近访问）
	cache     map[string]*list.Element      // 哈希表，提供O(1)时间复杂度查找
	OnEvicted func(key string, value Value) // 可选回调函数，在条目被淘汰时触发
}

// entry 表示缓存中的一个键值对条目
type entry struct {
	key   string
	value Value
}

// Value 接口定义缓存值的行为规范
// 要求所有缓存值类型必须实现Len()方法用于计算自身占用的内存字节数
type Value interface {
	Len() int
}

// New 创建LRU缓存实例的构造函数
// 参数：
//
//	maxBytes   - 最大内存容量（0表示无限制）
//	onEvicted  - 淘汰回调函数（可选）
//
// 返回值：
//
//	*Cache     - 初始化后的缓存实例指针
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Add 添加/更新缓存条目
// 核心逻辑：
//  1. 已存在键：移动到链表头部（标记为最近使用），更新值并重新计算内存占用
//  2. 新键：插入链表头部，更新哈希表索引
//  3. 内存超限时循环淘汰最久未使用的条目（链表尾部）
func (c *Cache) Add(key string, value Value) {
	// 存在性检查
	if ele, ok := c.cache[key]; ok {
		// 移动已有元素到链表头部
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)

		// 更新内存占用：新值大小 - 旧值大小
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		// 创建新条目插入链表头部
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele

		// 增加内存占用：key长度 + value长度
		c.nbytes += int64(len(key)) + int64(value.Len())
	}

	// 内存容量检查与淘汰处理
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Get 获取缓存值
// 返回值：
//
//	value - 缓存值（存在时）
//	ok    - 是否存在标识
//
// 副作用：
//
//	访问存在的条目时会被移动到链表头部（维护LRU特性）
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		// 将命中条目移动到链表头部
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return nil, false
}

// RemoveOldest 执行LRU淘汰策略
// 移除链表尾部元素（最久未使用），并同步更新内存计数和哈希表
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back() // 获取链表尾部元素
	if ele != nil {
		// 从链表中移除
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)

		// 从哈希表删除索引
		delete(c.cache, kv.key)

		// 更新内存占用
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())

		// 触发淘汰回调（如果设置）
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Len 获取当前缓存条目数量
// 通过链表长度实现O(1)时间复杂度查询
func (c *Cache) Len() int {
	return c.ll.Len()
}
