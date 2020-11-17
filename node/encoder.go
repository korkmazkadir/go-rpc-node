package node

import (
	"bytes"
	"encoding/gob"
	"sync"
)

// PayloadCodec implements a payload codec which uses same buffer
type PayloadCodec struct {
	buffer *bytes.Buffer
	mutex  *sync.Mutex
}

// NewPayloadCodec creates a new PayloadCodec and initilizes it
func NewPayloadCodec() *PayloadCodec {
	codec := new(PayloadCodec)
	codec.buffer = new(bytes.Buffer)
	codec.mutex = &sync.Mutex{}
	return codec
}

// EncodeToByte encodes object to byte
func (pc *PayloadCodec) EncodeToByte(o interface{}) []byte {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	pc.buffer.Reset()

	enc := gob.NewEncoder(pc.buffer)
	err := enc.Encode(o)
	if err != nil {
		panic(err)
	}
	return pc.buffer.Bytes()
}

// DecodeFromByte decodes bytes to object
func (pc *PayloadCodec) DecodeFromByte(data []byte, o interface{}) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	pc.buffer.Reset()

	pc.buffer.Write(data)
	dec := gob.NewDecoder(pc.buffer)
	err := dec.Decode(o)
	if err != nil {
		panic(err)
	}
}

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
