// Code generated by protoc-gen-gogo.
// source: events.proto
// DO NOT EDIT!

/*
	Package invite is a generated protocol buffer package.

	It is generated from these files:
		events.proto

	It has these top-level messages:
		Event
		InvitedColleagues
		InvitedPatients
*/
package invite

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"

import strconv "strconv"

import strings "strings"
import github_com_gogo_protobuf_proto "github.com/gogo/protobuf/proto"
import sort "sort"
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

type Event_Type int32

const (
	Event_INVALID            Event_Type = 0
	Event_INVITED_COLLEAGUES Event_Type = 1
	Event_INVITED_PATIENTS   Event_Type = 2
)

var Event_Type_name = map[int32]string{
	0: "INVALID",
	1: "INVITED_COLLEAGUES",
	2: "INVITED_PATIENTS",
}
var Event_Type_value = map[string]int32{
	"INVALID":            0,
	"INVITED_COLLEAGUES": 1,
	"INVITED_PATIENTS":   2,
}

func (Event_Type) EnumDescriptor() ([]byte, []int) { return fileDescriptorEvents, []int{0, 0} }

type Event struct {
	Type Event_Type `protobuf:"varint,1,opt,name=type,proto3,enum=invite.Event_Type" json:"type,omitempty"`
	// Types that are valid to be assigned to Details:
	//	*Event_InvitedColleagues
	//	*Event_InvitedPatients
	Details isEvent_Details `protobuf_oneof:"details"`
}

func (m *Event) Reset()                    { *m = Event{} }
func (*Event) ProtoMessage()               {}
func (*Event) Descriptor() ([]byte, []int) { return fileDescriptorEvents, []int{0} }

type isEvent_Details interface {
	isEvent_Details()
	Equal(interface{}) bool
	MarshalTo([]byte) (int, error)
	Size() int
}

type Event_InvitedColleagues struct {
	InvitedColleagues *InvitedColleagues `protobuf:"bytes,10,opt,name=invited_colleagues,json=invitedColleagues,oneof"`
}
type Event_InvitedPatients struct {
	InvitedPatients *InvitedPatients `protobuf:"bytes,11,opt,name=invited_patients,json=invitedPatients,oneof"`
}

func (*Event_InvitedColleagues) isEvent_Details() {}
func (*Event_InvitedPatients) isEvent_Details()   {}

func (m *Event) GetDetails() isEvent_Details {
	if m != nil {
		return m.Details
	}
	return nil
}

func (m *Event) GetInvitedColleagues() *InvitedColleagues {
	if x, ok := m.GetDetails().(*Event_InvitedColleagues); ok {
		return x.InvitedColleagues
	}
	return nil
}

func (m *Event) GetInvitedPatients() *InvitedPatients {
	if x, ok := m.GetDetails().(*Event_InvitedPatients); ok {
		return x.InvitedPatients
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*Event) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _Event_OneofMarshaler, _Event_OneofUnmarshaler, _Event_OneofSizer, []interface{}{
		(*Event_InvitedColleagues)(nil),
		(*Event_InvitedPatients)(nil),
	}
}

func _Event_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*Event)
	// details
	switch x := m.Details.(type) {
	case *Event_InvitedColleagues:
		_ = b.EncodeVarint(10<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.InvitedColleagues); err != nil {
			return err
		}
	case *Event_InvitedPatients:
		_ = b.EncodeVarint(11<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.InvitedPatients); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("Event.Details has unexpected type %T", x)
	}
	return nil
}

func _Event_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*Event)
	switch tag {
	case 10: // details.invited_colleagues
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(InvitedColleagues)
		err := b.DecodeMessage(msg)
		m.Details = &Event_InvitedColleagues{msg}
		return true, err
	case 11: // details.invited_patients
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(InvitedPatients)
		err := b.DecodeMessage(msg)
		m.Details = &Event_InvitedPatients{msg}
		return true, err
	default:
		return false, nil
	}
}

