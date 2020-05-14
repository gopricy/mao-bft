// Code generated by protoc-gen-go. DO NOT EDIT.
// source: maobft.proto

package pb

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// A merkle proof is a data structure that proves a content is stored in the Merkle tree.
type MerkleProof struct {
	// The root of Merkle tree, it's a SHA256 hash.
	Root []byte `protobuf:"bytes,1,opt,name=root,proto3" json:"root,omitempty"`
	// the proof pairs from bottom up.
	ProofPairs           []*ProofPair `protobuf:"bytes,2,rep,name=proof_pairs,json=proofPairs,proto3" json:"proof_pairs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *MerkleProof) Reset()         { *m = MerkleProof{} }
func (m *MerkleProof) String() string { return proto.CompactTextString(m) }
func (*MerkleProof) ProtoMessage()    {}
func (*MerkleProof) Descriptor() ([]byte, []int) {
	return fileDescriptor_e58dda516c73d392, []int{0}
}

func (m *MerkleProof) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MerkleProof.Unmarshal(m, b)
}
func (m *MerkleProof) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MerkleProof.Marshal(b, m, deterministic)
}
func (m *MerkleProof) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MerkleProof.Merge(m, src)
}
func (m *MerkleProof) XXX_Size() int {
	return xxx_messageInfo_MerkleProof.Size(m)
}
func (m *MerkleProof) XXX_DiscardUnknown() {
	xxx_messageInfo_MerkleProof.DiscardUnknown(m)
}

var xxx_messageInfo_MerkleProof proto.InternalMessageInfo

func (m *MerkleProof) GetRoot() []byte {
	if m != nil {
		return m.Root
	}
	return nil
}

func (m *MerkleProof) GetProofPairs() []*ProofPair {
	if m != nil {
		return m.ProofPairs
	}
	return nil
}

// ProofPair defines 2 hash values in the same layer of Merkle tree, that jointly calculate the parent.
// For example:
// * (parent primary)
// | \
// *  * secondary
// primary
type ProofPair struct {
	// primary is the content's ancestor node hash value.
	Primary []byte `protobuf:"bytes,1,opt,name=primary,proto3" json:"primary,omitempty"`
	// secondary is the helper of primary, in order to get the parent node's hash.
	Secondary            []byte   `protobuf:"bytes,2,opt,name=secondary,proto3" json:"secondary,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ProofPair) Reset()         { *m = ProofPair{} }
func (m *ProofPair) String() string { return proto.CompactTextString(m) }
func (*ProofPair) ProtoMessage()    {}
func (*ProofPair) Descriptor() ([]byte, []int) {
	return fileDescriptor_e58dda516c73d392, []int{1}
}

func (m *ProofPair) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProofPair.Unmarshal(m, b)
}
func (m *ProofPair) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProofPair.Marshal(b, m, deterministic)
}
func (m *ProofPair) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProofPair.Merge(m, src)
}
func (m *ProofPair) XXX_Size() int {
	return xxx_messageInfo_ProofPair.Size(m)
}
func (m *ProofPair) XXX_DiscardUnknown() {
	xxx_messageInfo_ProofPair.DiscardUnknown(m)
}

var xxx_messageInfo_ProofPair proto.InternalMessageInfo

func (m *ProofPair) GetPrimary() []byte {
	if m != nil {
		return m.Primary
	}
	return nil
}

func (m *ProofPair) GetSecondary() []byte {
	if m != nil {
		return m.Secondary
	}
	return nil
}

