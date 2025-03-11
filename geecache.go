package geecache

import (
	"fmt"
	"log"
	"sync"
)

// Group 表示一个逻辑独立的缓存命名空间
// 核心职责：
//  1. 管理缓存键的命名隔离
//  2. 协调缓存未命中时的数据加载流程
//  3. 集成底层缓存存储与数据获取逻辑
type Group struct {
	name      string // 缓存组唯一标识（命名空间）
	getter    Getter // 数据源获取接口（缓存未命中时调用）
	mainCache cache  // 并发安全缓存实例
}

// Getter 定义数据加载器接口规范
// 设计目标：解耦缓存系统与具体数据源，提供扩展能力
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 函数类型适配器，允许普通函数实现Getter接口
// 设计模式：适配器模式（函数 -> 接口）
type GetterFunc func(key string) ([]byte, error)

// Get 实现Getter接口方法（函数式适配器）
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key) // 直接委托给底层函数
}

var (
	mu     sync.RWMutex              // 全局读写锁，保护groups映射
	groups = make(map[string]*Group) // 全局缓存组注册表
)

// NewGroup 创建并注册新的缓存组（工厂方法）
// 安全机制：
//  1. 互斥锁保证并发安全
//  2. getter非空校验（防止空指针异常）
//
// 典型用法：
//
//	NewGroup("users", 1<<30, GetterFunc(func(key string) {...}))
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter") // 严格校验防止错误配置
	}

	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes}, // 初始化容量但延迟创建LRU
	}
	groups[name] = g // 注册到全局表
	return g
}

// GetGroup 按名称查找已注册的缓存组（安全并发读）
// 性能优化：使用读锁替代写锁，允许并发查询
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get 从缓存组获取键值（核心入口方法）
// 执行流程：
//  1. 参数校验 -> 2. 缓存查询 -> 3. 未命中时加载
//
// 设计特点：
//   - 对调用方隐藏加载细节
//   - 通过ByteView保证返回值的不可变性
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required") // 防御性编程
	}

	// 缓存命中路径
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	// 缓存未命中处理路径
	return g.load(key)
}

// load 统一控制缓存加载流程（预留分布式扩展点）
// 当前实现：直接本地加载，后续可扩展为多节点协同
func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key) // 当前仅本地加载，后续可添加分布式逻辑
}

// getLocally 本地数据加载实现
// 关键步骤：
//  1. 通过Getter获取原始数据
//  2. 数据格式转换与防御性拷贝
//  3. 回填缓存供后续请求使用
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, fmt.Errorf("getter failed: %w", err) // 错误包装
	}

	// 封装不可变视图并缓存
	value := ByteView{b: cloneBytes(bytes)} // 强制深拷贝
	g.populateCache(key, value)
	return value, nil
}

// populateCache 回填缓存的标准流程
// 分离设计：
//   - 独立方法便于后续添加缓存策略（如写穿透/异步更新）
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value) // 线程安全写入
}
