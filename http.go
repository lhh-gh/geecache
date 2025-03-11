package geecache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/_geecache/" // 默认HTTP路由前缀

// HTTPPool 实现PeerPicker接口的HTTP节点池
// 核心职责：
//  1. 作为HTTP服务端处理缓存请求
//  2. 提供节点间通信能力（后续可扩展为客户端功能）
//
// 设计特点：
//   - 固定路由前缀保证接口规范性
//   - 日志集成节点标识便于调试
type HTTPPool struct {
	self     string // 本节点地址（格式：协议://地址:端口）
	basePath string // 路由前缀（默认/_geecache/）
}

// NewHTTPPool 构造HTTP节点池实例
// 参数：
//
//	self - 本节点完整URL（如 "http://localhost:8000"）
//
// 典型用法：
//
//	pool := NewHTTPPool("http://localhost:8000")
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath, // 使用默认路由前缀
	}
}

// Log 带节点标识的日志方法
// 输出格式：[Server {self}] {message}
// 设计目的：
//  1. 多节点场景下快速定位日志来源
//  2. 统一日志格式便于分析
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP 处理所有HTTP请求（实现http.Handler接口）
// 请求路径规范：
//
//	{basePath}/{group}/{key}
//
// 处理流程：
//  1. 路径校验 -> 2. 参数解析 -> 3. 缓存查询 -> 4. 结果返回
//
// 安全机制：
//   - 严格路径前缀检查防止路由冲突
//   - 错误请求快速失败（400/404/500）
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. 路径验证
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path) // 设计约束：严格路径匹配
	}

	p.Log("%s %s", r.Method, r.URL.Path) // 记录访问日志

	// 2. 路径解析
	// 规范路径格式：/{basePath}/{group}/{key}
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0] // 缓存组名称
	key := parts[1]       // 缓存键

	// 3. 缓存组查询
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	// 4. 缓存数据获取
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. 响应处理
	w.Header().Set("Content-Type", "application/octet-stream") // 二进制流格式
	w.Write(view.ByteSlice())                                  // 返回数据的防御性拷贝
}
