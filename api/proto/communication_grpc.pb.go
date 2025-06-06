// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v4.25.3
// source: communication.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	CommunicationService_SendMessage_FullMethodName               = "/communication.CommunicationService/SendMessage"
	CommunicationService_SubscribeToMessages_FullMethodName       = "/communication.CommunicationService/SubscribeToMessages"
	CommunicationService_BroadcastMessage_FullMethodName          = "/communication.CommunicationService/BroadcastMessage"
	CommunicationService_GetMessageHistory_FullMethodName         = "/communication.CommunicationService/GetMessageHistory"
	CommunicationService_RegisterService_FullMethodName           = "/communication.CommunicationService/RegisterService"
	CommunicationService_UnregisterService_FullMethodName         = "/communication.CommunicationService/UnregisterService"
	CommunicationService_GetActiveServices_FullMethodName         = "/communication.CommunicationService/GetActiveServices"
	CommunicationService_SendNotification_FullMethodName          = "/communication.CommunicationService/SendNotification"
	CommunicationService_GetCommunicationAnalytics_FullMethodName = "/communication.CommunicationService/GetCommunicationAnalytics"
)

// CommunicationServiceClient is the client API for CommunicationService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// Communication Hub Service for inter-service messaging
type CommunicationServiceClient interface {
	// Send message between services
	SendMessage(ctx context.Context, in *SendMessageRequest, opts ...grpc.CallOption) (*SendMessageResponse, error)
	// Subscribe to message stream
	SubscribeToMessages(ctx context.Context, in *SubscribeRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[MessageEvent], error)
	// Broadcast message to multiple services
	BroadcastMessage(ctx context.Context, in *BroadcastMessageRequest, opts ...grpc.CallOption) (*BroadcastMessageResponse, error)
	// Get message history
	GetMessageHistory(ctx context.Context, in *GetMessageHistoryRequest, opts ...grpc.CallOption) (*GetMessageHistoryResponse, error)
	// Register service for communication
	RegisterService(ctx context.Context, in *RegisterServiceRequest, opts ...grpc.CallOption) (*RegisterServiceResponse, error)
	// Unregister service
	UnregisterService(ctx context.Context, in *UnregisterServiceRequest, opts ...grpc.CallOption) (*UnregisterServiceResponse, error)
	// Get active services
	GetActiveServices(ctx context.Context, in *GetActiveServicesRequest, opts ...grpc.CallOption) (*GetActiveServicesResponse, error)
	// Send notification with AI routing
	SendNotification(ctx context.Context, in *SendNotificationRequest, opts ...grpc.CallOption) (*SendNotificationResponse, error)
	// Get communication analytics
	GetCommunicationAnalytics(ctx context.Context, in *GetCommunicationAnalyticsRequest, opts ...grpc.CallOption) (*GetCommunicationAnalyticsResponse, error)
}

type communicationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCommunicationServiceClient(cc grpc.ClientConnInterface) CommunicationServiceClient {
	return &communicationServiceClient{cc}
}