type Payload struct {
	MerkleProof          *MerkleProof `protobuf:"bytes,1,opt,name=merkle_proof,json=merkleProof,proto3" json:"merkle_proof,omitempty"`
	Data                 []byte       `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *Payload) Reset()         { *m = Payload{} }
func (m *Payload) String() string { return proto.CompactTextString(m) }
func (*Payload) ProtoMessage()    {}
func (*Payload) Descriptor() ([]byte, []int) {
	return fileDescriptor_e58dda516c73d392, []int{2}
}

func (m *Payload) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Payload.Unmarshal(m, b)
}
func (m *Payload) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Payload.Marshal(b, m, deterministic)
}
func (m *Payload) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Payload.Merge(m, src)
}
func (m *Payload) XXX_Size() int {
	return xxx_messageInfo_Payload.Size(m)
}
func (m *Payload) XXX_DiscardUnknown() {
	xxx_messageInfo_Payload.DiscardUnknown(m)
}

var xxx_messageInfo_Payload proto.InternalMessageInfo

func (m *Payload) GetMerkleProof() *MerkleProof {
	if m != nil {
		return m.MerkleProof
	}
	return nil
}

func (m *Payload) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type PrepareResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PrepareResponse) Reset()         { *m = PrepareResponse{} }
func (m *PrepareResponse) String() string { return proto.CompactTextString(m) }
func (*PrepareResponse) ProtoMessage()    {}
func (*PrepareResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_e58dda516c73d392, []int{3}
}

func (m *PrepareResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PrepareResponse.Unmarshal(m, b)
}
func (m *PrepareResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PrepareResponse.Marshal(b, m, deterministic)
}
func (m *PrepareResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PrepareResponse.Merge(m, src)
}
func (m *PrepareResponse) XXX_Size() int {
	return xxx_messageInfo_PrepareResponse.Size(m)
}
func (m *PrepareResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PrepareResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PrepareResponse proto.InternalMessageInfo

type EchoResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *EchoResponse) Reset()         { *m = EchoResponse{} }
func (m *EchoResponse) String() string { return proto.CompactTextString(m) }
func (*EchoResponse) ProtoMessage()    {}
func (*EchoResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_e58dda516c73d392, []int{4}
}

func (m *EchoResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EchoResponse.Unmarshal(m, b)
}
func (m *EchoResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EchoResponse.Marshal(b, m, deterministic)
}
func (m *EchoResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EchoResponse.Merge(m, src)
}
func (m *EchoResponse) XXX_Size() int {
	return xxx_messageInfo_EchoResponse.Size(m)
}
func (m *EchoResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_EchoResponse.DiscardUnknown(m)
}

var xxx_messageInfo_EchoResponse proto.InternalMessageInfo

type ReadyRequest struct {
	MerkleRoot           []byte   `protobuf:"bytes,1,opt,name=merkle_root,json=merkleRoot,proto3" json:"merkle_root,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReadyRequest) Reset()         { *m = ReadyRequest{} }
func (m *ReadyRequest) String() string { return proto.CompactTextString(m) }
func (*ReadyRequest) ProtoMessage()    {}
func (*ReadyRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_e58dda516c73d392, []int{5}
}

func (m *ReadyRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReadyRequest.Unmarshal(m, b)
}
func (m *ReadyRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReadyRequest.Marshal(b, m, deterministic)
}
func (m *ReadyRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReadyRequest.Merge(m, src)
}
func (m *ReadyRequest) XXX_Size() int {
	return xxx_messageInfo_ReadyRequest.Size(m)
}
func (m *ReadyRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ReadyRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ReadyRequest proto.InternalMessageInfo

func (m *ReadyRequest) GetMerkleRoot() []byte {
	if m != nil {
		return m.MerkleRoot
	}
	return nil
}

type ReadyResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReadyResponse) Reset()         { *m = ReadyResponse{} }
func (m *ReadyResponse) String() string { return proto.CompactTextString(m) }
func (*ReadyResponse) ProtoMessage()    {}
func (*ReadyResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_e58dda516c73d392, []int{6}
}

func (m *ReadyResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReadyResponse.Unmarshal(m, b)
}
func (m *ReadyResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReadyResponse.Marshal(b, m, deterministic)
}
func (m *ReadyResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReadyResponse.Merge(m, src)
}
func (m *ReadyResponse) XXX_Size() int {
	return xxx_messageInfo_ReadyResponse.Size(m)
}
func (m *ReadyResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ReadyResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ReadyResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MerkleProof)(nil), "pb.MerkleProof")
	proto.RegisterType((*ProofPair)(nil), "pb.ProofPair")
	proto.RegisterType((*Payload)(nil), "pb.Payload")
	proto.RegisterType((*PrepareResponse)(nil), "pb.PrepareResponse")
	proto.RegisterType((*EchoResponse)(nil), "pb.EchoResponse")
	proto.RegisterType((*ReadyRequest)(nil), "pb.ReadyRequest")
	proto.RegisterType((*ReadyResponse)(nil), "pb.ReadyResponse")
}

func init() { proto.RegisterFile("maobft.proto", fileDescriptor_e58dda516c73d392) }

