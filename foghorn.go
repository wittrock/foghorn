package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"

	"encoding/json"

	datastore "cloud.google.com/go/datastore"
	ais "github.com/andmarios/aislib"
	"golang.org/x/net/context"
)

func readUDPStream(pc net.PacketConn, output chan string) {
	buffer := make([]byte, 4096)

	log.Println("Reading from udp stream....")
	for {
		nBytes, _, err := pc.ReadFrom(buffer)
		log.Printf("got message: %s", string(buffer[:nBytes]))
		if err != nil {
			log.Fatal(err)
		}

		output <- string(buffer[:nBytes-2]) // remove the CRLF
	}
}

func decodeAISMessages(aisByteStream chan string, positions chan ais.PositionReport) { // TODO(wittrock)
	receive := make(chan ais.Message, 1024*8)
	failed := make(chan ais.FailedSentence, 1024*8)

	done := make(chan bool)

	go ais.Router(aisByteStream, receive, failed)

	var message ais.Message
	var problematic ais.FailedSentence
	for {
		select {
		case message = <-receive:
			switch message.Type {
			case 1, 2, 3:
				t, _ := ais.DecodeClassAPositionReport(message.Payload)
				positions <- t.PositionReport
			case 4:
				_, _ = ais.DecodeBaseStationReport(message.Payload)
			case 5:
				t, _ := ais.DecodeStaticVoyageData(message.Payload)
				fmt.Println(t)
			case 8:
				_, _ = ais.DecodeBinaryBroadcast(message.Payload)
			case 18:
				t, _ := ais.DecodeClassBPositionReport(message.Payload)
				positions <- t.PositionReport
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

// Position is a datastore representation of a position report.
type Position struct {
	Timestamp      time.Time
	PositionReport string
}

func main() {
	datastoreContext := context.Background()
	datastoreClient, err := datastore.NewClient(datastoreContext, "foghorn-163114")

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
	positions := make(chan ais.PositionReport, 4096)
	go readUDPStream(pc, incomingAISChannel)
	go decodeAISMessages(incomingAISChannel, positions)

	for {
		position := <-positions
		json, _ := json.Marshal(position)
		log.Printf("Got position: %s\n", string(json))
		k := datastore.NewIncompleteKey(datastoreContext, "PositionReport", nil)
		datastorePosition := Position{
			Timestamp:      time.Now().UTC(),
			PositionReport: string(json),
		}

		datastoreClient.Put(datastoreContext, k, datastorePosition)
	}
}
