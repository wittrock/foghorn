package main

import (
	"github.com/admarios/aislib"
)

func readUDPStream(pc net.PacketConn, output <-chan byte) {
	buffer := make([]byte, 4096)
	for {
		nBytes, addr, err := pc.ReadFrom(buffer)
		output <- buffer[:nBytes]
	}
}

func decodeAISMessages(aisByteStream ->chan byte, aisMessages <-chan ais.Message) {
	// Implement
}

func main() {
	pc, err := net.ListenPacket("udp", ":10110")
	if err != nil {
		log.Fatal(err)
	}

	defer pc.Close()

	incomingAISChannel := make(chan byte, 4096)
	decodedMessages := make(chan ais.Message, 8192)
	go readUDPStream(pc, incomingAISChannel)
	go decodeAISMessages(incomingAISChannel, decodedMessages)

	for {
		msg := <-decodedMessages
		fmt.Printf("Message: %v\n", msg)
	}
}
