// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        (unknown)
// source: nerf.proto

package nerf

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Login *string `protobuf:"bytes,1,req,name=login" json:"login,omitempty"`
	Token *string `protobuf:"bytes,2,req,name=token" json:"token,omitempty"`
}

func (x *Request) Reset() {
	*x = Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nerf_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request) ProtoMessage() {}

func (x *Request) ProtoReflect() protoreflect.Message {
	mi := &file_nerf_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request.ProtoReflect.Descriptor instead.
func (*Request) Descriptor() ([]byte, []int) {
	return file_nerf_proto_rawDescGZIP(), []int{0}
}

func (x *Request) GetLogin() string {
	if x != nil && x.Login != nil {
		return *x.Login
	}
	return ""
}

func (x *Request) GetToken() string {
	if x != nil && x.Token != nil {
		return *x.Token
	}
	return ""
}

type Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ca  *string `protobuf:"bytes,1,req,name=ca" json:"ca,omitempty"`
	Crt *string `protobuf:"bytes,2,req,name=crt" json:"crt,omitempty"`
	Key *string `protobuf:"bytes,3,req,name=key" json:"key,omitempty"`
}

func (x *Response) Reset() {
	*x = Response{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nerf_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_nerf_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_nerf_proto_rawDescGZIP(), []int{1}
}

func (x *Response) GetCa() string {
	if x != nil && x.Ca != nil {
		return *x.Ca
	}
	return ""
}

func (x *Response) GetCrt() string {
	if x != nil && x.Crt != nil {
		return *x.Crt
	}
	return ""
}

func (x *Response) GetKey() string {
	if x != nil && x.Key != nil {
		return *x.Key
	}
	return ""
}

var File_nerf_proto protoreflect.FileDescriptor

var file_nerf_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x6e, 0x65, 0x72, 0x66, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x6e, 0x65,
	0x72, 0x66, 0x22, 0x35, 0x0a, 0x07, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a,
	0x05, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x02, 0x28, 0x09, 0x52, 0x05, 0x6c, 0x6f,
	0x67, 0x69, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x02,
	0x28, 0x09, 0x52, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x3e, 0x0a, 0x08, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x63, 0x61, 0x18, 0x01, 0x20, 0x02, 0x28,
	0x09, 0x52, 0x02, 0x63, 0x61, 0x12, 0x10, 0x0a, 0x03, 0x63, 0x72, 0x74, 0x18, 0x02, 0x20, 0x02,
	0x28, 0x09, 0x52, 0x03, 0x63, 0x72, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x03,
	0x20, 0x02, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x32, 0x3c, 0x0a, 0x06, 0x53, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x12, 0x32, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x65, 0x73, 0x12, 0x0d, 0x2e, 0x6e, 0x65, 0x72, 0x66, 0x2e, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0e, 0x2e, 0x6e, 0x65, 0x72, 0x66, 0x2e, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
}

var (
	file_nerf_proto_rawDescOnce sync.Once
	file_nerf_proto_rawDescData = file_nerf_proto_rawDesc
)

func file_nerf_proto_rawDescGZIP() []byte {
	file_nerf_proto_rawDescOnce.Do(func() {
		file_nerf_proto_rawDescData = protoimpl.X.CompressGZIP(file_nerf_proto_rawDescData)
	})
	return file_nerf_proto_rawDescData
}

var file_nerf_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_nerf_proto_goTypes = []interface{}{
	(*Request)(nil),  // 0: nerf.Request
	(*Response)(nil), // 1: nerf.Response
}
var file_nerf_proto_depIdxs = []int32{
	0, // 0: nerf.Server.GetCertificates:input_type -> nerf.Request
	1, // 1: nerf.Server.GetCertificates:output_type -> nerf.Response
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_nerf_proto_init() }
func file_nerf_proto_init() {
	if File_nerf_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_nerf_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Request); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_nerf_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Response); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_nerf_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_nerf_proto_goTypes,
		DependencyIndexes: file_nerf_proto_depIdxs,
		MessageInfos:      file_nerf_proto_msgTypes,
	}.Build()
	File_nerf_proto = out.File
	file_nerf_proto_rawDesc = nil
	file_nerf_proto_goTypes = nil
	file_nerf_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// ServerClient is the client API for Server service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ServerClient interface {
	GetCertificates(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
}

type serverClient struct {
	cc grpc.ClientConnInterface
}

func NewServerClient(cc grpc.ClientConnInterface) ServerClient {
	return &serverClient{cc}
}

func (c *serverClient) GetCertificates(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/nerf.Server/GetCertificates", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ServerServer is the server API for Server service.
type ServerServer interface {
	GetCertificates(context.Context, *Request) (*Response, error)
}

// UnimplementedServerServer can be embedded to have forward compatible implementations.
type UnimplementedServerServer struct {
}

func (*UnimplementedServerServer) GetCertificates(context.Context, *Request) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCertificates not implemented")
}

func RegisterServerServer(s *grpc.Server, srv ServerServer) {
	s.RegisterService(&_Server_serviceDesc, srv)
}

func _Server_GetCertificates_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServerServer).GetCertificates(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/nerf.Server/GetCertificates",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServerServer).GetCertificates(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

var _Server_serviceDesc = grpc.ServiceDesc{
	ServiceName: "nerf.Server",
	HandlerType: (*ServerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetCertificates",
			Handler:    _Server_GetCertificates_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "nerf.proto",
}
