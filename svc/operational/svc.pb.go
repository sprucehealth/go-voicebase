// Code generated by protoc-gen-gogo.
// source: svc.proto
// DO NOT EDIT!

/*
	Package operational is a generated protocol buffer package.

	It is generated from these files:
		svc.proto

	It has these top-level messages:
		BlockAccountRequest
		NewOrgCreatedEvent
*/
package operational

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"

import strings "strings"
import github_com_gogo_protobuf_proto "github.com/gogo/protobuf/proto"
import sort "sort"
import strconv "strconv"
import reflect "reflect"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type BlockAccountRequest struct {
	AccountID string `protobuf:"bytes,1,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty"`
}

func (m *BlockAccountRequest) Reset()                    { *m = BlockAccountRequest{} }
func (*BlockAccountRequest) ProtoMessage()               {}
func (*BlockAccountRequest) Descriptor() ([]byte, []int) { return fileDescriptorSvc, []int{0} }

type NewOrgCreatedEvent struct {
	InitialProviderEntityID string `protobuf:"bytes,1,opt,name=initial_provider_entity_id,json=initialProviderEntityId,proto3" json:"initial_provider_entity_id,omitempty"`
	OrgSupportThreadID      string `protobuf:"bytes,2,opt,name=org_support_thread_id,json=orgSupportThreadId,proto3" json:"org_support_thread_id,omitempty"`
	SpruceSupportThreadID   string `protobuf:"bytes,3,opt,name=spruce_support_thread_id,json=spruceSupportThreadId,proto3" json:"spruce_support_thread_id,omitempty"`
	OrgCreated              int64  `protobuf:"varint,4,opt,name=org_created,json=orgCreated,proto3" json:"org_created,omitempty"`
}

func (m *NewOrgCreatedEvent) Reset()                    { *m = NewOrgCreatedEvent{} }
func (*NewOrgCreatedEvent) ProtoMessage()               {}
func (*NewOrgCreatedEvent) Descriptor() ([]byte, []int) { return fileDescriptorSvc, []int{1} }

func init() {
	proto.RegisterType((*BlockAccountRequest)(nil), "operational.BlockAccountRequest")
	proto.RegisterType((*NewOrgCreatedEvent)(nil), "operational.NewOrgCreatedEvent")
}
func (this *BlockAccountRequest) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*BlockAccountRequest)
	if !ok {
		that2, ok := that.(BlockAccountRequest)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if this.AccountID != that1.AccountID {
		return false
	}
	return true
}
func (this *NewOrgCreatedEvent) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*NewOrgCreatedEvent)
	if !ok {
		that2, ok := that.(NewOrgCreatedEvent)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if this.InitialProviderEntityID != that1.InitialProviderEntityID {
		return false
	}
	if this.OrgSupportThreadID != that1.OrgSupportThreadID {
		return false
	}
	if this.SpruceSupportThreadID != that1.SpruceSupportThreadID {
		return false
	}
	if this.OrgCreated != that1.OrgCreated {
		return false
	}
	return true
}
func (this *BlockAccountRequest) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 5)
	s = append(s, "&operational.BlockAccountRequest{")
	s = append(s, "AccountID: "+fmt.Sprintf("%#v", this.AccountID)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *NewOrgCreatedEvent) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 8)
	s = append(s, "&operational.NewOrgCreatedEvent{")
	s = append(s, "InitialProviderEntityID: "+fmt.Sprintf("%#v", this.InitialProviderEntityID)+",\n")
	s = append(s, "OrgSupportThreadID: "+fmt.Sprintf("%#v", this.OrgSupportThreadID)+",\n")
	s = append(s, "SpruceSupportThreadID: "+fmt.Sprintf("%#v", this.SpruceSupportThreadID)+",\n")
	s = append(s, "OrgCreated: "+fmt.Sprintf("%#v", this.OrgCreated)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringSvc(v interface{}, typ string) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)
}
func extensionToGoStringSvc(m github_com_gogo_protobuf_proto.Message) string {
	e := github_com_gogo_protobuf_proto.GetUnsafeExtensionsMap(m)
	if e == nil {
		return "nil"
	}
	s := "proto.NewUnsafeXXX_InternalExtensions(map[int32]proto.Extension{"
	keys := make([]int, 0, len(e))
	for k := range e {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	ss := []string{}
	for _, k := range keys {
		ss = append(ss, strconv.Itoa(k)+": "+e[int32(k)].GoString())
	}
	s += strings.Join(ss, ",") + "})"
	return s
}
func (m *BlockAccountRequest) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *BlockAccountRequest) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.AccountID) > 0 {
		data[i] = 0xa
		i++
		i = encodeVarintSvc(data, i, uint64(len(m.AccountID)))
		i += copy(data[i:], m.AccountID)
	}
	return i, nil
}

