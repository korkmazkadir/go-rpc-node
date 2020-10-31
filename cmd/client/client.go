package main

import (
	"log"
	"net/rpc"
	"time"

	"../../node"
)

func main() {

	client, err := rpc.Dial("tcp", "127.0.0.1:54193")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	message := node.Message{Tag: "ABC", Payload: "Hello"}

	for {

		log.Println("sending a message")
		var response node.Response
		client.Call("GossipNode.Send", message, &response)

		time.Sleep(1 * time.Second)

	}

}
