package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"time"

	"encoding/json"

	datastore "cloud.google.com/go/datastore"
	ais "github.com/andmarios/aislib"
	"golang.org/x/net/context"
)

// Position is a datastore representation of a position report.
type Position struct {
	Timestamp      time.Time
	PositionReport string `datastore:",noindex"`
	MMSI           int32
	Lat            float64
	Lng            float64
}

var positionMap map[uint32]Position

func readUDPStream(pc net.PacketConn, output chan string) {
	buffer := make([]byte, 4096)

	log.Println("Reading from udp stream....")
	for {
		nBytes, _, err := pc.ReadFrom(buffer)
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
				_, _ = ais.DecodeStaticVoyageData(message.Payload)
			case 8:
				_, _ = ais.DecodeBinaryBroadcast(message.Payload)
			case 18:
				t, _ := ais.DecodeClassBPositionReport(message.Payload)
				positions <- t.PositionReport
			case 255:
				done <- true
			default:
			}
		case problematic = <-failed:
			log.Println(problematic)
		}
	}
}

type positionRequest struct {
	mmsi            int32
	responseChannel chan []Position
}

func cachePositions(positionUpdates chan Position, positionRequests chan positionRequest) {
	positionCache := make(map[int32]Position)
	for {
		select {
		case p := <-positionUpdates:
			log.Printf("Setting position for mmsi %d\n", p.MMSI)
			positionCache[p.MMSI] = p
		case r := <-positionRequests:
			if r.mmsi != 0 {
				// Only send back a single value.
				r.responseChannel <- []Position{positionCache[r.mmsi]}
				continue
			}

			// Dump the whole cache.
			positionIndex := 0
			response := make([]Position, len(positionCache))
			for _, pos := range positionCache {
				response[positionIndex] = pos
				positionIndex++
			}
			r.responseChannel <- response
		}

	}
}

var positionRequests chan positionRequest

func positionsHandler(w http.ResponseWriter, r *http.Request) {
	responseChan := make(chan []Position)

	request := positionRequest{
		mmsi:            0,
		responseChannel: responseChan,
	}

	positionRequests <- request

	response := <-responseChan

	json, _ := json.Marshal(response)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, "%s\n", json)
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

	positionUpdates := make(chan Position, 512)
	positionRequests = make(chan positionRequest, 512)
	go cachePositions(positionUpdates, positionRequests)

	http.HandleFunc("/positions", positionsHandler)
	go http.ListenAndServe(":8000", nil)

	go readUDPStream(pc, incomingAISChannel)
	go decodeAISMessages(incomingAISChannel, positions)

	for {
		position := <-positions
		json, _ := json.Marshal(position)
		k := datastore.IncompleteKey("PositionReport", nil)
		k.Namespace = "dev"
		datastorePosition := Position{
			Timestamp:      time.Now().UTC(),
			PositionReport: string(json),
			MMSI:           int32(position.MMSI),
			Lat:            position.Lat,
			Lng:            position.Lon,
		}

		// Save to cache
		positionUpdates <- datastorePosition

		// Output to datastore
		_ = datastoreClient
		// _, err := datastoreClient.Put(datastoreContext, k, &datastorePosition)
		// if err != nil {
		// 	log.Printf("Could not write to datastore: %v\n", err)
		// }
	}
}
