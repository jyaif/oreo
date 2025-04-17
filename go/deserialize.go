package value

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

func readType(buf *bytes.Buffer) (ValueType, error) {
	var t uint8
	err := binary.Read(buf, binary.LittleEndian, &t)
	return ValueType(t), err
}

func readBoolean(buf *bytes.Buffer) (bool, error) {
	var v int8
	err := binary.Read(buf, binary.LittleEndian, &v)
	if err != nil {
		return false, err
	}
	if v == 0 {
		return false, nil
	}
	return true, nil
}

func readInt32(buf *bytes.Buffer) (int32, error) {
	var v int32
	err := binary.Read(buf, binary.LittleEndian, &v)
	return v, err
}

func readInt64(buf *bytes.Buffer) (int64, error) {
	var v int64
	err := binary.Read(buf, binary.LittleEndian, &v)
	return v, err
}

func readString(buf *bytes.Buffer) (string, error) {
	len, lenErr := readInt32(buf)
	if lenErr != nil {
		return "", lenErr
	}
	n := buf.Cap() - buf.Len()
	if int32(n) < len {
		return "", fmt.Errorf("not enough bytes left: need %d, have %d", len, n)
	}
	var s = string(buf.Next(int(len)))
	return s, nil
}

func readList(buf *bytes.Buffer) ([]interface{}, error) {
	len, lenErr := readInt64(buf)
	if lenErr != nil {
		return nil, lenErr
	}
	var list []interface{}
	for i := 0; i < int(len); i++ {
		v, err := DeserializeInterface(buf)
		if err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, nil
}

func readDictionnary(buf *bytes.Buffer) (map[string]interface{}, error) {
	len, lenErr := readInt64(buf)
	if lenErr != nil {
		return nil, lenErr
	}
	dic := make(map[string]interface{})
	for i := 0; i < int(len); i++ {
		key, keyErr := readString(buf)
		if keyErr != nil {
			return nil, keyErr
		}
		val, valErr := DeserializeInterface(buf)
		if valErr != nil {
			return nil, valErr
		}
		dic[key] = val
	}
	return dic, nil
}

func DeserializeInterface(buf *bytes.Buffer) (interface{}, error) {
	t, typeErr := readType(buf)
	if typeErr != nil {
		return nil, typeErr
	}
	switch t {
	case Boolean:
		b, err := readBoolean(buf)
		if err != nil {
			return nil, err
		}
		return b, nil
	case Int64:
		i, err := readInt64(buf)
		if err != nil {
			return nil, err
		}
		return i, nil
	case String:
		s, err := readString(buf)
		if err != nil {
			return nil, err
		}
		return s, nil
	case List:
		l, err := readList(buf)
		if err != nil {
			return nil, err
		}
		return l, nil
	case Dictionnary:
		d, err := readDictionnary(buf)
		if err != nil {
			return nil, err
		}
		return d, nil

	default:
		return nil, errors.New("Unhandled type")
	}
}
