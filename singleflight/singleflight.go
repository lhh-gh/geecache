package singleflight

import "sync"

// call 表示一个正在执行或已完成的函数调用
// 设计要点：
//   - 通过WaitGroup实现调用结果的同步等待
//   - 统一保存返回值和错误信息供重复利用
type call struct {
	wg  sync.WaitGroup // 用于阻塞等待的同步原语
	val interface{}    // 函数调用返回的结果值
	err error          // 函数调用返回的错误信息
}

// Group 单飞机制的核心控制器
// 核心功能：
//   - 为相同key的并发请求提供去重机制
//   - 保证同一时刻单个key只有一个请求实际执行
//
// 内存管理：
//   - 采用延迟初始化策略减少内存占用
type Group struct {
	mu sync.Mutex       // 保护映射表的互斥锁
	m  map[string]*call // 按key跟踪正在执行的调用（延迟初始化）
}

// Do 执行并返回给定函数的结果
// 并发控制逻辑：
//  1. 首个请求：加锁创建调用记录 -> 解锁执行函数 -> 返回结果
//  2. 重复请求：直接等待已有调用结果 -> 共享返回值
//
// 关键特性：
//   - 协程安全：支持高并发场景下的重复请求抑制
//   - 自动清理：调用完成后立即删除映射条目
//   - 结果共享：相同key的并发调用获得相同结果
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// 第一阶段：加锁检查或创建调用记录
	g.mu.Lock()

	// 延迟初始化映射表
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	// 存在进行中的调用
	if c, ok := g.m[key]; ok {
		g.mu.Unlock() // 注意：必须先解锁再等待
		c.wg.Wait()   // 阻塞等待调用完成
		return c.val, c.err
	}

	// 创建新的调用记录
	c := new(call)
	c.wg.Add(1)   // 设置等待计数器
	g.m[key] = c  // 注册到映射表
	g.mu.Unlock() // 关键：提前释放锁，允许其他请求进入

	// 第二阶段：执行实际函数调用（无锁状态）
	c.val, c.err = fn()
	c.wg.Done() // 通知所有等待者调用完成

	// 第三阶段：清理调用记录
	g.mu.Lock()
	delete(g.m, key) // 及时移除已完成条目
	g.mu.Unlock()

	return c.val, c.err
}
