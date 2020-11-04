package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"time"

	"../../../node"
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
		message := node.NewMessage("ABC", []byte{index})

		var response node.Response
		client.Call("GossipNode.Send", message, &response)

		time.Sleep(100 * time.Millisecond)
		index++

	}

}
