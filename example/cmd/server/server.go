package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"../../../node"
	"../../application"
)

func main() {

	var peerAddresses []string

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		log.Println("data is being piped to stdin")

		peerAddresses = getPeerAddressesFromFile(os.Stdin)
	} else {
		log.Println("stdin is from a terminal")

		peerAddressesFlag := flag.String("peers", "", "semicolon seperated list peer addresses")
		flag.Parse()
		peerAddresses = getPeerAddressesFromFlag(peerAddressesFlag)
	}

	app := application.NewSimpleApp(1)
	n := node.NewGossipNode(app)
	app.Start()
	address, err := n.Start()
	if err != nil {
		panic(err)
	}

	for _, address := range peerAddresses {
		connectToPeer(address, n)
	}

	log.Println("node started ", address)

	// Outputs the address of the server
	fmt.Println(address)

	n.Wait()
}

func getPeerAddressesFromFile(file *os.File) []string {
	addresses := make([]string, 0)

	inputBytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	inputString := string(inputBytes)

	tokens := strings.Split(inputString, "\n")
	for _, token := range tokens {
		if token != "" {
			addresses = append(addresses, token)
		}
	}

	return addresses
}

func getPeerAddressesFromFlag(peerAddresses *string) []string {

	log.Println(*peerAddresses)
	if *peerAddresses != "" {
		addresses := strings.Split(*peerAddresses, ";")
		return addresses
	}

	return nil
}

func connectToPeer(peerAddress string, n *node.GossipNode) {

	peer, err := node.NewRemoteNode(peerAddress)
	if err != nil {
		panic(err)
	}

	err = n.AddPeer(peer)
	if err != nil {
		panic(err)
	}

}
