package main

import (
	"fmt"
	"log"
	"net"

	ais "github.com/andmarios/aislib"
)

func readUDPStream(pc net.PacketConn, output chan string) {
	buffer := make([]byte, 4096)
	for {
		nBytes, _, err := pc.ReadFrom(buffer)
		if err != nil {
			log.Fatal(err)
		}

		output <- string(buffer[:nBytes])
	}
}

func decodeAISMessages(aisByteStream chan string, aisMessages chan ais.Message) {
	// Implement
}

func main() {
	pc, err := net.ListenPacket("udp", ":10110")
	if err != nil {
		log.Fatal(err)
	}

	defer pc.Close()

	incomingAISChannel := make(chan string, 4096)
	decodedMessages := make(chan ais.Message, 8192)
	go readUDPStream(pc, incomingAISChannel)
	go decodeAISMessages(incomingAISChannel, decodedMessages)

	for {
		msg := <-decodedMessages
		fmt.Printf("Message: %v\n", msg)
	}
}
