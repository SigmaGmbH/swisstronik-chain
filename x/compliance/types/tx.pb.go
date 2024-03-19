// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: swisstronik/compliance/tx.proto

package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type MsgSetIssuerDetails struct {
	Signer        string         `protobuf:"bytes,1,opt,name=signer,proto3" json:"signer,omitempty"`
	IssuerAddress string         `protobuf:"bytes,2,opt,name=issuerAddress,proto3" json:"issuerAddress,omitempty"`
	Details       *IssuerDetails `protobuf:"bytes,3,opt,name=details,proto3" json:"details,omitempty"`
}

func (m *MsgSetIssuerDetails) Reset()         { *m = MsgSetIssuerDetails{} }
func (m *MsgSetIssuerDetails) String() string { return proto.CompactTextString(m) }
func (*MsgSetIssuerDetails) ProtoMessage()    {}
func (*MsgSetIssuerDetails) Descriptor() ([]byte, []int) {
	return fileDescriptor_b617e43f088d8eed, []int{0}
}
func (m *MsgSetIssuerDetails) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgSetIssuerDetails) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgSetIssuerDetails.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgSetIssuerDetails) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgSetIssuerDetails.Merge(m, src)
}
func (m *MsgSetIssuerDetails) XXX_Size() int {
	return m.Size()
}
func (m *MsgSetIssuerDetails) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgSetIssuerDetails.DiscardUnknown(m)
}

var xxx_messageInfo_MsgSetIssuerDetails proto.InternalMessageInfo

func (m *MsgSetIssuerDetails) GetSigner() string {
	if m != nil {
		return m.Signer
	}
	return ""
}

func (m *MsgSetIssuerDetails) GetIssuerAddress() string {
	if m != nil {
		return m.IssuerAddress
	}
	return ""
}

func (m *MsgSetIssuerDetails) GetDetails() *IssuerDetails {
	if m != nil {
		return m.Details
	}
	return nil
}

type MsgSetIssuerDetailsResponse struct {
}

func (m *MsgSetIssuerDetailsResponse) Reset()         { *m = MsgSetIssuerDetailsResponse{} }
func (m *MsgSetIssuerDetailsResponse) String() string { return proto.CompactTextString(m) }
func (*MsgSetIssuerDetailsResponse) ProtoMessage()    {}
func (*MsgSetIssuerDetailsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_b617e43f088d8eed, []int{1}
}
func (m *MsgSetIssuerDetailsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgSetIssuerDetailsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgSetIssuerDetailsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgSetIssuerDetailsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgSetIssuerDetailsResponse.Merge(m, src)
}
func (m *MsgSetIssuerDetailsResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgSetIssuerDetailsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgSetIssuerDetailsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgSetIssuerDetailsResponse proto.InternalMessageInfo

type MsgUpdateIssuerDetails struct {
	Signer        string         `protobuf:"bytes,1,opt,name=signer,proto3" json:"signer,omitempty"`
	IssuerAddress string         `protobuf:"bytes,2,opt,name=issuerAddress,proto3" json:"issuerAddress,omitempty"`
	Details       *IssuerDetails `protobuf:"bytes,3,opt,name=details,proto3" json:"details,omitempty"`
}

func (m *MsgUpdateIssuerDetails) Reset()         { *m = MsgUpdateIssuerDetails{} }
func (m *MsgUpdateIssuerDetails) String() string { return proto.CompactTextString(m) }
func (*MsgUpdateIssuerDetails) ProtoMessage()    {}
func (*MsgUpdateIssuerDetails) Descriptor() ([]byte, []int) {
	return fileDescriptor_b617e43f088d8eed, []int{2}
}
func (m *MsgUpdateIssuerDetails) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateIssuerDetails) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateIssuerDetails.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateIssuerDetails) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateIssuerDetails.Merge(m, src)
}
func (m *MsgUpdateIssuerDetails) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateIssuerDetails) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateIssuerDetails.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateIssuerDetails proto.InternalMessageInfo

func (m *MsgUpdateIssuerDetails) GetSigner() string {
	if m != nil {
		return m.Signer
	}
	return ""
}

