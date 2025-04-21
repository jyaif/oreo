package oreo

import (
	"bytes"
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

func TestBooleanSerialization(t *testing.T) {
	CheckSerialization(t, false, []byte{0})
	CheckSerialization(t, true, []byte{1})
}

func TestInt8Serialization(t *testing.T) {
	CheckSerialization(t, int8(127), []byte{127})
	CheckSerialization(t, int8(-128), []byte{0x80})
}
func TestUint8Serialization(t *testing.T) {
	CheckSerialization(t, uint8(127), []byte{127})
	CheckSerialization(t, uint8(128), []byte{0x80})
	CheckSerialization(t, uint8(255), []byte{0xff})
}

func TestVariableLengthIntSerialization(t *testing.T) {
	CheckSerialization(t, 0, []byte{0})
	CheckSerialization(t, 1, []byte{1})
	CheckSerialization(t, -1, []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 1})
	CheckSerialization(t, -2, []byte{254, 255, 255, 255, 255, 255, 255, 255, 255, 1})
	CheckSerialization(t, 127, []byte{127})
	CheckSerialization(t, 128, []byte{128, 1})
	CheckSerialization(t, 200, []byte{200, 1})
	CheckSerialization(t, 255, []byte{255, 1})
	CheckSerialization(t, 256, []byte{128, 2})
	CheckSerialization(t, 300, []byte{172, 2})
	CheckSerialization(t, 32767, []byte{255, 255, 1})
	CheckSerialization(t, 32768, []byte{128, 128, 2})
	CheckSerialization(t, 65535, []byte{255, 255, 3})
	CheckSerialization(t, 65536, []byte{128, 128, 4})
	CheckSerialization(t, 0x7fffffff, []byte{255, 255, 255, 255, 7})
	CheckSerialization(t, 0x80000000, []byte{128, 128, 128, 128, 8})
	CheckSerialization(t, 0xffffffff, []byte{255, 255, 255, 255, 15})
	CheckSerialization(t, 0x111111111111111, []byte{145, 162, 196, 136, 145, 162, 196, 136, 1})
	CheckSerialization(t, 0x7fffffffffffffff, []byte{255, 255, 255, 255, 255, 255, 255, 255, 127})
}

func TestStringSerialization(t *testing.T) {
	CheckSerialization(t, "", []byte{0})
	CheckSerialization(t, "a", []byte{1, 'a'})
	CheckSerialization(t, "hello", []byte{5, 'h', 'e', 'l', 'l', 'o'})
	CheckSerialization(t, "hello world", []byte{11, 'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd'})
}

func TestArraySerialization(t *testing.T) {
	CheckSerialization(t, []int8{}, []byte{0})
	CheckSerialization(t, []int8{1}, []byte{1, 1})
	CheckSerialization(t, []int8{1, 2}, []byte{2, 1, 2})
	CheckSerialization(t, []int8{1, 2, 3}, []byte{3, 1, 2, 3})
	CheckSerialization(t, []int8{1, 2, 3, 4}, []byte{4, 1, 2, 3, 4})

	CheckSerialization(t, []string{}, []byte{0})
	CheckSerialization(t, []string{"a"}, []byte{1, 1, 'a'})
	CheckSerialization(t, []string{"a", "b"}, []byte{2, 1, 'a', 1, 'b'})
	CheckSerialization(t, []string{"hello", "world"}, []byte{2, 5, 'h', 'e', 'l', 'l', 'o', 5, 'w', 'o', 'r', 'l', 'd'})
}

func TestSerializeStruct(t *testing.T) {
	type TestStruct struct {
		A int8
		B string
		C []int32
	}
	testStruct := TestStruct{
		A: 1,
		B: "hello",
		C: []int32{1, 2, 3},
	}
	expected := []byte{
		1,                          // A
		5, 'h', 'e', 'l', 'l', 'o', // B
		3, 1, 2, 3, // C
	}
	CheckSerialization(t, testStruct, expected)
}
