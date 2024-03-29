// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package textclassification

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ClassifierClient is the client API for Classifier service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ClassifierClient interface {
	// ClassifyText classifies a given text.
	ClassifyText(ctx context.Context, in *ClassifyTextRequest, opts ...grpc.CallOption) (*ClassifyTextReply, error)
}

type classifierClient struct {
	cc grpc.ClientConnInterface
}

func NewClassifierClient(cc grpc.ClientConnInterface) ClassifierClient {
	return &classifierClient{cc}
}

func (c *classifierClient) ClassifyText(ctx context.Context, in *ClassifyTextRequest, opts ...grpc.CallOption) (*ClassifyTextReply, error) {
	out := new(ClassifyTextReply)
	err := c.cc.Invoke(ctx, "/textclassification.Classifier/ClassifyText", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ClassifierServer is the server API for Classifier service.
// All implementations must embed UnimplementedClassifierServer
// for forward compatibility
type ClassifierServer interface {
	// ClassifyText classifies a given text.
	ClassifyText(context.Context, *ClassifyTextRequest) (*ClassifyTextReply, error)
	mustEmbedUnimplementedClassifierServer()
}

// UnimplementedClassifierServer must be embedded to have forward compatible implementations.
type UnimplementedClassifierServer struct {
}

func (UnimplementedClassifierServer) ClassifyText(context.Context, *ClassifyTextRequest) (*ClassifyTextReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClassifyText not implemented")
}
func (UnimplementedClassifierServer) mustEmbedUnimplementedClassifierServer() {}

// UnsafeClassifierServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ClassifierServer will
// result in compilation errors.
type UnsafeClassifierServer interface {
	mustEmbedUnimplementedClassifierServer()
}

func RegisterClassifierServer(s grpc.ServiceRegistrar, srv ClassifierServer) {
	s.RegisterService(&Classifier_ServiceDesc, srv)
}

func _Classifier_ClassifyText_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClassifyTextRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClassifierServer).ClassifyText(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/textclassification.Classifier/ClassifyText",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClassifierServer).ClassifyText(ctx, req.(*ClassifyTextRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Classifier_ServiceDesc is the grpc.ServiceDesc for Classifier service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Classifier_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "textclassification.Classifier",
	HandlerType: (*ClassifierServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ClassifyText",
			Handler:    _Classifier_ClassifyText_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "textclassification.proto",
}
