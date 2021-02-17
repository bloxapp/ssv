// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.13.0
// source: github.com/bloxapp/ssv/ibft/proto/beacon.proto

package proto

import (
	reflect "reflect"
	sync "sync"

	proto "github.com/golang/protobuf/proto"
	v1alpha1 "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
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

type InputValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Data:
	//	*InputValue_AttestationData
	//	*InputValue_AggregationData
	//	*InputValue_BeaconBlock
	Data isInputValue_Data `protobuf_oneof:"data"`
	// Types that are assignable to SignedData:
	//	*InputValue_Attestation
	//	*InputValue_Aggregation
	//	*InputValue_Block
	SignedData isInputValue_SignedData `protobuf_oneof:"signed_data"`
}

func (x *InputValue) Reset() {
	*x = InputValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InputValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InputValue) ProtoMessage() {}

func (x *InputValue) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InputValue.ProtoReflect.Descriptor instead.
func (*InputValue) Descriptor() ([]byte, []int) {
	return file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_rawDescGZIP(), []int{0}
}

func (m *InputValue) GetData() isInputValue_Data {
	if m != nil {
		return m.Data
	}
	return nil
}

func (x *InputValue) GetAttestationData() *v1alpha1.AttestationData {
	if x, ok := x.GetData().(*InputValue_AttestationData); ok {
		return x.AttestationData
	}
	return nil
}

func (x *InputValue) GetAggregationData() *v1alpha1.AggregateAttestationAndProof {
	if x, ok := x.GetData().(*InputValue_AggregationData); ok {
		return x.AggregationData
	}
	return nil
}

func (x *InputValue) GetBeaconBlock() *v1alpha1.BeaconBlock {
	if x, ok := x.GetData().(*InputValue_BeaconBlock); ok {
		return x.BeaconBlock
	}
	return nil
}

func (m *InputValue) GetSignedData() isInputValue_SignedData {
	if m != nil {
		return m.SignedData
	}
	return nil
}

func (x *InputValue) GetAttestation() *v1alpha1.Attestation {
	if x, ok := x.GetSignedData().(*InputValue_Attestation); ok {
		return x.Attestation
	}
	return nil
}

func (x *InputValue) GetAggregation() *v1alpha1.SignedAggregateAttestationAndProof {
	if x, ok := x.GetSignedData().(*InputValue_Aggregation); ok {
		return x.Aggregation
	}
	return nil
}

func (x *InputValue) GetBlock() *v1alpha1.SignedBeaconBlock {
	if x, ok := x.GetSignedData().(*InputValue_Block); ok {
		return x.Block
	}
	return nil
}

type isInputValue_Data interface {
	isInputValue_Data()
}

type InputValue_AttestationData struct {
	AttestationData *v1alpha1.AttestationData `protobuf:"bytes,2,opt,name=attestation_data,json=attestationData,proto3,oneof"`
}

type InputValue_AggregationData struct {
	AggregationData *v1alpha1.AggregateAttestationAndProof `protobuf:"bytes,3,opt,name=aggregation_data,json=aggregationData,proto3,oneof"`
}

type InputValue_BeaconBlock struct {
	BeaconBlock *v1alpha1.BeaconBlock `protobuf:"bytes,4,opt,name=beacon_block,json=beaconBlock,proto3,oneof"`
}

func (*InputValue_AttestationData) isInputValue_Data() {}

func (*InputValue_AggregationData) isInputValue_Data() {}

func (*InputValue_BeaconBlock) isInputValue_Data() {}

type isInputValue_SignedData interface {
	isInputValue_SignedData()
}

type InputValue_Attestation struct {
	Attestation *v1alpha1.Attestation `protobuf:"bytes,5,opt,name=attestation,proto3,oneof"`
}

type InputValue_Aggregation struct {
	Aggregation *v1alpha1.SignedAggregateAttestationAndProof `protobuf:"bytes,6,opt,name=aggregation,proto3,oneof"`
}

type InputValue_Block struct {
	Block *v1alpha1.SignedBeaconBlock `protobuf:"bytes,7,opt,name=block,proto3,oneof"`
}

func (*InputValue_Attestation) isInputValue_SignedData() {}

func (*InputValue_Aggregation) isInputValue_SignedData() {}

func (*InputValue_Block) isInputValue_SignedData() {}

var File_github_com_bloxapp_ssv_ibft_proto_beacon_proto protoreflect.FileDescriptor

