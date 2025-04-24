package oreo

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
)

func ReadBool(buf *bytes.Buffer, i *bool) error {
	b, err := buf.ReadByte()
	if err != nil {
		return err
	}
	*i = b != 0
	return nil
}

func ReadVariableLengthInteger(buf *bytes.Buffer, i *uint64) error {
	shift := int(0)
	for {
		b, err := buf.ReadByte()
		if err != nil {
			return err
		}
		*i |= uint64(b&0x7F) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
	}
	return nil
}

func ReadInt(buf *bytes.Buffer, i *int) error {
	var u uint64
	err := ReadVariableLengthInteger(buf, &u)
	if err != nil {
		return err
	}
	*i = int(u)
	return nil
}

func ReadInt8(buf *bytes.Buffer, i *int8) error {
	b, err := buf.ReadByte()
	if err != nil {
		return err
	}
	*i = int8(b)
	return nil
}

func ReadInt16(buf *bytes.Buffer, i *int16) error {
	var u uint64
	err := ReadVariableLengthInteger(buf, &u)
	if err != nil {
		return err
	}
	*i = int16(u)
	return nil
}

func ReadInt32(buf *bytes.Buffer, i *int32) error {
	var u uint64
	err := ReadVariableLengthInteger(buf, &u)
	if err != nil {
		return err
	}
	*i = int32(u)
	return nil
}

func ReadInt64(buf *bytes.Buffer, i *int64) error {
	var u uint64
	err := ReadVariableLengthInteger(buf, &u)
	if err != nil {
		return err
	}
	*i = int64(u)
	return nil
}

func ReadUint(buf *bytes.Buffer, i *uint) error {
	var u uint64
	err := ReadVariableLengthInteger(buf, &u)
	if err != nil {
		return err
	}
	*i = uint(u)
	return nil
}

func ReadUint8(buf *bytes.Buffer, i *uint8) error {
	b, err := buf.ReadByte()
	if err != nil {
		return err
	}
	*i = uint8(b)
	return nil
}

func ReadUint16(buf *bytes.Buffer, i *uint16) error {
	var u uint64
	err := ReadVariableLengthInteger(buf, &u)
	if err != nil {
		return err
	}
	*i = uint16(u)
	return nil
}

func ReadUint32(buf *bytes.Buffer, i *uint32) error {
	var u uint64
	err := ReadVariableLengthInteger(buf, &u)
	if err != nil {
		return err
	}
	*i = uint32(u)
	return nil
}

func ReadUint64(buf *bytes.Buffer, i *uint64) error {
	return ReadVariableLengthInteger(buf, i)
}

func ReadString(buf *bytes.Buffer, i *string) error {
	var length uint64
	err := ReadVariableLengthInteger(buf, &length)
	if err != nil {
		return err
	}
	strBytes := make([]byte, length)
	_, err = buf.Read(strBytes)
	if err != nil {
		return err
	}
	*i = string(strBytes)
	return nil
}

func ReadArray(buf *bytes.Buffer, i interface{}) error {
	// 1. Validate input 'i' is a pointer to a slice
	ptrVal := reflect.ValueOf(i)
	if ptrVal.Kind() != reflect.Ptr {
		return errors.New("ReadArray: input is not a pointer")
	}

	sliceVal := ptrVal.Elem() // Get the value the pointer points to (the slice itself)
	if sliceVal.Kind() != reflect.Slice {
		return errors.New("ReadArray: input is not a pointer to a slice")
	}

	// 2. Read the length of the array
	var length uint64
	err := ReadVariableLengthInteger(buf, &length)
	if err != nil {
		return fmt.Errorf("ReadArray: failed to read array length: %w", err)
	}

	// Check for potential overflow that would cause memory issues.
	const maxLen = 10000
	if length > maxLen {
		return fmt.Errorf("ReadArray: array length %d exceeds maximum allowed %d", length, maxLen)
	}
	intLen := int(length)

	// 3. Create a new slice of the correct type and length
	sliceType := sliceVal.Type()                             // Type of the slice (e.g., []int32)
	newSlice := reflect.MakeSlice(sliceType, intLen, intLen) // Create new slice

	// 4. Deserialize each element into the new slice
	for j := 0; j < intLen; j++ {
		// Get a pointer to the j-th element in the new slice
		elemPtr := newSlice.Index(j).Addr().Interface()
		// Deserialize into the element pointer
		err := Deserialize(buf, elemPtr)
		if err != nil {
			return fmt.Errorf("ReadArray: failed to deserialize element %d (0-based): %w", j, err)
		}
	}

	sliceVal.Set(newSlice)
	return nil
}