func (m *MsgUpdateIssuerDetails) GetIssuerAddress() string {
	if m != nil {
		return m.IssuerAddress
	}
	return ""
}

func (m *MsgUpdateIssuerDetails) GetDetails() *IssuerDetails {
	if m != nil {
		return m.Details
	}
	return nil
}

type MsgUpdateIssuerDetailsResponse struct {
}

func (m *MsgUpdateIssuerDetailsResponse) Reset()         { *m = MsgUpdateIssuerDetailsResponse{} }
func (m *MsgUpdateIssuerDetailsResponse) String() string { return proto.CompactTextString(m) }
func (*MsgUpdateIssuerDetailsResponse) ProtoMessage()    {}
func (*MsgUpdateIssuerDetailsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_b617e43f088d8eed, []int{3}
}
func (m *MsgUpdateIssuerDetailsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateIssuerDetailsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateIssuerDetailsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateIssuerDetailsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateIssuerDetailsResponse.Merge(m, src)
}
func (m *MsgUpdateIssuerDetailsResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateIssuerDetailsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateIssuerDetailsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateIssuerDetailsResponse proto.InternalMessageInfo

type MsgRemoveIssuer struct {
	Signer        string `protobuf:"bytes,1,opt,name=signer,proto3" json:"signer,omitempty"`
	IssuerAddress string `protobuf:"bytes,2,opt,name=issuerAddress,proto3" json:"issuerAddress,omitempty"`
}

func (m *MsgRemoveIssuer) Reset()         { *m = MsgRemoveIssuer{} }
func (m *MsgRemoveIssuer) String() string { return proto.CompactTextString(m) }
func (*MsgRemoveIssuer) ProtoMessage()    {}
func (*MsgRemoveIssuer) Descriptor() ([]byte, []int) {
	return fileDescriptor_b617e43f088d8eed, []int{4}
}
func (m *MsgRemoveIssuer) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgRemoveIssuer) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgRemoveIssuer.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgRemoveIssuer) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRemoveIssuer.Merge(m, src)
}
func (m *MsgRemoveIssuer) XXX_Size() int {
	return m.Size()
}
func (m *MsgRemoveIssuer) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRemoveIssuer.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRemoveIssuer proto.InternalMessageInfo

func (m *MsgRemoveIssuer) GetSigner() string {
	if m != nil {
		return m.Signer
	}
	return ""
}

func (m *MsgRemoveIssuer) GetIssuerAddress() string {
	if m != nil {
		return m.IssuerAddress
	}
	return ""
}

type MsgRemoveIssuerResponse struct {
}

func (m *MsgRemoveIssuerResponse) Reset()         { *m = MsgRemoveIssuerResponse{} }
func (m *MsgRemoveIssuerResponse) String() string { return proto.CompactTextString(m) }
func (*MsgRemoveIssuerResponse) ProtoMessage()    {}
func (*MsgRemoveIssuerResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_b617e43f088d8eed, []int{5}
}
func (m *MsgRemoveIssuerResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgRemoveIssuerResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgRemoveIssuerResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgRemoveIssuerResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRemoveIssuerResponse.Merge(m, src)
}
func (m *MsgRemoveIssuerResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgRemoveIssuerResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRemoveIssuerResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRemoveIssuerResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgSetIssuerDetails)(nil), "swisstronik.compliance.MsgSetIssuerDetails")
	proto.RegisterType((*MsgSetIssuerDetailsResponse)(nil), "swisstronik.compliance.MsgSetIssuerDetailsResponse")
	proto.RegisterType((*MsgUpdateIssuerDetails)(nil), "swisstronik.compliance.MsgUpdateIssuerDetails")
	proto.RegisterType((*MsgUpdateIssuerDetailsResponse)(nil), "swisstronik.compliance.MsgUpdateIssuerDetailsResponse")
	proto.RegisterType((*MsgRemoveIssuer)(nil), "swisstronik.compliance.MsgRemoveIssuer")
	proto.RegisterType((*MsgRemoveIssuerResponse)(nil), "swisstronik.compliance.MsgRemoveIssuerResponse")
}