func _Event_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*Event)
	// details
	switch x := m.Details.(type) {
	case *Event_InvitedColleagues:
		s := proto.Size(x.InvitedColleagues)
		n += proto.SizeVarint(10<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Event_InvitedPatients:
		s := proto.Size(x.InvitedPatients)
		n += proto.SizeVarint(11<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

type InvitedColleagues struct {
	OrganizationEntityID string `protobuf:"bytes,1,opt,name=organization_entity_id,json=organizationEntityId,proto3" json:"organization_entity_id,omitempty"`
	InviterEntityID      string `protobuf:"bytes,2,opt,name=inviter_entity_id,json=inviterEntityId,proto3" json:"inviter_entity_id,omitempty"`
}

func (m *InvitedColleagues) Reset()                    { *m = InvitedColleagues{} }
func (*InvitedColleagues) ProtoMessage()               {}
func (*InvitedColleagues) Descriptor() ([]byte, []int) { return fileDescriptorEvents, []int{1} }

type InvitedPatients struct {
	OrganizationEntityID string `protobuf:"bytes,1,opt,name=organization_entity_id,json=organizationEntityId,proto3" json:"organization_entity_id,omitempty"`
	InviterEntityID      string `protobuf:"bytes,2,opt,name=inviter_entity_id,json=inviterEntityId,proto3" json:"inviter_entity_id,omitempty"`
}

func (m *InvitedPatients) Reset()                    { *m = InvitedPatients{} }
func (*InvitedPatients) ProtoMessage()               {}
func (*InvitedPatients) Descriptor() ([]byte, []int) { return fileDescriptorEvents, []int{2} }

func init() {
	proto.RegisterType((*Event)(nil), "invite.Event")
	proto.RegisterType((*InvitedColleagues)(nil), "invite.InvitedColleagues")
	proto.RegisterType((*InvitedPatients)(nil), "invite.InvitedPatients")
	proto.RegisterEnum("invite.Event_Type", Event_Type_name, Event_Type_value)
}
func (x Event_Type) String() string {
	s, ok := Event_Type_name[int32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}
func (this *Event) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*Event)
	if !ok {
		that2, ok := that.(Event)
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
	if this.Type != that1.Type {
		return false
	}
	if that1.Details == nil {
		if this.Details != nil {
			return false
		}
	} else if this.Details == nil {
		return false
	} else if !this.Details.Equal(that1.Details) {
		return false
	}
	return true
}
func (this *Event_InvitedColleagues) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*Event_InvitedColleagues)
	if !ok {
		that2, ok := that.(Event_InvitedColleagues)
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
	if !this.InvitedColleagues.Equal(that1.InvitedColleagues) {
		return false
	}
	return true
}
func (this *Event_InvitedPatients) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*Event_InvitedPatients)
	if !ok {
		that2, ok := that.(Event_InvitedPatients)
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
	if !this.InvitedPatients.Equal(that1.InvitedPatients) {
		return false
	}
	return true
}
func (this *InvitedColleagues) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*InvitedColleagues)
	if !ok {
		that2, ok := that.(InvitedColleagues)
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
	if this.OrganizationEntityID != that1.OrganizationEntityID {
		return false
	}
	if this.InviterEntityID != that1.InviterEntityID {
		return false
	}
	return true
}
func (this *InvitedPatients) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*InvitedPatients)
	if !ok {
		that2, ok := that.(InvitedPatients)
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
	if this.OrganizationEntityID != that1.OrganizationEntityID {
		return false
	}
	if this.InviterEntityID != that1.InviterEntityID {
		return false
	}
	return true
}
func (this *Event) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 7)
	s = append(s, "&invite.Event{")
	s = append(s, "Type: "+fmt.Sprintf("%#v", this.Type)+",\n")
	if this.Details != nil {
		s = append(s, "Details: "+fmt.Sprintf("%#v", this.Details)+",\n")
	}
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *Event_InvitedColleagues) GoString() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&invite.Event_InvitedColleagues{` +
		`InvitedColleagues:` + fmt.Sprintf("%#v", this.InvitedColleagues) + `}`}, ", ")
	return s
}
func (this *Event_InvitedPatients) GoString() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&invite.Event_InvitedPatients{` +
		`InvitedPatients:` + fmt.Sprintf("%#v", this.InvitedPatients) + `}`}, ", ")
	return s
}
func (this *InvitedColleagues) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 6)
	s = append(s, "&invite.InvitedColleagues{")
	s = append(s, "OrganizationEntityID: "+fmt.Sprintf("%#v", this.OrganizationEntityID)+",\n")
	s = append(s, "InviterEntityID: "+fmt.Sprintf("%#v", this.InviterEntityID)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *InvitedPatients) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 6)
	s = append(s, "&invite.InvitedPatients{")
	s = append(s, "OrganizationEntityID: "+fmt.Sprintf("%#v", this.OrganizationEntityID)+",\n")
	s = append(s, "InviterEntityID: "+fmt.Sprintf("%#v", this.InviterEntityID)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringEvents(v interface{}, typ string) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)
}
func extensionToGoStringEvents(m github_com_gogo_protobuf_proto.Message) string {
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
func (m *Event) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *Event) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Type != 0 {
		data[i] = 0x8
		i++
		i = encodeVarintEvents(data, i, uint64(m.Type))
	}
	if m.Details != nil {
		nn1, err := m.Details.MarshalTo(data[i:])
		if err != nil {
			return 0, err
		}
		i += nn1
	}
	return i, nil
}