var fileDescriptor_e58dda516c73d392 = []byte{
	// 303 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x51, 0xc1, 0x6e, 0xea, 0x30,
	0x10, 0x04, 0x1e, 0xaf, 0x88, 0x75, 0x28, 0xc5, 0xbd, 0x44, 0xa8, 0x52, 0x91, 0x2f, 0xe5, 0xd2,
	0x20, 0xb9, 0x87, 0xf6, 0x5e, 0xf5, 0x58, 0x29, 0xf8, 0x07, 0x22, 0x87, 0x18, 0x35, 0x2a, 0x61,
	0x5d, 0xdb, 0x3d, 0xe4, 0xef, 0xab, 0x38, 0x0e, 0x24, 0xa7, 0xcc, 0xee, 0x64, 0x66, 0x67, 0xd7,
	0x10, 0x55, 0x12, 0xf3, 0xa3, 0x4b, 0xb4, 0x41, 0x87, 0x74, 0xa2, 0x73, 0xb6, 0x07, 0xf2, 0xa9,
	0xcc, 0xf7, 0x49, 0xa5, 0x06, 0xf1, 0x48, 0x29, 0x4c, 0x0d, 0xa2, 0x8b, 0xc7, 0x9b, 0xf1, 0x36,
	0x12, 0x1e, 0xd3, 0x04, 0x88, 0x6e, 0xc8, 0x4c, 0xcb, 0xd2, 0xd8, 0x78, 0xb2, 0xf9, 0xb7, 0x25,
	0x7c, 0x91, 0xe8, 0x3c, 0xf1, 0x9a, 0x54, 0x96, 0x46, 0x80, 0xee, 0xa0, 0x65, 0xef, 0x30, 0xbf,
	0x10, 0x34, 0x86, 0x99, 0x36, 0x65, 0x25, 0x4d, 0x1d, 0x3c, 0xbb, 0x92, 0x3e, 0xc0, 0xdc, 0xaa,
	0x03, 0x9e, 0x8b, 0x86, 0x9b, 0x78, 0xee, 0xda, 0x60, 0x7b, 0x98, 0xa5, 0xb2, 0x3e, 0xa1, 0x2c,
	0x28, 0x87, 0xa8, 0xf2, 0x11, 0x33, 0x3f, 0xc4, 0xfb, 0x10, 0xbe, 0x6c, 0x02, 0xf4, 0xa2, 0x0b,
	0x52, 0x0d, 0xf7, 0x28, 0xa4, 0x93, 0xc1, 0xd7, 0x63, 0xb6, 0x82, 0x65, 0x6a, 0x94, 0x96, 0x46,
	0x09, 0x65, 0x35, 0x9e, 0xad, 0x62, 0xb7, 0x10, 0x7d, 0x1c, 0xbe, 0xf0, 0x52, 0xef, 0x20, 0x12,
	0x4a, 0x16, 0xb5, 0x50, 0x3f, 0xbf, 0xca, 0x3a, 0xfa, 0x08, 0xc1, 0x35, 0xeb, 0x5d, 0x05, 0xda,
	0x96, 0x40, 0x74, 0x6c, 0x09, 0x8b, 0x20, 0x68, 0x1d, 0xf8, 0x1b, 0xcc, 0xc2, 0x10, 0xfa, 0x7c,
	0x85, 0xc4, 0x5f, 0xab, 0xdd, 0x67, 0x7d, 0xdf, 0x9e, 0x6e, 0x98, 0x64, 0xc4, 0x77, 0x30, 0x6d,
	0xb2, 0xd0, 0xa7, 0xf0, 0x1d, 0x68, 0xee, 0x9a, 0x62, 0x10, 0x75, 0xc4, 0x5f, 0xe1, 0xbf, 0x9f,
	0x4d, 0x93, 0x0e, 0xf8, 0xbf, 0xfa, 0x0b, 0xac, 0x57, 0xbd, 0x4e, 0x27, 0xcc, 0x6f, 0xfc, 0xf3,
	0xbf, 0xfc, 0x05, 0x00, 0x00, 0xff, 0xff, 0xcd, 0x5b, 0xf1, 0x9b, 0x0e, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// PrepareClient is the client API for Prepare service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type PrepareClient interface {
	Prepare(ctx context.Context, in *Payload, opts ...grpc.CallOption) (*PrepareResponse, error)
}

