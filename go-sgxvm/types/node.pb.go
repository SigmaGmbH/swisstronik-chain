// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.12.4
// source: node.proto

package types

import (
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

type InitializeMasterKeyRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ShouldReset bool `protobuf:"varint,1,opt,name=shouldReset,proto3" json:"shouldReset,omitempty"`
}

func (x *InitializeMasterKeyRequest) Reset() {
	*x = InitializeMasterKeyRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_node_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InitializeMasterKeyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InitializeMasterKeyRequest) ProtoMessage() {}

func (x *InitializeMasterKeyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_node_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InitializeMasterKeyRequest.ProtoReflect.Descriptor instead.
func (*InitializeMasterKeyRequest) Descriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{0}
}

func (x *InitializeMasterKeyRequest) GetShouldReset() bool {
	if x != nil {
		return x.ShouldReset
	}
	return false
}

type InitializeMasterKeyResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *InitializeMasterKeyResponse) Reset() {
	*x = InitializeMasterKeyResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_node_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InitializeMasterKeyResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InitializeMasterKeyResponse) ProtoMessage() {}

func (x *InitializeMasterKeyResponse) ProtoReflect() protoreflect.Message {
	mi := &file_node_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InitializeMasterKeyResponse.ProtoReflect.Descriptor instead.
func (*InitializeMasterKeyResponse) Descriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{1}
}

// Attestation server messages
type PeerAttestationRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Fd     int32 `protobuf:"varint,1,opt,name=fd,proto3" json:"fd,omitempty"`
	IsDCAP bool  `protobuf:"varint,2,opt,name=isDCAP,proto3" json:"isDCAP,omitempty"`
}

func (x *PeerAttestationRequest) Reset() {
	*x = PeerAttestationRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_node_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PeerAttestationRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PeerAttestationRequest) ProtoMessage() {}

func (x *PeerAttestationRequest) ProtoReflect() protoreflect.Message {
	mi := &file_node_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PeerAttestationRequest.ProtoReflect.Descriptor instead.
func (*PeerAttestationRequest) Descriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{2}
}

func (x *PeerAttestationRequest) GetFd() int32 {
	if x != nil {
		return x.Fd
	}
	return 0
}

func (x *PeerAttestationRequest) GetIsDCAP() bool {
	if x != nil {
		return x.IsDCAP
	}
	return false
}

type PeerAttestationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *PeerAttestationResponse) Reset() {
	*x = PeerAttestationResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_node_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PeerAttestationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PeerAttestationResponse) ProtoMessage() {}

func (x *PeerAttestationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_node_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PeerAttestationResponse.ProtoReflect.Descriptor instead.
func (*PeerAttestationResponse) Descriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{3}
}

// Remote Attestation Request
type RemoteAttestationRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Fd       int32  `protobuf:"varint,1,opt,name=fd,proto3" json:"fd,omitempty"`
	Hostname string `protobuf:"bytes,2,opt,name=hostname,proto3" json:"hostname,omitempty"`
	IsDCAP   bool   `protobuf:"varint,3,opt,name=isDCAP,proto3" json:"isDCAP,omitempty"`
}

func (x *RemoteAttestationRequest) Reset() {
	*x = RemoteAttestationRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_node_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RemoteAttestationRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoteAttestationRequest) ProtoMessage() {}

func (x *RemoteAttestationRequest) ProtoReflect() protoreflect.Message {
	mi := &file_node_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoteAttestationRequest.ProtoReflect.Descriptor instead.
func (*RemoteAttestationRequest) Descriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{4}
}

func (x *RemoteAttestationRequest) GetFd() int32 {
	if x != nil {
		return x.Fd
	}
	return 0
}

func (x *RemoteAttestationRequest) GetHostname() string {
	if x != nil {
		return x.Hostname
	}
	return ""
}

func (x *RemoteAttestationRequest) GetIsDCAP() bool {
	if x != nil {
		return x.IsDCAP
	}
	return false
}

type RemoteAttestationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *RemoteAttestationResponse) Reset() {
	*x = RemoteAttestationResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_node_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RemoteAttestationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoteAttestationResponse) ProtoMessage() {}

func (x *RemoteAttestationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_node_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoteAttestationResponse.ProtoReflect.Descriptor instead.
func (*RemoteAttestationResponse) Descriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{5}
}

type IsInitializedRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *IsInitializedRequest) Reset() {
	*x = IsInitializedRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_node_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IsInitializedRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IsInitializedRequest) ProtoMessage() {}

func (x *IsInitializedRequest) ProtoReflect() protoreflect.Message {
	mi := &file_node_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IsInitializedRequest.ProtoReflect.Descriptor instead.
func (*IsInitializedRequest) Descriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{6}
}

type IsInitializedResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IsInitialized bool `protobuf:"varint,1,opt,name=isInitialized,proto3" json:"isInitialized,omitempty"`
}

func (x *IsInitializedResponse) Reset() {
	*x = IsInitializedResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_node_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IsInitializedResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IsInitializedResponse) ProtoMessage() {}

func (x *IsInitializedResponse) ProtoReflect() protoreflect.Message {
	mi := &file_node_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IsInitializedResponse.ProtoReflect.Descriptor instead.
func (*IsInitializedResponse) Descriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{7}
}

func (x *IsInitializedResponse) GetIsInitialized() bool {
	if x != nil {
		return x.IsInitialized
	}
	return false
}

type NodeStatusRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *NodeStatusRequest) Reset() {
	*x = NodeStatusRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_node_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeStatusRequest) ProtoMessage() {}

func (x *NodeStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_node_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeStatusRequest.ProtoReflect.Descriptor instead.
func (*NodeStatusRequest) Descriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{8}
}

type NodeStatusResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *NodeStatusResponse) Reset() {
	*x = NodeStatusResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_node_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeStatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeStatusResponse) ProtoMessage() {}

func (x *NodeStatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_node_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeStatusResponse.ProtoReflect.Descriptor instead.
func (*NodeStatusResponse) Descriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{9}
}

type SetupRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Req:
	//
	//	*SetupRequest_InitializeMasterKey
	//	*SetupRequest_PeerAttestationRequest
	//	*SetupRequest_RemoteAttestationRequest
	//	*SetupRequest_IsInitialized
	//	*SetupRequest_NodeStatus
	Req isSetupRequest_Req `protobuf_oneof:"req"`
}

func (x *SetupRequest) Reset() {
	*x = SetupRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_node_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetupRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetupRequest) ProtoMessage() {}

func (x *SetupRequest) ProtoReflect() protoreflect.Message {
	mi := &file_node_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetupRequest.ProtoReflect.Descriptor instead.
func (*SetupRequest) Descriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{10}
}

func (m *SetupRequest) GetReq() isSetupRequest_Req {
	if m != nil {
		return m.Req
	}
	return nil
}

func (x *SetupRequest) GetInitializeMasterKey() *InitializeMasterKeyRequest {
	if x, ok := x.GetReq().(*SetupRequest_InitializeMasterKey); ok {
		return x.InitializeMasterKey
	}
	return nil
}

func (x *SetupRequest) GetPeerAttestationRequest() *PeerAttestationRequest {
	if x, ok := x.GetReq().(*SetupRequest_PeerAttestationRequest); ok {
		return x.PeerAttestationRequest
	}
	return nil
}

func (x *SetupRequest) GetRemoteAttestationRequest() *RemoteAttestationRequest {
	if x, ok := x.GetReq().(*SetupRequest_RemoteAttestationRequest); ok {
		return x.RemoteAttestationRequest
	}
	return nil
}

func (x *SetupRequest) GetIsInitialized() *IsInitializedRequest {
	if x, ok := x.GetReq().(*SetupRequest_IsInitialized); ok {
		return x.IsInitialized
	}
	return nil
}

func (x *SetupRequest) GetNodeStatus() *NodeStatusRequest {
	if x, ok := x.GetReq().(*SetupRequest_NodeStatus); ok {
		return x.NodeStatus
	}
	return nil
}

type isSetupRequest_Req interface {
	isSetupRequest_Req()
}

type SetupRequest_InitializeMasterKey struct {
	InitializeMasterKey *InitializeMasterKeyRequest `protobuf:"bytes,1,opt,name=initializeMasterKey,proto3,oneof"`
}

type SetupRequest_PeerAttestationRequest struct {
	PeerAttestationRequest *PeerAttestationRequest `protobuf:"bytes,2,opt,name=peerAttestationRequest,proto3,oneof"`
}

type SetupRequest_RemoteAttestationRequest struct {
	RemoteAttestationRequest *RemoteAttestationRequest `protobuf:"bytes,3,opt,name=remoteAttestationRequest,proto3,oneof"`
}

type SetupRequest_IsInitialized struct {
	IsInitialized *IsInitializedRequest `protobuf:"bytes,4,opt,name=isInitialized,proto3,oneof"`
}

type SetupRequest_NodeStatus struct {
	NodeStatus *NodeStatusRequest `protobuf:"bytes,5,opt,name=nodeStatus,proto3,oneof"`
}

func (*SetupRequest_InitializeMasterKey) isSetupRequest_Req() {}

func (*SetupRequest_PeerAttestationRequest) isSetupRequest_Req() {}

func (*SetupRequest_RemoteAttestationRequest) isSetupRequest_Req() {}

