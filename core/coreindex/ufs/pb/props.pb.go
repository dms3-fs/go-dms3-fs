// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: props.proto

package ufs_pb

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"

import bytes "bytes"

import strings "strings"
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

type IndexNode_DataType int32

const (
	IndexNode_Repo    IndexNode_DataType = 0
	IndexNode_Reposet IndexNode_DataType = 1
)

var IndexNode_DataType_name = map[int32]string{
	0: "Repo",
	1: "Reposet",
}
var IndexNode_DataType_value = map[string]int32{
	"Repo":    0,
	"Reposet": 1,
}

func (x IndexNode_DataType) String() string {
	return proto.EnumName(IndexNode_DataType_name, int32(x))
}
func (IndexNode_DataType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_props_4e64e2efd6209fbf, []int{0, 0}
}

type IndexNode struct {
	Type                 IndexNode_DataType `protobuf:"varint,1,opt,name=Type,proto3,enum=ufs.pb.IndexNode_DataType" json:"Type,omitempty"`
	Data                 []byte             `protobuf:"bytes,2,opt,name=Data,proto3" json:"Data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *IndexNode) Reset()      { *m = IndexNode{} }
func (*IndexNode) ProtoMessage() {}
func (*IndexNode) Descriptor() ([]byte, []int) {
	return fileDescriptor_props_4e64e2efd6209fbf, []int{0}
}
func (m *IndexNode) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *IndexNode) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_IndexNode.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (dst *IndexNode) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IndexNode.Merge(dst, src)
}
func (m *IndexNode) XXX_Size() int {
	return m.Size()
}
func (m *IndexNode) XXX_DiscardUnknown() {
	xxx_messageInfo_IndexNode.DiscardUnknown(m)
}

var xxx_messageInfo_IndexNode proto.InternalMessageInfo

func (m *IndexNode) GetType() IndexNode_DataType {
	if m != nil {
		return m.Type
	}
	return IndexNode_Repo
}

func (m *IndexNode) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterType((*IndexNode)(nil), "ufs.pb.IndexNode")
	proto.RegisterEnum("ufs.pb.IndexNode_DataType", IndexNode_DataType_name, IndexNode_DataType_value)
}
func (this *IndexNode) VerboseEqual(that interface{}) error {
	if that == nil {
		if this == nil {
			return nil
		}
		return fmt.Errorf("that == nil && this != nil")
	}

	that1, ok := that.(*IndexNode)
	if !ok {
		that2, ok := that.(IndexNode)
		if ok {
			that1 = &that2
		} else {
			return fmt.Errorf("that is not of type *IndexNode")
		}
	}
	if that1 == nil {
		if this == nil {
			return nil
		}
		return fmt.Errorf("that is type *IndexNode but is nil && this != nil")
	} else if this == nil {
		return fmt.Errorf("that is type *IndexNode but is not nil && this == nil")
	}
	if this.Type != that1.Type {
		return fmt.Errorf("Type this(%v) Not Equal that(%v)", this.Type, that1.Type)
	}
	if !bytes.Equal(this.Data, that1.Data) {
		return fmt.Errorf("Data this(%v) Not Equal that(%v)", this.Data, that1.Data)
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return fmt.Errorf("XXX_unrecognized this(%v) Not Equal that(%v)", this.XXX_unrecognized, that1.XXX_unrecognized)
	}
	return nil
}
func (this *IndexNode) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*IndexNode)
	if !ok {
		that2, ok := that.(IndexNode)
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
	if this.Type != that1.Type {
		return false
	}
	if !bytes.Equal(this.Data, that1.Data) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *IndexNode) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 6)
	s = append(s, "&ufs_pb.IndexNode{")
	s = append(s, "Type: "+fmt.Sprintf("%#v", this.Type)+",\n")
	s = append(s, "Data: "+fmt.Sprintf("%#v", this.Data)+",\n")
	if this.XXX_unrecognized != nil {
		s = append(s, "XXX_unrecognized:"+fmt.Sprintf("%#v", this.XXX_unrecognized)+",\n")
	}
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringProps(v interface{}, typ string) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)
}
func (m *IndexNode) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *IndexNode) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Type != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintProps(dAtA, i, uint64(m.Type))
	}
	if len(m.Data) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintProps(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func encodeVarintProps(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func NewPopulatedIndexNode(r randyProps, easy bool) *IndexNode {
	this := &IndexNode{}
	this.Type = IndexNode_DataType([]int32{0, 1}[r.Intn(2)])
	v1 := r.Intn(100)
	this.Data = make([]byte, v1)
	for i := 0; i < v1; i++ {
		this.Data[i] = byte(r.Intn(256))
	}
	if !easy && r.Intn(10) != 0 {
		this.XXX_unrecognized = randUnrecognizedProps(r, 3)
	}
	return this
}

type randyProps interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

func randUTF8RuneProps(r randyProps) rune {
	ru := r.Intn(62)
	if ru < 10 {
		return rune(ru + 48)
	} else if ru < 36 {
		return rune(ru + 55)
	}
	return rune(ru + 61)
}
func randStringProps(r randyProps) string {
	v2 := r.Intn(100)
	tmps := make([]rune, v2)
	for i := 0; i < v2; i++ {
		tmps[i] = randUTF8RuneProps(r)
	}
	return string(tmps)
}
func randUnrecognizedProps(r randyProps, maxFieldNumber int) (dAtA []byte) {
	l := r.Intn(5)
	for i := 0; i < l; i++ {
		wire := r.Intn(4)
		if wire == 3 {
			wire = 5
		}
		fieldNumber := maxFieldNumber + r.Intn(100)
		dAtA = randFieldProps(dAtA, r, fieldNumber, wire)
	}
	return dAtA
}
func randFieldProps(dAtA []byte, r randyProps, fieldNumber int, wire int) []byte {
	key := uint32(fieldNumber)<<3 | uint32(wire)
	switch wire {
	case 0:
		dAtA = encodeVarintPopulateProps(dAtA, uint64(key))
		v3 := r.Int63()
		if r.Intn(2) == 0 {
			v3 *= -1
		}
		dAtA = encodeVarintPopulateProps(dAtA, uint64(v3))
	case 1:
		dAtA = encodeVarintPopulateProps(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	case 2:
		dAtA = encodeVarintPopulateProps(dAtA, uint64(key))
		ll := r.Intn(100)
		dAtA = encodeVarintPopulateProps(dAtA, uint64(ll))
		for j := 0; j < ll; j++ {
			dAtA = append(dAtA, byte(r.Intn(256)))
		}
	default:
		dAtA = encodeVarintPopulateProps(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	}
	return dAtA
}
func encodeVarintPopulateProps(dAtA []byte, v uint64) []byte {
	for v >= 1<<7 {
		dAtA = append(dAtA, uint8(uint64(v)&0x7f|0x80))
		v >>= 7
	}
	dAtA = append(dAtA, uint8(v))
	return dAtA
}
func (m *IndexNode) Size() (n int) {
	var l int
	_ = l
	if m.Type != 0 {
		n += 1 + sovProps(uint64(m.Type))
	}
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovProps(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovProps(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozProps(x uint64) (n int) {
	return sovProps(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *IndexNode) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&IndexNode{`,
		`Type:` + fmt.Sprintf("%v", this.Type) + `,`,
		`Data:` + fmt.Sprintf("%v", this.Data) + `,`,
		`XXX_unrecognized:` + fmt.Sprintf("%v", this.XXX_unrecognized) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringProps(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *IndexNode) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProps
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: IndexNode: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: IndexNode: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProps
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Type |= (IndexNode_DataType(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProps
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthProps
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProps(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthProps
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipProps(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowProps
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
					return 0, ErrIntOverflowProps
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
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
					return 0, ErrIntOverflowProps
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
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthProps
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowProps
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
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
				next, err := skipProps(dAtA[start:])
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
	ErrInvalidLengthProps = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowProps   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("props.proto", fileDescriptor_props_4e64e2efd6209fbf) }

var fileDescriptor_props_4e64e2efd6209fbf = []byte{
	// 216 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2e, 0x28, 0xca, 0x2f,
	0x28, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2b, 0x4d, 0x2b, 0xd6, 0x2b, 0x48, 0x92,
	0xd2, 0x4d, 0xcf, 0x2c, 0xc9, 0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x4f, 0xcf, 0x4f, 0xcf,
	0xd7, 0x07, 0x4b, 0x27, 0x95, 0xa6, 0x81, 0x79, 0x60, 0x0e, 0x98, 0x05, 0xd1, 0xa6, 0x54, 0xc4,
	0xc5, 0xe9, 0x99, 0x97, 0x92, 0x5a, 0xe1, 0x97, 0x9f, 0x92, 0x2a, 0xa4, 0xc7, 0xc5, 0x12, 0x52,
	0x59, 0x90, 0x2a, 0xc1, 0xa8, 0xc0, 0xa8, 0xc1, 0x67, 0x24, 0xa5, 0x07, 0x31, 0x52, 0x0f, 0xae,
	0x40, 0xcf, 0x25, 0xb1, 0x24, 0x11, 0xa4, 0x22, 0x08, 0xac, 0x4e, 0x48, 0x88, 0x8b, 0x05, 0x24,
	0x22, 0xc1, 0xa4, 0xc0, 0xa8, 0xc1, 0x13, 0x04, 0x66, 0x2b, 0x29, 0x72, 0x71, 0xc0, 0x54, 0x09,
	0x71, 0x70, 0xb1, 0x04, 0xa5, 0x16, 0xe4, 0x0b, 0x30, 0x08, 0x71, 0x73, 0xb1, 0x83, 0x58, 0xc5,
	0xa9, 0x25, 0x02, 0x8c, 0x4e, 0x3a, 0x37, 0x1e, 0xca, 0x31, 0x3c, 0x78, 0x28, 0xc7, 0xf8, 0xe1,
	0xa1, 0x1c, 0xe3, 0x8f, 0x87, 0x72, 0x8c, 0x0d, 0x8f, 0xe4, 0x18, 0x57, 0x3c, 0x92, 0x63, 0xdc,
	0xf1, 0x48, 0x8e, 0xf1, 0xc0, 0x23, 0x39, 0xc6, 0x13, 0x8f, 0xe4, 0x18, 0x2f, 0x3c, 0x92, 0x63,
	0x7c, 0xf0, 0x48, 0x8e, 0x31, 0x89, 0x0d, 0xec, 0x50, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff,
	0x54, 0x14, 0x9f, 0x45, 0xee, 0x00, 0x00, 0x00,
}