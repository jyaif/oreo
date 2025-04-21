package oreo

import (
	"bytes"
	"reflect"
)

func WriteBool(v bool, buf *bytes.Buffer) error {
	b := func() int8 {
		if v {
			return int8(1)
		}
		return int8(0)
	}()
	return buf.WriteByte(byte(b))
}

func WriteInt8(v int8, buf *bytes.Buffer) error {
	return buf.WriteByte(byte(v))
}

func WriteVariableLengthInt(v uint64, buf *bytes.Buffer) error {
	for v >= 0b10000000 {
		err := buf.WriteByte(byte(v | 0b10000000))
		if err != nil {
			return err
		}
		v >>= 7
	}
	return buf.WriteByte(byte(v))
}

func WriteString(s string, buf *bytes.Buffer) error {
	// Write the length of the string
	err := WriteVariableLengthInt(uint64(len(s)), buf)
	if err != nil {
		return err
	}
	// Write the string itself
	_, err = buf.WriteString(s)
	return err
}

func WriteArray(i interface{}, buf *bytes.Buffer) error {
	length := reflect.ValueOf(i).Len()
	err := WriteVariableLengthInt(uint64(length), buf)
	if err != nil {
		return err
	}
	for j := 0; j < length; j++ {
		err = Serialize(reflect.ValueOf(i).Index(j).Interface(), buf)
		if err != nil {
			return err
		}
	}
	return nil
}

func Serialize(i interface{}, buf *bytes.Buffer) error {
	t := reflect.TypeOf(i)
	kind := t.Kind()

	// Handle pointers
	// if kind == reflect.Ptr {
	// 	t = t.Elem()
	// 	return Serialize(t, buf)
	// }
	if kind == reflect.Bool {
		return WriteBool(i.(bool), buf)
	}
	if kind == reflect.Int8 {
		return WriteInt8(i.(int8), buf)
	}
	if kind == reflect.Uint8 {
		return WriteInt8(int8(i.(uint8)), buf)
	}
	if kind == reflect.Uint || kind == reflect.Uint16 || kind == reflect.Uint32 || kind == reflect.Uint64 {
		return WriteVariableLengthInt(reflect.ValueOf(i).Uint(), buf)
	}
	if kind == reflect.Int || kind == reflect.Int16 || kind == reflect.Int32 || kind == reflect.Int64 {
		return WriteVariableLengthInt(uint64(reflect.ValueOf(i).Int()), buf)
	}
	if kind == reflect.String {
		return WriteString(i.(string), buf)
	}
	if kind == reflect.Array || kind == reflect.Slice {
		return WriteArray(i, buf)
	}

	if kind == reflect.Struct {
		for j := 0; j < t.NumField(); j++ {
			err := Serialize(reflect.ValueOf(i).Field(j).Interface(), buf)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