type prepareClient struct {
	cc *grpc.ClientConn
}

func NewPrepareClient(cc *grpc.ClientConn) PrepareClient {
	return &prepareClient{cc}
}

func (c *prepareClient) Prepare(ctx context.Context, in *Payload, opts ...grpc.CallOption) (*PrepareResponse, error) {
	out := new(PrepareResponse)
	err := c.cc.Invoke(ctx, "/pb.Prepare/Prepare", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PrepareServer is the server API for Prepare service.
type PrepareServer interface {
	Prepare(context.Context, *Payload) (*PrepareResponse, error)
}

// UnimplementedPrepareServer can be embedded to have forward compatible implementations.
type UnimplementedPrepareServer struct {
}

func (*UnimplementedPrepareServer) Prepare(ctx context.Context, req *Payload) (*PrepareResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Prepare not implemented")
}

func RegisterPrepareServer(s *grpc.Server, srv PrepareServer) {
	s.RegisterService(&_Prepare_serviceDesc, srv)
}

func _Prepare_Prepare_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Payload)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PrepareServer).Prepare(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Prepare/Prepare",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PrepareServer).Prepare(ctx, req.(*Payload))
	}
	return interceptor(ctx, in, info, handler)
}

var _Prepare_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.Prepare",
	HandlerType: (*PrepareServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Prepare",
			Handler:    _Prepare_Prepare_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "maobft.proto",
}

// EchoClient is the client API for Echo service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type EchoClient interface {
	Echo(ctx context.Context, in *Payload, opts ...grpc.CallOption) (*EchoResponse, error)
}

type echoClient struct {
	cc *grpc.ClientConn
}

func NewEchoClient(cc *grpc.ClientConn) EchoClient {
	return &echoClient{cc}
}

func (c *echoClient) Echo(ctx context.Context, in *Payload, opts ...grpc.CallOption) (*EchoResponse, error) {
	out := new(EchoResponse)
	err := c.cc.Invoke(ctx, "/pb.Echo/Echo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EchoServer is the server API for Echo service.
type EchoServer interface {
	Echo(context.Context, *Payload) (*EchoResponse, error)
}

// UnimplementedEchoServer can be embedded to have forward compatible implementations.
type UnimplementedEchoServer struct {
}

func (*UnimplementedEchoServer) Echo(ctx context.Context, req *Payload) (*EchoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Echo not implemented")
}

func RegisterEchoServer(s *grpc.Server, srv EchoServer) {
	s.RegisterService(&_Echo_serviceDesc, srv)
}

func _Echo_Echo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Payload)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EchoServer).Echo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Echo/Echo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EchoServer).Echo(ctx, req.(*Payload))
	}
	return interceptor(ctx, in, info, handler)
}

var _Echo_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.Echo",
	HandlerType: (*EchoServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Echo",
			Handler:    _Echo_Echo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "maobft.proto",
}

// ReadyClient is the client API for Ready service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ReadyClient interface {
	Ready(ctx context.Context, in *ReadyRequest, opts ...grpc.CallOption) (*ReadyResponse, error)
}

type readyClient struct {
	cc *grpc.ClientConn
}

func NewReadyClient(cc *grpc.ClientConn) ReadyClient {
	return &readyClient{cc}
}

func (c *readyClient) Ready(ctx context.Context, in *ReadyRequest, opts ...grpc.CallOption) (*ReadyResponse, error) {
	out := new(ReadyResponse)
	err := c.cc.Invoke(ctx, "/pb.Ready/Ready", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReadyServer is the server API for Ready service.
type ReadyServer interface {
	Ready(context.Context, *ReadyRequest) (*ReadyResponse, error)
}

// UnimplementedReadyServer can be embedded to have forward compatible implementations.
type UnimplementedReadyServer struct {
}

func (*UnimplementedReadyServer) Ready(ctx context.Context, req *ReadyRequest) (*ReadyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ready not implemented")
}

func RegisterReadyServer(s *grpc.Server, srv ReadyServer) {
	s.RegisterService(&_Ready_serviceDesc, srv)
}

func _Ready_Ready_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReadyServer).Ready(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Ready/Ready",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReadyServer).Ready(ctx, req.(*ReadyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Ready_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.Ready",
	HandlerType: (*ReadyServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ready",
			Handler:    _Ready_Ready_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "maobft.proto",
}
