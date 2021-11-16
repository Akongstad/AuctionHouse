// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.17.3
// source: Auction/auction.proto

package Auction

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

type User struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        int64  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Name      string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Timestamp int32  `protobuf:"varint,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *User) Reset() {
	*x = User{}
	if protoimpl.UnsafeEnabled {
		mi := &file_Auction_auction_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *User) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*User) ProtoMessage() {}

func (x *User) ProtoReflect() protoreflect.Message {
	mi := &file_Auction_auction_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use User.ProtoReflect.Descriptor instead.
func (*User) Descriptor() ([]byte, []int) {
	return file_Auction_auction_proto_rawDescGZIP(), []int{0}
}

func (x *User) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *User) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *User) GetTimestamp() int32 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	User      *User `protobuf:"bytes,1,opt,name=user,proto3" json:"user,omitempty"`
	Timestamp int32 `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_Auction_auction_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_Auction_auction_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_Auction_auction_proto_rawDescGZIP(), []int{1}
}

func (x *Message) GetUser() *User {
	if x != nil {
		return x.User
	}
	return nil
}

func (x *Message) GetTimestamp() int32 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

type ResultMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Amount      int32 `protobuf:"varint,1,opt,name=amount,proto3" json:"amount,omitempty"`
	User        *User `protobuf:"bytes,2,opt,name=user,proto3" json:"user,omitempty"`
	Timestamp   int32 `protobuf:"varint,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	StillActive bool  `protobuf:"varint,4,opt,name=stillActive,proto3" json:"stillActive,omitempty"`
}

func (x *ResultMessage) Reset() {
	*x = ResultMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_Auction_auction_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResultMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResultMessage) ProtoMessage() {}

func (x *ResultMessage) ProtoReflect() protoreflect.Message {
	mi := &file_Auction_auction_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResultMessage.ProtoReflect.Descriptor instead.
func (*ResultMessage) Descriptor() ([]byte, []int) {
	return file_Auction_auction_proto_rawDescGZIP(), []int{2}
}

func (x *ResultMessage) GetAmount() int32 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *ResultMessage) GetUser() *User {
	if x != nil {
		return x.User
	}
	return nil
}

func (x *ResultMessage) GetTimestamp() int32 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *ResultMessage) GetStillActive() bool {
	if x != nil {
		return x.StillActive
	}
	return false
}

//1 for succes, 2 for fail, 3 for exception
type BidReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ReturnType int32 `protobuf:"varint,1,opt,name=returnType,proto3" json:"returnType,omitempty"`
	Timestamp  int32 `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *BidReply) Reset() {
	*x = BidReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_Auction_auction_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BidReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BidReply) ProtoMessage() {}

