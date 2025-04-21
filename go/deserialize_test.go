package oreo

import (
	"bytes"
	"reflect"
	"testing"
)

func CheckDeserialization(t *testing.T, buffer []byte, expected interface{}) {
	t.Helper()

	expectedType := reflect.TypeOf(expected)
	if expectedType == nil {
		t.Fatal("CheckDeserialization: 'expected' value cannot be a typeless nil")
		return
	}

	targetPtrVal := reflect.New(expectedType)
	// Get the actual pointer interface (e.g., *Foo, *string, *int)
	targetPtrInterface := targetPtrVal.Interface()

	buf := bytes.NewBuffer(buffer)
	initialLen := buf.Len()

	err := Deserialize(buf, targetPtrInterface)

	if err != nil {
		t.Fatalf("Deserialize failed unexpectedly:\n Error: %v\n Expected Type: %T\n Buffer (len %d): %x",
			err, expected, initialLen, buffer)
	}
	if buf.Len() > 0 {
		t.Fatalf("Buffer not fully consumed after Deserialize:\n Bytes remaining: %d\n Initial buffer: %x",
			buf.Len(), buffer)
	}
	actualValue := targetPtrVal.Elem().Interface()
	if !reflect.DeepEqual(actualValue, expected) {
		t.Fatalf("Deserialized value mismatch:\n Expected: %#v (%T)\n Actual:   %#v (%T)\n Buffer (len %d): %x",
			expected, expected,
			actualValue, actualValue,
			initialLen, buffer)
	}
}

func TestBooleanDeserialization(t *testing.T) {
	CheckDeserialization(t, []byte{0}, false)
	CheckDeserialization(t, []byte{1}, true)
}

func TestVariableLengthIntegerDeserialization(t *testing.T) {
	CheckDeserialization(t, []byte{0}, uint64(0))
	CheckDeserialization(t, []byte{1}, uint64(1))
	CheckDeserialization(t, []byte{127}, uint64(127))
	CheckDeserialization(t, []byte{128, 1}, uint64(128))
	CheckDeserialization(t, []byte{200, 1}, uint64(200))
	CheckDeserialization(t, []byte{255, 1}, uint64(255))
	CheckDeserialization(t, []byte{128, 2}, uint64(256))
	CheckDeserialization(t, []byte{172, 2}, uint64(300))
	CheckDeserialization(t, []byte{255, 255, 1}, uint64(32767))
	CheckDeserialization(t, []byte{128, 128, 2}, uint64(32768))
	CheckDeserialization(t, []byte{255, 255, 3}, uint64(65535))
	CheckDeserialization(t, []byte{128, 128, 4}, uint64(65536))
	CheckDeserialization(t, []byte{128, 128, 4}, uint32(65536))
	CheckDeserialization(t, []byte{128, 128, 4}, int32(65536))
	CheckDeserialization(t, []byte{255, 255, 255, 255, 7}, 0x7fffffff)
	CheckDeserialization(t, []byte{255, 255, 255, 255, 7}, uint64(0x7fffffff))
	CheckDeserialization(t, []byte{128, 128, 128, 128, 8}, 0x80000000)
	CheckDeserialization(t, []byte{255, 255, 255, 255, 15}, 0xffffffff)
	CheckDeserialization(t, []byte{145, 162, 196, 136, 145, 162, 196, 136, 1}, 0x111111111111111)
	CheckDeserialization(t, []byte{255, 255, 255, 255, 255, 255, 255, 255, 127}, 0x7fffffffffffffff)
}

func TestStringDeserialization(t *testing.T) {
	CheckDeserialization(t, []byte{0}, "")
	CheckDeserialization(t, []byte{1, 'a'}, "a")
	CheckDeserialization(t, []byte{5, 'h', 'e', 'l', 'l', 'o'}, "hello")
}

func TestStructDeserialization(t *testing.T) {
	type TestStruct struct {
		A int
		B string
		C bool
	}
	expected := TestStruct{A: 42, B: "hello", C: true}
	buffer := []byte{42, 5, 'h', 'e', 'l', 'l', 'o', 1}
	CheckDeserialization(t, buffer, expected)
}

func TestArrayDeserialization(t *testing.T) {
	CheckDeserialization(t, []byte{0}, []int{})
	CheckDeserialization(t, []byte{1, 1}, []int{1})
	CheckDeserialization(t, []byte{2, 1, 2}, []int{1, 2})
	CheckDeserialization(t, []byte{3, 1, 2, 3}, []int{1, 2, 3})
	CheckDeserialization(t, []byte{4, 1, 2, 3, 4}, []int{1, 2, 3, 4})
	CheckDeserialization(t, []byte{2, 5, 'h', 'e', 'l', 'l', 'o', 5, 'w', 'o', 'r', 'l', 'd'}, []string{"hello", "world"})
}
