package value

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicDeserialization(t *testing.T) {
	{
		b := []byte{1, 1}
		var buf bytes.Buffer
		buf.Write(b)
		v, err := DeserializeInterface(&buf)
		assert.Equal(t, nil, err)
		assert.Equal(t, true, v)
	}
	{
		b := []byte{2, 43, 0, 0, 0, 0, 0, 0, 0}
		var buf bytes.Buffer
		buf.Write(b)
		v, err := DeserializeInterface(&buf)
		assert.Equal(t, nil, err)
		assert.Equal(t, int64(43), v)
	}
	{
		b := []byte{3, 4, 0, 0, 0, 'f', 'd', 's', 'f'}
		var buf bytes.Buffer
		buf.Write(b)
		s, err := DeserializeInterface(&buf)
		assert.Equal(t, nil, err)
		assert.Equal(t, "fdsf", s)
	}
}

func TestListDeserialization(t *testing.T) {
	b := []byte{5, 2, 0, 0, 0, 0, 0, 0, 0, 1, 1, 2, 43, 0, 0, 0, 0, 0, 0, 0}
	var buf bytes.Buffer
	buf.Write(b)
	l, err := DeserializeInterface(&buf)
	list := l.([]interface{})
	assert.Equal(t, nil, err)
	assert.Equal(t, len(list), 2)
	assert.Equal(t, true, list[0])
	assert.Equal(t, int64(43), list[1])
}

func TestMapDeserialization(t *testing.T) {
	b := []byte{4, 2, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0,
		'f', 'o', 'o', 2, 44, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 'b', 'a',
		'r', 3, 6, 0, 0, 0, 'p', 'e', 'w', 'p', 'e', 'w'}
	var buf bytes.Buffer
	buf.Write(b)
	d, err := DeserializeInterface(&buf)
	dic := d.(map[string]interface{})
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(dic))
	assert.Equal(t, "pewpew", dic["bar"])
	assert.Equal(t, int64(44), dic["foo"])
}

func TestAdvancedDeserialization(t *testing.T) {
	b := []byte{4, 2, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 108, 100, 4, 1, 0, 0, 0, 0, 0, 0, 0, 26, 0, 0,
		0, 47, 119, 101, 98, 47, 55, 105, 55, 74, 99, 49, 48, 74, 88, 49, 54, 70, 107, 90, 117, 55, 71,
		90, 70, 120, 112, 4, 5, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 108, 110, 3, 40, 0, 0, 0, 35, 102, 102,
		102, 102, 48, 48, 102, 102, 83, 97, 109, 112, 108, 101, 58, 32, 35, 102, 102, 102, 102, 102, 102,
		102, 102, 65, 100, 118, 97, 110, 99, 101, 100, 32, 108, 101, 118, 101, 108, 2, 0, 0, 0, 109, 100,
		2, 2, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 112, 99, 2, 5, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 115, 115, 2,
		119, 16, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 116, 112, 2, 64, 6, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 114, 101,
		112, 5, 7, 0, 0, 0, 0, 0, 0, 0, 5, 2, 0, 0, 0, 0, 0, 0, 0, 3, 12, 0, 0, 0, 121, 109, 72, 110, 74, 81,
		67, 102, 80, 113, 103, 97, 2, 2, 0, 0, 0, 0, 0, 0, 0, 5, 2, 0, 0, 0, 0, 0, 0, 0, 3, 12, 0, 0, 0, 51, 81, 100,
		81, 120, 72, 51, 52, 82, 56, 83, 110, 2, 2, 0, 0, 0, 0, 0, 0, 0, 5, 2, 0, 0, 0, 0, 0, 0, 0, 3, 12, 0, 0, 0, 71, 109, 57,
		99, 68, 68, 79, 107, 65, 117, 72, 98, 2, 2, 0, 0, 0, 0, 0, 0, 0, 5, 2, 0, 0, 0, 0, 0, 0, 0, 3, 12, 0, 0, 0, 50, 86, 74,
		121, 56, 105, 51, 55, 116, 120, 45, 117, 2, 5, 0, 0, 0, 0, 0, 0, 0, 5, 2, 0, 0, 0, 0, 0, 0, 0, 3, 12, 0, 0, 0, 104, 104,
		89, 67, 122, 52, 111, 109, 49, 70, 71, 69, 2, 20, 0, 0, 0, 0, 0, 0, 0, 5, 2, 0, 0, 0, 0, 0, 0, 0, 3, 12, 0, 0, 0, 76, 105,
		50, 85, 114, 71, 67, 66, 117, 70, 90, 114, 2, 1, 0, 0, 0, 0, 0, 0, 0, 5, 2, 0, 0, 0, 0, 0, 0, 0, 3, 12, 0, 0, 0, 101, 81, 88,
		54, 90, 105, 67, 80, 70, 103, 100, 113, 2, 1, 0, 0, 0, 0, 0, 0, 0,
	}
	var buf bytes.Buffer
	buf.Write(b)
	v, err := DeserializeInterface(&buf)
	assert.Equal(t, nil, err)
	fmt.Printf("v= %v\n", v)

	jsonData, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(jsonData))

}
