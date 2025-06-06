// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.12.4
// source: contract.proto

package contract

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Contract represents a smart contract
type Contract struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	UserId    string                 `protobuf:"bytes,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Name      string                 `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Address   string                 `protobuf:"bytes,4,opt,name=address,proto3" json:"address,omitempty"`
	Chain     string                 `protobuf:"bytes,5,opt,name=chain,proto3" json:"chain,omitempty"`
	Abi       string                 `protobuf:"bytes,6,opt,name=abi,proto3" json:"abi,omitempty"`
	Bytecode  string                 `protobuf:"bytes,7,opt,name=bytecode,proto3" json:"bytecode,omitempty"`
	CreatedAt *timestamppb.Timestamp `protobuf:"bytes,8,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt *timestamppb.Timestamp `protobuf:"bytes,9,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
}

func (x *Contract) Reset() {
	*x = Contract{}
	if protoimpl.UnsafeEnabled {
		mi := &file_contract_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Contract) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Contract) ProtoMessage() {}

func (x *Contract) ProtoReflect() protoreflect.Message {
	mi := &file_contract_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Contract.ProtoReflect.Descriptor instead.
func (*Contract) Descriptor() ([]byte, []int) {
	return file_contract_proto_rawDescGZIP(), []int{0}
}

func (x *Contract) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Contract) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *Contract) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Contract) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *Contract) GetChain() string {
	if x != nil {
		return x.Chain
	}
	return ""
}

func (x *Contract) GetAbi() string {
	if x != nil {
		return x.Abi
	}
	return ""
}

func (x *Contract) GetBytecode() string {
	if x != nil {
		return x.Bytecode
	}
	return ""
}

func (x *Contract) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *Contract) GetUpdatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdatedAt
	}
	return nil
}

// Transaction represents a blockchain transaction (simplified for this service)
type Transaction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Hash     string `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
	From     string `protobuf:"bytes,3,opt,name=from,proto3" json:"from,omitempty"`
	To       string `protobuf:"bytes,4,opt,name=to,proto3" json:"to,omitempty"`
	Value    string `protobuf:"bytes,5,opt,name=value,proto3" json:"value,omitempty"`
	Gas      uint64 `protobuf:"varint,6,opt,name=gas,proto3" json:"gas,omitempty"`
	GasPrice string `protobuf:"bytes,7,opt,name=gas_price,json=gasPrice,proto3" json:"gas_price,omitempty"`
	Status   string `protobuf:"bytes,8,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *Transaction) Reset() {
	*x = Transaction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_contract_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Transaction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Transaction) ProtoMessage() {}

func (x *Transaction) ProtoReflect() protoreflect.Message {
	mi := &file_contract_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Transaction.ProtoReflect.Descriptor instead.
func (*Transaction) Descriptor() ([]byte, []int) {
	return file_contract_proto_rawDescGZIP(), []int{1}
}

func (x *Transaction) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Transaction) GetHash() string {
	if x != nil {
		return x.Hash
	}
	return ""
}

func (x *Transaction) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *Transaction) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

func (x *Transaction) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *Transaction) GetGas() uint64 {
	if x != nil {
		return x.Gas
	}
	return 0
}

func (x *Transaction) GetGasPrice() string {
	if x != nil {
		return x.GasPrice
	}
	return ""
}

func (x *Transaction) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

// DeployContractRequest represents a request to deploy a contract
type DeployContractRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserId     string   `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	WalletId   string   `protobuf:"bytes,2,opt,name=wallet_id,json=walletId,proto3" json:"wallet_id,omitempty"`
	Name       string   `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Chain      string   `protobuf:"bytes,4,opt,name=chain,proto3" json:"chain,omitempty"`
	Type       string   `protobuf:"bytes,5,opt,name=type,proto3" json:"type,omitempty"`
	Abi        string   `protobuf:"bytes,6,opt,name=abi,proto3" json:"abi,omitempty"`
	Bytecode   string   `protobuf:"bytes,7,opt,name=bytecode,proto3" json:"bytecode,omitempty"`
	Arguments  []string `protobuf:"bytes,8,rep,name=arguments,proto3" json:"arguments,omitempty"`
	Gas        uint64   `protobuf:"varint,9,opt,name=gas,proto3" json:"gas,omitempty"`
	GasPrice   string   `protobuf:"bytes,10,opt,name=gas_price,json=gasPrice,proto3" json:"gas_price,omitempty"`
	Passphrase string   `protobuf:"bytes,11,opt,name=passphrase,proto3" json:"passphrase,omitempty"`
}