func (*SetupRequest_IsInitialized) isSetupRequest_Req() {}

func (*SetupRequest_NodeStatus) isSetupRequest_Req() {}

var File_node_proto protoreflect.FileDescriptor

var file_node_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x6e, 0x6f, 0x64, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x6e, 0x6f,
	0x64, 0x65, 0x2e, 0x6e, 0x6f, 0x64, 0x65, 0x22, 0x3e, 0x0a, 0x1a, 0x49, 0x6e, 0x69, 0x74, 0x69,
	0x61, 0x6c, 0x69, 0x7a, 0x65, 0x4d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x4b, 0x65, 0x79, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x20, 0x0a, 0x0b, 0x73, 0x68, 0x6f, 0x75, 0x6c, 0x64, 0x52,
	0x65, 0x73, 0x65, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x73, 0x68, 0x6f, 0x75,
	0x6c, 0x64, 0x52, 0x65, 0x73, 0x65, 0x74, 0x22, 0x1d, 0x0a, 0x1b, 0x49, 0x6e, 0x69, 0x74, 0x69,
	0x61, 0x6c, 0x69, 0x7a, 0x65, 0x4d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x4b, 0x65, 0x79, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x40, 0x0a, 0x16, 0x50, 0x65, 0x65, 0x72, 0x41, 0x74,
	0x74, 0x65, 0x73, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x0e, 0x0a, 0x02, 0x66, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x02, 0x66, 0x64,
	0x12, 0x16, 0x0a, 0x06, 0x69, 0x73, 0x44, 0x43, 0x41, 0x50, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x06, 0x69, 0x73, 0x44, 0x43, 0x41, 0x50, 0x22, 0x19, 0x0a, 0x17, 0x50, 0x65, 0x65, 0x72,
	0x41, 0x74, 0x74, 0x65, 0x73, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x5e, 0x0a, 0x18, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x41, 0x74, 0x74,
	0x65, 0x73, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x0e, 0x0a, 0x02, 0x66, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x02, 0x66, 0x64, 0x12,
	0x1a, 0x0a, 0x08, 0x68, 0x6f, 0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x68, 0x6f, 0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x69,
	0x73, 0x44, 0x43, 0x41, 0x50, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x69, 0x73, 0x44,
	0x43, 0x41, 0x50, 0x22, 0x1b, 0x0a, 0x19, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x41, 0x74, 0x74,
	0x65, 0x73, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x16, 0x0a, 0x14, 0x49, 0x73, 0x49, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65,
	0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x3d, 0x0a, 0x15, 0x49, 0x73, 0x49, 0x6e,
	0x69, 0x74, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x24, 0x0a, 0x0d, 0x69, 0x73, 0x49, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x69, 0x7a,
	0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0d, 0x69, 0x73, 0x49, 0x6e, 0x69, 0x74,
	0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x22, 0x13, 0x0a, 0x11, 0x4e, 0x6f, 0x64, 0x65, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x14, 0x0a, 0x12,
	0x4e, 0x6f, 0x64, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0xb9, 0x03, 0x0a, 0x0c, 0x53, 0x65, 0x74, 0x75, 0x70, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x59, 0x0a, 0x13, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x69, 0x7a,
	0x65, 0x4d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x4b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x25, 0x2e, 0x6e, 0x6f, 0x64, 0x65, 0x2e, 0x6e, 0x6f, 0x64, 0x65, 0x2e, 0x49, 0x6e, 0x69,
	0x74, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x4d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x4b, 0x65, 0x79,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x13, 0x69, 0x6e, 0x69, 0x74, 0x69,
	0x61, 0x6c, 0x69, 0x7a, 0x65, 0x4d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x4b, 0x65, 0x79, 0x12, 0x5b,
	0x0a, 0x16, 0x70, 0x65, 0x65, 0x72, 0x41, 0x74, 0x74, 0x65, 0x73, 0x74, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21,
	0x2e, 0x6e, 0x6f, 0x64, 0x65, 0x2e, 0x6e, 0x6f, 0x64, 0x65, 0x2e, 0x50, 0x65, 0x65, 0x72, 0x41,
	0x74, 0x74, 0x65, 0x73, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x48, 0x00, 0x52, 0x16, 0x70, 0x65, 0x65, 0x72, 0x41, 0x74, 0x74, 0x65, 0x73, 0x74, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x61, 0x0a, 0x18, 0x72,
	0x65, 0x6d, 0x6f, 0x74, 0x65, 0x41, 0x74, 0x74, 0x65, 0x73, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e,
	0x6e, 0x6f, 0x64, 0x65, 0x2e, 0x6e, 0x6f, 0x64, 0x65, 0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65,
	0x41, 0x74, 0x74, 0x65, 0x73, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x48, 0x00, 0x52, 0x18, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x41, 0x74, 0x74, 0x65,
	0x73, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x47,
	0x0a, 0x0d, 0x69, 0x73, 0x49, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x6e, 0x6f, 0x64, 0x65, 0x2e, 0x6e, 0x6f, 0x64,
	0x65, 0x2e, 0x49, 0x73, 0x49, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x0d, 0x69, 0x73, 0x49, 0x6e, 0x69, 0x74,
	0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x12, 0x3e, 0x0a, 0x0a, 0x6e, 0x6f, 0x64, 0x65, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x6e, 0x6f,
	0x64, 0x65, 0x2e, 0x6e, 0x6f, 0x64, 0x65, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x0a, 0x6e, 0x6f, 0x64,
	0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x42, 0x05, 0x0a, 0x03, 0x72, 0x65, 0x71, 0x42, 0x26,
	0x5a, 0x24, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x53, 0x69, 0x67,
	0x6d, 0x61, 0x47, 0x6d, 0x62, 0x48, 0x2f, 0x6c, 0x69, 0x62, 0x72, 0x75, 0x73, 0x74, 0x67, 0x6f,
	0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_node_proto_rawDescOnce sync.Once
	file_node_proto_rawDescData = file_node_proto_rawDesc
)