// ReadPointer deserializes a pointer from the buffer.
// It first reads a boolean flag:
// - If true: allocates memory for the pointed-to type and deserializes into it.
// - If false: sets the pointer pointed to by 'i' to nil.
func ReadPointer(buf *bytes.Buffer, i interface{}) error {
	ptrVal := reflect.ValueOf(i)
	if ptrVal.IsNil() {
		return errors.New("ReadPointer: input pointer is nil")
	}

	// Get the pointer that we need to potentially set (e.g., *int in **int)
	innerPtrVal := ptrVal.Elem()
	if innerPtrVal.Kind() != reflect.Ptr {
		// This check ensures we have a pointer *to* a pointer
		return fmt.Errorf("ReadPointer: input is not a pointer to a pointer (got %T)", i)
	}
	if !innerPtrVal.CanSet() {
		return errors.New("ReadPointer: inner pointer is not settable")
	}

	// 2. Read the validity flag (boolean indicating if the pointer is non-nil)
	var isValid bool
	err := ReadBool(buf, &isValid)
	if err != nil {
		return fmt.Errorf("ReadPointer: failed to read validity flag: %w", err)
	}

	// 3. Handle based on the validity flag
	if !isValid {
		// Pointer is nil, set the target pointer to its zero value (nil)
		innerPtrVal.Set(reflect.Zero(innerPtrVal.Type())) // Set the *int (or *MyStruct etc.) to nil
		return nil
	} else {
		// Pointer is valid, need to allocate and deserialize the underlying value

		// Get the type of the element the pointer should point to (e.g., int from *int)
		elemType := innerPtrVal.Type().Elem()

		// Create a new value of the element type (allocates memory)
		// reflect.New returns a Value representing a pointer to a new zero value for 'elemType'.
		// So, if elemType is int, newValue is a Value representing *int.
		newValue := reflect.New(elemType)

		// Set the inner pointer (e.g., the *int) to point to this newly allocated memory.
		innerPtrVal.Set(newValue)

		// Now, deserialize the data *into* the newly allocated memory.
		// We pass the interface{} representation of the pointer to the new memory.
		return Deserialize(buf, newValue.Interface())
	}
}

// Populates the variable pointed to by `v` by reading data from `buf“.
// If `v` points to a struct, it deserializes field by field based on struct order.
// If `v` points to a basic type (bool, int, string...), it deserializes directly into it.
func Deserialize(buf *bytes.Buffer, v interface{}) error {
	ptrVal := reflect.ValueOf(v)

	// We need a pointer to modify the original variable.
	if ptrVal.Kind() != reflect.Ptr {
		return fmt.Errorf("Deserialize: expected a pointer, got %T", v)
	}
	if ptrVal.IsNil() {
		return fmt.Errorf("Deserialize: expected a non-nil pointer, got nil %T", v)
	}

	targetVal := ptrVal.Elem()

	// Check if the pointed-to element is settable.
	if !targetVal.CanSet() {
		return fmt.Errorf("Deserialize: cannot set value of type %s (is it addressable/exported?)", targetVal.Type())
	}

	if targetVal.Kind() == reflect.Struct {
		structType := targetVal.Type()
		for i := 0; i < targetVal.NumField(); i++ {
			fieldVal := targetVal.Field(i)
			if !fieldVal.CanSet() { // Skip unexported/unsettable fields
				continue
			}
			fieldPtrVal := fieldVal.Addr()
			fieldPtrInterface := fieldPtrVal.Interface()
			err := Deserialize(buf, fieldPtrInterface)
			if err != nil {
				fieldType := structType.Field(i) // Field metadata
				return fmt.Errorf("error deserializing struct field '%s' (%s): %w",
					fieldType.Name, fieldVal.Kind(), err)
			}
		}
		return nil
	} else {
		var err error
		switch targetVal.Kind() {
		case reflect.Bool:
			err = ReadBool(buf, v.(*bool))
		case reflect.String:
			err = ReadString(buf, v.(*string))
		case reflect.Uint:
			err = ReadUint(buf, v.(*uint))
		case reflect.Uint8:
			err = ReadUint8(buf, v.(*uint8))
		case reflect.Uint16:
			err = ReadUint16(buf, v.(*uint16))
		case reflect.Uint32:
			err = ReadUint32(buf, v.(*uint32))
		case reflect.Uint64:
			err = ReadUint64(buf, v.(*uint64))
		case reflect.Int:
			err = ReadInt(buf, v.(*int))
		case reflect.Int8:
			err = ReadInt8(buf, v.(*int8))
		case reflect.Int16:
			err = ReadInt16(buf, v.(*int16))
		case reflect.Int32:
			err = ReadInt32(buf, v.(*int32))
		case reflect.Int64:
			err = ReadInt64(buf, v.(*int64))
		case reflect.Slice:
			err = ReadArray(buf, v)
		case reflect.Pointer:
			err = ReadPointer(buf, v)
		default:
			err = fmt.Errorf("unsupported type for direct deserialization: %s", targetVal.Kind())
		}
		if err != nil {
			return fmt.Errorf("error deserializing %s: %w", targetVal.Kind(), err)
		}
		return nil
	}
}
