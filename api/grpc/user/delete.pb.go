package user

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

type DeleteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Username string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
}

func (x *DeleteRequest) Reset() {
	*x = DeleteRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_user_delete_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteRequest) String() string { return protoimpl.X.MessageStringOf(x) }
func (*DeleteRequest) ProtoMessage()    {}

func (x *DeleteRequest) ProtoReflect() protoreflect.Message {
	mi := &file_user_delete_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*DeleteRequest) Descriptor() ([]byte, []int) {
	return file_user_delete_proto_rawDescGZIP(), []int{0}
}

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

func (x *DeleteResponse) Reset() {
	*x = DeleteResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_user_delete_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteResponse) String() string { return protoimpl.X.MessageStringOf(x) }
func (*DeleteResponse) ProtoMessage()    {}

func (x *DeleteResponse) ProtoReflect() protoreflect.Message {
	mi := &file_user_delete_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*DeleteResponse) Descriptor() ([]byte, []int) {
	return file_user_delete_proto_rawDescGZIP(), []int{1}
}

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

var File_user_delete_proto protoreflect.FileDescriptor

var file_user_delete_proto_rawDescOnce sync.Once
var file_user_delete_proto_rawDescData = []byte{}

func file_user_delete_proto_rawDescGZIP() []byte {
	file_user_delete_proto_rawDescOnce.Do(func() {
		file_user_delete_proto_rawDescData = protoimpl.X.CompressGZIP(file_user_delete_proto_rawDescData)
	})
	return file_user_delete_proto_rawDescData
}

var file_user_delete_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_user_delete_proto_goTypes = []any{
	(*DeleteRequest)(nil),
	(*DeleteResponse)(nil),
}
var file_user_delete_proto_depIdxs = []int32{}

func init() { file_user_delete_proto_init() }
func file_user_delete_proto_init() {
	if File_user_delete_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_user_delete_proto_rawDescData,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_user_delete_proto_goTypes,
		DependencyIndexes: file_user_delete_proto_depIdxs,
		MessageInfos:      file_user_delete_proto_msgTypes,
	}.Build()
	File_user_delete_proto = out.File
	file_user_delete_proto_rawDescData = nil
	file_user_delete_proto_goTypes = nil
	file_user_delete_proto_depIdxs = nil
}
