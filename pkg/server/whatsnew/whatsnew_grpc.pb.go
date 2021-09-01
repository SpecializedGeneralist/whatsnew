// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package whatsnew

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

// WhatsnewClient is the client API for Whatsnew service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WhatsnewClient interface {
	GetFeeds(ctx context.Context, in *GetFeedsRequest, opts ...grpc.CallOption) (*GetFeedsResponse, error)
	CreateFeeds(ctx context.Context, in *CreateFeedsRequest, opts ...grpc.CallOption) (*CreateFeedsResponse, error)
	CreateFeed(ctx context.Context, in *CreateFeedRequest, opts ...grpc.CallOption) (*CreateFeedResponse, error)
	GetFeed(ctx context.Context, in *GetFeedRequest, opts ...grpc.CallOption) (*GetFeedResponse, error)
	UpdateFeed(ctx context.Context, in *UpdateFeedRequest, opts ...grpc.CallOption) (*UpdateFeedResponse, error)
	DeleteFeed(ctx context.Context, in *DeleteFeedRequest, opts ...grpc.CallOption) (*DeleteFeedResponse, error)
	GetUserTwitterSources(ctx context.Context, in *GetUserTwitterSourcesRequest, opts ...grpc.CallOption) (*GetUserTwitterSourcesResponse, error)
	CreateUserTwitterSources(ctx context.Context, in *CreateUserTwitterSourcesRequest, opts ...grpc.CallOption) (*CreateUserTwitterSourcesResponse, error)
	CreateUserTwitterSource(ctx context.Context, in *CreateUserTwitterSourceRequest, opts ...grpc.CallOption) (*CreateUserTwitterSourceResponse, error)
	GetUserTwitterSource(ctx context.Context, in *GetUserTwitterSourceRequest, opts ...grpc.CallOption) (*GetUserTwitterSourceResponse, error)
	UpdateUserTwitterSource(ctx context.Context, in *UpdateUserTwitterSourceRequest, opts ...grpc.CallOption) (*UpdateUserTwitterSourceResponse, error)
	DeleteUserTwitterSource(ctx context.Context, in *DeleteUserTwitterSourceRequest, opts ...grpc.CallOption) (*DeleteUserTwitterSourceResponse, error)
	GetQueryTwitterSources(ctx context.Context, in *GetQueryTwitterSourcesRequest, opts ...grpc.CallOption) (*GetQueryTwitterSourcesResponse, error)
	CreateQueryTwitterSources(ctx context.Context, in *CreateQueryTwitterSourcesRequest, opts ...grpc.CallOption) (*CreateQueryTwitterSourcesResponse, error)
	CreateQueryTwitterSource(ctx context.Context, in *CreateQueryTwitterSourceRequest, opts ...grpc.CallOption) (*CreateQueryTwitterSourceResponse, error)
	GetQueryTwitterSource(ctx context.Context, in *GetQueryTwitterSourceRequest, opts ...grpc.CallOption) (*GetQueryTwitterSourceResponse, error)
	UpdateQueryTwitterSource(ctx context.Context, in *UpdateQueryTwitterSourceRequest, opts ...grpc.CallOption) (*UpdateQueryTwitterSourceResponse, error)
	DeleteQueryTwitterSource(ctx context.Context, in *DeleteQueryTwitterSourceRequest, opts ...grpc.CallOption) (*DeleteQueryTwitterSourceResponse, error)
}

type whatsnewClient struct {
	cc grpc.ClientConnInterface
}

func NewWhatsnewClient(cc grpc.ClientConnInterface) WhatsnewClient {
	return &whatsnewClient{cc}
}

