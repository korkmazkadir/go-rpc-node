package node

import (
	"encoding/base32"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const folder = "/tmp/message-inventory"

// FileBackedMessageInventory defines a message inventory
type FileBackedMessageInventory struct {
	pid int
}

// NewFileBackedMessageInventory creates a message inventory
func NewFileBackedMessageInventory() FileBackedMessageInventory {
	inventory := FileBackedMessageInventory{pid: os.Getpid()}
	os.MkdirAll(filepath.Join(".", folder, fmt.Sprint(inventory.pid)), os.ModePerm)
	return inventory
}

// Put writes message to disc
func (f FileBackedMessageInventory) Put(message *Message) string {

	hash := message.Base32EncodedHash()
	file, err := os.OpenFile(fmt.Sprintf("./%s/%d/%s", folder, f.pid, hash), os.O_CREATE|os.O_WRONLY, 0644)
	//if err == os.ErrExist {
	if err != nil {
		log.Println(err)
		return hash
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")

	err = encoder.Encode(message)
	if err != nil {
		panic(err)
	}

	err = file.Close()
	if err != nil {
		panic(err)
	}

	return hash
}

// Get read message from the disc
func (f FileBackedMessageInventory) Get(hash string) (*Message, error) {

	key := base32.StdEncoding.EncodeToString([]byte(hash))

	file, err := os.OpenFile(fmt.Sprintf("./%s/%d/%s", folder, f.pid, key), os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(file)
	message := new(Message)
	decoder.Decode(message)

	return message, nil
}
