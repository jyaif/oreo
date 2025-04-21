package oreo

import (
	"bytes"
	"reflect"
	"testing"
)

func CheckDeserialization(t *testing.T, buffer []byte, expected interface{}) {
	// Mark this function as a test helper.
	// Error/Fatal messages will report the caller's line number.
	t.Helper()

	// 1. Validate the expected value and get its type.
	// We cannot deserialize into a nil interface directly.
	if expected == nil {
		t.Fatal("CheckDeserialization: 'expected' value cannot be nil")
		return // Or panic, as this is a test setup error
	}
	expectedType := reflect.TypeOf(expected)
	// Further check in case expected was an explicit `interface{}(nil)`
	if expectedType == nil {
		t.Fatal("CheckDeserialization: 'expected' value cannot be a typeless nil")
		return
	}

	// 2. Create a pointer to a new zero value of the expected type.
	// reflect.New returns a Value representing a pointer to a new zero value.
	targetPtrVal := reflect.New(expectedType)
	// Get the actual pointer interface (e.g., *Foo, *string, *int)
	targetPtrInterface := targetPtrVal.Interface()

	// 3. Prepare the buffer
	buf := bytes.NewBuffer(buffer)
	initialLen := buf.Len() // For debugging output

	// 4. Perform the deserialization
	err := Deserialize(buf, targetPtrInterface) // Pass the pointer

	// 5. Check for deserialization errors
	if err != nil {
		// Use Fatalf as the test cannot proceed meaningfully.
		t.Fatalf("Deserialize failed unexpectedly:\n Error: %v\n Expected Type: %T\n Buffer (len %d): %x",
			err, expected, initialLen, buffer)
		return // Keep compiler happy, Fatalf exits
	}

	// --- Optional Check: Buffer Consumption ---
	// Decide if you want to enforce full buffer consumption.
	// Sometimes partial reads might be valid depending on the protocol.
	// If full consumption is always expected:

	if buf.Len() > 0 {
		t.Errorf("Buffer not fully consumed after Deserialize:\n Bytes remaining: %d\n Initial buffer: %x",
			buf.Len(), buffer)
		// Don't return here, still check the value if no error occurred
	}

	// 6. Get the actual deserialized value
	// targetPtrVal is the reflect.Value of the pointer (*T).
	// Elem() gets the reflect.Value of the pointed-to value (T).
	actualValue := targetPtrVal.Elem().Interface() // Get the actual value (T) as interface{}

	// 7. Compare actual vs. expected
	// reflect.DeepEqual handles comparison for structs, slices, maps, etc.
	if !reflect.DeepEqual(actualValue, expected) {
		// Use Errorf to report the mismatch but allow other tests to run.
		t.Errorf("Deserialized value mismatch:\n Expected: %#v (%T)\n Actual:   %#v (%T)\n Buffer (len %d): %x",
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
