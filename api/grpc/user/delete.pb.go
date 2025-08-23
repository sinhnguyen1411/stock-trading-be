// Code generated manually for Delete user messages.
// This is a simplified placeholder and may not contain full descriptor info.
package user

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

type DeleteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Username string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
}

func (x *DeleteRequest) Reset()         { *x = DeleteRequest{} }
func (x *DeleteRequest) String() string { return protoimpl.X.MessageStringOf(x) }
func (*DeleteRequest) ProtoMessage()    {}

var deleteRequestMessageInfo = protoimpl.MessageInfo{
	GoReflectType: reflect.TypeOf((*DeleteRequest)(nil)).Elem(),
}

func (x *DeleteRequest) ProtoReflect() protoreflect.Message {
	mi := &deleteRequestMessageInfo
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*DeleteRequest) Descriptor() ([]byte, []int) { return nil, []int{} }

func (x *DeleteRequest) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

type DeleteResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code    uint32 `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *DeleteResponse) Reset()         { *x = DeleteResponse{} }
func (x *DeleteResponse) String() string { return protoimpl.X.MessageStringOf(x) }
func (*DeleteResponse) ProtoMessage()    {}

var deleteResponseMessageInfo = protoimpl.MessageInfo{
	GoReflectType: reflect.TypeOf((*DeleteResponse)(nil)).Elem(),
}

func (x *DeleteResponse) ProtoReflect() protoreflect.Message {
	mi := &deleteResponseMessageInfo
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*DeleteResponse) Descriptor() ([]byte, []int) { return nil, []int{} }

func (x *DeleteResponse) GetCode() uint32 {
	if x != nil {
		return x.Code
	}
	return 0
}

func (x *DeleteResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var (
	_ sync.Mutex // to avoid unused import errors with sync
)