func init() { proto.RegisterFile("swisstronik/compliance/tx.proto", fileDescriptor_b617e43f088d8eed) }

var fileDescriptor_b617e43f088d8eed = []byte{
	// 388 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xcc, 0x94, 0x41, 0x4b, 0xe3, 0x40,
	0x18, 0x86, 0x3b, 0x2d, 0x74, 0xd9, 0x59, 0x96, 0x85, 0xec, 0x92, 0x6d, 0xb3, 0xec, 0x6c, 0x09,
	0x5b, 0xb6, 0xb0, 0x90, 0xd0, 0x16, 0xc4, 0x9b, 0x28, 0x1e, 0xf4, 0x10, 0x84, 0x88, 0x17, 0x6f,
	0x69, 0xf2, 0x11, 0x06, 0xdb, 0x99, 0x90, 0x6f, 0xac, 0xd5, 0xab, 0x7f, 0xc0, 0x83, 0xe0, 0x5f,
	0xf2, 0xd8, 0xa3, 0x47, 0x69, 0x7f, 0x84, 0x57, 0x31, 0x69, 0x6a, 0xa3, 0x69, 0xb1, 0x78, 0xf1,
	0x96, 0x64, 0x9e, 0xbc, 0xef, 0x33, 0x33, 0xcc, 0xd0, 0x3f, 0x78, 0xc6, 0x11, 0x55, 0x2c, 0x05,
	0x3f, 0xb1, 0x7d, 0x39, 0x88, 0xfa, 0xdc, 0x13, 0x3e, 0xd8, 0x6a, 0x64, 0x45, 0xb1, 0x54, 0x52,
	0xd3, 0x17, 0x00, 0xeb, 0x19, 0x30, 0x7e, 0x84, 0x32, 0x94, 0x09, 0x62, 0x3f, 0x3d, 0xa5, 0xb4,
	0xc1, 0x7c, 0x89, 0x03, 0x89, 0x76, 0xcf, 0x43, 0xb0, 0x87, 0xed, 0x1e, 0x28, 0xaf, 0x6d, 0xfb,
	0x92, 0x8b, 0xd9, 0x78, 0x73, 0x49, 0x1d, 0x08, 0xc5, 0x15, 0x07, 0x4c, 0x31, 0xf3, 0x9a, 0xd0,
	0xef, 0x0e, 0x86, 0x87, 0xa0, 0xf6, 0x11, 0x4f, 0x21, 0xde, 0x05, 0xe5, 0xf1, 0x3e, 0x6a, 0x3a,
	0xad, 0x22, 0x0f, 0x05, 0xc4, 0x35, 0xd2, 0x20, 0xad, 0xcf, 0xee, 0xec, 0x4d, 0xfb, 0x4b, 0xbf,
	0xf2, 0x04, 0xdc, 0x0e, 0x82, 0x18, 0x10, 0x6b, 0xe5, 0x64, 0x38, 0xff, 0x51, 0xdb, 0xa2, 0x9f,
	0x82, 0x34, 0xa8, 0x56, 0x69, 0x90, 0xd6, 0x97, 0x4e, 0xd3, 0x2a, 0x9e, 0x9c, 0x95, 0x6b, 0x75,
	0xb3, 0xbf, 0xcc, 0xdf, 0xf4, 0x57, 0x81, 0x95, 0x0b, 0x18, 0x49, 0x81, 0x60, 0xde, 0x10, 0xaa,
	0x3b, 0x18, 0x1e, 0x45, 0x81, 0xa7, 0xe0, 0x43, 0x89, 0x37, 0x28, 0x2b, 0x16, 0x9b, 0xbb, 0x1f,
	0xd0, 0x6f, 0x0e, 0x86, 0x2e, 0x0c, 0xe4, 0x70, 0x46, 0xbc, 0xcf, 0xd9, 0xac, 0xd3, 0x9f, 0x2f,
	0x02, 0xb3, 0xae, 0xce, 0x43, 0x99, 0x56, 0x1c, 0x0c, 0xb5, 0x0b, 0xaa, 0xef, 0x79, 0x22, 0xe8,
	0xc3, 0xab, 0x7d, 0xfe, 0xbf, 0x6c, 0x7e, 0x05, 0xcb, 0x6f, 0x74, 0xd7, 0x80, 0x33, 0x07, 0xed,
	0x92, 0xd0, 0x7a, 0x5a, 0x5e, 0xb4, 0x5d, 0xd6, 0x8a, 0xc8, 0x02, 0xde, 0xd8, 0x58, 0x8f, 0x9f,
	0x5b, 0x08, 0xaa, 0xa5, 0x12, 0xb9, 0x85, 0xff, 0xb7, 0x22, 0x6d, 0x11, 0x34, 0xec, 0x37, 0x82,
	0x59, 0xdf, 0xce, 0xe6, 0xed, 0x84, 0x91, 0xf1, 0x84, 0x91, 0xfb, 0x09, 0x23, 0x57, 0x53, 0x56,
	0x1a, 0x4f, 0x59, 0xe9, 0x6e, 0xca, 0x4a, 0xc7, 0x6c, 0xf1, 0x60, 0x8e, 0x72, 0x37, 0xc1, 0x79,
	0x04, 0xd8, 0xab, 0x26, 0x07, 0xb3, 0xfb, 0x18, 0x00, 0x00, 0xff, 0xff, 0x7c, 0x72, 0xa6, 0x42,
	0x30, 0x04, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	HandleSetIssuerDetails(ctx context.Context, in *MsgSetIssuerDetails, opts ...grpc.CallOption) (*MsgSetIssuerDetailsResponse, error)
	HandleUpdateIssuerDetails(ctx context.Context, in *MsgUpdateIssuerDetails, opts ...grpc.CallOption) (*MsgUpdateIssuerDetailsResponse, error)
	HandleRemoveIssuer(ctx context.Context, in *MsgRemoveIssuer, opts ...grpc.CallOption) (*MsgRemoveIssuerResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) HandleSetIssuerDetails(ctx context.Context, in *MsgSetIssuerDetails, opts ...grpc.CallOption) (*MsgSetIssuerDetailsResponse, error) {
	out := new(MsgSetIssuerDetailsResponse)
	err := c.cc.Invoke(ctx, "/swisstronik.compliance.Msg/HandleSetIssuerDetails", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) HandleUpdateIssuerDetails(ctx context.Context, in *MsgUpdateIssuerDetails, opts ...grpc.CallOption) (*MsgUpdateIssuerDetailsResponse, error) {
	out := new(MsgUpdateIssuerDetailsResponse)
	err := c.cc.Invoke(ctx, "/swisstronik.compliance.Msg/HandleUpdateIssuerDetails", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) HandleRemoveIssuer(ctx context.Context, in *MsgRemoveIssuer, opts ...grpc.CallOption) (*MsgRemoveIssuerResponse, error) {
	out := new(MsgRemoveIssuerResponse)
	err := c.cc.Invoke(ctx, "/swisstronik.compliance.Msg/HandleRemoveIssuer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	HandleSetIssuerDetails(context.Context, *MsgSetIssuerDetails) (*MsgSetIssuerDetailsResponse, error)
	HandleUpdateIssuerDetails(context.Context, *MsgUpdateIssuerDetails) (*MsgUpdateIssuerDetailsResponse, error)
	HandleRemoveIssuer(context.Context, *MsgRemoveIssuer) (*MsgRemoveIssuerResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) HandleSetIssuerDetails(ctx context.Context, req *MsgSetIssuerDetails) (*MsgSetIssuerDetailsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HandleSetIssuerDetails not implemented")
}
func (*UnimplementedMsgServer) HandleUpdateIssuerDetails(ctx context.Context, req *MsgUpdateIssuerDetails) (*MsgUpdateIssuerDetailsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HandleUpdateIssuerDetails not implemented")
}
func (*UnimplementedMsgServer) HandleRemoveIssuer(ctx context.Context, req *MsgRemoveIssuer) (*MsgRemoveIssuerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HandleRemoveIssuer not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_HandleSetIssuerDetails_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSetIssuerDetails)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).HandleSetIssuerDetails(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/swisstronik.compliance.Msg/HandleSetIssuerDetails",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).HandleSetIssuerDetails(ctx, req.(*MsgSetIssuerDetails))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_HandleUpdateIssuerDetails_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateIssuerDetails)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).HandleUpdateIssuerDetails(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/swisstronik.compliance.Msg/HandleUpdateIssuerDetails",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).HandleUpdateIssuerDetails(ctx, req.(*MsgUpdateIssuerDetails))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_HandleRemoveIssuer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRemoveIssuer)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).HandleRemoveIssuer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/swisstronik.compliance.Msg/HandleRemoveIssuer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).HandleRemoveIssuer(ctx, req.(*MsgRemoveIssuer))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "swisstronik.compliance.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "HandleSetIssuerDetails",
			Handler:    _Msg_HandleSetIssuerDetails_Handler,
		},
		{
			MethodName: "HandleUpdateIssuerDetails",
			Handler:    _Msg_HandleUpdateIssuerDetails_Handler,
		},
		{
			MethodName: "HandleRemoveIssuer",
			Handler:    _Msg_HandleRemoveIssuer_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "swisstronik/compliance/tx.proto",
}

