package value

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicSerialization(t *testing.T) {
	{
		buf := new(bytes.Buffer)
		serializeBool(true, buf)
		serializeBool(false, buf)
		assert.Equal(t, []byte{1, 1, 1, 0}, buf.Bytes())
	}
	{
		buf := new(bytes.Buffer)
		v := int64(43)
		serializeInt64(v, buf)
		assert.Equal(t, []byte{2, 43, 0, 0, 0, 0, 0, 0, 0}, buf.Bytes())
	}
	{
		buf := new(bytes.Buffer)
		v := "pewpew"
		serializeString(v, buf)
		assert.Equal(t, []byte{3, 6, 0, 0, 0, 'p', 'e', 'w', 'p', 'e', 'w'}, buf.Bytes())
	}
}

func TestSerializeInterface(t *testing.T) {
	{
		buf := new(bytes.Buffer)
		serializeInterface(true, buf)
		serializeInterface(false, buf)
		assert.Equal(t, []byte{1, 1, 1, 0}, buf.Bytes())
	}
	{
		buf := new(bytes.Buffer)
		v := int64(43)
		serializeInterface(v, buf)
		assert.Equal(t, []byte{2, 43, 0, 0, 0, 0, 0, 0, 0}, buf.Bytes())
	}
	{
		buf := new(bytes.Buffer)
		v := "pewpew"
		serializeInterface(v, buf)
		assert.Equal(t, []byte{3, 6, 0, 0, 0, 'p', 'e', 'w', 'p', 'e', 'w'}, buf.Bytes())
	}
}

func TestListSerialization(t *testing.T) {
	// Empty list
	{
		buf := new(bytes.Buffer)
		var list []interface{}
		serializeInterface(list, buf)
		assert.Equal(t, []byte{5, 0, 0, 0, 0, 0, 0, 0, 0}, buf.Bytes())
	}
	// List with 2 elements
	{
		buf := new(bytes.Buffer)
		var list []interface{}
		list = append(list, true)
		list = append(list, int64(43))
		serializeInterface(list, buf)
		assert.Equal(t, []byte{5, 2, 0, 0, 0, 0, 0, 0, 0, 1, 1, 2, 43, 0, 0, 0, 0, 0, 0, 0}, buf.Bytes())
	}
}

func TestMapSerialization(t *testing.T) {
	// Empty map
	{
		buf := new(bytes.Buffer)
		var m map[string]interface{}
		serializeInterface(m, buf)
		assert.Equal(t, []byte{4, 0, 0, 0, 0, 0, 0, 0, 0}, buf.Bytes())
	}
	// Map with 2 elements
	{
		buf := new(bytes.Buffer)
		// var m map[string]interface{}{}
		m := map[string]interface{}{}
		// app
		m["foo"] = int64(44)
		m["bar"] = "pewpew"
		serializeInterface(m, buf)
		assert.Equal(t, []byte{4, 2, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0,
			'f', 'o', 'o', 2, 44, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 'b', 'a',
			'r', 3, 6, 0, 0, 0, 'p', 'e', 'w', 'p', 'e', 'w'}, buf.Bytes())
	}
}
