// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: strangelove_ventures/poa/v1/params.proto

package poa

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	_ "google.golang.org/protobuf/types/known/durationpb"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Params defines the parameters for the module.
type Params struct {
	// Array of addresses that are allowed to control the chains validators power.
	Admins []string `protobuf:"bytes,1,rep,name=admins,proto3" json:"admins,omitempty"`
	// Array of validator base wallet addresses which are whitelisted to be able to MsgCreateValidator.
	ValidatorWhitelist []*PendingValidator `protobuf:"bytes,2,rep,name=validator_whitelist,json=validatorWhitelist,proto3" json:"validator_whitelist,omitempty"`
}

func (m *Params) Reset()      { *m = Params{} }
func (*Params) ProtoMessage() {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_b1333a19bedb70c3, []int{0}
}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return m.Size()
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func (m *Params) GetAdmins() []string {
	if m != nil {
		return m.Admins
	}
	return nil
}

func (m *Params) GetValidatorWhitelist() []*PendingValidator {
	if m != nil {
		return m.ValidatorWhitelist
	}
	return nil
}

type PendingValidator struct {
	// cosmos-sdk accountaddress
	Address []byte `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Info    string `protobuf:"bytes,2,opt,name=info,proto3" json:"info,omitempty"`
}

func (m *PendingValidator) Reset()         { *m = PendingValidator{} }
func (m *PendingValidator) String() string { return proto.CompactTextString(m) }
func (*PendingValidator) ProtoMessage()    {}
func (*PendingValidator) Descriptor() ([]byte, []int) {
	return fileDescriptor_b1333a19bedb70c3, []int{1}
}
func (m *PendingValidator) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PendingValidator) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PendingValidator.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PendingValidator) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PendingValidator.Merge(m, src)
}
func (m *PendingValidator) XXX_Size() int {
	return m.Size()
}
func (m *PendingValidator) XXX_DiscardUnknown() {
	xxx_messageInfo_PendingValidator.DiscardUnknown(m)
}

var xxx_messageInfo_PendingValidator proto.InternalMessageInfo

func (m *PendingValidator) GetAddress() []byte {
	if m != nil {
		return m.Address
	}
	return nil
}

func (m *PendingValidator) GetInfo() string {
	if m != nil {
		return m.Info
	}
	return ""
}

func init() {
	proto.RegisterType((*Params)(nil), "strangelove_ventures.poa.v1.Params")
	proto.RegisterType((*PendingValidator)(nil), "strangelove_ventures.poa.v1.PendingValidator")
}

func init() {
	proto.RegisterFile("strangelove_ventures/poa/v1/params.proto", fileDescriptor_b1333a19bedb70c3)
}

var fileDescriptor_b1333a19bedb70c3 = []byte{
	// 331 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x91, 0xc1, 0x4a, 0xc3, 0x40,
	0x10, 0x86, 0xb3, 0xad, 0x54, 0xba, 0x7a, 0xd0, 0x28, 0x1a, 0x2b, 0x6c, 0x4b, 0xbd, 0x04, 0xa1,
	0x59, 0xaa, 0x37, 0x41, 0x10, 0x9f, 0xa0, 0xe4, 0xa0, 0xe0, 0xc1, 0xb2, 0x6d, 0xb6, 0xdb, 0x85,
	0x64, 0x27, 0xec, 0x6e, 0xe2, 0x3b, 0x78, 0xf2, 0xa8, 0xb7, 0x3e, 0x82, 0x8f, 0xe1, 0xb1, 0x47,
	0x8f, 0xd2, 0x1e, 0xf4, 0x31, 0xa4, 0x49, 0x2b, 0x45, 0xc4, 0xcb, 0x32, 0xff, 0xce, 0xcc, 0xc7,
	0xcc, 0x3f, 0xd8, 0x37, 0x56, 0x33, 0x25, 0x78, 0x0c, 0x39, 0xef, 0xe7, 0x5c, 0xd9, 0x4c, 0x73,
	0x43, 0x53, 0x60, 0x34, 0xef, 0xd2, 0x94, 0x69, 0x96, 0x98, 0x20, 0xd5, 0x60, 0xc1, 0x3d, 0xfe,
	0xab, 0x32, 0x48, 0x81, 0x05, 0x79, 0xb7, 0xb1, 0x2f, 0x40, 0x40, 0x51, 0x47, 0x17, 0x51, 0xd9,
	0xd2, 0xd8, 0x65, 0x89, 0x54, 0x40, 0x8b, 0x77, 0xf9, 0x45, 0x04, 0x80, 0x88, 0x39, 0x2d, 0xd4,
	0x20, 0x1b, 0xd1, 0x28, 0xd3, 0xcc, 0x4a, 0x50, 0xcb, 0xfc, 0xd1, 0x10, 0x4c, 0x02, 0xa6, 0x5f,
	0xb2, 0x4a, 0x51, 0xa6, 0xda, 0x2f, 0x08, 0xd7, 0x7a, 0xc5, 0x44, 0xee, 0x01, 0xae, 0xb1, 0x28,
	0x91, 0xca, 0x78, 0xa8, 0x55, 0xf5, 0xeb, 0xe1, 0x52, 0xb9, 0xf7, 0x78, 0x2f, 0x67, 0xb1, 0x8c,
	0x98, 0x05, 0xdd, 0x7f, 0x18, 0x4b, 0xcb, 0x63, 0x69, 0xac, 0x57, 0x69, 0x55, 0xfd, 0xad, 0xb3,
	0x4e, 0xf0, 0xcf, 0x06, 0x41, 0x8f, 0xab, 0x48, 0x2a, 0x71, 0xb3, 0x6a, 0x0f, 0xdd, 0x1f, 0xd2,
	0xed, 0x0a, 0x74, 0x71, 0xf8, 0x3c, 0x69, 0x3a, 0x5f, 0x93, 0x26, 0x7a, 0xfc, 0x7c, 0x3d, 0xc5,
	0x0b, 0x97, 0x4a, 0x8b, 0xda, 0x57, 0x78, 0xe7, 0x37, 0xc0, 0xf5, 0xf0, 0x26, 0x8b, 0x22, 0xcd,
	0xcd, 0x62, 0x4a, 0xe4, 0x6f, 0x87, 0x2b, 0xe9, 0xba, 0x78, 0x43, 0xaa, 0x11, 0x78, 0x95, 0x16,
	0xf2, 0xeb, 0x61, 0x11, 0x5f, 0x5f, 0xbe, 0xcd, 0x08, 0x9a, 0xce, 0x08, 0xfa, 0x98, 0x11, 0xf4,
	0x34, 0x27, 0xce, 0x74, 0x4e, 0x9c, 0xf7, 0x39, 0x71, 0xee, 0x4e, 0x84, 0xb4, 0xe3, 0x6c, 0x10,
	0x0c, 0x21, 0xa1, 0x6b, 0x1b, 0x74, 0xd6, 0xaf, 0x35, 0xa8, 0x15, 0x1e, 0x9d, 0x7f, 0x07, 0x00,
	0x00, 0xff, 0xff, 0xd4, 0x7c, 0x5f, 0x08, 0xd0, 0x01, 0x00, 0x00,
}

func (this *Params) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Params)
	if !ok {
		that2, ok := that.(Params)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if len(this.Admins) != len(that1.Admins) {
		return false
	}
	for i := range this.Admins {
		if this.Admins[i] != that1.Admins[i] {
			return false
		}
	}
	if len(this.ValidatorWhitelist) != len(that1.ValidatorWhitelist) {
		return false
	}
	for i := range this.ValidatorWhitelist {
		if !this.ValidatorWhitelist[i].Equal(that1.ValidatorWhitelist[i]) {
			return false
		}
	}
	return true
}
func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ValidatorWhitelist) > 0 {
		for iNdEx := len(m.ValidatorWhitelist) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ValidatorWhitelist[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintParams(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.Admins) > 0 {
		for iNdEx := len(m.Admins) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.Admins[iNdEx])
			copy(dAtA[i:], m.Admins[iNdEx])
			i = encodeVarintParams(dAtA, i, uint64(len(m.Admins[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *PendingValidator) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PendingValidator) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PendingValidator) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Info) > 0 {
		i -= len(m.Info)
		copy(dAtA[i:], m.Info)
		i = encodeVarintParams(dAtA, i, uint64(len(m.Info)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintParams(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintParams(dAtA []byte, offset int, v uint64) int {
	offset -= sovParams(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Admins) > 0 {
		for _, s := range m.Admins {
			l = len(s)
			n += 1 + l + sovParams(uint64(l))
		}
	}
	if len(m.ValidatorWhitelist) > 0 {
		for _, e := range m.ValidatorWhitelist {
			l = e.Size()
			n += 1 + l + sovParams(uint64(l))
		}
	}
	return n
}

func (m *PendingValidator) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovParams(uint64(l))
	}
	l = len(m.Info)
	if l > 0 {
		n += 1 + l + sovParams(uint64(l))
	}
	return n
}

func sovParams(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozParams(x uint64) (n int) {
	return sovParams(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowParams
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Admins", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Admins = append(m.Admins, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorWhitelist", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ValidatorWhitelist = append(m.ValidatorWhitelist, &PendingValidator{})
			if err := m.ValidatorWhitelist[len(m.ValidatorWhitelist)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipParams(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthParams
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *PendingValidator) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowParams
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: PendingValidator: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PendingValidator: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = append(m.Address[:0], dAtA[iNdEx:postIndex]...)
			if m.Address == nil {
				m.Address = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Info", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Info = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipParams(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthParams
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipParams(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowParams
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowParams
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowParams
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthParams
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupParams
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthParams
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthParams        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowParams          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupParams = fmt.Errorf("proto: unexpected end of group")
)
