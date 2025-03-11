package geecachepb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// 协议缓冲区版本兼容性检查（自动生成）
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// 编译时proto包版本验证（自动生成安全检查）
const _ = proto.ProtoPackageIsVersion3

// Request 定义gRPC请求格式，对应HTTP路径参数：
// 示例：/_geecache/{group}/{key} -> group=group, key=key
type Request struct {
	Group string `protobuf:"bytes,1,opt,name=group,proto3" json:"group,omitempty"` // 缓存组标识，对应URL路径第一部分
	Key   string `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`     // 缓存键，对应URL路径第二部分
	// Protobuf内部字段（自动生成，无需直接操作）
	XXX_NoUnkeyedLiteral struct{} `json:"-"` // 空结构体用于编译器优化
	XXX_unrecognized     []byte   `json:"-"` // 未识别字段暂存
	XXX_sizecache        int32    `json:"-"` // 缓存大小计算优化
}

// 实现proto.Message接口方法（自动生成）
func (m *Request) Reset()         { *m = Request{} }
func (m *Request) String() string { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()    {}
func (*Request) Descriptor() ([]byte, []int) {
	return fileDescriptor_889d0a4ad37a0d42, []int{0}
}

// Response 定义gRPC响应格式，包含原始字节数据
type Response struct {
	Value []byte `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"` // 缓存值字节数据，使用ByteView包装
	// Protobuf内部字段（同上）
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

// 实现proto.Message接口方法（自动生成）
func (m *Response) Reset()         { *m = Response{} }
func (m *Response) String() string { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()    {}
func (*Response) Descriptor() ([]byte, []int) {
	return fileDescriptor_889d0a4ad37a0d42, []int{1}
}

// 注册协议类型到proto库（自动生成）
func init() {
	proto.RegisterType((*Request)(nil), "geecachepb.Request")
	proto.RegisterType((*Response)(nil), "geecachepb.Response")
}

// 文件描述符元数据（自动生成，对应原始proto文件结构）
// 该二进制数据包含proto文件的结构化描述信息，用于反射和编解码
// 原始proto文件内容：
// syntax = "proto3";
// package geecachepb;
// message Request { string group = 1; string key = 2; }
// message Response { bytes value = 1; }
// service GroupCache { rpc Get(Request) returns (Response); }
var fileDescriptor_889d0a4ad37a0d42 = []byte{
	// 压缩后的文件描述符（十六进制格式）
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x48, 0x4f, 0x4d, 0x4d,
	// [...]（其余二进制数据省略）
}
