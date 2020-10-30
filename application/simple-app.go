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
}

func NewSimpleApp(appID int) SimpleApp {

	app := SimpleApp{
		id:                appID,
		networkReadySig:   make(chan struct{}, 1),
		incommingMessages: make(chan node.Message, 10),
		outgoingMessages:  make(chan node.Message, 10),
	}

	return app
}

func (app SimpleApp) mainLoop() {
	for {
		select {
		case message := <-app.incommingMessages:
			log.Printf("[SimpleApp-%d] handling message tag: %s, payload: %s \n", app.id, message.Tag, message.Payload)
			// Forwards the message
			app.outgoingMessages <- message
			//default:
			//	time.Sleep(5 * time.Second)
		}

	}
}

func (app SimpleApp) HandleMessage(message node.Message) {
	app.incommingMessages <- message
}

func (app SimpleApp) OutgoingMessageChannel() chan node.Message {
	return app.outgoingMessages
}

func (app SimpleApp) SignalChannel() chan struct{} {
	return app.networkReadySig
}

func (app SimpleApp) Start() {

	go func() {
		//waits for network signal
		<-app.networkReadySig

		log.Println("App started...")
		app.mainLoop()
	}()

}
