package node

import (
	"bytes"
	"encoding/gob"
)

// EncodeToByte gets an object and returns bytes
func EncodeToByte(o interface{}) []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(o)
	if err != nil {
		panic(err)
	}
	return buffer.Bytes()
}

// DecodeFromByte gets byte array and decodes to struct
func DecodeFromByte(data []byte, o interface{}) {
	var buffer bytes.Buffer
	buffer.Write(data)
	dec := gob.NewDecoder(&buffer)
	err := dec.Decode(o)
	if err != nil {
		panic(err)
	}
}
