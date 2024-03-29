package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os/exec"
	"time"

	"encoding/json"

	"cloud.google.com/go/datastore"
	ais "github.com/andmarios/aislib"
	"golang.org/x/net/context"
	"google.golang.org/grpc/grpclog"
)

// Position is a datastore representation of a position report.
type Position struct {
	Timestamp      time.Time          `datastore:"Timestamp"`
	PositionReport ais.PositionReport `datastore:"PositionReport,flatten"`
}

type positionRequest struct {
	mmsi            uint32
	responseChannel chan []Position
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

func cachePositions(positionUpdates chan Position, positionRequests chan positionRequest) {
	positionCache := make(map[uint32]Position)
	maintenanceChan := time.NewTicker(time.Duration(20 * time.Second)).C
	for {
		select {
		case p := <-positionUpdates:
			positionCache[p.PositionReport.MMSI] = p
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
		case <-maintenanceChan:
			// Loop over the cache and delete things older than a minute.
			toDelete := []uint32{}
			now := time.Now()
			for mmsi, pos := range positionCache {
				if now.Sub(pos.Timestamp) > (10 * time.Minute) {
					toDelete = append(toDelete, mmsi)
				}
			}

			for _, mmsi := range toDelete {
				delete(positionCache, mmsi)
			}
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
	grpclog.SetLogger(log.New(ioutil.Discard, "", 0))
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
		k := datastore.IncompleteKey("PositionReport", nil)
		k.Namespace = "dev"
		datastorePosition := Position{
			Timestamp:      time.Now().UTC(),
			PositionReport: position,
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