func (m *Event_InvitedColleagues) MarshalTo(data []byte) (int, error) {
	i := 0
	if m.InvitedColleagues != nil {
		data[i] = 0x52
		i++
		i = encodeVarintEvents(data, i, uint64(m.InvitedColleagues.Size()))
		n2, err := m.InvitedColleagues.MarshalTo(data[i:])
		if err != nil {
			return 0, err
		}
		i += n2
	}
	return i, nil
}
func (m *Event_InvitedPatients) MarshalTo(data []byte) (int, error) {
	i := 0
	if m.InvitedPatients != nil {
		data[i] = 0x5a
		i++
		i = encodeVarintEvents(data, i, uint64(m.InvitedPatients.Size()))
		n3, err := m.InvitedPatients.MarshalTo(data[i:])
		if err != nil {
			return 0, err
		}
		i += n3
	}
	return i, nil
}
func (m *InvitedColleagues) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *InvitedColleagues) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.OrganizationEntityID) > 0 {
		data[i] = 0xa
		i++
		i = encodeVarintEvents(data, i, uint64(len(m.OrganizationEntityID)))
		i += copy(data[i:], m.OrganizationEntityID)
	}
	if len(m.InviterEntityID) > 0 {
		data[i] = 0x12
		i++
		i = encodeVarintEvents(data, i, uint64(len(m.InviterEntityID)))
		i += copy(data[i:], m.InviterEntityID)
	}
	return i, nil
}

func (m *InvitedPatients) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *InvitedPatients) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.OrganizationEntityID) > 0 {
		data[i] = 0xa
		i++
		i = encodeVarintEvents(data, i, uint64(len(m.OrganizationEntityID)))
		i += copy(data[i:], m.OrganizationEntityID)
	}
	if len(m.InviterEntityID) > 0 {
		data[i] = 0x12
		i++
		i = encodeVarintEvents(data, i, uint64(len(m.InviterEntityID)))
		i += copy(data[i:], m.InviterEntityID)
	}
	return i, nil
}

