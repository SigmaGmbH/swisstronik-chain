// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: swisstronik/vesting/tx.proto

package types

import (
	context "context"
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/types/msgservice"
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

// MsgCreateMonthlyVestingAccount defines a message that enables creating a monthly vesting
// account with cliff feature.
type MsgCreateMonthlyVestingAccount struct {
	// from_address is a signer address that funds tokens
	FromAddress string `protobuf:"bytes,1,opt,name=from_address,json=fromAddress,proto3" json:"from_address,omitempty"`
	// to_address defines vesting address that receives funds
	ToAddress string `protobuf:"bytes,2,opt,name=to_address,json=toAddress,proto3" json:"to_address,omitempty"`
	// cliff_days defines the days relative to start time
	CliffDays int64 `protobuf:"varint,3,opt,name=cliff_days,json=cliffDays,proto3" json:"cliff_days,omitempty"`
	// months defines number of months for linear vesting
	Months int64                                    `protobuf:"varint,4,opt,name=months,proto3" json:"months,omitempty"`
	Amount github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,5,rep,name=amount,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"amount"`
}

func (m *MsgCreateMonthlyVestingAccount) Reset()         { *m = MsgCreateMonthlyVestingAccount{} }
func (m *MsgCreateMonthlyVestingAccount) String() string { return proto.CompactTextString(m) }
func (*MsgCreateMonthlyVestingAccount) ProtoMessage()    {}
func (*MsgCreateMonthlyVestingAccount) Descriptor() ([]byte, []int) {
	return fileDescriptor_899431b613aba5b5, []int{0}
}
func (m *MsgCreateMonthlyVestingAccount) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgCreateMonthlyVestingAccount) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgCreateMonthlyVestingAccount.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgCreateMonthlyVestingAccount) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgCreateMonthlyVestingAccount.Merge(m, src)
}
func (m *MsgCreateMonthlyVestingAccount) XXX_Size() int {
	return m.Size()
}
func (m *MsgCreateMonthlyVestingAccount) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgCreateMonthlyVestingAccount.DiscardUnknown(m)
}

var xxx_messageInfo_MsgCreateMonthlyVestingAccount proto.InternalMessageInfo

func (m *MsgCreateMonthlyVestingAccount) GetFromAddress() string {
	if m != nil {
		return m.FromAddress
	}
	return ""
}

func (m *MsgCreateMonthlyVestingAccount) GetToAddress() string {
	if m != nil {
		return m.ToAddress
	}
	return ""
}

func (m *MsgCreateMonthlyVestingAccount) GetCliffDays() int64 {
	if m != nil {
		return m.CliffDays
	}
	return 0
}

func (m *MsgCreateMonthlyVestingAccount) GetMonths() int64 {
	if m != nil {
		return m.Months
	}
	return 0
}

func (m *MsgCreateMonthlyVestingAccount) GetAmount() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Amount
	}
	return nil
}

// MsgCreateMonthlyVestingAccountResponse defines MsgCreateMonthlyVestingAccount response type.
type MsgCreateMonthlyVestingAccountResponse struct {
}

func (m *MsgCreateMonthlyVestingAccountResponse) Reset() {
	*m = MsgCreateMonthlyVestingAccountResponse{}
}
func (m *MsgCreateMonthlyVestingAccountResponse) String() string { return proto.CompactTextString(m) }
func (*MsgCreateMonthlyVestingAccountResponse) ProtoMessage()    {}
func (*MsgCreateMonthlyVestingAccountResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_899431b613aba5b5, []int{1}
}
func (m *MsgCreateMonthlyVestingAccountResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgCreateMonthlyVestingAccountResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgCreateMonthlyVestingAccountResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgCreateMonthlyVestingAccountResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgCreateMonthlyVestingAccountResponse.Merge(m, src)
}
func (m *MsgCreateMonthlyVestingAccountResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgCreateMonthlyVestingAccountResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgCreateMonthlyVestingAccountResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgCreateMonthlyVestingAccountResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgCreateMonthlyVestingAccount)(nil), "swisstronik.vesting.MsgCreateMonthlyVestingAccount")
	proto.RegisterType((*MsgCreateMonthlyVestingAccountResponse)(nil), "swisstronik.vesting.MsgCreateMonthlyVestingAccountResponse")
}

func init() { proto.RegisterFile("swisstronik/vesting/tx.proto", fileDescriptor_899431b613aba5b5) }

var fileDescriptor_899431b613aba5b5 = []byte{
	// 400 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x92, 0xbd, 0x8e, 0xd3, 0x40,
	0x10, 0xc7, 0xbd, 0x67, 0x88, 0x94, 0x3d, 0x1a, 0x0c, 0x02, 0x13, 0x60, 0x2f, 0x97, 0x02, 0x59,
	0x48, 0xec, 0x92, 0x44, 0x34, 0x50, 0x25, 0xa1, 0xa0, 0x49, 0xe3, 0x82, 0x82, 0x26, 0x5a, 0xdb,
	0x1b, 0xc7, 0x4a, 0xbc, 0x13, 0x79, 0x36, 0x21, 0x6e, 0x79, 0x02, 0x28, 0x28, 0x79, 0x01, 0x2a,
	0x1e, 0x23, 0x65, 0x4a, 0x2a, 0x40, 0x49, 0xc1, 0x6b, 0x20, 0x7f, 0x00, 0x41, 0x42, 0x29, 0xae,
	0x9a, 0xdd, 0xf9, 0xcd, 0xec, 0xcc, 0x7f, 0x76, 0xe8, 0x03, 0x7c, 0x9b, 0x20, 0x9a, 0x0c, 0x74,
	0x32, 0x17, 0x6b, 0x85, 0x26, 0xd1, 0xb1, 0x30, 0x1b, 0xbe, 0xcc, 0xc0, 0x80, 0x73, 0xeb, 0x88,
	0xf2, 0x9a, 0xb6, 0x6e, 0xc7, 0x10, 0x43, 0xc9, 0x45, 0x71, 0xaa, 0x42, 0x5b, 0x2c, 0x04, 0x4c,
	0x01, 0x45, 0x20, 0x51, 0x89, 0x75, 0x37, 0x50, 0x46, 0x76, 0x45, 0x08, 0x89, 0xae, 0xf9, 0xdd,
	0x9a, 0xa7, 0x18, 0x8b, 0x75, 0xb7, 0x30, 0x15, 0xe8, 0x7c, 0x38, 0xa3, 0x6c, 0x8c, 0xf1, 0x28,
	0x53, 0xd2, 0xa8, 0x31, 0x68, 0x33, 0x5b, 0xe4, 0xaf, 0xab, 0x52, 0x83, 0x30, 0x84, 0x95, 0x36,
	0xce, 0x25, 0xbd, 0x31, 0xcd, 0x20, 0x9d, 0xc8, 0x28, 0xca, 0x14, 0xa2, 0x4b, 0xda, 0xc4, 0x6b,
	0xfa, 0xe7, 0x85, 0x6f, 0x50, 0xb9, 0x9c, 0x87, 0x94, 0x1a, 0xf8, 0x13, 0x70, 0x56, 0x06, 0x34,
	0x0d, 0x1c, 0xe1, 0x70, 0x91, 0x4c, 0xa7, 0x93, 0x48, 0xe6, 0xe8, 0xda, 0x6d, 0xe2, 0xd9, 0x7e,
	0xb3, 0xf4, 0xbc, 0x94, 0x39, 0x3a, 0x77, 0x68, 0x23, 0x2d, 0x2a, 0xa3, 0x7b, 0xad, 0x44, 0xf5,
	0xcd, 0x09, 0x69, 0x43, 0xa6, 0x45, 0x0b, 0xee, 0xf5, 0xb6, 0xed, 0x9d, 0xf7, 0xee, 0xf1, 0x4a,
	0x05, 0x2f, 0x54, 0xf2, 0x5a, 0x25, 0x1f, 0x41, 0xa2, 0x87, 0x4f, 0xb7, 0xdf, 0x2e, 0xac, 0xcf,
	0xdf, 0x2f, 0xbc, 0x38, 0x31, 0xb3, 0x55, 0xc0, 0x43, 0x48, 0x45, 0x2d, 0xb9, 0x32, 0x4f, 0x30,
	0x9a, 0x0b, 0x93, 0x2f, 0x15, 0x96, 0x09, 0xe8, 0xd7, 0x4f, 0x3f, 0xbf, 0xf9, 0xee, 0xe7, 0x97,
	0xc7, 0xff, 0x08, 0xec, 0x78, 0xf4, 0xd1, 0xe9, 0x91, 0xf8, 0x0a, 0x97, 0xa0, 0x51, 0xf5, 0x3e,
	0x11, 0x6a, 0x8f, 0x31, 0x76, 0x3e, 0x12, 0x7a, 0xf9, 0x4a, 0xea, 0x68, 0xa1, 0x4e, 0x0d, 0xb2,
	0xcf, 0xff, 0xf3, 0xa1, 0xfc, 0x74, 0xa9, 0xd6, 0x8b, 0x2b, 0x24, 0xfd, 0xee, 0x6f, 0xf8, 0x6c,
	0xbb, 0x67, 0x64, 0xb7, 0x67, 0xe4, 0xc7, 0x9e, 0x91, 0xf7, 0x07, 0x66, 0xed, 0x0e, 0xcc, 0xfa,
	0x7a, 0x60, 0xd6, 0x9b, 0xfb, 0xc7, 0x9b, 0xb7, 0xf9, 0xbb, 0x7b, 0xc5, 0x84, 0x82, 0x46, 0xb9,
	0x1b, 0xfd, 0x5f, 0x01, 0x00, 0x00, 0xff, 0xff, 0xe0, 0x86, 0x6f, 0xbf, 0x9f, 0x02, 0x00, 0x00,
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
	// CreateMonthlyVestingAccount defines a method that enables creating a monthly vesting account
	// with cliff feature.
	HandleCreateMonthlyVestingAccount(ctx context.Context, in *MsgCreateMonthlyVestingAccount, opts ...grpc.CallOption) (*MsgCreateMonthlyVestingAccountResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) HandleCreateMonthlyVestingAccount(ctx context.Context, in *MsgCreateMonthlyVestingAccount, opts ...grpc.CallOption) (*MsgCreateMonthlyVestingAccountResponse, error) {
	out := new(MsgCreateMonthlyVestingAccountResponse)
	err := c.cc.Invoke(ctx, "/swisstronik.vesting.Msg/HandleCreateMonthlyVestingAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	// CreateMonthlyVestingAccount defines a method that enables creating a monthly vesting account
	// with cliff feature.
	HandleCreateMonthlyVestingAccount(context.Context, *MsgCreateMonthlyVestingAccount) (*MsgCreateMonthlyVestingAccountResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) HandleCreateMonthlyVestingAccount(ctx context.Context, req *MsgCreateMonthlyVestingAccount) (*MsgCreateMonthlyVestingAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HandleCreateMonthlyVestingAccount not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_HandleCreateMonthlyVestingAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateMonthlyVestingAccount)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).HandleCreateMonthlyVestingAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/swisstronik.vesting.Msg/HandleCreateMonthlyVestingAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).HandleCreateMonthlyVestingAccount(ctx, req.(*MsgCreateMonthlyVestingAccount))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "swisstronik.vesting.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "HandleCreateMonthlyVestingAccount",
			Handler:    _Msg_HandleCreateMonthlyVestingAccount_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "swisstronik/vesting/tx.proto",
}

func (m *MsgCreateMonthlyVestingAccount) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgCreateMonthlyVestingAccount) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgCreateMonthlyVestingAccount) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Amount) > 0 {
		for iNdEx := len(m.Amount) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Amount[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTx(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x2a
		}
	}
	if m.Months != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Months))
		i--
		dAtA[i] = 0x20
	}
	if m.CliffDays != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.CliffDays))
		i--
		dAtA[i] = 0x18
	}
	if len(m.ToAddress) > 0 {
		i -= len(m.ToAddress)
		copy(dAtA[i:], m.ToAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ToAddress)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.FromAddress) > 0 {
		i -= len(m.FromAddress)
		copy(dAtA[i:], m.FromAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.FromAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgCreateMonthlyVestingAccountResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgCreateMonthlyVestingAccountResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgCreateMonthlyVestingAccountResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
func (m *MsgCreateMonthlyVestingAccount) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.FromAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.ToAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.CliffDays != 0 {
		n += 1 + sovTx(uint64(m.CliffDays))
	}
	if m.Months != 0 {
		n += 1 + sovTx(uint64(m.Months))
	}
	if len(m.Amount) > 0 {
		for _, e := range m.Amount {
			l = e.Size()
			n += 1 + l + sovTx(uint64(l))
		}
	}
	return n
}

func (m *MsgCreateMonthlyVestingAccountResponse) Size() (n int) {
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
func (m *MsgCreateMonthlyVestingAccount) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgCreateMonthlyVestingAccount: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgCreateMonthlyVestingAccount: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FromAddress", wireType)
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
			m.FromAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ToAddress", wireType)
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
			m.ToAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CliffDays", wireType)
			}
			m.CliffDays = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CliffDays |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Months", wireType)
			}
			m.Months = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Months |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
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
			m.Amount = append(m.Amount, types.Coin{})
			if err := m.Amount[len(m.Amount)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *MsgCreateMonthlyVestingAccountResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgCreateMonthlyVestingAccountResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgCreateMonthlyVestingAccountResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