func (c *whatsnewClient) GetFeeds(ctx context.Context, in *GetFeedsRequest, opts ...grpc.CallOption) (*GetFeedsResponse, error) {
	out := new(GetFeedsResponse)
	err := c.cc.Invoke(ctx, "/whatsnew.Whatsnew/GetFeeds", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *whatsnewClient) CreateFeeds(ctx context.Context, in *CreateFeedsRequest, opts ...grpc.CallOption) (*CreateFeedsResponse, error) {
	out := new(CreateFeedsResponse)
	err := c.cc.Invoke(ctx, "/whatsnew.Whatsnew/CreateFeeds", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *whatsnewClient) CreateFeed(ctx context.Context, in *CreateFeedRequest, opts ...grpc.CallOption) (*CreateFeedResponse, error) {
	out := new(CreateFeedResponse)
	err := c.cc.Invoke(ctx, "/whatsnew.Whatsnew/CreateFeed", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *whatsnewClient) GetFeed(ctx context.Context, in *GetFeedRequest, opts ...grpc.CallOption) (*GetFeedResponse, error) {
	out := new(GetFeedResponse)
	err := c.cc.Invoke(ctx, "/whatsnew.Whatsnew/GetFeed", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *whatsnewClient) UpdateFeed(ctx context.Context, in *UpdateFeedRequest, opts ...grpc.CallOption) (*UpdateFeedResponse, error) {
	out := new(UpdateFeedResponse)
	err := c.cc.Invoke(ctx, "/whatsnew.Whatsnew/UpdateFeed", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *whatsnewClient) DeleteFeed(ctx context.Context, in *DeleteFeedRequest, opts ...grpc.CallOption) (*DeleteFeedResponse, error) {
	out := new(DeleteFeedResponse)
	err := c.cc.Invoke(ctx, "/whatsnew.Whatsnew/DeleteFeed", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *whatsnewClient) GetUserTwitterSources(ctx context.Context, in *GetUserTwitterSourcesRequest, opts ...grpc.CallOption) (*GetUserTwitterSourcesResponse, error) {
	out := new(GetUserTwitterSourcesResponse)
	err := c.cc.Invoke(ctx, "/whatsnew.Whatsnew/GetUserTwitterSources", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *whatsnewClient) CreateUserTwitterSources(ctx context.Context, in *CreateUserTwitterSourcesRequest, opts ...grpc.CallOption) (*CreateUserTwitterSourcesResponse, error) {
	out := new(CreateUserTwitterSourcesResponse)
	err := c.cc.Invoke(ctx, "/whatsnew.Whatsnew/CreateUserTwitterSources", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *whatsnewClient) CreateUserTwitterSource(ctx context.Context, in *CreateUserTwitterSourceRequest, opts ...grpc.CallOption) (*CreateUserTwitterSourceResponse, error) {
	out := new(CreateUserTwitterSourceResponse)
	err := c.cc.Invoke(ctx, "/whatsnew.Whatsnew/CreateUserTwitterSource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *whatsnewClient) GetUserTwitterSource(ctx context.Context, in *GetUserTwitterSourceRequest, opts ...grpc.CallOption) (*GetUserTwitterSourceResponse, error) {
	out := new(GetUserTwitterSourceResponse)
	err := c.cc.Invoke(ctx, "/whatsnew.Whatsnew/GetUserTwitterSource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *whatsnewClient) UpdateUserTwitterSource(ctx context.Context, in *UpdateUserTwitterSourceRequest, opts ...grpc.CallOption) (*UpdateUserTwitterSourceResponse, error) {
	out := new(UpdateUserTwitterSourceResponse)
	err := c.cc.Invoke(ctx, "/whatsnew.Whatsnew/UpdateUserTwitterSource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *whatsnewClient) DeleteUserTwitterSource(ctx context.Context, in *DeleteUserTwitterSourceRequest, opts ...grpc.CallOption) (*DeleteUserTwitterSourceResponse, error) {
	out := new(DeleteUserTwitterSourceResponse)
	err := c.cc.Invoke(ctx, "/whatsnew.Whatsnew/DeleteUserTwitterSource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *whatsnewClient) GetQueryTwitterSources(ctx context.Context, in *GetQueryTwitterSourcesRequest, opts ...grpc.CallOption) (*GetQueryTwitterSourcesResponse, error) {
	out := new(GetQueryTwitterSourcesResponse)
	err := c.cc.Invoke(ctx, "/whatsnew.Whatsnew/GetQueryTwitterSources", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *whatsnewClient) CreateQueryTwitterSources(ctx context.Context, in *CreateQueryTwitterSourcesRequest, opts ...grpc.CallOption) (*CreateQueryTwitterSourcesResponse, error) {
	out := new(CreateQueryTwitterSourcesResponse)
	err := c.cc.Invoke(ctx, "/whatsnew.Whatsnew/CreateQueryTwitterSources", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *whatsnewClient) CreateQueryTwitterSource(ctx context.Context, in *CreateQueryTwitterSourceRequest, opts ...grpc.CallOption) (*CreateQueryTwitterSourceResponse, error) {
	out := new(CreateQueryTwitterSourceResponse)
	err := c.cc.Invoke(ctx, "/whatsnew.Whatsnew/CreateQueryTwitterSource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *whatsnewClient) GetQueryTwitterSource(ctx context.Context, in *GetQueryTwitterSourceRequest, opts ...grpc.CallOption) (*GetQueryTwitterSourceResponse, error) {
	out := new(GetQueryTwitterSourceResponse)
	err := c.cc.Invoke(ctx, "/whatsnew.Whatsnew/GetQueryTwitterSource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *whatsnewClient) UpdateQueryTwitterSource(ctx context.Context, in *UpdateQueryTwitterSourceRequest, opts ...grpc.CallOption) (*UpdateQueryTwitterSourceResponse, error) {
	out := new(UpdateQueryTwitterSourceResponse)
	err := c.cc.Invoke(ctx, "/whatsnew.Whatsnew/UpdateQueryTwitterSource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *whatsnewClient) DeleteQueryTwitterSource(ctx context.Context, in *DeleteQueryTwitterSourceRequest, opts ...grpc.CallOption) (*DeleteQueryTwitterSourceResponse, error) {
	out := new(DeleteQueryTwitterSourceResponse)
	err := c.cc.Invoke(ctx, "/whatsnew.Whatsnew/DeleteQueryTwitterSource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WhatsnewServer is the server API for Whatsnew service.
// All implementations must embed UnimplementedWhatsnewServer
// for forward compatibility
type WhatsnewServer interface {
	GetFeeds(context.Context, *GetFeedsRequest) (*GetFeedsResponse, error)
	CreateFeeds(context.Context, *CreateFeedsRequest) (*CreateFeedsResponse, error)
	CreateFeed(context.Context, *CreateFeedRequest) (*CreateFeedResponse, error)
	GetFeed(context.Context, *GetFeedRequest) (*GetFeedResponse, error)
	UpdateFeed(context.Context, *UpdateFeedRequest) (*UpdateFeedResponse, error)
	DeleteFeed(context.Context, *DeleteFeedRequest) (*DeleteFeedResponse, error)
	GetUserTwitterSources(context.Context, *GetUserTwitterSourcesRequest) (*GetUserTwitterSourcesResponse, error)
	CreateUserTwitterSources(context.Context, *CreateUserTwitterSourcesRequest) (*CreateUserTwitterSourcesResponse, error)
	CreateUserTwitterSource(context.Context, *CreateUserTwitterSourceRequest) (*CreateUserTwitterSourceResponse, error)
	GetUserTwitterSource(context.Context, *GetUserTwitterSourceRequest) (*GetUserTwitterSourceResponse, error)
	UpdateUserTwitterSource(context.Context, *UpdateUserTwitterSourceRequest) (*UpdateUserTwitterSourceResponse, error)
	DeleteUserTwitterSource(context.Context, *DeleteUserTwitterSourceRequest) (*DeleteUserTwitterSourceResponse, error)
	GetQueryTwitterSources(context.Context, *GetQueryTwitterSourcesRequest) (*GetQueryTwitterSourcesResponse, error)
	CreateQueryTwitterSources(context.Context, *CreateQueryTwitterSourcesRequest) (*CreateQueryTwitterSourcesResponse, error)
	CreateQueryTwitterSource(context.Context, *CreateQueryTwitterSourceRequest) (*CreateQueryTwitterSourceResponse, error)
	GetQueryTwitterSource(context.Context, *GetQueryTwitterSourceRequest) (*GetQueryTwitterSourceResponse, error)
	UpdateQueryTwitterSource(context.Context, *UpdateQueryTwitterSourceRequest) (*UpdateQueryTwitterSourceResponse, error)
	DeleteQueryTwitterSource(context.Context, *DeleteQueryTwitterSourceRequest) (*DeleteQueryTwitterSourceResponse, error)
	mustEmbedUnimplementedWhatsnewServer()
}

// UnimplementedWhatsnewServer must be embedded to have forward compatible implementations.
type UnimplementedWhatsnewServer struct {
}

func (UnimplementedWhatsnewServer) GetFeeds(context.Context, *GetFeedsRequest) (*GetFeedsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFeeds not implemented")
}
func (UnimplementedWhatsnewServer) CreateFeeds(context.Context, *CreateFeedsRequest) (*CreateFeedsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateFeeds not implemented")
}
func (UnimplementedWhatsnewServer) CreateFeed(context.Context, *CreateFeedRequest) (*CreateFeedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateFeed not implemented")
}
func (UnimplementedWhatsnewServer) GetFeed(context.Context, *GetFeedRequest) (*GetFeedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFeed not implemented")
}
func (UnimplementedWhatsnewServer) UpdateFeed(context.Context, *UpdateFeedRequest) (*UpdateFeedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateFeed not implemented")
}
func (UnimplementedWhatsnewServer) DeleteFeed(context.Context, *DeleteFeedRequest) (*DeleteFeedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteFeed not implemented")
}
func (UnimplementedWhatsnewServer) GetUserTwitterSources(context.Context, *GetUserTwitterSourcesRequest) (*GetUserTwitterSourcesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserTwitterSources not implemented")
}
func (UnimplementedWhatsnewServer) CreateUserTwitterSources(context.Context, *CreateUserTwitterSourcesRequest) (*CreateUserTwitterSourcesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateUserTwitterSources not implemented")
}
func (UnimplementedWhatsnewServer) CreateUserTwitterSource(context.Context, *CreateUserTwitterSourceRequest) (*CreateUserTwitterSourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateUserTwitterSource not implemented")
}
func (UnimplementedWhatsnewServer) GetUserTwitterSource(context.Context, *GetUserTwitterSourceRequest) (*GetUserTwitterSourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserTwitterSource not implemented")
}
func (UnimplementedWhatsnewServer) UpdateUserTwitterSource(context.Context, *UpdateUserTwitterSourceRequest) (*UpdateUserTwitterSourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUserTwitterSource not implemented")
}
func (UnimplementedWhatsnewServer) DeleteUserTwitterSource(context.Context, *DeleteUserTwitterSourceRequest) (*DeleteUserTwitterSourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUserTwitterSource not implemented")
}
func (UnimplementedWhatsnewServer) GetQueryTwitterSources(context.Context, *GetQueryTwitterSourcesRequest) (*GetQueryTwitterSourcesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetQueryTwitterSources not implemented")
}
func (UnimplementedWhatsnewServer) CreateQueryTwitterSources(context.Context, *CreateQueryTwitterSourcesRequest) (*CreateQueryTwitterSourcesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateQueryTwitterSources not implemented")
}
func (UnimplementedWhatsnewServer) CreateQueryTwitterSource(context.Context, *CreateQueryTwitterSourceRequest) (*CreateQueryTwitterSourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateQueryTwitterSource not implemented")
}
func (UnimplementedWhatsnewServer) GetQueryTwitterSource(context.Context, *GetQueryTwitterSourceRequest) (*GetQueryTwitterSourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetQueryTwitterSource not implemented")
}
func (UnimplementedWhatsnewServer) UpdateQueryTwitterSource(context.Context, *UpdateQueryTwitterSourceRequest) (*UpdateQueryTwitterSourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateQueryTwitterSource not implemented")
}
func (UnimplementedWhatsnewServer) DeleteQueryTwitterSource(context.Context, *DeleteQueryTwitterSourceRequest) (*DeleteQueryTwitterSourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteQueryTwitterSource not implemented")
}
func (UnimplementedWhatsnewServer) mustEmbedUnimplementedWhatsnewServer() {}

// UnsafeWhatsnewServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WhatsnewServer will
// result in compilation errors.
type UnsafeWhatsnewServer interface {
	mustEmbedUnimplementedWhatsnewServer()
}

func RegisterWhatsnewServer(s grpc.ServiceRegistrar, srv WhatsnewServer) {
	s.RegisterService(&Whatsnew_ServiceDesc, srv)
}

func _Whatsnew_GetFeeds_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFeedsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WhatsnewServer).GetFeeds(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whatsnew.Whatsnew/GetFeeds",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WhatsnewServer).GetFeeds(ctx, req.(*GetFeedsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Whatsnew_CreateFeeds_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateFeedsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WhatsnewServer).CreateFeeds(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whatsnew.Whatsnew/CreateFeeds",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WhatsnewServer).CreateFeeds(ctx, req.(*CreateFeedsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Whatsnew_CreateFeed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateFeedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WhatsnewServer).CreateFeed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whatsnew.Whatsnew/CreateFeed",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WhatsnewServer).CreateFeed(ctx, req.(*CreateFeedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Whatsnew_GetFeed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFeedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WhatsnewServer).GetFeed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whatsnew.Whatsnew/GetFeed",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WhatsnewServer).GetFeed(ctx, req.(*GetFeedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Whatsnew_UpdateFeed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateFeedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WhatsnewServer).UpdateFeed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whatsnew.Whatsnew/UpdateFeed",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WhatsnewServer).UpdateFeed(ctx, req.(*UpdateFeedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Whatsnew_DeleteFeed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteFeedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WhatsnewServer).DeleteFeed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whatsnew.Whatsnew/DeleteFeed",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WhatsnewServer).DeleteFeed(ctx, req.(*DeleteFeedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Whatsnew_GetUserTwitterSources_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserTwitterSourcesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WhatsnewServer).GetUserTwitterSources(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whatsnew.Whatsnew/GetUserTwitterSources",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WhatsnewServer).GetUserTwitterSources(ctx, req.(*GetUserTwitterSourcesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Whatsnew_CreateUserTwitterSources_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateUserTwitterSourcesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WhatsnewServer).CreateUserTwitterSources(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whatsnew.Whatsnew/CreateUserTwitterSources",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WhatsnewServer).CreateUserTwitterSources(ctx, req.(*CreateUserTwitterSourcesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Whatsnew_CreateUserTwitterSource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateUserTwitterSourceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WhatsnewServer).CreateUserTwitterSource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whatsnew.Whatsnew/CreateUserTwitterSource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WhatsnewServer).CreateUserTwitterSource(ctx, req.(*CreateUserTwitterSourceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Whatsnew_GetUserTwitterSource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserTwitterSourceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WhatsnewServer).GetUserTwitterSource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whatsnew.Whatsnew/GetUserTwitterSource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WhatsnewServer).GetUserTwitterSource(ctx, req.(*GetUserTwitterSourceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Whatsnew_UpdateUserTwitterSource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateUserTwitterSourceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WhatsnewServer).UpdateUserTwitterSource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whatsnew.Whatsnew/UpdateUserTwitterSource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WhatsnewServer).UpdateUserTwitterSource(ctx, req.(*UpdateUserTwitterSourceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Whatsnew_DeleteUserTwitterSource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteUserTwitterSourceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WhatsnewServer).DeleteUserTwitterSource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whatsnew.Whatsnew/DeleteUserTwitterSource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WhatsnewServer).DeleteUserTwitterSource(ctx, req.(*DeleteUserTwitterSourceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Whatsnew_GetQueryTwitterSources_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetQueryTwitterSourcesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WhatsnewServer).GetQueryTwitterSources(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whatsnew.Whatsnew/GetQueryTwitterSources",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WhatsnewServer).GetQueryTwitterSources(ctx, req.(*GetQueryTwitterSourcesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Whatsnew_CreateQueryTwitterSources_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateQueryTwitterSourcesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WhatsnewServer).CreateQueryTwitterSources(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whatsnew.Whatsnew/CreateQueryTwitterSources",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WhatsnewServer).CreateQueryTwitterSources(ctx, req.(*CreateQueryTwitterSourcesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Whatsnew_CreateQueryTwitterSource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateQueryTwitterSourceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WhatsnewServer).CreateQueryTwitterSource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whatsnew.Whatsnew/CreateQueryTwitterSource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WhatsnewServer).CreateQueryTwitterSource(ctx, req.(*CreateQueryTwitterSourceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Whatsnew_GetQueryTwitterSource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetQueryTwitterSourceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WhatsnewServer).GetQueryTwitterSource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whatsnew.Whatsnew/GetQueryTwitterSource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WhatsnewServer).GetQueryTwitterSource(ctx, req.(*GetQueryTwitterSourceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Whatsnew_UpdateQueryTwitterSource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateQueryTwitterSourceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WhatsnewServer).UpdateQueryTwitterSource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whatsnew.Whatsnew/UpdateQueryTwitterSource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WhatsnewServer).UpdateQueryTwitterSource(ctx, req.(*UpdateQueryTwitterSourceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Whatsnew_DeleteQueryTwitterSource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteQueryTwitterSourceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WhatsnewServer).DeleteQueryTwitterSource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/whatsnew.Whatsnew/DeleteQueryTwitterSource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WhatsnewServer).DeleteQueryTwitterSource(ctx, req.(*DeleteQueryTwitterSourceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Whatsnew_ServiceDesc is the grpc.ServiceDesc for Whatsnew service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Whatsnew_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "whatsnew.Whatsnew",
	HandlerType: (*WhatsnewServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetFeeds",
			Handler:    _Whatsnew_GetFeeds_Handler,
		},
		{
			MethodName: "CreateFeeds",
			Handler:    _Whatsnew_CreateFeeds_Handler,
		},
		{
			MethodName: "CreateFeed",
			Handler:    _Whatsnew_CreateFeed_Handler,
		},
		{
			MethodName: "GetFeed",
			Handler:    _Whatsnew_GetFeed_Handler,
		},
		{
			MethodName: "UpdateFeed",
			Handler:    _Whatsnew_UpdateFeed_Handler,
		},
		{
			MethodName: "DeleteFeed",
			Handler:    _Whatsnew_DeleteFeed_Handler,
		},
		{
			MethodName: "GetUserTwitterSources",
			Handler:    _Whatsnew_GetUserTwitterSources_Handler,
		},
		{
			MethodName: "CreateUserTwitterSources",
			Handler:    _Whatsnew_CreateUserTwitterSources_Handler,
		},
		{
			MethodName: "CreateUserTwitterSource",
			Handler:    _Whatsnew_CreateUserTwitterSource_Handler,
		},
		{
			MethodName: "GetUserTwitterSource",
			Handler:    _Whatsnew_GetUserTwitterSource_Handler,
		},
		{
			MethodName: "UpdateUserTwitterSource",
			Handler:    _Whatsnew_UpdateUserTwitterSource_Handler,
		},
		{
			MethodName: "DeleteUserTwitterSource",
			Handler:    _Whatsnew_DeleteUserTwitterSource_Handler,
		},
		{
			MethodName: "GetQueryTwitterSources",
			Handler:    _Whatsnew_GetQueryTwitterSources_Handler,
		},
		{
			MethodName: "CreateQueryTwitterSources",
			Handler:    _Whatsnew_CreateQueryTwitterSources_Handler,
		},
		{
			MethodName: "CreateQueryTwitterSource",
			Handler:    _Whatsnew_CreateQueryTwitterSource_Handler,
		},
		{
			MethodName: "GetQueryTwitterSource",
			Handler:    _Whatsnew_GetQueryTwitterSource_Handler,
		},
		{
			MethodName: "UpdateQueryTwitterSource",
			Handler:    _Whatsnew_UpdateQueryTwitterSource_Handler,
		},
		{
			MethodName: "DeleteQueryTwitterSource",
			Handler:    _Whatsnew_DeleteQueryTwitterSource_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "whatsnew.proto",
}
