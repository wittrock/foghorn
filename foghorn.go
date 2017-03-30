package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"

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

func decodeAISMessages(aisByteStream chan string, payloads chan ais.Message) { // TODO(wittrock)
	send := make(chan string, 1024*8)
	receive := make(chan ais.Message, 1024*8)
	failed := make(chan ais.FailedSentence, 1024*8)

	done := make(chan bool)

	go ais.Router(send, receive, failed)

	var message ais.Message
	var problematic ais.FailedSentence
	for {
		select {
		case message = <-receive:
			switch message.Type {
			case 1, 2, 3:
				t, _ := ais.DecodeClassAPositionReport(message.Payload)
				fmt.Println(t)
			case 4:
				t, _ := ais.DecodeBaseStationReport(message.Payload)
				fmt.Println(t)
			case 5:
				t, _ := ais.DecodeStaticVoyageData(message.Payload)
				fmt.Println(t)
			case 8:
				t, _ := ais.DecodeBinaryBroadcast(message.Payload)
				fmt.Println(t)
			case 18:
				t, _ := ais.DecodeClassBPositionReport(message.Payload)
				fmt.Println(t)
			case 255:
				done <- true
			default:
				fmt.Printf("=== Message Type %2d ===\n", message.Type)
				fmt.Printf(" Unsupported type \n\n")
			}
		case problematic = <-failed:
			log.Println(problematic)
		}
	}
}

func main() {
	cmd := exec.Command("/home/jwittrock/src/rtl-ais/rtl_ais")
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Running rtl_ais as pid %d\n", cmd.Process.Pid)

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
