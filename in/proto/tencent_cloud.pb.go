// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.19.4
// source: proto/tencent_cloud.proto

package proto

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

type TencentCloudCreateRecRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AudioUrl    string `protobuf:"bytes,1,opt,name=audio_url,json=audioUrl,proto3" json:"audio_url,omitempty"`
	CallbackUrl string `protobuf:"bytes,2,opt,name=callback_url,json=callbackUrl,proto3" json:"callback_url,omitempty"`
}

func (x *TencentCloudCreateRecRequest) Reset() {
	*x = TencentCloudCreateRecRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_tencent_cloud_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TencentCloudCreateRecRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TencentCloudCreateRecRequest) ProtoMessage() {}

func (x *TencentCloudCreateRecRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_tencent_cloud_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TencentCloudCreateRecRequest.ProtoReflect.Descriptor instead.
func (*TencentCloudCreateRecRequest) Descriptor() ([]byte, []int) {
	return file_proto_tencent_cloud_proto_rawDescGZIP(), []int{0}
}

func (x *TencentCloudCreateRecRequest) GetAudioUrl() string {
	if x != nil {
		return x.AudioUrl
	}
	return ""
}

func (x *TencentCloudCreateRecRequest) GetCallbackUrl() string {
	if x != nil {
		return x.CallbackUrl
	}
	return ""
}

type TencentCloudResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Msg  string `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
	Code int32  `protobuf:"varint,2,opt,name=code,proto3" json:"code,omitempty"`
	Data uint64 `protobuf:"varint,3,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *TencentCloudResponse) Reset() {
	*x = TencentCloudResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_tencent_cloud_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TencentCloudResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TencentCloudResponse) ProtoMessage() {}

func (x *TencentCloudResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_tencent_cloud_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TencentCloudResponse.ProtoReflect.Descriptor instead.
func (*TencentCloudResponse) Descriptor() ([]byte, []int) {
	return file_proto_tencent_cloud_proto_rawDescGZIP(), []int{1}
}

func (x *TencentCloudResponse) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

func (x *TencentCloudResponse) GetCode() int32 {
	if x != nil {
		return x.Code
	}
	return 0
}

func (x *TencentCloudResponse) GetData() uint64 {
	if x != nil {
		return x.Data
	}
	return 0
}

var File_proto_tencent_cloud_proto protoreflect.FileDescriptor

var file_proto_tencent_cloud_proto_rawDesc = []byte{
	0x0a, 0x19, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65, 0x6e, 0x63, 0x65, 0x6e, 0x74, 0x5f,
	0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x5e, 0x0a, 0x1c, 0x54,
	0x65, 0x6e, 0x63, 0x65, 0x6e, 0x74, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x52, 0x65, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x61,
	0x75, 0x64, 0x69, 0x6f, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x61, 0x75, 0x64, 0x69, 0x6f, 0x55, 0x72, 0x6c, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x61, 0x6c, 0x6c,
	0x62, 0x61, 0x63, 0x6b, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b,
	0x63, 0x61, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x55, 0x72, 0x6c, 0x22, 0x50, 0x0a, 0x14, 0x54,
	0x65, 0x6e, 0x63, 0x65, 0x6e, 0x74, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6d, 0x73, 0x67, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74,
	0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x32, 0x57, 0x0a,
	0x0c, 0x54, 0x65, 0x6e, 0x63, 0x65, 0x6e, 0x74, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x12, 0x47, 0x0a,
	0x0d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x63, 0x54, 0x61, 0x73, 0x6b, 0x12, 0x1d,
	0x2e, 0x54, 0x65, 0x6e, 0x63, 0x65, 0x6e, 0x74, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x52, 0x65, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e,
	0x54, 0x65, 0x6e, 0x63, 0x65, 0x6e, 0x74, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x09, 0x5a, 0x07, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_tencent_cloud_proto_rawDescOnce sync.Once
	file_proto_tencent_cloud_proto_rawDescData = file_proto_tencent_cloud_proto_rawDesc
)

func file_proto_tencent_cloud_proto_rawDescGZIP() []byte {
	file_proto_tencent_cloud_proto_rawDescOnce.Do(func() {
		file_proto_tencent_cloud_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_tencent_cloud_proto_rawDescData)
	})
	return file_proto_tencent_cloud_proto_rawDescData
}

var file_proto_tencent_cloud_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proto_tencent_cloud_proto_goTypes = []interface{}{
	(*TencentCloudCreateRecRequest)(nil), // 0: TencentCloudCreateRecRequest
	(*TencentCloudResponse)(nil),         // 1: TencentCloudResponse
}
var file_proto_tencent_cloud_proto_depIdxs = []int32{
	0, // 0: TencentCloud.CreateRecTask:input_type -> TencentCloudCreateRecRequest
	1, // 1: TencentCloud.CreateRecTask:output_type -> TencentCloudResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_tencent_cloud_proto_init() }
func file_proto_tencent_cloud_proto_init() {
	if File_proto_tencent_cloud_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_tencent_cloud_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TencentCloudCreateRecRequest); i {
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
		file_proto_tencent_cloud_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TencentCloudResponse); i {
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
			RawDescriptor: file_proto_tencent_cloud_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_tencent_cloud_proto_goTypes,
		DependencyIndexes: file_proto_tencent_cloud_proto_depIdxs,
		MessageInfos:      file_proto_tencent_cloud_proto_msgTypes,
	}.Build()
	File_proto_tencent_cloud_proto = out.File
	file_proto_tencent_cloud_proto_rawDesc = nil
	file_proto_tencent_cloud_proto_goTypes = nil
	file_proto_tencent_cloud_proto_depIdxs = nil
}