func (x *BidReply) ProtoReflect() protoreflect.Message {
	mi := &file_Auction_auction_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BidReply.ProtoReflect.Descriptor instead.
func (*BidReply) Descriptor() ([]byte, []int) {
	return file_Auction_auction_proto_rawDescGZIP(), []int{3}
}

func (x *BidReply) GetReturnType() int32 {
	if x != nil {
		return x.ReturnType
	}
	return 0
}

func (x *BidReply) GetTimestamp() int32 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

type BidMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Amount    int32 `protobuf:"varint,1,opt,name=amount,proto3" json:"amount,omitempty"`
	User      *User `protobuf:"bytes,2,opt,name=user,proto3" json:"user,omitempty"`
	Timestamp int32 `protobuf:"varint,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *BidMessage) Reset() {
	*x = BidMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_Auction_auction_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BidMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BidMessage) ProtoMessage() {}

func (x *BidMessage) ProtoReflect() protoreflect.Message {
	mi := &file_Auction_auction_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BidMessage.ProtoReflect.Descriptor instead.
func (*BidMessage) Descriptor() ([]byte, []int) {
	return file_Auction_auction_proto_rawDescGZIP(), []int{4}
}

func (x *BidMessage) GetAmount() int32 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *BidMessage) GetUser() *User {
	if x != nil {
		return x.User
	}
	return nil
}

func (x *BidMessage) GetTimestamp() int32 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

type Connect struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	User   *User `protobuf:"bytes,1,opt,name=user,proto3" json:"user,omitempty"`
	Active bool  `protobuf:"varint,2,opt,name=active,proto3" json:"active,omitempty"`
}

func (x *Connect) Reset() {
	*x = Connect{}
	if protoimpl.UnsafeEnabled {
		mi := &file_Auction_auction_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Connect) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Connect) ProtoMessage() {}

func (x *Connect) ProtoReflect() protoreflect.Message {
	mi := &file_Auction_auction_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Connect.ProtoReflect.Descriptor instead.
func (*Connect) Descriptor() ([]byte, []int) {
	return file_Auction_auction_proto_rawDescGZIP(), []int{5}
}

func (x *Connect) GetUser() *User {
	if x != nil {
		return x.User
	}
	return nil
}

func (x *Connect) GetActive() bool {
	if x != nil {
		return x.Active
	}
	return false
}

type Close struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Close) Reset() {
	*x = Close{}
	if protoimpl.UnsafeEnabled {
		mi := &file_Auction_auction_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Close) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Close) ProtoMessage() {}

func (x *Close) ProtoReflect() protoreflect.Message {
	mi := &file_Auction_auction_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Close.ProtoReflect.Descriptor instead.
func (*Close) Descriptor() ([]byte, []int) {
	return file_Auction_auction_proto_rawDescGZIP(), []int{6}
}

var File_Auction_auction_proto protoreflect.FileDescriptor

var file_Auction_auction_proto_rawDesc = []byte{
	0x0a, 0x15, 0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x61, 0x75, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x22, 0x48, 0x0a, 0x04, 0x55, 0x73, 0x65, 0x72, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x22, 0x4a, 0x0a, 0x07, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x21, 0x0a, 0x04, 0x75, 0x73, 0x65, 0x72, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x55, 0x73,
	0x65, 0x72, 0x52, 0x04, 0x75, 0x73, 0x65, 0x72, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x74, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x22, 0x8a, 0x01, 0x0a, 0x0d, 0x52, 0x65, 0x73, 0x75, 0x6c,
	0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75,
	0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74,
	0x12, 0x21, 0x0a, 0x04, 0x75, 0x73, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d,
	0x2e, 0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x52, 0x04, 0x75,
	0x73, 0x65, 0x72, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x12, 0x20, 0x0a, 0x0b, 0x73, 0x74, 0x69, 0x6c, 0x6c, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x73, 0x74, 0x69, 0x6c, 0x6c, 0x41, 0x63, 0x74,
	0x69, 0x76, 0x65, 0x22, 0x48, 0x0a, 0x08, 0x42, 0x69, 0x64, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12,
	0x1e, 0x0a, 0x0a, 0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x0a, 0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x12,
	0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x22, 0x65, 0x0a,
	0x0a, 0x42, 0x69, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x61,
	0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x61, 0x6d, 0x6f,
	0x75, 0x6e, 0x74, 0x12, 0x21, 0x0a, 0x04, 0x75, 0x73, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x0d, 0x2e, 0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x55, 0x73, 0x65, 0x72,
	0x52, 0x04, 0x75, 0x73, 0x65, 0x72, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x22, 0x44, 0x0a, 0x07, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x12,
	0x21, 0x0a, 0x04, 0x75, 0x73, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e,
	0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x52, 0x04, 0x75, 0x73,
	0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x06, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x22, 0x07, 0x0a, 0x05, 0x43, 0x6c,
	0x6f, 0x73, 0x65, 0x32, 0x9a, 0x02, 0x0a, 0x0c, 0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x48,
	0x6f, 0x75, 0x73, 0x65, 0x12, 0x38, 0x0a, 0x0e, 0x4f, 0x70, 0x65, 0x6e, 0x43, 0x6f, 0x6e, 0x6e,
	0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x10, 0x2e, 0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x2e, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x1a, 0x10, 0x2e, 0x41, 0x75, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0x32,
	0x0a, 0x0f, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x0d, 0x2e, 0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x55, 0x73, 0x65, 0x72,
	0x1a, 0x0e, 0x2e, 0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x43, 0x6c, 0x6f, 0x73, 0x65,
	0x22, 0x00, 0x12, 0x2f, 0x0a, 0x03, 0x42, 0x69, 0x64, 0x12, 0x13, 0x2e, 0x41, 0x75, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x42, 0x69, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x1a, 0x11,
	0x2e, 0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x42, 0x69, 0x64, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x22, 0x00, 0x12, 0x3a, 0x0a, 0x06, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x16, 0x2e,
	0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x1a, 0x16, 0x2e, 0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e,
	0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x00, 0x12,
	0x2f, 0x0a, 0x09, 0x42, 0x72, 0x6f, 0x61, 0x64, 0x63, 0x61, 0x73, 0x74, 0x12, 0x10, 0x2e, 0x41,
	0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x1a, 0x0e,
	0x2e, 0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x22, 0x00,
	0x42, 0x3e, 0x5a, 0x3c, 0x2f, 0x55, 0x73, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x6e, 0x64, 0x72, 0x65,
	0x61, 0x73, 0x6b, 0x6f, 0x6e, 0x67, 0x73, 0x74, 0x61, 0x64, 0x2f, 0x44, 0x6f, 0x63, 0x75, 0x6d,
	0x65, 0x6e, 0x74, 0x73, 0x2f, 0x67, 0x6f, 0x4c, 0x61, 0x6e, 0x67, 0x2f, 0x41, 0x75, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x48, 0x6f, 0x75, 0x73, 0x65, 0x2f, 0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_Auction_auction_proto_rawDescOnce sync.Once
	file_Auction_auction_proto_rawDescData = file_Auction_auction_proto_rawDesc
)

func file_Auction_auction_proto_rawDescGZIP() []byte {
	file_Auction_auction_proto_rawDescOnce.Do(func() {
		file_Auction_auction_proto_rawDescData = protoimpl.X.CompressGZIP(file_Auction_auction_proto_rawDescData)
	})
	return file_Auction_auction_proto_rawDescData
}

var file_Auction_auction_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_Auction_auction_proto_goTypes = []interface{}{
	(*User)(nil),          // 0: Auction.User
	(*Message)(nil),       // 1: Auction.Message
	(*ResultMessage)(nil), // 2: Auction.ResultMessage
	(*BidReply)(nil),      // 3: Auction.BidReply
	(*BidMessage)(nil),    // 4: Auction.BidMessage
	(*Connect)(nil),       // 5: Auction.Connect
	(*Close)(nil),         // 6: Auction.Close
}
var file_Auction_auction_proto_depIdxs = []int32{
	0, // 0: Auction.Message.user:type_name -> Auction.User
	0, // 1: Auction.ResultMessage.user:type_name -> Auction.User
	0, // 2: Auction.BidMessage.user:type_name -> Auction.User
	0, // 3: Auction.Connect.user:type_name -> Auction.User
	5, // 4: Auction.AuctionHouse.OpenConnection:input_type -> Auction.Connect
	0, // 5: Auction.AuctionHouse.CloseConnection:input_type -> Auction.User
	4, // 6: Auction.AuctionHouse.Bid:input_type -> Auction.BidMessage
	2, // 7: Auction.AuctionHouse.Result:input_type -> Auction.ResultMessage
	1, // 8: Auction.AuctionHouse.Broadcast:input_type -> Auction.Message
	1, // 9: Auction.AuctionHouse.OpenConnection:output_type -> Auction.Message
	6, // 10: Auction.AuctionHouse.CloseConnection:output_type -> Auction.Close
	3, // 11: Auction.AuctionHouse.Bid:output_type -> Auction.BidReply
	2, // 12: Auction.AuctionHouse.Result:output_type -> Auction.ResultMessage
	6, // 13: Auction.AuctionHouse.Broadcast:output_type -> Auction.Close
	9, // [9:14] is the sub-list for method output_type
	4, // [4:9] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_Auction_auction_proto_init() }
func file_Auction_auction_proto_init() {
	if File_Auction_auction_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_Auction_auction_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*User); i {
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
		file_Auction_auction_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Message); i {
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
		file_Auction_auction_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResultMessage); i {
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
		file_Auction_auction_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BidReply); i {
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
		file_Auction_auction_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BidMessage); i {
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
		file_Auction_auction_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Connect); i {
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
		file_Auction_auction_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Close); i {
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
			RawDescriptor: file_Auction_auction_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_Auction_auction_proto_goTypes,
		DependencyIndexes: file_Auction_auction_proto_depIdxs,
		MessageInfos:      file_Auction_auction_proto_msgTypes,
	}.Build()
	File_Auction_auction_proto = out.File
	file_Auction_auction_proto_rawDesc = nil
	file_Auction_auction_proto_goTypes = nil
	file_Auction_auction_proto_depIdxs = nil
}