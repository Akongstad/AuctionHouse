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

const (
	port = ":9000"
)

var client Auction.AuctionHouseClient
var wait *sync.WaitGroup

func init() {
	wait = &sync.WaitGroup{}
}

func connect(user *Auction.User) error {
	var streamError error

	user.Timestamp++
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
			log.Printf("%v:,(%d)", msg.GetUser().GetName(), user.GetTimestamp())
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
		Id:        int64(userId),
		Name:      *clientName,
		Timestamp: 0,
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(port, grpc.WithInsecure())
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

			if strings.ToLower(scanner.Text()) == "exit" {
				client.CloseConnection(context.Background(), clientUser)
				os.Exit(1)
			}
			if strings.TrimSpace(scanner.Text()) != "" {
				if strings.ToLower(scanner.Text()) == "bid" {
					bid, err := strconv.Atoi(scanner.Text())
					if err != nil {
						log.Printf("Bid: Error %v", err)
						break
						Bidmsg := &Auction.BidMessage{
							Amount:    int32(bid),
							User:      clientUser,
							Timestamp: clientUser.Timestamp,
						}
						_, err := client.Bid(context.Background(), Bidmsg)
						if err != nil {
							log.Printf("Error publishing Message: %v", err)
							break
						}
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
