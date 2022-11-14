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
	requestTimestamp	int64
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
		currentlyUsing: false,
		currentlyRequesting: false,
	}

	// Listen to port
	list, err := net.Listen("tcp", fmt.Sprintf(":%v", *ownPort))
	if err != nil {
		log.Fatalf("Failed to listen on port: %v", err)
	}

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

		// mark myself as currently requesting access
		p.currentlyRequesting = true
		p.requestTimestamp = p.timestamp
		log.Printf("%v is now currently requesting, and will say NO to incoming requests", p.port)

		// request each other peer for access to the restricted function
		for peerId, peer := range p.clients {
			log.Printf("L(%v): %v is requesting %v\n", p.timestamp, p.port, peerId)
			reply, err := peer.RequestAccess(p.ctx, &exclusion.Request{
				Port:      p.port,
				Timestamp: p.timestamp,
			})
			

			if err != nil {
				log.Fatalf("L(%v): FATAL ERROR: %s", p.timestamp, err.Error())
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
	// Pretend restricting function
	log.Printf("RESTRICTED ACCESS BY %v\n", id)
	for i := 0; i < 5; i++ {
		log.Print(".")
		time.Sleep(time.Second)
	}
	log.Println()
	log.Printf("Access by %v complete\n\n", id)
}

func (p *peer) RequestAccess(ctx context.Context, req *exclusion.Request) (*exclusion.Reply, error) {
	log.Printf("L(%v): %v requests %v access to critical section\n", p.timestamp, req.Port, p.port)

	// if p is using, wait until done
	for p.currentlyUsing {
		time.Sleep(200 * time.Millisecond) // Making sure it doesn't use 100% CPU
	} // Blocking

	// if p is requesting, compare timestamps:
	// Lower timestamp is granted access
	// If timestamp is equal, lower portnumber gets access (tiebreaker)

		for p.currentlyRequesting {
			if req.Timestamp < p.requestTimestamp {
				log.Printf("L(%v): %v has an earlier timestamp for their request, and is therefore granted access\n", p.timestamp, req.Port,)
			}
			if req.Timestamp == p.requestTimestamp {
				if req.Port < p.port {
					log.Printf("L(%v): %v has the same timestamp as %v but a lower portnumber\n", p.timestamp, req.Port, p.port)
					break
				}
			} else {
				time.Sleep(200 * time.Millisecond)
				// Blocking requester from gaining access, untill this peer is done.
			}
		}

	// req wants access to critical section
	p.timestamp = Max(req.Timestamp, p.timestamp) + 1

	// If the above does not hold, wait until it does, otherwise send reply.
	log.Printf("L(%v): %v gains access to critical section by %v\n", p.timestamp, req.Port, p.port)
	p.timestamp++
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
