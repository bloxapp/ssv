// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.13.0
// source: github.com/bloxapp/ssv/ibft/proto/msgs.proto

package proto

import (
	reflect "reflect"
	sync "sync"

	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type RoundState int32

const (
	RoundState_NotStarted  RoundState = 0
	RoundState_PrePrepare  RoundState = 1
	RoundState_Prepare     RoundState = 2
	RoundState_Commit      RoundState = 3
	RoundState_ChangeRound RoundState = 4
	RoundState_Decided     RoundState = 5
)

// Enum value maps for RoundState.
var (
	RoundState_name = map[int32]string{
		0: "NotStarted",
		1: "PrePrepare",
		2: "Prepare",
		3: "Commit",
		4: "ChangeRound",
		5: "Decided",
	}
	RoundState_value = map[string]int32{
		"NotStarted":  0,
		"PrePrepare":  1,
		"Prepare":     2,
		"Commit":      3,
		"ChangeRound": 4,
		"Decided":     5,
	}
)

func (x RoundState) Enum() *RoundState {
	p := new(RoundState)
	*p = x
	return p
}

func (x RoundState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (RoundState) Descriptor() protoreflect.EnumDescriptor {
	return file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_enumTypes[0].Descriptor()
}

func (RoundState) Type() protoreflect.EnumType {
	return &file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_enumTypes[0]
}

func (x RoundState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use RoundState.Descriptor instead.
func (RoundState) EnumDescriptor() ([]byte, []int) {
	return file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_rawDescGZIP(), []int{0}
}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type           RoundState `protobuf:"varint,1,opt,name=type,proto3,enum=proto.RoundState" json:"type,omitempty"`
	Round          uint64     `protobuf:"varint,2,opt,name=round,proto3" json:"round,omitempty"`
	Lambda         []byte     `protobuf:"bytes,3,opt,name=lambda,proto3" json:"lambda,omitempty"`
	PreviousLambda []byte     `protobuf:"bytes,4,opt,name=previous_lambda,json=previousLambda,proto3" json:"previous_lambda,omitempty"` // previous_lambda could be compared to prev block hash, to build instances as a chain
	Value          []byte     `protobuf:"bytes,5,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_msgTypes[0]
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
	return file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_rawDescGZIP(), []int{0}
}

func (x *Message) GetType() RoundState {
	if x != nil {
		return x.Type
	}
	return RoundState_NotStarted
}

func (x *Message) GetRound() uint64 {
	if x != nil {
		return x.Round
	}
	return 0
}

func (x *Message) GetLambda() []byte {
	if x != nil {
		return x.Lambda
	}
	return nil
}

func (x *Message) GetPreviousLambda() []byte {
	if x != nil {
		return x.PreviousLambda
	}
	return nil
}

func (x *Message) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}

type SignedMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message   *Message `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	Signature []byte   `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
	SignerIds []uint64 `protobuf:"varint,3,rep,packed,name=signer_ids,json=signerIds,proto3" json:"signer_ids,omitempty"`
}

func (x *SignedMessage) Reset() {
	*x = SignedMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignedMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignedMessage) ProtoMessage() {}

func (x *SignedMessage) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignedMessage.ProtoReflect.Descriptor instead.
func (*SignedMessage) Descriptor() ([]byte, []int) {
	return file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_rawDescGZIP(), []int{1}
}

func (x *SignedMessage) GetMessage() *Message {
	if x != nil {
		return x.Message
	}
	return nil
}

func (x *SignedMessage) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *SignedMessage) GetSignerIds() []uint64 {
	if x != nil {
		return x.SignerIds
	}
	return nil
}

type ChangeRoundData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PreparedRound    uint64   `protobuf:"varint,1,opt,name=prepared_round,json=preparedRound,proto3" json:"prepared_round,omitempty"`
	PreparedValue    []byte   `protobuf:"bytes,2,opt,name=prepared_value,json=preparedValue,proto3" json:"prepared_value,omitempty"`
	JustificationMsg *Message `protobuf:"bytes,3,opt,name=justification_msg,json=justificationMsg,proto3" json:"justification_msg,omitempty"`
	JustificationSig []byte   `protobuf:"bytes,4,opt,name=justification_sig,json=justificationSig,proto3" json:"justification_sig,omitempty"`
	SignerIds        []uint64 `protobuf:"varint,5,rep,packed,name=signer_ids,json=signerIds,proto3" json:"signer_ids,omitempty"`
}

func (x *ChangeRoundData) Reset() {
	*x = ChangeRoundData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChangeRoundData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChangeRoundData) ProtoMessage() {}

func (x *ChangeRoundData) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChangeRoundData.ProtoReflect.Descriptor instead.
func (*ChangeRoundData) Descriptor() ([]byte, []int) {
	return file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_rawDescGZIP(), []int{2}
}

func (x *ChangeRoundData) GetPreparedRound() uint64 {
	if x != nil {
		return x.PreparedRound
	}
	return 0
}

func (x *ChangeRoundData) GetPreparedValue() []byte {
	if x != nil {
		return x.PreparedValue
	}
	return nil
}

func (x *ChangeRoundData) GetJustificationMsg() *Message {
	if x != nil {
		return x.JustificationMsg
	}
	return nil
}

func (x *ChangeRoundData) GetJustificationSig() []byte {
	if x != nil {
		return x.JustificationSig
	}
	return nil
}

func (x *ChangeRoundData) GetSignerIds() []uint64 {
	if x != nil {
		return x.SignerIds
	}
	return nil
}

var File_github_com_bloxapp_ssv_ibft_proto_msgs_proto protoreflect.FileDescriptor

var file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_rawDesc = []byte{
	0x0a, 0x2c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x6c, 0x6f,
	0x78, 0x61, 0x70, 0x70, 0x2f, 0x73, 0x73, 0x76, 0x2f, 0x69, 0x62, 0x66, 0x74, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x67, 0x6f, 0x67, 0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x67, 0x6f, 0x67, 0x6f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x67, 0x6f, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x9d, 0x01, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x12, 0x25, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x11,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x6f, 0x75, 0x6e, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x72, 0x6f, 0x75, 0x6e, 0x64, 0x12, 0x16, 0x0a,
	0x06, 0x6c, 0x61, 0x6d, 0x62, 0x64, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x6c,
	0x61, 0x6d, 0x62, 0x64, 0x61, 0x12, 0x27, 0x0a, 0x0f, 0x70, 0x72, 0x65, 0x76, 0x69, 0x6f, 0x75,
	0x73, 0x5f, 0x6c, 0x61, 0x6d, 0x62, 0x64, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e,
	0x70, 0x72, 0x65, 0x76, 0x69, 0x6f, 0x75, 0x73, 0x4c, 0x61, 0x6d, 0x62, 0x64, 0x61, 0x12, 0x14,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x22, 0x82, 0x01, 0x0a, 0x0d, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x2e, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x42, 0x04, 0xc8, 0xde, 0x1f, 0x00, 0x52, 0x07, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x22, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74,
	0x75, 0x72, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x04, 0xc8, 0xde, 0x1f, 0x00, 0x52,
	0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x69,
	0x67, 0x6e, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x04, 0x52, 0x09,
	0x73, 0x69, 0x67, 0x6e, 0x65, 0x72, 0x49, 0x64, 0x73, 0x22, 0xfa, 0x01, 0x0a, 0x0f, 0x43, 0x68,
	0x61, 0x6e, 0x67, 0x65, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x44, 0x61, 0x74, 0x61, 0x12, 0x25, 0x0a,
	0x0e, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x64, 0x5f, 0x72, 0x6f, 0x75, 0x6e, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0d, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x64, 0x52,
	0x6f, 0x75, 0x6e, 0x64, 0x12, 0x2b, 0x0a, 0x0e, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x64,
	0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x04, 0xc8, 0xde,
	0x1f, 0x00, 0x52, 0x0d, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x64, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x12, 0x41, 0x0a, 0x11, 0x6a, 0x75, 0x73, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x5f, 0x6d, 0x73, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x42, 0x04, 0xc8, 0xde,
	0x1f, 0x00, 0x52, 0x10, 0x6a, 0x75, 0x73, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x4d, 0x73, 0x67, 0x12, 0x31, 0x0a, 0x11, 0x6a, 0x75, 0x73, 0x74, 0x69, 0x66, 0x69, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x73, 0x69, 0x67, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x42,
	0x04, 0xc8, 0xde, 0x1f, 0x00, 0x52, 0x10, 0x6a, 0x75, 0x73, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x53, 0x69, 0x67, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x69, 0x67, 0x6e, 0x65,
	0x72, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x04, 0x52, 0x09, 0x73, 0x69, 0x67,
	0x6e, 0x65, 0x72, 0x49, 0x64, 0x73, 0x2a, 0x63, 0x0a, 0x0a, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x53,
	0x74, 0x61, 0x74, 0x65, 0x12, 0x0e, 0x0a, 0x0a, 0x4e, 0x6f, 0x74, 0x53, 0x74, 0x61, 0x72, 0x74,
	0x65, 0x64, 0x10, 0x00, 0x12, 0x0e, 0x0a, 0x0a, 0x50, 0x72, 0x65, 0x50, 0x72, 0x65, 0x70, 0x61,
	0x72, 0x65, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x50, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x10,
	0x02, 0x12, 0x0a, 0x0a, 0x06, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x10, 0x03, 0x12, 0x0f, 0x0a,
	0x0b, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x10, 0x04, 0x12, 0x0b,
	0x0a, 0x07, 0x44, 0x65, 0x63, 0x69, 0x64, 0x65, 0x64, 0x10, 0x05, 0x42, 0x23, 0x5a, 0x21, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x6c, 0x6f, 0x78, 0x61, 0x70,
	0x70, 0x2f, 0x73, 0x73, 0x76, 0x2f, 0x69, 0x62, 0x66, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_rawDescOnce sync.Once
	file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_rawDescData = file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_rawDesc
)

func file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_rawDescGZIP() []byte {
	file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_rawDescOnce.Do(func() {
		file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_rawDescData)
	})
	return file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_rawDescData
}

var file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_goTypes = []interface{}{
	(RoundState)(0),         // 0: proto.RoundState
	(*Message)(nil),         // 1: proto.Message
	(*SignedMessage)(nil),   // 2: proto.SignedMessage
	(*ChangeRoundData)(nil), // 3: proto.ChangeRoundData
}
var file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_depIdxs = []int32{
	0, // 0: proto.Message.type:type_name -> proto.RoundState
	1, // 1: proto.SignedMessage.message:type_name -> proto.Message
	1, // 2: proto.ChangeRoundData.justification_msg:type_name -> proto.Message
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_init() }
func file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_init() {
	if File_github_com_bloxapp_ssv_ibft_proto_msgs_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignedMessage); i {
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
		file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChangeRoundData); i {
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
			RawDescriptor: file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_goTypes,
		DependencyIndexes: file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_depIdxs,
		EnumInfos:         file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_enumTypes,
		MessageInfos:      file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_msgTypes,
	}.Build()
	File_github_com_bloxapp_ssv_ibft_proto_msgs_proto = out.File
	file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_rawDesc = nil
	file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_goTypes = nil
	file_github_com_bloxapp_ssv_ibft_proto_msgs_proto_depIdxs = nil
}
