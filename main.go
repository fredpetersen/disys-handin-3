package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	exclusion "github.com/fredpetersen/disys-handin-4/grpc"
	"google.golang.org/grpc"
)

type peer struct {
	exclusion.UnimplementedExclusionServer

	id int
	clients map[int32]exclusion.ExclusionClient
	ctx context.Context
}
var ownPort = flag.Int("port", 5000, "the port for the client")


func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background());
	
	p := &peer{
		id: *ownPort, // own Port???
		clients: make(map[int32]exclusion.ExclusionClient),
		ctx: ctx, 
	}

	grpcServer := grpc.NewServer()
	exclusion.RegisterExclusionServer(grpcServer, p)
	
	// exit gracefully when interrupting
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Exiting gracefully...")
		cancel()
		grpcServer.GracefulStop()
		os.Exit(0)
	}()

	go p.readInput()
	
	//Connect to other peers
	
}
func (p *peer) connect() {
	// Find first available port starting at 5000

	// ping all peers from 5000 to ownPort that p now exist

	// Make sure every peer adds clients correctly in p.clients
}

func (p *peer) readInput() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// read input (clicking enter in terminal)
		
		for peerId, peer := range p.clients {
			log.Printf("🤮🤮🤮🤮🤮🤮🤮🤮🤮🤮 requesting %v\n", peerId);
			reply, err := peer.RequestAccess(p.ctx, &exclusion.Request{
				Id: int32(p.id),
			})

			if err != nil {
				log.Printf("FATAL ERROR: %s", err.Error())
			}
			if reply.Granted {
				log.Printf("%v got access to the restricted function", p.id)
				restrictedFunc(p.id)
			}
		}
	}
}

func restrictedFunc(id int) {
	log.Printf("Getting accessed by %v\n", id)
	for i := 0; i<5;i++ {
		log.Print(".")
		time.Sleep(time.Second)
	}
	log.Println()
	log.Printf("Access by %v complete\n\n", id)
}