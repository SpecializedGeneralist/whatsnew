// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: textclassification.proto

package textclassification

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

// ClassifyTextRequest is the request for text classification.
type ClassifyTextRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The text to be classified.
	Text string `protobuf:"bytes,1,opt,name=text,proto3" json:"text,omitempty"`
}

func (x *ClassifyTextRequest) Reset() {
	*x = ClassifyTextRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_textclassification_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClassifyTextRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClassifyTextRequest) ProtoMessage() {}

func (x *ClassifyTextRequest) ProtoReflect() protoreflect.Message {
	mi := &file_textclassification_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClassifyTextRequest.ProtoReflect.Descriptor instead.
func (*ClassifyTextRequest) Descriptor() ([]byte, []int) {
	return file_textclassification_proto_rawDescGZIP(), []int{0}
}

func (x *ClassifyTextRequest) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

// ClassifyTextRequest is the response for text classification.
type ClassifyTextReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// List of text classification results.
	Classes []*Class `protobuf:"bytes,1,rep,name=classes,proto3" json:"classes,omitempty"`
}

func (x *ClassifyTextReply) Reset() {
	*x = ClassifyTextReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_textclassification_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClassifyTextReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClassifyTextReply) ProtoMessage() {}

func (x *ClassifyTextReply) ProtoReflect() protoreflect.Message {
	mi := &file_textclassification_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClassifyTextReply.ProtoReflect.Descriptor instead.
func (*ClassifyTextReply) Descriptor() ([]byte, []int) {
	return file_textclassification_proto_rawDescGZIP(), []int{1}
}

func (x *ClassifyTextReply) GetClasses() []*Class {
	if x != nil {
		return x.Classes
	}
	return nil
}

// Class is a single text classification result.
type Class struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A label describing the type of this class (e.g. "sentiment").
	Type string `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	// A label representing the actual class (e.g. "positive" or "negative").
	Label string `protobuf:"bytes,2,opt,name=label,proto3" json:"label,omitempty"`
	// Prediction confidence, for example in case of a machine-learning system
	// being used. It should be a number between 0 and 1.
	Confidence float32 `protobuf:"fixed32,3,opt,name=confidence,proto3" json:"confidence,omitempty"`
}

func (x *Class) Reset() {
	*x = Class{}
	if protoimpl.UnsafeEnabled {
		mi := &file_textclassification_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Class) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Class) ProtoMessage() {}

func (x *Class) ProtoReflect() protoreflect.Message {
	mi := &file_textclassification_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Class.ProtoReflect.Descriptor instead.
func (*Class) Descriptor() ([]byte, []int) {
	return file_textclassification_proto_rawDescGZIP(), []int{2}
}

func (x *Class) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Class) GetLabel() string {
	if x != nil {
		return x.Label
	}
	return ""
}

func (x *Class) GetConfidence() float32 {
	if x != nil {
		return x.Confidence
	}
	return 0
}

var File_textclassification_proto protoreflect.FileDescriptor

var file_textclassification_proto_rawDesc = []byte{
	0x0a, 0x18, 0x74, 0x65, 0x78, 0x74, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x74, 0x65, 0x78, 0x74,
	0x63, 0x6c, 0x61, 0x73, 0x73, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x29,
	0x0a, 0x13, 0x43, 0x6c, 0x61, 0x73, 0x73, 0x69, 0x66, 0x79, 0x54, 0x65, 0x78, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x65, 0x78, 0x74, 0x22, 0x48, 0x0a, 0x11, 0x43, 0x6c, 0x61,
	0x73, 0x73, 0x69, 0x66, 0x79, 0x54, 0x65, 0x78, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x33,
	0x0a, 0x07, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x19, 0x2e, 0x74, 0x65, 0x78, 0x74, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x43, 0x6c, 0x61, 0x73, 0x73, 0x52, 0x07, 0x63, 0x6c, 0x61, 0x73,
	0x73, 0x65, 0x73, 0x22, 0x51, 0x0a, 0x05, 0x43, 0x6c, 0x61, 0x73, 0x73, 0x12, 0x12, 0x0a, 0x04,
	0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x12, 0x14, 0x0a, 0x05, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x64,
	0x65, 0x6e, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x02, 0x52, 0x0a, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x64, 0x65, 0x6e, 0x63, 0x65, 0x32, 0x6e, 0x0a, 0x0a, 0x43, 0x6c, 0x61, 0x73, 0x73, 0x69,
	0x66, 0x69, 0x65, 0x72, 0x12, 0x60, 0x0a, 0x0c, 0x43, 0x6c, 0x61, 0x73, 0x73, 0x69, 0x66, 0x79,
	0x54, 0x65, 0x78, 0x74, 0x12, 0x27, 0x2e, 0x74, 0x65, 0x78, 0x74, 0x63, 0x6c, 0x61, 0x73, 0x73,
	0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x43, 0x6c, 0x61, 0x73, 0x73, 0x69,
	0x66, 0x79, 0x54, 0x65, 0x78, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x25, 0x2e,
	0x74, 0x65, 0x78, 0x74, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x2e, 0x43, 0x6c, 0x61, 0x73, 0x73, 0x69, 0x66, 0x79, 0x54, 0x65, 0x78, 0x74, 0x52,
	0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x42, 0x42, 0x5a, 0x40, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x53, 0x70, 0x65, 0x63, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64,
	0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x6c, 0x69, 0x73, 0x74, 0x2f, 0x77, 0x68, 0x61, 0x74, 0x73,
	0x6e, 0x65, 0x77, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x65, 0x78, 0x74, 0x63, 0x6c, 0x61, 0x73,
	0x73, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_textclassification_proto_rawDescOnce sync.Once
	file_textclassification_proto_rawDescData = file_textclassification_proto_rawDesc
)

func file_textclassification_proto_rawDescGZIP() []byte {
	file_textclassification_proto_rawDescOnce.Do(func() {
		file_textclassification_proto_rawDescData = protoimpl.X.CompressGZIP(file_textclassification_proto_rawDescData)
	})
	return file_textclassification_proto_rawDescData
}

var file_textclassification_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_textclassification_proto_goTypes = []interface{}{
	(*ClassifyTextRequest)(nil), // 0: textclassification.ClassifyTextRequest
	(*ClassifyTextReply)(nil),   // 1: textclassification.ClassifyTextReply
	(*Class)(nil),               // 2: textclassification.Class
}
var file_textclassification_proto_depIdxs = []int32{
	2, // 0: textclassification.ClassifyTextReply.classes:type_name -> textclassification.Class
	0, // 1: textclassification.Classifier.ClassifyText:input_type -> textclassification.ClassifyTextRequest
	1, // 2: textclassification.Classifier.ClassifyText:output_type -> textclassification.ClassifyTextReply
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_textclassification_proto_init() }
func file_textclassification_proto_init() {
	if File_textclassification_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_textclassification_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClassifyTextRequest); i {
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
		file_textclassification_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClassifyTextReply); i {
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
		file_textclassification_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Class); i {
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
			RawDescriptor: file_textclassification_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_textclassification_proto_goTypes,
		DependencyIndexes: file_textclassification_proto_depIdxs,
		MessageInfos:      file_textclassification_proto_msgTypes,
	}.Build()
	File_textclassification_proto = out.File
	file_textclassification_proto_rawDesc = nil
	file_textclassification_proto_goTypes = nil
	file_textclassification_proto_depIdxs = nil
}