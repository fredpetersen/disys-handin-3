package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	exclusion "github.com/fredpetersen/disys-handin-4/grpc"
	"google.golang.org/grpc"
)

type peer struct {
	exclusion.UnimplementedExclusionServer

	port                int64
	timestamp           int64
	currentlyRequesting bool
	currentlyUsing      bool
	clients             map[int64]exclusion.ExclusionClient
	ctx                 context.Context
}

var ownPort = flag.Int64("port", 5000, "the port for the client")

func main() {
	flag.Parse()
	fmt.Printf("Creating on port %v\n", *ownPort)

	// enable logging to a file
	f, err := os.OpenFile(fmt.Sprintf("peer_%v.log", *ownPort), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("error opening logfile: %v", err)
	}
	log.SetOutput(f)

	ctx, cancel := context.WithCancel(context.Background())

	p := &peer{
		port:      *ownPort,
		clients:   make(map[int64]exclusion.ExclusionClient),
		timestamp: 1,
		ctx:       ctx,
	}

	// Listen to port
	list, err := net.Listen("tcp", fmt.Sprintf(":%v", *ownPort))
	if err != nil {
		log.Fatalf("Failed to listen on port: %v", err)
	}

	// p.findAvailablePort()

	grpcServer := grpc.NewServer()
	exclusion.RegisterExclusionServer(grpcServer, p)

	go func() {
		err := grpcServer.Serve(list)
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Connect to peers

	for i := 0; i < 3; i++ {
		port := int64(5000 + i)

		if port == *ownPort {
			continue
		}

		var conn *grpc.ClientConn
		fmt.Printf("Trying to dial: %v\n", port)
		conn, err := grpc.Dial(fmt.Sprintf(":%v", port), grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}
		defer conn.Close()
		c := exclusion.NewExclusionClient(conn)
		p.clients[port] = c
	}

	// exit gracefully when interrupting
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("Exiting gracefully...")
		cancel()
		grpcServer.GracefulStop()
		os.Exit(0)
	}()

	go p.readInput()

	time.Sleep(time.Minute)
}

func (p *peer) readInput() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// read input (clicking enter in terminal)
		p.currentlyRequesting = true
		for peerId, peer := range p.clients {
			log.Printf("L(%v): %v is requesting %v\n", p.timestamp, p.port, peerId)
			reply, err := peer.RequestAccess(p.ctx, &exclusion.Request{
				Port:      p.port,
				Timestamp: p.timestamp, // TODO
			})

			if err != nil {
				log.Printf("L(%v): FATAL ERROR: %s", p.timestamp, err.Error())
				// Exit system as this is a fatal error
			}

			log.Printf("L(%v): %v got access to the restricted function by %v", p.timestamp, p.port, reply.Port)

		}

		p.timestamp++

		// mark myself as using
		p.currentlyUsing = true

		// use restricted function
		restrictedFunc(p.port)

		// release restricted function
		p.currentlyUsing = false
		p.currentlyRequesting = false
	}
}

func restrictedFunc(id int64) {
	log.Printf("RESTRICTED ACCESS BY %v\n", id)
	for i := 0; i < 5; i++ {
		log.Print(".")
		time.Sleep(time.Second)
	}
	log.Println()
	log.Printf("Access by %v complete\n\n", id)
}

func (p *peer) RequestAccess(ctx context.Context, req *exclusion.Request) (*exclusion.Reply, error) {
	// req wants access to critical section
	p.timestamp = Max(req.Timestamp, p.timestamp+1)
	log.Printf("L(%v): %v requests %v access to critical section\n", p.timestamp, req.Port, p.port)

	// if p is using, wait until done
	for p.currentlyUsing {
		time.Sleep(time.Second) // Making sure it doesn't use 100% CPU
	} // Blocking

	// if p is requesting, compare timestamps:
	// Lower timestamp is granted access
	// If timestamp is equal, lower portnumber gets access (tiebreaker)

	if p.currentlyRequesting {
		for req.Timestamp >= p.timestamp {
			if req.Timestamp == p.timestamp {
				if req.Port < p.port {
					log.Printf("L(%v): %v has the same timestamp as %v but a lower portnumber\n", p.timestamp, req.Port, p.port)
					break
				}
			}
		}
	}

	// If the above does not hold, wait until it does, otherwise send reply.
	log.Printf("L(%v): %v gains access to critical section by %v\n", p.timestamp, req.Port, p.port)
	return &exclusion.Reply{
		Port: p.port,
	}, nil
}

func Max(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}
