package application

import (
	"log"

	"../node"
)

type SimpleApp struct {
	id                int
	networkReadySig   chan struct{}
	incommingMessages chan node.Message
	outgoingMessages  chan node.Message

	processedMessageDigests []string
}

const MaxSizeProcessedMessageDigest = 100

func NewSimpleApp(appID int) *SimpleApp {

	app := SimpleApp{
		id:                      appID,
		networkReadySig:         make(chan struct{}, 1),
		incommingMessages:       make(chan node.Message, 10),
		outgoingMessages:        make(chan node.Message, 10),
		processedMessageDigests: make([]string, 0, MaxSizeProcessedMessageDigest),
	}

	return &app
}

func (app *SimpleApp) isMessageProcessedBefore(message node.Message) bool {

	messageHash := message.Hash()
	for _, digest := range app.processedMessageDigests {
		if digest == messageHash {
			return true
		}
	}

	return false
}

func (app *SimpleApp) addToProcessedMessages(message node.Message) {

	messageHash := message.Hash()
	if len(app.processedMessageDigests) == (MaxSizeProcessedMessageDigest - 1) {
		app.processedMessageDigests = app.processedMessageDigests[1:]
	}

	app.processedMessageDigests = append(app.processedMessageDigests, messageHash)

}

func (app *SimpleApp) mainLoop() {
	for {

		message := <-app.incommingMessages

		if app.isMessageProcessedBefore(message) == false {
			app.addToProcessedMessages(message)
			log.Printf("[SimpleApp-%d] handling message tag: %s, payload: %s \n", app.id, message.Tag, message.Payload)
			// Forwards the message
			message.Forward()
		}

	}
}

func (app *SimpleApp) HandleMessage(message node.Message) {
	app.incommingMessages <- message
}

func (app *SimpleApp) OutgoingMessageChannel() chan node.Message {
	return app.outgoingMessages
}

func (app *SimpleApp) SignalChannel() chan struct{} {
	return app.networkReadySig
}

func (app *SimpleApp) Start() {

	go func() {
		//waits for network signal
		<-app.networkReadySig

		log.Println("App started...")
		app.mainLoop()
	}()

}
