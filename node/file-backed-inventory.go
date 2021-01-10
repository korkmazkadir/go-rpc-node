package node

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const folder = "message-inventory"

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
	file, err := os.OpenFile(fmt.Sprintf("./%s/%d/%s", folder, f.pid, hash), os.O_EXCL|os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0644)
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

	file, err := os.OpenFile(fmt.Sprintf("./%s/%d/%s", folder, f.pid, hash), os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(file)
	message := new(Message)
	decoder.Decode(message)

	return message, nil
}
