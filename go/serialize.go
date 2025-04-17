package value

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

func writeInt64(v int64, buf *bytes.Buffer) error {
	return binary.Write(buf, binary.LittleEndian, v)
}

func writeInt32(v int32, buf *bytes.Buffer) error {
	return binary.Write(buf, binary.LittleEndian, v)
}

func writeString(v string, buf *bytes.Buffer) error {
	writeInt32(int32(len(v)), buf)
	strBytes := []byte(v)
	return binary.Write(buf, binary.LittleEndian, strBytes)
}

func serializeInt64(v int64, buf *bytes.Buffer) error {
	binary.Write(buf, binary.LittleEndian, int8(Int64))
	return writeInt64(v, buf)
}

func serializeBool(v bool, buf *bytes.Buffer) error {
	binary.Write(buf, binary.LittleEndian, int8(Boolean))
	b := func() int8 {
		if v {
			return int8(1)
		}
		return int8(0)
	}()
	return binary.Write(buf, binary.LittleEndian, b)
}

func serializeString(v string, buf *bytes.Buffer) error {
	binary.Write(buf, binary.LittleEndian, int8(String))
	return writeString(v, buf)
}

func serializeList(v []interface{}, buf *bytes.Buffer) error {
	val := reflect.ValueOf(v)
	binary.Write(buf, binary.LittleEndian, int8(List))
	writeInt64(int64(val.Len()), buf)
	for i := 0; i < val.Len(); i++ {
		serializeInterface(val.Index(i).Interface(), buf)
	}
	return nil
}

func serializeMap(m map[string]interface{}, buf *bytes.Buffer) error {
	binary.Write(buf, binary.LittleEndian, int8(Dictionnary))
	writeInt64(int64(len(m)), buf)
	for k, v := range m {
		writeString(k, buf)
		serializeInterface(v, buf)
	}
	return nil
}

func serializeInterface(v interface{}, buf *bytes.Buffer) {
	t := reflect.TypeOf(v)
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		serializeList(v.([]interface{}), buf)
	case reflect.Bool:
		serializeBool(v.(bool), buf)
	case reflect.Int64:
		serializeInt64(v.(int64), buf)
	case reflect.String:
		serializeString(v.(string), buf)
	case reflect.Map:
		serializeMap(v.(map[string]interface{}), buf)
	default:
		panic(fmt.Sprintf("Type %v not supported", t.Kind()))
	}
}
