package geecache

// ByteView 表示一个不可变的字节数据视图  示缓存值
// 设计目标：确保缓存值的只读特性，防止外部修改导致数据不一致
type ByteView struct {
	b []byte // 底层字节切片，通过封装实现访问控制
}

// Len 返回字节视图的当前长度
// 时间复杂度：O(1)，直接返回切片长度属性
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 返回字节数据的副本（防御性拷贝）
// 核心机制：
//  1. 使用 cloneBytes 深度复制底层数据
//  2. 确保外部修改不影响缓存原始值
//
// 典型场景：需要修改返回值的业务逻辑
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// String 将字节数据转换为字符串（自动处理拷贝）
// 安全特性：
//   - 直接转换时会拷贝数据（因string不可变）
//   - 与 ByteSlice() 显式拷贝形成双重保障
//
// 适用场景：文本数据的直接使用
func (v ByteView) String() string {
	return string(v.b)
}

// cloneBytes 实现安全的数据拷贝基础方法
// 设计要点：
//  1. 独立函数封装拷贝逻辑，统一维护
//  2. 预分配目标切片避免append的多次扩容
//  3. copy内置函数保证高效内存复制
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
