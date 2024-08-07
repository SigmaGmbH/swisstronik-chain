// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: swisstronik/compliance/genesis.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
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

// GenesisState defines the compliance module's genesis state.
type GenesisState struct {
	Params              Params                        `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	IssuerDetails       []*GenesisIssuerDetails       `protobuf:"bytes,2,rep,name=issuerDetails,proto3" json:"issuerDetails,omitempty"`
	AddressDetails      []*GenesisAddressDetails      `protobuf:"bytes,3,rep,name=addressDetails,proto3" json:"addressDetails,omitempty"`
	VerificationDetails []*GenesisVerificationDetails `protobuf:"bytes,4,rep,name=verificationDetails,proto3" json:"verificationDetails,omitempty"`
	Operators           []*OperatorDetails            `protobuf:"bytes,5,rep,name=operators,proto3" json:"operators,omitempty"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_d430e46e02363948, []int{0}
}
func (m *GenesisState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisState.Merge(m, src)
}
func (m *GenesisState) XXX_Size() int {
	return m.Size()
}
func (m *GenesisState) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisState.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisState proto.InternalMessageInfo

func (m *GenesisState) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func (m *GenesisState) GetIssuerDetails() []*GenesisIssuerDetails {
	if m != nil {
		return m.IssuerDetails
	}
	return nil
}

func (m *GenesisState) GetAddressDetails() []*GenesisAddressDetails {
	if m != nil {
		return m.AddressDetails
	}
	return nil
}

func (m *GenesisState) GetVerificationDetails() []*GenesisVerificationDetails {
	if m != nil {
		return m.VerificationDetails
	}
	return nil
}

func (m *GenesisState) GetOperators() []*OperatorDetails {
	if m != nil {
		return m.Operators
	}
	return nil
}

type GenesisIssuerDetails struct {
	Address string         `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Details *IssuerDetails `protobuf:"bytes,2,opt,name=details,proto3" json:"details,omitempty"`
}

func (m *GenesisIssuerDetails) Reset()         { *m = GenesisIssuerDetails{} }
func (m *GenesisIssuerDetails) String() string { return proto.CompactTextString(m) }
func (*GenesisIssuerDetails) ProtoMessage()    {}
func (*GenesisIssuerDetails) Descriptor() ([]byte, []int) {
	return fileDescriptor_d430e46e02363948, []int{1}
}
func (m *GenesisIssuerDetails) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisIssuerDetails) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisIssuerDetails.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisIssuerDetails) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisIssuerDetails.Merge(m, src)
}
func (m *GenesisIssuerDetails) XXX_Size() int {
	return m.Size()
}
func (m *GenesisIssuerDetails) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisIssuerDetails.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisIssuerDetails proto.InternalMessageInfo

func (m *GenesisIssuerDetails) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *GenesisIssuerDetails) GetDetails() *IssuerDetails {
	if m != nil {
		return m.Details
	}
	return nil
}

type GenesisAddressDetails struct {
	Address string          `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Details *AddressDetails `protobuf:"bytes,2,opt,name=details,proto3" json:"details,omitempty"`
}

func (m *GenesisAddressDetails) Reset()         { *m = GenesisAddressDetails{} }
func (m *GenesisAddressDetails) String() string { return proto.CompactTextString(m) }
func (*GenesisAddressDetails) ProtoMessage()    {}
func (*GenesisAddressDetails) Descriptor() ([]byte, []int) {
	return fileDescriptor_d430e46e02363948, []int{2}
}
func (m *GenesisAddressDetails) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisAddressDetails) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisAddressDetails.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisAddressDetails) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisAddressDetails.Merge(m, src)
}
func (m *GenesisAddressDetails) XXX_Size() int {
	return m.Size()
}
func (m *GenesisAddressDetails) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisAddressDetails.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisAddressDetails proto.InternalMessageInfo

func (m *GenesisAddressDetails) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *GenesisAddressDetails) GetDetails() *AddressDetails {
	if m != nil {
		return m.Details
	}
	return nil
}