var file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_rawDesc = []byte{
	0x0a, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x6c, 0x6f,
	0x78, 0x61, 0x70, 0x70, 0x2f, 0x73, 0x73, 0x76, 0x2f, 0x69, 0x62, 0x66, 0x74, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x62, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x65, 0x74, 0x68, 0x2f, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x61, 0x74, 0x74, 0x65, 0x73, 0x74, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x65, 0x74, 0x68, 0x2f, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x62, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x5f, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8c, 0x04, 0x0a, 0x0a, 0x49, 0x6e, 0x70,
	0x75, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x53, 0x0a, 0x10, 0x61, 0x74, 0x74, 0x65, 0x73,
	0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x26, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68,
	0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x41, 0x74, 0x74, 0x65, 0x73, 0x74,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x44, 0x61, 0x74, 0x61, 0x48, 0x00, 0x52, 0x0f, 0x61, 0x74, 0x74,
	0x65, 0x73, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x44, 0x61, 0x74, 0x61, 0x12, 0x60, 0x0a, 0x10,
	0x61, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x33, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75,
	0x6d, 0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x41,
	0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x41, 0x74, 0x74, 0x65, 0x73, 0x74, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x41, 0x6e, 0x64, 0x50, 0x72, 0x6f, 0x6f, 0x66, 0x48, 0x00, 0x52, 0x0f, 0x61,
	0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x44, 0x61, 0x74, 0x61, 0x12, 0x47,
	0x0a, 0x0c, 0x62, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e,
	0x65, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x42, 0x65, 0x61,
	0x63, 0x6f, 0x6e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x00, 0x52, 0x0b, 0x62, 0x65, 0x61, 0x63,
	0x6f, 0x6e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x46, 0x0a, 0x0b, 0x61, 0x74, 0x74, 0x65, 0x73,
	0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x65,
	0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x2e, 0x41, 0x74, 0x74, 0x65, 0x73, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x48, 0x01, 0x52, 0x0b, 0x61, 0x74, 0x74, 0x65, 0x73, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x5d, 0x0a, 0x0b, 0x61, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x39, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e,
	0x65, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x69, 0x67,
	0x6e, 0x65, 0x64, 0x41, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x41, 0x74, 0x74, 0x65,
	0x73, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x41, 0x6e, 0x64, 0x50, 0x72, 0x6f, 0x6f, 0x66, 0x48,
	0x01, 0x52, 0x0b, 0x61, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x40,
	0x0a, 0x05, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x28, 0x2e,
	0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x42, 0x65, 0x61, 0x63,
	0x6f, 0x6e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x01, 0x52, 0x05, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x42, 0x06, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x42, 0x0d, 0x0a, 0x0b, 0x73, 0x69, 0x67, 0x6e,
	0x65, 0x64, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x42, 0x23, 0x5a, 0x21, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x6c, 0x6f, 0x78, 0x61, 0x70, 0x70, 0x2f, 0x73, 0x73,
	0x76, 0x2f, 0x69, 0x62, 0x66, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_rawDescOnce sync.Once
	file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_rawDescData = file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_rawDesc
)

func file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_rawDescGZIP() []byte {
	file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_rawDescOnce.Do(func() {
		file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_rawDescData)
	})
	return file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_rawDescData
}

var file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_goTypes = []interface{}{
	(*InputValue)(nil),                                  // 0: proto.InputValue
	(*v1alpha1.AttestationData)(nil),                    // 1: ethereum.eth.v1alpha1.AttestationData
	(*v1alpha1.AggregateAttestationAndProof)(nil),       // 2: ethereum.eth.v1alpha1.AggregateAttestationAndProof
	(*v1alpha1.BeaconBlock)(nil),                        // 3: ethereum.eth.v1alpha1.BeaconBlock
	(*v1alpha1.Attestation)(nil),                        // 4: ethereum.eth.v1alpha1.Attestation
	(*v1alpha1.SignedAggregateAttestationAndProof)(nil), // 5: ethereum.eth.v1alpha1.SignedAggregateAttestationAndProof
	(*v1alpha1.SignedBeaconBlock)(nil),                  // 6: ethereum.eth.v1alpha1.SignedBeaconBlock
}
var file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_depIdxs = []int32{
	1, // 0: proto.InputValue.attestation_data:type_name -> ethereum.eth.v1alpha1.AttestationData
	2, // 1: proto.InputValue.aggregation_data:type_name -> ethereum.eth.v1alpha1.AggregateAttestationAndProof
	3, // 2: proto.InputValue.beacon_block:type_name -> ethereum.eth.v1alpha1.BeaconBlock
	4, // 3: proto.InputValue.attestation:type_name -> ethereum.eth.v1alpha1.Attestation
	5, // 4: proto.InputValue.aggregation:type_name -> ethereum.eth.v1alpha1.SignedAggregateAttestationAndProof
	6, // 5: proto.InputValue.block:type_name -> ethereum.eth.v1alpha1.SignedBeaconBlock
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_init() }
func file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_init() {
	if File_github_com_bloxapp_ssv_ibft_proto_beacon_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InputValue); i {
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
	file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*InputValue_AttestationData)(nil),
		(*InputValue_AggregationData)(nil),
		(*InputValue_BeaconBlock)(nil),
		(*InputValue_Attestation)(nil),
		(*InputValue_Aggregation)(nil),
		(*InputValue_Block)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_goTypes,
		DependencyIndexes: file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_depIdxs,
		MessageInfos:      file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_msgTypes,
	}.Build()
	File_github_com_bloxapp_ssv_ibft_proto_beacon_proto = out.File
	file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_rawDesc = nil
	file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_goTypes = nil
	file_github_com_bloxapp_ssv_ibft_proto_beacon_proto_depIdxs = nil
}