func file_node_proto_rawDescGZIP() []byte {
	file_node_proto_rawDescOnce.Do(func() {
		file_node_proto_rawDescData = protoimpl.X.CompressGZIP(file_node_proto_rawDescData)
	})
	return file_node_proto_rawDescData
}

var file_node_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_node_proto_goTypes = []interface{}{
	(*InitializeMasterKeyRequest)(nil),  // 0: node.node.InitializeMasterKeyRequest
	(*InitializeMasterKeyResponse)(nil), // 1: node.node.InitializeMasterKeyResponse
	(*PeerAttestationRequest)(nil),      // 2: node.node.PeerAttestationRequest
	(*PeerAttestationResponse)(nil),     // 3: node.node.PeerAttestationResponse
	(*RemoteAttestationRequest)(nil),    // 4: node.node.RemoteAttestationRequest
	(*RemoteAttestationResponse)(nil),   // 5: node.node.RemoteAttestationResponse
	(*IsInitializedRequest)(nil),        // 6: node.node.IsInitializedRequest
	(*IsInitializedResponse)(nil),       // 7: node.node.IsInitializedResponse
	(*NodeStatusRequest)(nil),           // 8: node.node.NodeStatusRequest
	(*NodeStatusResponse)(nil),          // 9: node.node.NodeStatusResponse
	(*SetupRequest)(nil),                // 10: node.node.SetupRequest
}
var file_node_proto_depIdxs = []int32{
	0, // 0: node.node.SetupRequest.initializeMasterKey:type_name -> node.node.InitializeMasterKeyRequest
	2, // 1: node.node.SetupRequest.peerAttestationRequest:type_name -> node.node.PeerAttestationRequest
	4, // 2: node.node.SetupRequest.remoteAttestationRequest:type_name -> node.node.RemoteAttestationRequest
	6, // 3: node.node.SetupRequest.isInitialized:type_name -> node.node.IsInitializedRequest
	8, // 4: node.node.SetupRequest.nodeStatus:type_name -> node.node.NodeStatusRequest
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_node_proto_init() }
func file_node_proto_init() {
	if File_node_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_node_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InitializeMasterKeyRequest); i {
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
		file_node_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InitializeMasterKeyResponse); i {
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
		file_node_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PeerAttestationRequest); i {
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
		file_node_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PeerAttestationResponse); i {
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
		file_node_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RemoteAttestationRequest); i {
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
		file_node_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RemoteAttestationResponse); i {
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
		file_node_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IsInitializedRequest); i {
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
		file_node_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IsInitializedResponse); i {
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
		file_node_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NodeStatusRequest); i {
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
		file_node_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NodeStatusResponse); i {
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
		file_node_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetupRequest); i {
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
	file_node_proto_msgTypes[10].OneofWrappers = []interface{}{
		(*SetupRequest_InitializeMasterKey)(nil),
		(*SetupRequest_PeerAttestationRequest)(nil),
		(*SetupRequest_RemoteAttestationRequest)(nil),
		(*SetupRequest_IsInitialized)(nil),
		(*SetupRequest_NodeStatus)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_node_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_node_proto_goTypes,
		DependencyIndexes: file_node_proto_depIdxs,
		MessageInfos:      file_node_proto_msgTypes,
	}.Build()
	File_node_proto = out.File
	file_node_proto_rawDesc = nil
	file_node_proto_goTypes = nil
	file_node_proto_depIdxs = nil
}