type GenesisVerificationDetails struct {
	Id      []byte               `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Details *VerificationDetails `protobuf:"bytes,2,opt,name=details,proto3" json:"details,omitempty"`
}

func (m *GenesisVerificationDetails) Reset()         { *m = GenesisVerificationDetails{} }
func (m *GenesisVerificationDetails) String() string { return proto.CompactTextString(m) }
func (*GenesisVerificationDetails) ProtoMessage()    {}
func (*GenesisVerificationDetails) Descriptor() ([]byte, []int) {
	return fileDescriptor_d430e46e02363948, []int{3}
}
func (m *GenesisVerificationDetails) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisVerificationDetails) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisVerificationDetails.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisVerificationDetails) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisVerificationDetails.Merge(m, src)
}
func (m *GenesisVerificationDetails) XXX_Size() int {
	return m.Size()
}
func (m *GenesisVerificationDetails) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisVerificationDetails.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisVerificationDetails proto.InternalMessageInfo

func (m *GenesisVerificationDetails) GetId() []byte {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *GenesisVerificationDetails) GetDetails() *VerificationDetails {
	if m != nil {
		return m.Details
	}
	return nil
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "swisstronik.compliance.GenesisState")
	proto.RegisterType((*GenesisIssuerDetails)(nil), "swisstronik.compliance.GenesisIssuerDetails")
	proto.RegisterType((*GenesisAddressDetails)(nil), "swisstronik.compliance.GenesisAddressDetails")
	proto.RegisterType((*GenesisVerificationDetails)(nil), "swisstronik.compliance.GenesisVerificationDetails")
}

func init() {
	proto.RegisterFile("swisstronik/compliance/genesis.proto", fileDescriptor_d430e46e02363948)
}

var fileDescriptor_d430e46e02363948 = []byte{
	// 397 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x93, 0x4f, 0x4f, 0xe2, 0x40,
	0x18, 0xc6, 0x5b, 0x60, 0x21, 0x0c, 0x2c, 0x87, 0x59, 0x76, 0xd3, 0xf4, 0x30, 0x4b, 0xd8, 0x45,
	0x49, 0xd4, 0x92, 0xe0, 0xc5, 0x83, 0x89, 0x4a, 0x24, 0xc6, 0x93, 0xa6, 0x46, 0x0f, 0xde, 0x46,
	0x3a, 0x92, 0x89, 0xd0, 0xa9, 0x33, 0xe3, 0xbf, 0x6f, 0xe1, 0x87, 0xf0, 0xc3, 0x70, 0xe4, 0xe8,
	0xc9, 0x18, 0xf8, 0x22, 0x86, 0xa1, 0x95, 0x52, 0x3b, 0x72, 0x6b, 0x93, 0xe7, 0xf9, 0x3d, 0xef,
	0xbc, 0x4f, 0x5e, 0xf0, 0x5f, 0x3c, 0x50, 0x21, 0x24, 0x67, 0x3e, 0xbd, 0x69, 0xf5, 0xd8, 0x30,
	0x18, 0x50, 0xec, 0xf7, 0x48, 0xab, 0x4f, 0x7c, 0x22, 0xa8, 0x70, 0x02, 0xce, 0x24, 0x83, 0x7f,
	0x62, 0x2a, 0x67, 0xa1, 0xb2, 0xab, 0x7d, 0xd6, 0x67, 0x4a, 0xd2, 0x9a, 0x7d, 0xcd, 0xd5, 0xf6,
	0x3f, 0x0d, 0x33, 0xc0, 0x1c, 0x0f, 0x43, 0xa4, 0xdd, 0xd0, 0x88, 0x88, 0x2f, 0xa9, 0xa4, 0x24,
	0x94, 0xd5, 0x5f, 0xb2, 0xa0, 0x7c, 0x34, 0x9f, 0xe5, 0x4c, 0x62, 0x49, 0xe0, 0x2e, 0xc8, 0xcf,
	0x39, 0x96, 0x59, 0x33, 0x9b, 0xa5, 0x36, 0x72, 0xd2, 0x67, 0x73, 0x4e, 0x95, 0xaa, 0x93, 0x1b,
	0xbd, 0xfd, 0x35, 0xdc, 0xd0, 0x03, 0x5d, 0xf0, 0x93, 0x0a, 0x71, 0x47, 0xf8, 0x21, 0x91, 0x98,
	0x0e, 0x84, 0x95, 0xa9, 0x65, 0x9b, 0xa5, 0xf6, 0xa6, 0x0e, 0x12, 0x46, 0x1f, 0xc7, 0x3d, 0xee,
	0x32, 0x02, 0x9e, 0x83, 0x0a, 0xf6, 0x3c, 0x4e, 0x84, 0x88, 0xa0, 0x59, 0x05, 0xdd, 0x5a, 0x01,
	0x3d, 0x58, 0x32, 0xb9, 0x09, 0x08, 0xf4, 0xc0, 0xaf, 0x7b, 0xc2, 0xe9, 0x35, 0xed, 0x61, 0x49,
	0x99, 0x1f, 0xb1, 0x73, 0x8a, 0xdd, 0x5e, 0xc1, 0xbe, 0xf8, 0xea, 0x74, 0xd3, 0x70, 0xb0, 0x0b,
	0x8a, 0x2c, 0x20, 0x1c, 0x4b, 0xc6, 0x85, 0xf5, 0x43, 0xb1, 0xd7, 0x75, 0xec, 0x93, 0x50, 0x18,
	0x01, 0x17, 0xce, 0xfa, 0x2d, 0xa8, 0xa6, 0xad, 0x0a, 0x5a, 0xa0, 0x10, 0x3e, 0x4b, 0xd5, 0x55,
	0x74, 0xa3, 0x5f, 0xb8, 0x07, 0x0a, 0xde, 0x67, 0x07, 0xb3, 0x22, 0x1b, 0xba, 0xd8, 0xe5, 0xe5,
	0x47, 0xae, 0xba, 0x00, 0xbf, 0x53, 0x17, 0xf9, 0x4d, 0xe6, 0x7e, 0x32, 0x73, 0x4d, 0x97, 0x99,
	0xe8, 0x26, 0x16, 0x6a, 0xeb, 0x37, 0x0c, 0x2b, 0x20, 0x43, 0x3d, 0x15, 0x5a, 0x76, 0x33, 0xd4,
	0x83, 0xdd, 0x64, 0xde, 0x86, 0x2e, 0x2f, 0xad, 0xaf, 0xc8, 0xdb, 0xd9, 0x19, 0x4d, 0x90, 0x39,
	0x9e, 0x20, 0xf3, 0x7d, 0x82, 0xcc, 0xe7, 0x29, 0x32, 0xc6, 0x53, 0x64, 0xbc, 0x4e, 0x91, 0x71,
	0x89, 0xe2, 0x47, 0xf4, 0x18, 0x3f, 0x23, 0xf9, 0x14, 0x10, 0x71, 0x95, 0x57, 0x47, 0xb4, 0xfd,
	0x11, 0x00, 0x00, 0xff, 0xff, 0x96, 0xf3, 0x85, 0xab, 0xe6, 0x03, 0x00, 0x00,
}

func (m *GenesisState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Operators) > 0 {
		for iNdEx := len(m.Operators) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Operators[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x2a
		}
	}
	if len(m.VerificationDetails) > 0 {
		for iNdEx := len(m.VerificationDetails) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.VerificationDetails[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.AddressDetails) > 0 {
		for iNdEx := len(m.AddressDetails) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.AddressDetails[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.IssuerDetails) > 0 {
		for iNdEx := len(m.IssuerDetails) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.IssuerDetails[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *GenesisIssuerDetails) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisIssuerDetails) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisIssuerDetails) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
			i = encodeVarintGenesis(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *GenesisAddressDetails) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisAddressDetails) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisAddressDetails) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
			i = encodeVarintGenesis(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *GenesisVerificationDetails) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisVerificationDetails) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisVerificationDetails) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
			i = encodeVarintGenesis(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if len(m.Id) > 0 {
		i -= len(m.Id)
		copy(dAtA[i:], m.Id)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Id)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintGenesis(dAtA []byte, offset int, v uint64) int {
	offset -= sovGenesis(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if len(m.IssuerDetails) > 0 {
		for _, e := range m.IssuerDetails {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.AddressDetails) > 0 {
		for _, e := range m.AddressDetails {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.VerificationDetails) > 0 {
		for _, e := range m.VerificationDetails {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.Operators) > 0 {
		for _, e := range m.Operators {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *GenesisIssuerDetails) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.Details != nil {
		l = m.Details.Size()
		n += 1 + l + sovGenesis(uint64(l))
	}
	return n
}

func (m *GenesisAddressDetails) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.Details != nil {
		l = m.Details.Size()
		n += 1 + l + sovGenesis(uint64(l))
	}
	return n
}

func (m *GenesisVerificationDetails) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Id)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.Details != nil {
		l = m.Details.Size()
		n += 1 + l + sovGenesis(uint64(l))
	}
	return n
}

func sovGenesis(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenesis(x uint64) (n int) {
	return sovGenesis(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GenesisState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: GenesisState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IssuerDetails", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IssuerDetails = append(m.IssuerDetails, &GenesisIssuerDetails{})
			if err := m.IssuerDetails[len(m.IssuerDetails)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AddressDetails", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AddressDetails = append(m.AddressDetails, &GenesisAddressDetails{})
			if err := m.AddressDetails[len(m.AddressDetails)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field VerificationDetails", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.VerificationDetails = append(m.VerificationDetails, &GenesisVerificationDetails{})
			if err := m.VerificationDetails[len(m.VerificationDetails)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Operators", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Operators = append(m.Operators, &OperatorDetails{})
			if err := m.Operators[len(m.Operators)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func (m *GenesisIssuerDetails) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: GenesisIssuerDetails: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisIssuerDetails: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Details", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
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
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func (m *GenesisAddressDetails) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: GenesisAddressDetails: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisAddressDetails: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Details", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Details == nil {
				m.Details = &AddressDetails{}
			}
			if err := m.Details.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func (m *GenesisVerificationDetails) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: GenesisVerificationDetails: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisVerificationDetails: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Id = append(m.Id[:0], dAtA[iNdEx:postIndex]...)
			if m.Id == nil {
				m.Id = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Details", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Details == nil {
				m.Details = &VerificationDetails{}
			}
			if err := m.Details.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func skipGenesis(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenesis
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
					return 0, ErrIntOverflowGenesis
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
					return 0, ErrIntOverflowGenesis
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
				return 0, ErrInvalidLengthGenesis
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGenesis
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGenesis
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGenesis        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenesis          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGenesis = fmt.Errorf("proto: unexpected end of group")
)
