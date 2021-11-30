package main

import (
	"bufio"
	"context"
	"flag"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Akongstad/AuctionHouse/Auction"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

var client Auction.AuctionHouseClient

type gRPCClient struct {
	conn      *grpc.ClientConn
	wait      *sync.WaitGroup
	clock     Auction.LamportClock
	stream    Auction.AuctionHouse_OpenConnectionClient
	done      chan bool
	reconnect chan bool
	user      Auction.User
}

func translateReturn(nr int32) (msg string) {
	if nr == 1 {
		return "Success"
	} else if nr == 2 {
		return "Fail"
	} else {
		return "Exception"
	}
}

func (grpcclient *gRPCClient) ProcessRequests() {
	defer grpcclient.conn.Close()

	go grpcclient.process()
	for {
		select {
		case <-grpcclient.reconnect:
			if !grpcclient.waitUntilReady() {
				log.Fatal("failed to establish a connection within the defined timeout")
			}
			go grpcclient.process()
		case <-grpcclient.done:
			return
		}
	}
}

func (grpcclient *gRPCClient) process() {
	reqclient := grpcclient.GetStream()
	for {
		request, err := reqclient.Recv()
		if err == io.EOF {
			log.Printf("io error: %v", err)
			grpcclient.done <- true
			return
		} else if err != nil {
			log.Printf("Reconnecting...")
			grpcclient.reconnect <- true
			return

		} else {

			grpcclient.clock.SyncClocks(request.GetTimestamp())
			log.Printf("Auction House: %s: Has joined the auction(%d)", request.GetUser().GetName(), grpcclient.clock.GetTime())
		}
	}
}

func (grpcclient *gRPCClient) waitUntilReady() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	currentState := grpcclient.conn.GetState()
	stillConnecting := true

	for currentState != connectivity.Ready && stillConnecting {

		stillConnecting = grpcclient.conn.WaitForStateChange(ctx, currentState)
		currentState = grpcclient.conn.GetState()
		log.Printf("Attempting reconnection. State has changed to: %v", currentState)
	}

	if stillConnecting {
		log.Fatal("Connection attempt has timed out.")
		return false
	}

	return true
}

func (grpcclient *gRPCClient) GetStream() Auction.AuctionHouse_OpenConnectionClient {
	grpcclient.dialServer()
	connection, err := client.OpenConnection(context.Background(), &Auction.Connect{
		User:   &grpcclient.user,
		Active: true,
	})
	if err != nil {
		log.Fatalf("Failed to get stream: %v", err)
	}

	return connection
}

func (grpcclient *gRPCClient) dialServer() {
	if grpcclient.conn != nil {
		grpcclient.conn.Close()
	}
	var err error
	grpcclient.conn, err = grpc.Dial(":5001", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	client = Auction.NewAuctionHouseClient(grpcclient.conn)
}

func main() {

	//init channel
	done := make(chan int)

	//Get User info
	clientName := flag.String("U", "Anonymous", "ClientName")
	flag.Parse()
	userId := rand.Intn(999)
	clientUser := &Auction.User{
		Id:   int64(userId),
		Name: *clientName,
	}
	c := &gRPCClient{
		wait:      &sync.WaitGroup{},
		clock:     Auction.LamportClock{},
		user:      *clientUser,
		reconnect: make(chan bool),
		done:      make(chan bool),
	}

	log.Println(*clientName, "Connecting")

	// Set up a connection to the server.
	c.dialServer()
	go c.ProcessRequests()

	c.wait.Add(1)

	go func() {
		defer c.wait.Done()

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {

			inputArray := strings.Fields(scanner.Text())

			input := strings.ToLower(strings.TrimSpace(inputArray[0]))

			if input == "exit" {
				c.clock.Tick()
				client.CloseConnection(context.Background(), &Auction.Message{
					User:      clientUser,
					Timestamp: c.clock.GetTime()})
				os.Exit(1)
			}

			if input != "" {
				if input == "result" {
					c.clock.Tick()
					nullMsg := &Auction.Void{}

					resultReply, err := client.Result(context.Background(), nullMsg)
					if err != nil {
						log.Printf("Error receiving current auction result: %v", err)
						client.CloseConnection(context.Background(), &Auction.Message{
							User:      clientUser,
							Timestamp: c.clock.GetTime(),
						})
						os.Exit(1)
					}

					c.clock.SyncClocks(resultReply.Timestamp)

					if !resultReply.StillActive {
						log.Printf("The auction is no longer active. Winner: %s bid: %d(%d)", resultReply.User.GetName(), resultReply.Amount, c.clock.GetTime())
						client.CloseConnection(context.Background(), &Auction.Message{
							User:      clientUser,
							Timestamp: c.clock.GetTime()})
						os.Exit(1)
					} else {
						log.Printf("The auction is still active. Current highest bid: %d. By: %s(%d)", resultReply.Amount, resultReply.User.Name, c.clock.GetTime())
					}

				} else if input == "bid" {
					c.clock.Tick()
					if len(inputArray) > 1 {
						bid, err := strconv.Atoi(inputArray[1])

						if err != nil {

							log.Println("The Auction House only accepts real integers as currency")
						}

						Bidmsg := &Auction.BidMessage{
							Amount:    int32(bid),
							User:      clientUser,
							Timestamp: c.clock.GetTime(),
						}

						reply, err := client.Bid(context.Background(), Bidmsg)
						if err != nil {
							log.Printf("Something went wrong")
							log.Println("We are currently trying to reconnect you")

						} else {

							c.clock.SyncClocks(reply.Timestamp)
							log.Printf("Auction House: Ack(%d)", c.clock.GetTime())
							log.Printf("bid: %s", translateReturn(reply.ReturnType))
						}
					} else {
						log.Println("Please input a number after your bid.")
					}
				}
			}
		}
	}()

	go func() {
		c.wait.Wait()
		close(done)
	}()

	<-done
}
