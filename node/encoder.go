package node

import (
	"bytes"
	"encoding/gob"
)

func EncodeToByte(o interface{}) []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(o)
	if err != nil {
		panic(err)
	}
	return buffer.Bytes()
}

func DecodeFromByte(data []byte, o interface{}) {
	var buffer bytes.Buffer
	buffer.Write(data)
	dec := gob.NewDecoder(&buffer)
	err := dec.Decode(o)
	if err != nil {
		panic(err)
	}
}