func (c *communicationServiceClient) SendMessage(ctx context.Context, in *SendMessageRequest, opts ...grpc.CallOption) (*SendMessageResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SendMessageResponse)
	err := c.cc.Invoke(ctx, CommunicationService_SendMessage_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *communicationServiceClient) SubscribeToMessages(ctx context.Context, in *SubscribeRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[MessageEvent], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &CommunicationService_ServiceDesc.Streams[0], CommunicationService_SubscribeToMessages_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[SubscribeRequest, MessageEvent]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type CommunicationService_SubscribeToMessagesClient = grpc.ServerStreamingClient[MessageEvent]

func (c *communicationServiceClient) BroadcastMessage(ctx context.Context, in *BroadcastMessageRequest, opts ...grpc.CallOption) (*BroadcastMessageResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BroadcastMessageResponse)
	err := c.cc.Invoke(ctx, CommunicationService_BroadcastMessage_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *communicationServiceClient) GetMessageHistory(ctx context.Context, in *GetMessageHistoryRequest, opts ...grpc.CallOption) (*GetMessageHistoryResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetMessageHistoryResponse)
	err := c.cc.Invoke(ctx, CommunicationService_GetMessageHistory_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *communicationServiceClient) RegisterService(ctx context.Context, in *RegisterServiceRequest, opts ...grpc.CallOption) (*RegisterServiceResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RegisterServiceResponse)
	err := c.cc.Invoke(ctx, CommunicationService_RegisterService_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *communicationServiceClient) UnregisterService(ctx context.Context, in *UnregisterServiceRequest, opts ...grpc.CallOption) (*UnregisterServiceResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UnregisterServiceResponse)
	err := c.cc.Invoke(ctx, CommunicationService_UnregisterService_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *communicationServiceClient) GetActiveServices(ctx context.Context, in *GetActiveServicesRequest, opts ...grpc.CallOption) (*GetActiveServicesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetActiveServicesResponse)
	err := c.cc.Invoke(ctx, CommunicationService_GetActiveServices_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *communicationServiceClient) SendNotification(ctx context.Context, in *SendNotificationRequest, opts ...grpc.CallOption) (*SendNotificationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SendNotificationResponse)
	err := c.cc.Invoke(ctx, CommunicationService_SendNotification_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *communicationServiceClient) GetCommunicationAnalytics(ctx context.Context, in *GetCommunicationAnalyticsRequest, opts ...grpc.CallOption) (*GetCommunicationAnalyticsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetCommunicationAnalyticsResponse)
	err := c.cc.Invoke(ctx, CommunicationService_GetCommunicationAnalytics_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CommunicationServiceServer is the server API for CommunicationService service.
// All implementations must embed UnimplementedCommunicationServiceServer
// for forward compatibility.
//
// Communication Hub Service for inter-service messaging
type CommunicationServiceServer interface {
	// Send message between services
	SendMessage(context.Context, *SendMessageRequest) (*SendMessageResponse, error)
	// Subscribe to message stream
	SubscribeToMessages(*SubscribeRequest, grpc.ServerStreamingServer[MessageEvent]) error
	// Broadcast message to multiple services
	BroadcastMessage(context.Context, *BroadcastMessageRequest) (*BroadcastMessageResponse, error)
	// Get message history
	GetMessageHistory(context.Context, *GetMessageHistoryRequest) (*GetMessageHistoryResponse, error)
	// Register service for communication
	RegisterService(context.Context, *RegisterServiceRequest) (*RegisterServiceResponse, error)
	// Unregister service
	UnregisterService(context.Context, *UnregisterServiceRequest) (*UnregisterServiceResponse, error)
	// Get active services
	GetActiveServices(context.Context, *GetActiveServicesRequest) (*GetActiveServicesResponse, error)
	// Send notification with AI routing
	SendNotification(context.Context, *SendNotificationRequest) (*SendNotificationResponse, error)
	// Get communication analytics
	GetCommunicationAnalytics(context.Context, *GetCommunicationAnalyticsRequest) (*GetCommunicationAnalyticsResponse, error)
	mustEmbedUnimplementedCommunicationServiceServer()
}

// UnimplementedCommunicationServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedCommunicationServiceServer struct{}

func (UnimplementedCommunicationServiceServer) SendMessage(context.Context, *SendMessageRequest) (*SendMessageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMessage not implemented")
}
func (UnimplementedCommunicationServiceServer) SubscribeToMessages(*SubscribeRequest, grpc.ServerStreamingServer[MessageEvent]) error {
	return status.Errorf(codes.Unimplemented, "method SubscribeToMessages not implemented")
}
func (UnimplementedCommunicationServiceServer) BroadcastMessage(context.Context, *BroadcastMessageRequest) (*BroadcastMessageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BroadcastMessage not implemented")
}
func (UnimplementedCommunicationServiceServer) GetMessageHistory(context.Context, *GetMessageHistoryRequest) (*GetMessageHistoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMessageHistory not implemented")
}
func (UnimplementedCommunicationServiceServer) RegisterService(context.Context, *RegisterServiceRequest) (*RegisterServiceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterService not implemented")
}
func (UnimplementedCommunicationServiceServer) UnregisterService(context.Context, *UnregisterServiceRequest) (*UnregisterServiceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnregisterService not implemented")
}
func (UnimplementedCommunicationServiceServer) GetActiveServices(context.Context, *GetActiveServicesRequest) (*GetActiveServicesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetActiveServices not implemented")
}
func (UnimplementedCommunicationServiceServer) SendNotification(context.Context, *SendNotificationRequest) (*SendNotificationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendNotification not implemented")
}
func (UnimplementedCommunicationServiceServer) GetCommunicationAnalytics(context.Context, *GetCommunicationAnalyticsRequest) (*GetCommunicationAnalyticsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCommunicationAnalytics not implemented")
}
func (UnimplementedCommunicationServiceServer) mustEmbedUnimplementedCommunicationServiceServer() {}
func (UnimplementedCommunicationServiceServer) testEmbeddedByValue()                              {}

// UnsafeCommunicationServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CommunicationServiceServer will
// result in compilation errors.
type UnsafeCommunicationServiceServer interface {
	mustEmbedUnimplementedCommunicationServiceServer()
}

func RegisterCommunicationServiceServer(s grpc.ServiceRegistrar, srv CommunicationServiceServer) {
	// If the following call pancis, it indicates UnimplementedCommunicationServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&CommunicationService_ServiceDesc, srv)
}

func _CommunicationService_SendMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendMessageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommunicationServiceServer).SendMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommunicationService_SendMessage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommunicationServiceServer).SendMessage(ctx, req.(*SendMessageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommunicationService_SubscribeToMessages_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SubscribeRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(CommunicationServiceServer).SubscribeToMessages(m, &grpc.GenericServerStream[SubscribeRequest, MessageEvent]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type CommunicationService_SubscribeToMessagesServer = grpc.ServerStreamingServer[MessageEvent]

func _CommunicationService_BroadcastMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BroadcastMessageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommunicationServiceServer).BroadcastMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommunicationService_BroadcastMessage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommunicationServiceServer).BroadcastMessage(ctx, req.(*BroadcastMessageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommunicationService_GetMessageHistory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMessageHistoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommunicationServiceServer).GetMessageHistory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommunicationService_GetMessageHistory_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommunicationServiceServer).GetMessageHistory(ctx, req.(*GetMessageHistoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommunicationService_RegisterService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterServiceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommunicationServiceServer).RegisterService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommunicationService_RegisterService_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommunicationServiceServer).RegisterService(ctx, req.(*RegisterServiceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommunicationService_UnregisterService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UnregisterServiceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommunicationServiceServer).UnregisterService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommunicationService_UnregisterService_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommunicationServiceServer).UnregisterService(ctx, req.(*UnregisterServiceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommunicationService_GetActiveServices_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetActiveServicesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommunicationServiceServer).GetActiveServices(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommunicationService_GetActiveServices_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommunicationServiceServer).GetActiveServices(ctx, req.(*GetActiveServicesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommunicationService_SendNotification_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendNotificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommunicationServiceServer).SendNotification(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommunicationService_SendNotification_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommunicationServiceServer).SendNotification(ctx, req.(*SendNotificationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommunicationService_GetCommunicationAnalytics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCommunicationAnalyticsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommunicationServiceServer).GetCommunicationAnalytics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommunicationService_GetCommunicationAnalytics_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommunicationServiceServer).GetCommunicationAnalytics(ctx, req.(*GetCommunicationAnalyticsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CommunicationService_ServiceDesc is the grpc.ServiceDesc for CommunicationService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CommunicationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "communication.CommunicationService",
	HandlerType: (*CommunicationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendMessage",
			Handler:    _CommunicationService_SendMessage_Handler,
		},
		{
			MethodName: "BroadcastMessage",
			Handler:    _CommunicationService_BroadcastMessage_Handler,
		},
		{
			MethodName: "GetMessageHistory",
			Handler:    _CommunicationService_GetMessageHistory_Handler,
		},
		{
			MethodName: "RegisterService",
			Handler:    _CommunicationService_RegisterService_Handler,
		},
		{
			MethodName: "UnregisterService",
			Handler:    _CommunicationService_UnregisterService_Handler,
		},
		{
			MethodName: "GetActiveServices",
			Handler:    _CommunicationService_GetActiveServices_Handler,
		},
		{
			MethodName: "SendNotification",
			Handler:    _CommunicationService_SendNotification_Handler,
		},
		{
			MethodName: "GetCommunicationAnalytics",
			Handler:    _CommunicationService_GetCommunicationAnalytics_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SubscribeToMessages",
			Handler:       _CommunicationService_SubscribeToMessages_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "communication.proto",
}
