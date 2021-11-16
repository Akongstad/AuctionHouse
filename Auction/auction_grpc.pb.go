// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package Auction

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

// AuctionHouseClient is the client API for AuctionHouse service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AuctionHouseClient interface {
	OpenConnection(ctx context.Context, in *Connect, opts ...grpc.CallOption) (AuctionHouse_OpenConnectionClient, error)
	CloseConnection(ctx context.Context, in *User, opts ...grpc.CallOption) (*Close, error)
	Bid(ctx context.Context, in *BidMessage, opts ...grpc.CallOption) (*BidReply, error)
	Result(ctx context.Context, in *ResultMessage, opts ...grpc.CallOption) (*ResultMessage, error)
	Broadcast(ctx context.Context, in *Message, opts ...grpc.CallOption) (*Close, error)
}

type auctionHouseClient struct {
	cc grpc.ClientConnInterface
}

func NewAuctionHouseClient(cc grpc.ClientConnInterface) AuctionHouseClient {
	return &auctionHouseClient{cc}
}

func (c *auctionHouseClient) OpenConnection(ctx context.Context, in *Connect, opts ...grpc.CallOption) (AuctionHouse_OpenConnectionClient, error) {
	stream, err := c.cc.NewStream(ctx, &AuctionHouse_ServiceDesc.Streams[0], "/Auction.AuctionHouse/OpenConnection", opts...)
	if err != nil {
		return nil, err
	}
	x := &auctionHouseOpenConnectionClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type AuctionHouse_OpenConnectionClient interface {
	Recv() (*Message, error)
	grpc.ClientStream
}

type auctionHouseOpenConnectionClient struct {
	grpc.ClientStream
}

func (x *auctionHouseOpenConnectionClient) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *auctionHouseClient) CloseConnection(ctx context.Context, in *User, opts ...grpc.CallOption) (*Close, error) {
	out := new(Close)
	err := c.cc.Invoke(ctx, "/Auction.AuctionHouse/CloseConnection", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *auctionHouseClient) Bid(ctx context.Context, in *BidMessage, opts ...grpc.CallOption) (*BidReply, error) {
	out := new(BidReply)
	err := c.cc.Invoke(ctx, "/Auction.AuctionHouse/Bid", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *auctionHouseClient) Result(ctx context.Context, in *ResultMessage, opts ...grpc.CallOption) (*ResultMessage, error) {
	out := new(ResultMessage)
	err := c.cc.Invoke(ctx, "/Auction.AuctionHouse/Result", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *auctionHouseClient) Broadcast(ctx context.Context, in *Message, opts ...grpc.CallOption) (*Close, error) {
	out := new(Close)
	err := c.cc.Invoke(ctx, "/Auction.AuctionHouse/Broadcast", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuctionHouseServer is the server API for AuctionHouse service.
// All implementations must embed UnimplementedAuctionHouseServer
// for forward compatibility
type AuctionHouseServer interface {
	OpenConnection(*Connect, AuctionHouse_OpenConnectionServer) error
	CloseConnection(context.Context, *User) (*Close, error)
	Bid(context.Context, *BidMessage) (*BidReply, error)
	Result(context.Context, *ResultMessage) (*ResultMessage, error)
	Broadcast(context.Context, *Message) (*Close, error)
	mustEmbedUnimplementedAuctionHouseServer()
}

// UnimplementedAuctionHouseServer must be embedded to have forward compatible implementations.
type UnimplementedAuctionHouseServer struct {
}

func (UnimplementedAuctionHouseServer) OpenConnection(*Connect, AuctionHouse_OpenConnectionServer) error {
	return status.Errorf(codes.Unimplemented, "method OpenConnection not implemented")
}
func (UnimplementedAuctionHouseServer) CloseConnection(context.Context, *User) (*Close, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CloseConnection not implemented")
}
func (UnimplementedAuctionHouseServer) Bid(context.Context, *BidMessage) (*BidReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Bid not implemented")
}
func (UnimplementedAuctionHouseServer) Result(context.Context, *ResultMessage) (*ResultMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Result not implemented")
}
func (UnimplementedAuctionHouseServer) Broadcast(context.Context, *Message) (*Close, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Broadcast not implemented")
}
func (UnimplementedAuctionHouseServer) mustEmbedUnimplementedAuctionHouseServer() {}

// UnsafeAuctionHouseServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AuctionHouseServer will
// result in compilation errors.
type UnsafeAuctionHouseServer interface {
	mustEmbedUnimplementedAuctionHouseServer()
}

func RegisterAuctionHouseServer(s grpc.ServiceRegistrar, srv AuctionHouseServer) {
	s.RegisterService(&AuctionHouse_ServiceDesc, srv)
}

func _AuctionHouse_OpenConnection_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Connect)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(AuctionHouseServer).OpenConnection(m, &auctionHouseOpenConnectionServer{stream})
}

type AuctionHouse_OpenConnectionServer interface {
	Send(*Message) error
	grpc.ServerStream
}

type auctionHouseOpenConnectionServer struct {
	grpc.ServerStream
}

func (x *auctionHouseOpenConnectionServer) Send(m *Message) error {
	return x.ServerStream.SendMsg(m)
}

func _AuctionHouse_CloseConnection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(User)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuctionHouseServer).CloseConnection(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Auction.AuctionHouse/CloseConnection",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuctionHouseServer).CloseConnection(ctx, req.(*User))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuctionHouse_Bid_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BidMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuctionHouseServer).Bid(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Auction.AuctionHouse/Bid",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuctionHouseServer).Bid(ctx, req.(*BidMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuctionHouse_Result_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResultMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuctionHouseServer).Result(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Auction.AuctionHouse/Result",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuctionHouseServer).Result(ctx, req.(*ResultMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuctionHouse_Broadcast_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Message)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuctionHouseServer).Broadcast(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Auction.AuctionHouse/Broadcast",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuctionHouseServer).Broadcast(ctx, req.(*Message))
	}
	return interceptor(ctx, in, info, handler)
}

// AuctionHouse_ServiceDesc is the grpc.ServiceDesc for AuctionHouse service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AuctionHouse_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Auction.AuctionHouse",
	HandlerType: (*AuctionHouseServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CloseConnection",
			Handler:    _AuctionHouse_CloseConnection_Handler,
		},
		{
			MethodName: "Bid",
			Handler:    _AuctionHouse_Bid_Handler,
		},
		{
			MethodName: "Result",
			Handler:    _AuctionHouse_Result_Handler,
		},
		{
			MethodName: "Broadcast",
			Handler:    _AuctionHouse_Broadcast_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "OpenConnection",
			Handler:       _AuctionHouse_OpenConnection_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "Auction/auction.proto",
}
