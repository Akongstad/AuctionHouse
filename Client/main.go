package main

import (
	"bufio"
	"context"
	"errors"
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

/*
	VI SKAL HAVE EN NY STREAM
	GRPC OPRETTER FORBINDELSE AUTOMATISK
	JEG TROR VI SKAL BRUGE CONNECT METODEN IGEN
*/

/*func (grpcclient *gRPCClient) connect(user *Auction.User) error {
	var streamError error
	var err error
	grpcclient.stream, err = grpcclient.client.OpenConnection(context.Background(), &Auction.Connect{
		User:   user,
		Active: true,
	})

	if err != nil {
		log.Fatalf("Connect failed: %v", err)
		return err
	}

	grpcclient.wait.Add(1)
	go grpcclient.receiveStream(grpcclient.stream, user)

	return streamError
}*/

/*func (grpcclient *gRPCClient) receiveStream(str Auction.AuctionHouse_OpenConnectionClient, user *Auction.User) {
	defer grpcclient.wait.Done()

	for {

		msg, err := str.Recv()
		if err != nil {
			log.Println("Error reading message")
			log.Println("Retrying in 10 seconds")
			log.Println("----------------------------")
			// luk stream

			grpcclient.conn.Close()
			time.Sleep(time.Second * 10)

			grpcclient.conn.WaitForStateChange(ctx, connectivity.Ready)

			//åben stream igen
			//var err2 error
			//grpcclient.stream, err2 = grpcclient.client.OpenConnection(context.Background(), &Auction.Connect{
				//User:   user,
				//Active: true,
			//})
			//if err2 != nil {
				log.Printf("Connect failed: %v", err)
			//}

		} else {
			grpcclient.clock.SyncClocks(msg.GetTimestamp())
			log.Printf("Auction House: %s: Has joined the auction(%d)", msg.GetUser().GetName(), grpcclient.clock.GetTime())
		}

	}
}*/

func (grpcclient *gRPCClient) ProcessRequests() error {
	defer grpcclient.conn.Close()

	go grpcclient.process()
	for {
		select {
		case <-grpcclient.reconnect:
			if !grpcclient.waitUntilReady() {
				return errors.New("failed to establish a connection within the defined timeout")
			}
			go grpcclient.process()
		case <-grpcclient.done:
			return nil
		}
	}
}

func (grpcclient *gRPCClient) process() {
	reqclient := GetStream(&grpcclient.user) //always get a new stream
	for {
		request, err := reqclient.Recv()
		log.Println("Request received")
		if err == io.EOF {
			grpcclient.done <- true
			log.Printf("io error: %v", err)
			return
		}
		if err != nil {
			grpcclient.reconnect <- true
			log.Printf("Failed to reconnect: %v", err)
			return

		} else {
			//the happy path
			//code block to process any requests that are received
			grpcclient.clock.SyncClocks(request.GetTimestamp())
			log.Printf("Auction House: %s: Has joined the auction(%d)", request.GetUser().GetName(), grpcclient.clock.GetTime())
		}
	}
}

func (grpcclient *gRPCClient) waitUntilReady() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second) //define how long you want to wait for connection to be restored before giving up
	defer cancel()

	currentState := grpcclient.conn.GetState()
	stillConnecting := true

	for currentState != connectivity.Ready && stillConnecting {
		//will return true when state has changed from thisState, false if timeout
		stillConnecting = grpcclient.conn.WaitForStateChange(ctx, currentState)
		currentState = grpcclient.conn.GetState()
		log.Println("Attempting reconnection. State has changed to:")
		log.Printf("state: %v", currentState)
	}

	if stillConnecting {
		log.Fatal("Connection attempt has timed out.")
		return false
	}

	return true
}

func GetStream(user *Auction.User) Auction.AuctionHouse_OpenConnectionClient {
	connection, err := client.OpenConnection(context.Background(), &Auction.Connect{
		User:   user,
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
		log.Println("Closing connections")
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
		wait:  &sync.WaitGroup{},
		clock: Auction.LamportClock{},
		user:  *clientUser,
	}

	log.Println(*clientName, "Connecting")

	// Set up a connection to the server.
	c.dialServer()

	//Create stream
	go c.ProcessRequests()
	//c.connect(clientUser)
	log.Println(*clientName, "Connected")

	//Send messages
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
							//log.Printf("Bid: Error %v", err)
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
							log.Println("Trying to reconnect")

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