func encodeFixed64Events(data []byte, offset int, v uint64) int {
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
func encodeFixed32Events(data []byte, offset int, v uint32) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintEvents(data []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		data[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	data[offset] = uint8(v)
	return offset + 1
}
func (m *Event) Size() (n int) {
	var l int
	_ = l
	if m.Type != 0 {
		n += 1 + sovEvents(uint64(m.Type))
	}
	if m.Details != nil {
		n += m.Details.Size()
	}
	return n
}

func (m *Event_InvitedColleagues) Size() (n int) {
	var l int
	_ = l
	if m.InvitedColleagues != nil {
		l = m.InvitedColleagues.Size()
		n += 1 + l + sovEvents(uint64(l))
	}
	return n
}
func (m *Event_InvitedPatients) Size() (n int) {
	var l int
	_ = l
	if m.InvitedPatients != nil {
		l = m.InvitedPatients.Size()
		n += 1 + l + sovEvents(uint64(l))
	}
	return n
}
func (m *InvitedColleagues) Size() (n int) {
	var l int
	_ = l
	l = len(m.OrganizationEntityID)
	if l > 0 {
		n += 1 + l + sovEvents(uint64(l))
	}
	l = len(m.InviterEntityID)
	if l > 0 {
		n += 1 + l + sovEvents(uint64(l))
	}
	return n
}

func (m *InvitedPatients) Size() (n int) {
	var l int
	_ = l
	l = len(m.OrganizationEntityID)
	if l > 0 {
		n += 1 + l + sovEvents(uint64(l))
	}
	l = len(m.InviterEntityID)
	if l > 0 {
		n += 1 + l + sovEvents(uint64(l))
	}
	return n
}

func sovEvents(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozEvents(x uint64) (n int) {
	return sovEvents(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *Event) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&Event{`,
		`Type:` + fmt.Sprintf("%v", this.Type) + `,`,
		`Details:` + fmt.Sprintf("%v", this.Details) + `,`,
		`}`,
	}, "")
	return s
}
func (this *Event_InvitedColleagues) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&Event_InvitedColleagues{`,
		`InvitedColleagues:` + strings.Replace(fmt.Sprintf("%v", this.InvitedColleagues), "InvitedColleagues", "InvitedColleagues", 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *Event_InvitedPatients) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&Event_InvitedPatients{`,
		`InvitedPatients:` + strings.Replace(fmt.Sprintf("%v", this.InvitedPatients), "InvitedPatients", "InvitedPatients", 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *InvitedColleagues) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&InvitedColleagues{`,
		`OrganizationEntityID:` + fmt.Sprintf("%v", this.OrganizationEntityID) + `,`,
		`InviterEntityID:` + fmt.Sprintf("%v", this.InviterEntityID) + `,`,
		`}`,
	}, "")
	return s
}
func (this *InvitedPatients) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&InvitedPatients{`,
		`OrganizationEntityID:` + fmt.Sprintf("%v", this.OrganizationEntityID) + `,`,
		`InviterEntityID:` + fmt.Sprintf("%v", this.InviterEntityID) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringEvents(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *Event) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvents
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
			return fmt.Errorf("proto: Event: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Event: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				m.Type |= (Event_Type(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InvitedColleagues", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &InvitedColleagues{}
			if err := v.Unmarshal(data[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Details = &Event_InvitedColleagues{v}
			iNdEx = postIndex
		case 11:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InvitedPatients", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &InvitedPatients{}
			if err := v.Unmarshal(data[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Details = &Event_InvitedPatients{v}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipEvents(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthEvents
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
func (m *InvitedColleagues) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvents
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
			return fmt.Errorf("proto: InvitedColleagues: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: InvitedColleagues: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OrganizationEntityID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
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
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OrganizationEntityID = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InviterEntityID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
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
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.InviterEntityID = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipEvents(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthEvents
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
func (m *InvitedPatients) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvents
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
			return fmt.Errorf("proto: InvitedPatients: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: InvitedPatients: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OrganizationEntityID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
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
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OrganizationEntityID = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InviterEntityID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
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
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.InviterEntityID = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipEvents(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthEvents
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
func skipEvents(data []byte) (n int, err error) {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowEvents
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
					return 0, ErrIntOverflowEvents
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
					return 0, ErrIntOverflowEvents
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
				return 0, ErrInvalidLengthEvents
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowEvents
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
				next, err := skipEvents(data[start:])
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
	ErrInvalidLengthEvents = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowEvents   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("events.proto", fileDescriptorEvents) }

var fileDescriptorEvents = []byte{
	// 401 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xcc, 0x92, 0xc1, 0xae, 0xd2, 0x40,
	0x18, 0x85, 0x3b, 0xe4, 0x7a, 0x6f, 0xee, 0x60, 0xa4, 0x77, 0x24, 0x58, 0x5d, 0x0c, 0xa4, 0x0b,
	0xc3, 0x42, 0x4b, 0x82, 0x0f, 0x60, 0x0a, 0x6d, 0xa4, 0x86, 0x14, 0x52, 0x2a, 0xdb, 0xa6, 0xd0,
	0xb1, 0x4e, 0x82, 0x9d, 0xa6, 0x0c, 0x24, 0xb8, 0xf2, 0x0d, 0xf4, 0x19, 0x8c, 0x0b, 0x1f, 0xc5,
	0x25, 0x4b, 0x57, 0x44, 0xc6, 0x8d, 0x4b, 0x1e, 0xc1, 0x30, 0xa5, 0x04, 0xe1, 0x05, 0xee, 0xae,
	0x73, 0xce, 0xf9, 0xbf, 0x9c, 0xff, 0x4f, 0xe1, 0x43, 0xb2, 0x24, 0x09, 0x9f, 0x1b, 0x69, 0xc6,
	0x38, 0x43, 0xd7, 0x34, 0x59, 0x52, 0x4e, 0x9e, 0xbd, 0x8c, 0x29, 0xff, 0xb0, 0x98, 0x18, 0x53,
	0xf6, 0xb1, 0x15, 0xb3, 0x98, 0xb5, 0xa4, 0x3d, 0x59, 0xbc, 0x97, 0x2f, 0xf9, 0x90, 0x5f, 0xf9,
	0x98, 0xfe, 0xa5, 0x04, 0x1f, 0xd8, 0x7b, 0x0e, 0x7a, 0x0e, 0xaf, 0xf8, 0x2a, 0x25, 0x1a, 0x68,
	0x80, 0xe6, 0xa3, 0x36, 0x32, 0x72, 0x9e, 0x21, 0x4d, 0xc3, 0x5f, 0xa5, 0xc4, 0x93, 0x3e, 0x7a,
	0x0b, 0x51, 0x6e, 0x45, 0xc1, 0x94, 0xcd, 0x66, 0x24, 0x8c, 0x17, 0x64, 0xae, 0xc1, 0x06, 0x68,
	0x96, 0xdb, 0x4f, 0x8b, 0x29, 0x27, 0x4f, 0x74, 0x8f, 0x81, 0x9e, 0xe2, 0xdd, 0xd1, 0x73, 0x11,
	0x59, 0x50, 0x2d, 0x58, 0x69, 0xc8, 0xe9, 0x7e, 0x1d, 0xad, 0x2c, 0x49, 0x4f, 0xce, 0x48, 0xc3,
	0x83, 0xdd, 0x53, 0xbc, 0x0a, 0xfd, 0x5f, 0xd2, 0x4d, 0x78, 0xb5, 0xef, 0x87, 0xca, 0xf0, 0xc6,
	0x71, 0xc7, 0x66, 0xdf, 0xb1, 0x54, 0x05, 0xd5, 0x20, 0x72, 0xdc, 0xb1, 0xe3, 0xdb, 0x56, 0xd0,
	0x1d, 0xf4, 0xfb, 0xb6, 0xf9, 0xe6, 0x9d, 0x3d, 0x52, 0x01, 0xaa, 0x42, 0xb5, 0xd0, 0x87, 0xa6,
	0xef, 0xd8, 0xae, 0x3f, 0x52, 0x4b, 0x9d, 0x5b, 0x78, 0x13, 0x11, 0x1e, 0xd2, 0xd9, 0x5c, 0xff,
	0x0e, 0xe0, 0xdd, 0x45, 0x7d, 0xe4, 0xc2, 0x1a, 0xcb, 0xe2, 0x30, 0xa1, 0x9f, 0x42, 0x4e, 0x59,
	0x12, 0x90, 0x84, 0x53, 0xbe, 0x0a, 0x68, 0x24, 0xef, 0x75, 0xdb, 0xd1, 0xc4, 0xa6, 0x5e, 0x1d,
	0x9c, 0x24, 0x6c, 0x19, 0x70, 0x2c, 0xaf, 0xca, 0x2e, 0xd5, 0x08, 0xbd, 0x86, 0x87, 0x73, 0x64,
	0x27, 0xa8, 0x92, 0x44, 0x3d, 0x16, 0x9b, 0x7a, 0x25, 0x6f, 0x90, 0x1d, 0x29, 0x87, 0xa5, 0x0b,
	0x21, 0xd2, 0xbf, 0x01, 0x58, 0x39, 0xbb, 0xcd, 0xbd, 0x2b, 0xd9, 0x79, 0xb1, 0xde, 0x62, 0xf0,
	0x6b, 0x8b, 0x95, 0xdd, 0x16, 0x83, 0xcf, 0x02, 0x83, 0x1f, 0x02, 0x83, 0x9f, 0x02, 0x83, 0xb5,
	0xc0, 0xe0, 0xb7, 0xc0, 0xe0, 0xaf, 0xc0, 0xca, 0x4e, 0x60, 0xf0, 0xf5, 0x0f, 0x56, 0x26, 0xd7,
	0xf2, 0x97, 0x7c, 0xf5, 0x2f, 0x00, 0x00, 0xff, 0xff, 0x0a, 0x44, 0x9c, 0x50, 0xd9, 0x02, 0x00,
	0x00,
}
