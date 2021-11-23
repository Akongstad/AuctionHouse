package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/Akongstad/AuctionHouse/Auction"
	"google.golang.org/grpc"
)

var client Auction.AuctionHouseClient
var wait *sync.WaitGroup

func init() {
	wait = &sync.WaitGroup{}
}

var (
	clock Auction.LamportClock
)

func translateReturn(nr int32) (msg string) {
	if nr == 1 {
		return "Success"
	} else if nr == 2 {
		return "Fail"
	} else {
		return "Exception"
	}
}

func connect(user *Auction.User) error {
	var streamError error

	stream, err := client.OpenConnection(context.Background(), &Auction.Connect{
		User:   user,
		Active: true,
	})

	if err != nil {
		log.Fatalf("Connect failed: %v", err)
		return err
	}

	wait.Add(1)
	go func(str Auction.AuctionHouse_OpenConnectionClient) {
		defer wait.Done()

		for {
			msg, err := str.Recv()
			if err != nil {
				log.Fatalf("Error reading message, %v", err)
				streamError = err
				break
			}
			clock.SyncClocks(msg.GetTimestamp())
			log.Printf("Auction House: %s: Has joined the auction(%d)", msg.GetUser().GetName(), clock.GetTime())
		}
	}(stream)

	return streamError
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

	// Set up a connection to the server.
	conn, err := grpc.Dial(":5001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	client = Auction.NewAuctionHouseClient(conn)

	//Create stream
	log.Println(*clientName, " Connecting")
	connect(clientUser)

	//Send messages
	wait.Add(1)

	go func() {
		defer wait.Done()

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {

			inputArray := strings.Fields(scanner.Text())

			input := strings.ToLower(strings.TrimSpace(inputArray[0]))

			if input == "exit" {
				clock.Tick()
				client.CloseConnection(context.Background(), &Auction.Message{
					User:      clientUser,
					Timestamp: clock.GetTime()})
				os.Exit(1)
			}

			if input != "" {
				if input == "result" {
					clock.Tick()
					nullMsg := &Auction.Void{}

					resultReply, err := client.Result(context.Background(), nullMsg)
					clock.SyncClocks(resultReply.Timestamp)
					if err != nil {
						log.Printf("Error receiving current auction result: %v", err)
						client.CloseConnection(context.Background(), &Auction.Message{
							User:      clientUser,
							Timestamp: clock.GetTime()})
						os.Exit(1)
					}

					if !resultReply.StillActive {
						log.Printf("The auction is no longer active. Winner: %s bid: %d(%d)", resultReply.User.GetName(), resultReply.Amount, clock.GetTime() )
						client.CloseConnection(context.Background(), &Auction.Message{
							User:      clientUser,
							Timestamp: clock.GetTime()})
						os.Exit(1)
					} else {
						log.Printf("The auction is still active. Current highest bid: %d. By: %s(%d)", resultReply.Amount, resultReply.User.Name, clock.GetTime())
					}

				} else if input == "bid" {
					clock.Tick()
					if len(inputArray) > 0 {
						bid, err := strconv.Atoi(inputArray[1])

						if err != nil {
							//log.Printf("Bid: Error %v", err)
							log.Println("The Auction House only accepts real integers as currency")
						}

						Bidmsg := &Auction.BidMessage{
							Amount: int32(bid),
							User:   clientUser,
							Timestamp: clock.GetTime(),
						}

						reply, err := client.Bid(context.Background(), Bidmsg)
						if err != nil {
							log.Printf("Error publishing Message: %v", err)
							break
						}
						clock.SyncClocks(reply.Timestamp)
						log.Printf("Auction House: Ack(%d)", clock.GetTime())
						log.Printf("bid: %s", translateReturn(reply.ReturnType))

					} else {
						log.Println("Please input a number after your bid.")
					}
				}
			}
		}
	}()

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
}
