package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"git.sr.ht/~spc/go-log"
	pb "github.com/redhatinsights/yggdrasil/protocol"
	"google.golang.org/grpc"
)

// posterServer implements the Worker gRPC service as defined by the yggdrasil
// gRPC protocol. It accepts Assignment messages, unmarshals the data into a
// string, and echoes the content back to the Dispatch service by calling the
// "Finish" method.
type posterServer struct {
	pb.UnimplementedWorkerServer
	Url      string
	Username string
	Password string
}

type httpMessage struct {
	ResponseTo string
	Metadata   map[string]string
	Content    []byte
	Directive  string
}

// Send implements the "Send" method of the Worker gRPC service.
func (s *posterServer) Send(ctx context.Context, d *pb.Data) (*pb.Receipt, error) {
	go func() {
		log.Tracef("received data: %#v", d)

		// Dial the Dispatcher and call "Finish"
		conn, err := grpc.Dial(yggdDispatchSocketAddr, grpc.WithInsecure())
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		// Create a data message to send back to the dispatcher.
		data := httpMessage{
			ResponseTo: d.GetMessageId(),
			Metadata:   d.GetMetadata(),
			Content:    d.GetContent(),
			Directive:  d.GetDirective(),
		}

		dataJson, error := json.Marshal(data)
		if error != nil {
			log.Fatal(error)
		}
		log.Infof("sending %v", dataJson)

		// Call http post
		request, _ := http.NewRequest("POST", s.Url, bytes.NewBuffer(dataJson))
		request.Header.Set("Content-Type", "application/json")
		request.SetBasicAuth(s.Username, s.Password)

		client := &http.Client{}
		response, error := client.Do(request)
		if error != nil {
			log.Fatal(error)
		}
		defer response.Body.Close()

		log.Tracef("response Status: %v", response.Status)
		log.Tracef("response Headers: %+v", response.Header)
		body, _ := ioutil.ReadAll(response.Body)
		log.Tracef("response Body: %v", string(body))

	}()

	// Respond to the start request that the work was accepted.
	return &pb.Receipt{}, nil
}
