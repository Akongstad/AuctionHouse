package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Akongstad/AuctionHouse/Auction"

	"google.golang.org/grpc"
)

const (
	port = ":9000"
)

type Connection struct {
	stream Auction.AuctionHouse_OpenConnectionServer
	id     string
	Name   string
	active bool
	error  chan error
}

type Server struct {
	HighestBid  int32
	StartTime   time.Time
	Duration    time.Duration
	Connections []*Connection
	Auction.UnimplementedAuctionHouseServer
	ServerTimestamp int32
	lock            sync.Mutex
}

func (s *Server) Bid(ctx context.Context, bid *Auction.BidMessage) (*Auction.BidReply, error) {
	if s.HighestBid < bid.Amount {
		s.HighestBid = bid.Amount
		return &Auction.BidReply{
			ReturnType: 1,
			Timestamp:  s.ServerTimestamp + 1,
		}, nil
	} else {
		return &Auction.BidReply{
			ReturnType: 2,
			Timestamp:  s.ServerTimestamp + 1,
		}, nil
	}

	/* s.lock.Lock()
	defer s.lock.Unlock()
	bidReply:= Auction.BidReply{
		ReturnType: 1,
		Timestamp: s.ServerTimestamp,
	}

	s.Broadcast(ctx, &Auction.Message{User: bid.User, Timestamp: s.ServerTimestamp})
	return &bidReply, nil */
}

func (s *Server) Result(ctx context.Context, msg *Auction.ResultMessage) (*Auction.ResultMessage, error) {
	if time.Now().Before(s.StartTime.Add(s.Duration)) {
		return &Auction.ResultMessage{
			Amount:      s.HighestBid,
			Timestamp:   0,
			StillActive: true}, nil
	} else {
		return &Auction.ResultMessage{
			Amount:      s.HighestBid,
			User:        nil,
			Timestamp:   0,
			StillActive: false}, nil
	}
}

func (s *Server) Broadcast(ctx context.Context, msg *Auction.Message) (*Auction.Close, error) {
	wait := sync.WaitGroup{}
	done := make(chan int)

	for _, conn := range s.Connections {
		wait.Add(1)

		go func(msg *Auction.Message, conn *Connection) {
			defer wait.Done()

			if conn.active {

				err := conn.stream.Send(msg)
				log.Printf("Broadcasting message to: %s", conn.Name)
				if err != nil {
					conn.active = false
					conn.error <- err
				}
			}
		}(msg, conn)
	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
	return &Auction.Close{}, nil
}

func (s *Server) OpenConnection(connect *Auction.Connect, stream Auction.AuctionHouse_OpenConnectionServer) error {
	conn := &Connection{
		stream: stream,
		active: true,
		id:     connect.User.Name + strconv.Itoa(int(connect.User.Id)),
		Name:   connect.User.Name,
		error:  make(chan error),
	}

	s.Connections = append(s.Connections, conn)

	joinMessage := Auction.Message{
		User:      connect.User,
		Timestamp: s.ServerTimestamp,
	}
	log.Print("__________________________________")
	log.Printf("Auction House: %s has joined the auction!(%d)", connect.User.Name, connect.User.Timestamp)
	s.Broadcast(context.Background(), &joinMessage)

	return <-conn.error
}

func (s *Server) CloseConnection(ctx context.Context, user *Auction.User) (*Auction.Close, error) {

	var deleted *Connection

	for index, conn := range s.Connections {
		if conn.id == user.Name+strconv.Itoa(int(user.Id)) {
			s.Connections = remove(s.Connections, index)
			deleted = conn
		}
	}
	if deleted == nil {
		log.Print("No such connection to close")
		return &Auction.Close{}, nil
	}

	leaveMessage := Auction.Message{
		User:      &Auction.User{Name: "Auction House"},
		Timestamp: s.ServerTimestamp,
	}

	log.Print("__________________________________")
	log.Printf("Auction House: %s has left the auction(%d)", user.Name, user.Timestamp)
	s.Broadcast(context.Background(), &leaveMessage)

	return &Auction.Close{}, nil
}

func remove(slice []*Connection, i int) []*Connection {
	slice[i] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}

func main() {
	var connections []*Connection
	s := &Server{
		Connections:                     connections,
		UnimplementedAuctionHouseServer: Auction.UnimplementedAuctionHouseServer{},
		ServerTimestamp:                 0,
		lock:                            sync.Mutex{},
	}
	// If the file doesn't exist, create it or append to the file. For append functionality : os.O_APPEND
	file, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	//Create multiwriter
	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)

	grpcServer := grpc.NewServer()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen port: %v", err)
	}

	log.Printf("Auction open at: %v", lis.Addr())

	Auction.RegisterAuctionHouseServer(grpcServer, s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve server: %v", err)
	}
}

func Max(a int32, b int32) int32 {
	if b > a {
		return b
	}
	return a
}