func (m *MsgSetIssuerDetails) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgSetIssuerDetails) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgSetIssuerDetails) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Details != nil {
		{
			size, err := m.Details.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTx(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x1a
	}
	if len(m.IssuerAddress) > 0 {
		i -= len(m.IssuerAddress)
		copy(dAtA[i:], m.IssuerAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.IssuerAddress)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Signer) > 0 {
		i -= len(m.Signer)
		copy(dAtA[i:], m.Signer)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Signer)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgSetIssuerDetailsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgSetIssuerDetailsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgSetIssuerDetailsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *MsgUpdateIssuerDetails) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateIssuerDetails) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateIssuerDetails) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Details != nil {
		{
			size, err := m.Details.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTx(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x1a
	}
	if len(m.IssuerAddress) > 0 {
		i -= len(m.IssuerAddress)
		copy(dAtA[i:], m.IssuerAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.IssuerAddress)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Signer) > 0 {
		i -= len(m.Signer)
		copy(dAtA[i:], m.Signer)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Signer)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgUpdateIssuerDetailsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateIssuerDetailsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateIssuerDetailsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *MsgRemoveIssuer) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRemoveIssuer) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRemoveIssuer) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.IssuerAddress) > 0 {
		i -= len(m.IssuerAddress)
		copy(dAtA[i:], m.IssuerAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.IssuerAddress)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Signer) > 0 {
		i -= len(m.Signer)
		copy(dAtA[i:], m.Signer)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Signer)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgRemoveIssuerResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRemoveIssuerResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRemoveIssuerResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgSetIssuerDetails) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Signer)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.IssuerAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Details != nil {
		l = m.Details.Size()
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgSetIssuerDetailsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *MsgUpdateIssuerDetails) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Signer)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.IssuerAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Details != nil {
		l = m.Details.Size()
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgUpdateIssuerDetailsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *MsgRemoveIssuer) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Signer)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.IssuerAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgRemoveIssuerResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgSetIssuerDetails) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgSetIssuerDetails: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgSetIssuerDetails: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Signer", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Signer = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IssuerAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IssuerAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Details", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Details == nil {
				m.Details = &IssuerDetails{}
			}
			if err := m.Details.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgSetIssuerDetailsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgSetIssuerDetailsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgSetIssuerDetailsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgUpdateIssuerDetails) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgUpdateIssuerDetails: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateIssuerDetails: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Signer", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Signer = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IssuerAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IssuerAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Details", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Details == nil {
				m.Details = &IssuerDetails{}
			}
			if err := m.Details.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgUpdateIssuerDetailsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgUpdateIssuerDetailsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateIssuerDetailsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgRemoveIssuer) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgRemoveIssuer: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRemoveIssuer: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Signer", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Signer = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IssuerAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IssuerAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgRemoveIssuerResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgRemoveIssuerResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRemoveIssuerResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTx
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTx
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx = fmt.Errorf("proto: unexpected end of group")
)
