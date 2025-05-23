// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0-devel
// 	protoc        v3.21.1
// source: smarter.proto

package transfer

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Req struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	App    string `protobuf:"bytes,1,opt,name=app,proto3" json:"app,omitempty"`
	Method string `protobuf:"bytes,2,opt,name=method,proto3" json:"method,omitempty"`
	Param  []byte `protobuf:"bytes,3,opt,name=param,proto3" json:"param,omitempty"`
}

func (x *Req) Reset() {
	*x = Req{}
	if protoimpl.UnsafeEnabled {
		mi := &file_smarter_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Req) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Req) ProtoMessage() {}

func (x *Req) ProtoReflect() protoreflect.Message {
	mi := &file_smarter_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Req.ProtoReflect.Descriptor instead.
func (*Req) Descriptor() ([]byte, []int) {
	return file_smarter_proto_rawDescGZIP(), []int{0}
}

func (x *Req) GetApp() string {
	if x != nil {
		return x.App
	}
	return ""
}

func (x *Req) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *Req) GetParam() []byte {
	if x != nil {
		return x.Param
	}
	return nil
}

type Res struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code int32  `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Msg  string `protobuf:"bytes,2,opt,name=msg,proto3" json:"msg,omitempty"`
	Data []byte `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *Res) Reset() {
	*x = Res{}
	if protoimpl.UnsafeEnabled {
		mi := &file_smarter_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Res) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Res) ProtoMessage() {}

func (x *Res) ProtoReflect() protoreflect.Message {
	mi := &file_smarter_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Res.ProtoReflect.Descriptor instead.
func (*Res) Descriptor() ([]byte, []int) {
	return file_smarter_proto_rawDescGZIP(), []int{1}
}

func (x *Res) GetCode() int32 {
	if x != nil {
		return x.Code
	}
	return 0
}

func (x *Res) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

func (x *Res) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_smarter_proto protoreflect.FileDescriptor

var file_smarter_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x73, 0x6d, 0x61, 0x72, 0x74, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x07, 0x73, 0x6d, 0x61, 0x72, 0x74, 0x65, 0x72, 0x22, 0x45, 0x0a, 0x03, 0x52, 0x65, 0x71, 0x12,
	0x10, 0x0a, 0x03, 0x61, 0x70, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x61, 0x70,
	0x70, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x61, 0x72,
	0x61, 0x6d, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x22,
	0x3f, 0x0a, 0x03, 0x52, 0x65, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x73,
	0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x12, 0x12, 0x0a, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x32, 0x2f, 0x0a, 0x07, 0x53, 0x6d, 0x61, 0x72, 0x74, 0x65, 0x72, 0x12, 0x24, 0x0a, 0x04, 0x43,
	0x61, 0x6c, 0x6c, 0x12, 0x0c, 0x2e, 0x73, 0x6d, 0x61, 0x72, 0x74, 0x65, 0x72, 0x2e, 0x52, 0x65,
	0x71, 0x1a, 0x0c, 0x2e, 0x73, 0x6d, 0x61, 0x72, 0x74, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x73, 0x22,
	0x00, 0x42, 0x0a, 0x5a, 0x08, 0x2f, 0x73, 0x6d, 0x61, 0x72, 0x74, 0x65, 0x72, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_smarter_proto_rawDescOnce sync.Once
	file_smarter_proto_rawDescData = file_smarter_proto_rawDesc
)

func file_smarter_proto_rawDescGZIP() []byte {
	file_smarter_proto_rawDescOnce.Do(func() {
		file_smarter_proto_rawDescData = protoimpl.X.CompressGZIP(file_smarter_proto_rawDescData)
	})
	return file_smarter_proto_rawDescData
}

var file_smarter_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_smarter_proto_goTypes = []interface{}{
	(*Req)(nil), // 0: smarter.Req
	(*Res)(nil), // 1: smarter.Res
}
var file_smarter_proto_depIdxs = []int32{
	0, // 0: smarter.Smarter.Call:input_type -> smarter.Req
	1, // 1: smarter.Smarter.Call:output_type -> smarter.Res
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_smarter_proto_init() }
func file_smarter_proto_init() {
	if File_smarter_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_smarter_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Req); i {
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
		file_smarter_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Res); i {
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
			RawDescriptor: file_smarter_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_smarter_proto_goTypes,
		DependencyIndexes: file_smarter_proto_depIdxs,
		MessageInfos:      file_smarter_proto_msgTypes,
	}.Build()
	File_smarter_proto = out.File
	file_smarter_proto_rawDesc = nil
	file_smarter_proto_goTypes = nil
	file_smarter_proto_depIdxs = nil
}
