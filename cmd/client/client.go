package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"

	"../../node"
)

func main() {

	serverAddress := flag.String("server", "", "server address")
	flag.Parse()

	if *serverAddress == "" {
		fmt.Println("you should provide a server address using server flag")
	}

	client, err := rpc.Dial("tcp", *serverAddress)

	if err != nil {
		log.Fatal("dialing:", err)
	}

	var index byte

	for {

		log.Println("sending a message")
		message := node.Message{Layer: node.APPLICATION, Tag: "ABC", Payload: []byte{index}}

		var response node.Response
		client.Call("GossipNode.Send", message, &response)

		//time.Sleep(1 * time.Millisecond)
		index++

	}

}
