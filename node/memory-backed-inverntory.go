package node

import (
	"encoding/json"

	simplekeyvalue "github.com/korkmazkadir/simple-key-value-store"
)

type Inventory struct {
	client *simplekeyvalue.Client
}

func NewInventory() *Inventory {

	client, err := simplekeyvalue.NewClient()
	if err != nil {
		panic(err)
	}

	inventory := &Inventory{client: client}

	return inventory
}

// Put writes message to disc
func (inv *Inventory) Put(message *Message) string {

	hash := string(message.hash())

	b, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}

	inv.client.Put(hash, b)

	return hash
}

// Get read message from the disc
func (inv *Inventory) Get(hash string) (*Message, error) {

	b := inv.client.Get(hash)

	message := new(Message)
	err := json.Unmarshal(b, message)

	return message, err
}
