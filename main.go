package main

import (
	"context"
	"net"
	"os"
	"time"
	"fmt"
	"strings"

	"git.sr.ht/~spc/go-log"
	"github.com/pelletier/go-toml"
	pb "github.com/redhatinsights/yggdrasil/protocol"
	"google.golang.org/grpc"
)

var yggdDispatchSocketAddr string
var yggdHandler string

func main() {
	var ok bool

	configFile, ok := os.LookupEnv("CONFIG_FILE")
	if ok {
		config, err := toml.LoadFile(configFile)
		if err != nil {
			fmt.Errorf("cannot load config: %w", err)
			return
		}

		for _, value := range config.GetArray("env").([]string) {
			split := strings.Split(value, "=")
			os.Setenv(split[0], split[1])
		}

		os.Setenv("FORWARDER_HANDLER", strings.TrimSuffix(filepath.Base(configFile), filepath.Ext(configFile)))
	} else {
		log.Fatal("CONFIG_FILE not set")
	}

	// Get initialization values from the environment.
	yggdDispatchSocketAddr, ok = os.LookupEnv("YGG_SOCKET_ADDR")
	if !ok {
		log.Fatal("Missing YGG_SOCKET_ADDR environment variable")
	}

	yggdHandler, ok = os.LookupEnv("FORWARDER_HANDLER")
	if !ok {
		log.Fatal("Missing FORWARDER_HANDLER environment variable")
	}

	postUrl, ok := os.LookupEnv("FORWARDER_URL")
	if !ok {
		log.Fatal("Missing FORWARDER_URL environment variable")
	}

	postUser, ok := os.LookupEnv("FORWARDER_USER")
	if !ok {
		log.Fatal("Missing FORWARDER_USER environment variable")
	}

	postPassword, ok := os.LookupEnv("FORWARDER_PASSWORD")
	if !ok {
		log.Fatal("Missing FORWARDER_PASSWORD environment variable")
	}

	// Dial the dispatcher on its well-known address.
	conn, err := grpc.Dial(yggdDispatchSocketAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Create a dispatcher client
	c := pb.NewDispatcherClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Register as a handler as defined in the ENV variable.
	r, err := c.Register(ctx, &pb.RegistrationRequest{Handler: yggdHandler, Pid: int64(os.Getpid())})
	if err != nil {
		log.Fatal(err)
	}
	if !r.GetRegistered() {
		log.Fatalf("handler registration failed: %v", err)
	}

	// Listen on the provided socket address.
	l, err := net.Listen("unix", r.GetAddress())
	if err != nil {
		log.Fatal(err)
	}

	// Register as a Worker service with gRPC and start accepting connections.
	s := grpc.NewServer()
	pb.RegisterWorkerServer(s, &forwarderServer{Url: postUrl, Username: postUser, Password: postPassword})
	if err := s.Serve(l); err != nil {
		log.Fatal(err)
	}

	log.Infof("Successfully registered to the server")
}
