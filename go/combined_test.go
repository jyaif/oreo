package oreo

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func CheckSerialization(t *testing.T, i interface{}, expected []byte) {
	t.Helper()

	buf := new(bytes.Buffer)
	err := Serialize(i, buf)
	if err != nil {
		t.Fatalf("Serialize failed unexpectedly:\n Error: %v\n", err)
	}

	assert.Equal(t, len(expected), buf.Len(), "Buffer length should match expected length")
	assert.Equal(t, expected, buf.Bytes(), "Serialized value should match expected value")
}

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

func Check(t *testing.T, i interface{}, expected []byte) {
	t.Helper()

	CheckSerialization(t, i, expected)
	CheckDeserialization(t, expected, i)
}

func TestBoolean(t *testing.T) {
	Check(t, false, []byte{0})
	Check(t, true, []byte{1})
}

func TestInt8(t *testing.T) {
	Check(t, int8(127), []byte{127})
	Check(t, int8(-128), []byte{0x80})
}
func TestUint8(t *testing.T) {
	Check(t, uint8(127), []byte{127})
	Check(t, uint8(128), []byte{0x80})
	Check(t, uint8(255), []byte{0xff})
}

func TestVariableLengthInt(t *testing.T) {
	Check(t, 0, []byte{0})
	Check(t, 1, []byte{1})
	Check(t, -1, []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 1})
	Check(t, -2, []byte{254, 255, 255, 255, 255, 255, 255, 255, 255, 1})
	Check(t, 127, []byte{127})
	Check(t, 128, []byte{128, 1})
	Check(t, 200, []byte{200, 1})
	Check(t, 255, []byte{255, 1})
	Check(t, 256, []byte{128, 2})
	Check(t, 300, []byte{172, 2})
	Check(t, 32767, []byte{255, 255, 1})
	Check(t, 32768, []byte{128, 128, 2})
	Check(t, 65535, []byte{255, 255, 3})
	Check(t, 65536, []byte{128, 128, 4})
	Check(t, 0x7fffffff, []byte{255, 255, 255, 255, 7})
	Check(t, 0x80000000, []byte{128, 128, 128, 128, 8})
	Check(t, 0xffffffff, []byte{255, 255, 255, 255, 15})
	Check(t, 0x111111111111111, []byte{145, 162, 196, 136, 145, 162, 196, 136, 1})
	Check(t, 0x7fffffffffffffff, []byte{255, 255, 255, 255, 255, 255, 255, 255, 127})
}

func TestString(t *testing.T) {
	Check(t, "", []byte{0})
	Check(t, "a", []byte{1, 'a'})
	Check(t, "hello", []byte{5, 'h', 'e', 'l', 'l', 'o'})
	Check(t, "hello world", []byte{11, 'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd'})
}

func TestArray(t *testing.T) {
	Check(t, []int8{}, []byte{0})
	Check(t, []int8{1}, []byte{1, 1})
	Check(t, []int8{1, 2}, []byte{2, 1, 2})
	Check(t, []int8{1, 2, 3}, []byte{3, 1, 2, 3})
	Check(t, []int8{1, 2, 3, 4}, []byte{4, 1, 2, 3, 4})

	Check(t, []string{}, []byte{0})
	Check(t, []string{"a"}, []byte{1, 1, 'a'})
	Check(t, []string{"a", "b"}, []byte{2, 1, 'a', 1, 'b'})
	Check(t, []string{"hello", "world"}, []byte{2, 5, 'h', 'e', 'l', 'l', 'o', 5, 'w', 'o', 'r', 'l', 'd'})
}

func TestStruct(t *testing.T) {
	type InnerStruct struct {
		X int8
		Y string
	}
	type TestStruct struct {
		A int8
		B string
		C []int32
		D InnerStruct
		E []InnerStruct
	}
	testStruct := TestStruct{
		A: 1,
		B: "hello",
		C: []int32{1, 2, 3},
		D: InnerStruct{
			X: 2,
			Y: "world",
		},
		E: []InnerStruct{
			{X: 5, Y: "qux"},
			{X: 6, Y: "foo"},
		},
	}
	expected := []byte{
		1,                          // A
		5, 'h', 'e', 'l', 'l', 'o', // B
		3, 1, 2, 3, // C
		2, 5, 'w', 'o', 'r', 'l', 'd', // D
		2,                   // len(E) size
		5, 3, 'q', 'u', 'x', // E[0]
		6, 3, 'f', 'o', 'o', // E[1]
	}
	Check(t, testStruct, expected)
}

func TestPointers(t *testing.T) {
	{
		var p *string
		expected := []byte{0}
		Check(t, p, expected)
	}
	{
		p := new(string)
		*p = "booya"
		expected := []byte{
			1, 5, 'b', 'o', 'o', 'y', 'a',
		}
		Check(t, p, expected)
	}
	{
		str := new(string)
		*str = "foo"

		type InnerStruct struct {
			X int8
			Y string
		}
		type TestStruct struct {
			A *int8
			B *string
			C *[]int32
			D *InnerStruct
			E *[]InnerStruct
		}
		testStruct := TestStruct{
			A: nil,
			B: str,
			C: nil,
			D: nil,
			E: nil,
		}
		expected := []byte{
			0,                   // A
			1, 3, 'f', 'o', 'o', // B
			0, // C
			0, // D
			0, // E
		}
		Check(t, testStruct, expected)
	}
}
