package dbutil

import (
	"reflect"
	"testing"
)

func TestAppendInt64(t *testing.T) {
	ints := []int64{0, 1, 2, 3, 4}
	ifs := AppendInt64sToInterfaceSlice(nil, ints)

	expected := []interface{}{int64(0), int64(1), int64(2), int64(3), int64(4)}
	if !reflect.DeepEqual(ifs, expected) {
		t.Fatalf("Expected %#v, got %#v", expected, ifs)
	}

	intsA := []int64{10, 20}
	intsB := []int64{30, 40}
	ifs = AppendInt64sToInterfaceSlice(nil, intsA)
	ifs = AppendInt64sToInterfaceSlice(ifs, intsB)

	expected = []interface{}{int64(10), int64(20), int64(30), int64(40)}
	if !reflect.DeepEqual(ifs, expected) {
		t.Fatalf("Expected %#v, got %#v", expected, ifs)
	}
}

func TestAppendUInt64(t *testing.T) {
	ints := []uint64{0, 1, 2, 3, 4}
	ifs := AppendUint64sToInterfaceSlice(nil, ints)

	expected := []interface{}{uint64(0), uint64(1), uint64(2), uint64(3), uint64(4)}
	if !reflect.DeepEqual(ifs, expected) {
		t.Fatalf("Expected %#v, got %#v", expected, ifs)
	}

	intsA := []uint64{10, 20}
	intsB := []uint64{30, 40}
	ifs = AppendUint64sToInterfaceSlice(nil, intsA)
	ifs = AppendUint64sToInterfaceSlice(ifs, intsB)

	expected = []interface{}{uint64(10), uint64(20), uint64(30), uint64(40)}
	if !reflect.DeepEqual(ifs, expected) {
		t.Fatalf("Expected %#v, got %#v", expected, ifs)
	}
}

func TestAppendString(t *testing.T) {
	strings := []string{"A", "B", "C", "D"}
	ifs := AppendStringsToInterfaceSlice(nil, strings)

	expected := []interface{}{"A", "B", "C", "D"}
	if !reflect.DeepEqual(ifs, expected) {
		t.Fatalf("Expected %#v, got %#v", expected, ifs)
	}

	stringsA := []string{"10", "20"}
	stringsB := []string{"30", "40"}
	ifs = AppendStringsToInterfaceSlice(nil, stringsA)
	ifs = AppendStringsToInterfaceSlice(ifs, stringsB)

	expected = []interface{}{"10", "20", "30", "40"}
	if !reflect.DeepEqual(ifs, expected) {
		t.Fatalf("Expected %#v, got %#v", expected, ifs)
	}

}