func (x *DeployContractRequest) Reset() {
	*x = DeployContractRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_contract_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeployContractRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeployContractRequest) ProtoMessage() {}

func (x *DeployContractRequest) ProtoReflect() protoreflect.Message {
	mi := &file_contract_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeployContractRequest.ProtoReflect.Descriptor instead.
func (*DeployContractRequest) Descriptor() ([]byte, []int) {
	return file_contract_proto_rawDescGZIP(), []int{2}
}

func (x *DeployContractRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *DeployContractRequest) GetWalletId() string {
	if x != nil {
		return x.WalletId
	}
	return ""
}

func (x *DeployContractRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *DeployContractRequest) GetChain() string {
	if x != nil {
		return x.Chain
	}
	return ""
}

func (x *DeployContractRequest) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *DeployContractRequest) GetAbi() string {
	if x != nil {
		return x.Abi
	}
	return ""
}

func (x *DeployContractRequest) GetBytecode() string {
	if x != nil {
		return x.Bytecode
	}
	return ""
}

func (x *DeployContractRequest) GetArguments() []string {
	if x != nil {
		return x.Arguments
	}
	return nil
}

func (x *DeployContractRequest) GetGas() uint64 {
	if x != nil {
		return x.Gas
	}
	return 0
}

func (x *DeployContractRequest) GetGasPrice() string {
	if x != nil {
		return x.GasPrice
	}
	return ""
}

func (x *DeployContractRequest) GetPassphrase() string {
	if x != nil {
		return x.Passphrase
	}
	return ""
}

// DeployContractResponse represents a response to a deploy contract request
type DeployContractResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Contract    *Contract    `protobuf:"bytes,1,opt,name=contract,proto3" json:"contract,omitempty"`
	Transaction *Transaction `protobuf:"bytes,2,opt,name=transaction,proto3" json:"transaction,omitempty"`
}

func (x *DeployContractResponse) Reset() {
	*x = DeployContractResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_contract_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeployContractResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeployContractResponse) ProtoMessage() {}

func (x *DeployContractResponse) ProtoReflect() protoreflect.Message {
	mi := &file_contract_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeployContractResponse.ProtoReflect.Descriptor instead.
func (*DeployContractResponse) Descriptor() ([]byte, []int) {
	return file_contract_proto_rawDescGZIP(), []int{3}
}

func (x *DeployContractResponse) GetContract() *Contract {
	if x != nil {
		return x.Contract
	}
	return nil
}

func (x *DeployContractResponse) GetTransaction() *Transaction {
	if x != nil {
		return x.Transaction
	}
	return nil
}

// SmartContractServiceClient is the client API for SmartContractService service.
type SmartContractServiceClient interface {
	// DeployContract deploys a new smart contract
	DeployContract(ctx context.Context, in *DeployContractRequest, opts ...grpc.CallOption) (*DeployContractResponse, error)
}

type smartContractServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewSmartContractServiceClient(cc grpc.ClientConnInterface) SmartContractServiceClient {
	return &smartContractServiceClient{cc}
}

func (c *smartContractServiceClient) DeployContract(ctx context.Context, in *DeployContractRequest, opts ...grpc.CallOption) (*DeployContractResponse, error) {
	out := new(DeployContractResponse)
	err := c.cc.Invoke(ctx, "/contract.SmartContractService/DeployContract", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SmartContractServiceServer is the server API for SmartContractService service.
type SmartContractServiceServer interface {
	// DeployContract deploys a new smart contract
	DeployContract(context.Context, *DeployContractRequest) (*DeployContractResponse, error)
}

// UnimplementedSmartContractServiceServer can be embedded to have forward compatible implementations.
type UnimplementedSmartContractServiceServer struct {
}

func (UnimplementedSmartContractServiceServer) DeployContract(context.Context, *DeployContractRequest) (*DeployContractResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeployContract not implemented")
}

func RegisterSmartContractServiceServer(s grpc.ServiceRegistrar, srv SmartContractServiceServer) {
	s.RegisterService(&SmartContractService_ServiceDesc, srv)
}

func _SmartContractService_DeployContract_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeployContractRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SmartContractServiceServer).DeployContract(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/contract.SmartContractService/DeployContract",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SmartContractServiceServer).DeployContract(ctx, req.(*DeployContractRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var SmartContractService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "contract.SmartContractService",
	HandlerType: (*SmartContractServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "DeployContract",
			Handler:    _SmartContractService_DeployContract_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "contract.proto",
}

var file_contract_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_contract_proto_rawDescGZIP = func() []byte { return nil }