func (m *NewOrgCreatedEvent) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *NewOrgCreatedEvent) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.InitialProviderEntityID) > 0 {
		data[i] = 0xa
		i++
		i = encodeVarintSvc(data, i, uint64(len(m.InitialProviderEntityID)))
		i += copy(data[i:], m.InitialProviderEntityID)
	}
	if len(m.OrgSupportThreadID) > 0 {
		data[i] = 0x12
		i++
		i = encodeVarintSvc(data, i, uint64(len(m.OrgSupportThreadID)))
		i += copy(data[i:], m.OrgSupportThreadID)
	}
	if len(m.SpruceSupportThreadID) > 0 {
		data[i] = 0x1a
		i++
		i = encodeVarintSvc(data, i, uint64(len(m.SpruceSupportThreadID)))
		i += copy(data[i:], m.SpruceSupportThreadID)
	}
	if m.OrgCreated != 0 {
		data[i] = 0x20
		i++
		i = encodeVarintSvc(data, i, uint64(m.OrgCreated))
	}
	return i, nil
}

func encodeFixed64Svc(data []byte, offset int, v uint64) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	data[offset+4] = uint8(v >> 32)
	data[offset+5] = uint8(v >> 40)
	data[offset+6] = uint8(v >> 48)
	data[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Svc(data []byte, offset int, v uint32) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintSvc(data []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		data[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	data[offset] = uint8(v)
	return offset + 1
}
func (m *BlockAccountRequest) Size() (n int) {
	var l int
	_ = l
	l = len(m.AccountID)
	if l > 0 {
		n += 1 + l + sovSvc(uint64(l))
	}
	return n
}

func (m *NewOrgCreatedEvent) Size() (n int) {
	var l int
	_ = l
	l = len(m.InitialProviderEntityID)
	if l > 0 {
		n += 1 + l + sovSvc(uint64(l))
	}
	l = len(m.OrgSupportThreadID)
	if l > 0 {
		n += 1 + l + sovSvc(uint64(l))
	}
	l = len(m.SpruceSupportThreadID)
	if l > 0 {
		n += 1 + l + sovSvc(uint64(l))
	}
	if m.OrgCreated != 0 {
		n += 1 + sovSvc(uint64(m.OrgCreated))
	}
	return n
}

func sovSvc(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozSvc(x uint64) (n int) {
	return sovSvc(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *BlockAccountRequest) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&BlockAccountRequest{`,
		`AccountID:` + fmt.Sprintf("%v", this.AccountID) + `,`,
		`}`,
	}, "")
	return s
}
func (this *NewOrgCreatedEvent) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&NewOrgCreatedEvent{`,
		`InitialProviderEntityID:` + fmt.Sprintf("%v", this.InitialProviderEntityID) + `,`,
		`OrgSupportThreadID:` + fmt.Sprintf("%v", this.OrgSupportThreadID) + `,`,
		`SpruceSupportThreadID:` + fmt.Sprintf("%v", this.SpruceSupportThreadID) + `,`,
		`OrgCreated:` + fmt.Sprintf("%v", this.OrgCreated) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringSvc(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *BlockAccountRequest) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSvc
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: BlockAccountRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BlockAccountRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AccountID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSvc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSvc
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AccountID = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipSvc(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSvc
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
func (m *NewOrgCreatedEvent) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSvc
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: NewOrgCreatedEvent: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: NewOrgCreatedEvent: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InitialProviderEntityID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSvc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSvc
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.InitialProviderEntityID = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OrgSupportThreadID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSvc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSvc
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OrgSupportThreadID = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SpruceSupportThreadID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSvc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSvc
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SpruceSupportThreadID = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field OrgCreated", wireType)
			}
			m.OrgCreated = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSvc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				m.OrgCreated |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipSvc(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSvc
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
func skipSvc(data []byte) (n int, err error) {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowSvc
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
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
					return 0, ErrIntOverflowSvc
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if data[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowSvc
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthSvc
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowSvc
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := data[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipSvc(data[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthSvc = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowSvc   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("svc.proto", fileDescriptorSvc) }

var fileDescriptorSvc = []byte{
	// 350 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x6c, 0x91, 0xbf, 0x4e, 0xeb, 0x30,
	0x14, 0xc6, 0xeb, 0xf6, 0xea, 0x4a, 0x71, 0x75, 0x17, 0x5f, 0x95, 0x96, 0x22, 0x39, 0x55, 0xa7,
	0x0e, 0xa5, 0x1d, 0x78, 0x02, 0xfa, 0x67, 0xc8, 0x42, 0x51, 0xca, 0xc0, 0x16, 0xa5, 0xb6, 0x49,
	0x2d, 0x4a, 0x4e, 0x70, 0x9c, 0x22, 0x36, 0x1e, 0x81, 0xc7, 0x60, 0xe2, 0x39, 0x18, 0x3b, 0x32,
	0x45, 0xd4, 0x2c, 0x8c, 0x7d, 0x04, 0x54, 0xa7, 0x02, 0xa9, 0xea, 0xe6, 0xf3, 0x7d, 0x3f, 0xff,
	0x64, 0x1d, 0x63, 0x27, 0x5d, 0xb2, 0x5e, 0xa2, 0x40, 0x03, 0xa9, 0x42, 0x22, 0x54, 0xa8, 0x25,
	0xc4, 0xe1, 0xa2, 0x79, 0x1a, 0x49, 0x3d, 0xcf, 0x66, 0x3d, 0x06, 0x77, 0xfd, 0x08, 0x22, 0xe8,
	0x5b, 0x66, 0x96, 0xdd, 0xd8, 0xc9, 0x0e, 0xf6, 0x54, 0xdc, 0x6d, 0x0f, 0xf1, 0xff, 0xc1, 0x02,
	0xd8, 0xed, 0x39, 0x63, 0x90, 0xc5, 0xda, 0x17, 0xf7, 0x99, 0x48, 0x35, 0xe9, 0x62, 0x1c, 0x16,
	0x49, 0x20, 0x79, 0x03, 0xb5, 0x50, 0xc7, 0x19, 0xfc, 0x33, 0xb9, 0xeb, 0xec, 0x38, 0x6f, 0xe4,
	0x3b, 0x3b, 0xc0, 0xe3, 0xed, 0xd7, 0x32, 0x26, 0x17, 0xe2, 0x61, 0xa2, 0xa2, 0xa1, 0x12, 0xa1,
	0x16, 0x7c, 0xbc, 0x14, 0xb1, 0x26, 0xd7, 0xb8, 0x29, 0x63, 0xa9, 0x65, 0xb8, 0x08, 0x12, 0x05,
	0x4b, 0xc9, 0x85, 0x0a, 0x44, 0xac, 0xa5, 0x7e, 0xfc, 0x95, 0x9e, 0x98, 0xdc, 0xad, 0x7b, 0x05,
	0x75, 0xb9, 0x83, 0xc6, 0x96, 0xf1, 0x46, 0x7e, 0x5d, 0x1e, 0x2c, 0x38, 0xf1, 0x70, 0x0d, 0x54,
	0x14, 0xa4, 0x59, 0x92, 0x80, 0xd2, 0x81, 0x9e, 0x2b, 0x11, 0xf2, 0xad, 0xb4, 0x6c, 0xa5, 0x47,
	0x26, 0x77, 0xc9, 0x44, 0x45, 0xd3, 0xa2, 0xbf, 0xb2, 0xb5, 0x37, 0xf2, 0x09, 0xec, 0x67, 0x9c,
	0xf8, 0xb8, 0x91, 0x26, 0x2a, 0x63, 0xe2, 0x80, 0xad, 0x62, 0x6d, 0xc7, 0x26, 0x77, 0x6b, 0x53,
	0xcb, 0xec, 0x0b, 0x6b, 0xe9, 0x81, 0x98, 0x13, 0x17, 0x57, 0xb7, 0xcf, 0x63, 0xc5, 0x32, 0x1a,
	0x7f, 0x5a, 0xa8, 0x53, 0xf1, 0x31, 0xfc, 0xac, 0x67, 0xd0, 0x5d, 0xad, 0x29, 0x7a, 0x5f, 0xd3,
	0xd2, 0x66, 0x4d, 0xd1, 0x93, 0xa1, 0xe8, 0xc5, 0x50, 0xf4, 0x66, 0x28, 0x5a, 0x19, 0x8a, 0x3e,
	0x0c, 0x45, 0x5f, 0x86, 0x96, 0x36, 0x86, 0xa2, 0xe7, 0x4f, 0x5a, 0x9a, 0xfd, 0xb5, 0x5f, 0x75,
	0xf6, 0x1d, 0x00, 0x00, 0xff, 0xff, 0x86, 0xbd, 0x7f, 0x33, 0xf3, 0x01, 0x00, 0x00,
}
