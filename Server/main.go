package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"math"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Akongstad/AuctionHouse/Auction"

	"google.golang.org/grpc"
)

type Connection struct {
	stream Auction.AuctionHouse_OpenConnectionServer
	id     string
	Name   string
	active bool
	error  chan error
}

type Server struct {
	ID            int32
	HighestBid    int32
	HighestBidder Auction.User
	StartTime     time.Time
	Duration      time.Duration
	Connections   []*Connection
	PrimePulse    time.Time
	PrimeId       int32
	Auction.UnimplementedAuctionHouseServer
	ServerTimestamp int32
	lock            sync.Mutex
	Ports           []int32
	Port            int32
}

func (s *Server) Pulse() {
	for {
		time.Sleep(time.Second * 10)
		if s.ID == s.PrimeId {
			log.Print("Prime replica: Pulse")
			go func() {
				for i := 0; i < len(s.Ports); i++ {
					if i != int(s.ID) {
						conn, err := grpc.Dial(":" + strconv.Itoa(int(s.Ports[i])), grpc.WithInsecure())

						if err != nil {
							log.Fatalf("Failed to dial this port(Message all): %v", err)
						}

						defer conn.Close()
						client := Auction.NewAuctionHouseClient(conn)
						msg := Auction.Message{
							Timestamp: s.ServerTimestamp,
						}
						client.RegisterPulse(context.Background(), &msg)

					}
				}
			}()
		} else if math.Abs(float64(time.Now().Second()-s.PrimePulse.Second())) > 19 {
			log.Printf("Missing pulse from prime replica. Last Pulse: %v seconds ago", math.Abs(float64(time.Now().Second()-s.PrimePulse.Second())))
			s.CallRingElection(context.Background(), s.Ports[s.PrimeId])
			log.Printf("Leader election called")
		}
	}
}
func (s *Server) RegisterPulse(ctx context.Context, msg *Auction.Message) (*Auction.Void, error) {
	log.Printf("Received pulse from prime replica")
	s.PrimePulse = time.Now()
	return &Auction.Void{}, nil

}

func (s *Server) ReplicateBackups(ctx context.Context, HighestBid int32, HighestBidder string) {
	for i := 0; i < len(s.Ports); i++ {
		if i != int(s.ID) {
			conn, err := grpc.Dial(":" + strconv.Itoa(int(s.Ports[i])), grpc.WithInsecure())

			if err != nil {
				log.Fatalf("Failed to dial this port(Message all): %v", err)
			}

			defer conn.Close()
			client := Auction.NewAuctionHouseClient(conn)
			repMsg := &Auction.BidMessage{Amount: s.HighestBid, User: &s.HighestBidder, Timestamp: s.ServerTimestamp}
			client.Replicate(ctx, repMsg)
		}
	}
}

func (s *Server) CallRingElection(ctx context.Context, brokenPort int32) {

	listOfPorts := make([]int32, 0)

	listOfPorts = append(listOfPorts, s.Port)

	index := s.FindIndex(s.Port)

	nextPort := ":" + strconv.Itoa(int(s.Ports[index + 1 % len(s.Ports)]))

	conn, err := grpc.Dial(nextPort, grpc.WithInsecure())
	if err != nil {
		log.Printf("Election: Failed to dial this port: %v", err)
	} else {
		defer conn.Close()
		client := Auction.NewAuctionHouseClient(conn)

		protoListOfPorts := Auction.RingMessage{
			ListOfPorts: listOfPorts,
		}

		client.RingElection(ctx, &protoListOfPorts)
	}
}

func (s *Server) RingElection(ctx context.Context, msg *Auction.RingMessage) (*Auction.Void, error) {

	listOfPorts := msg.ListOfPorts

	if listOfPorts[0] == s.Port {
		highestPort := 0
		for i := 0; i < len(listOfPorts); i++ {
			if int(listOfPorts[i]) > highestPort {
				highestPort = int(listOfPorts[i])
			}
		}

		//Call other ports with the new leader


	} else {
		msg.ListOfPorts = append(msg.ListOfPorts, s.ID)

		//Call RingElection på alle andre
		index := s.FindIndex(s.Port)

		nextPort := ":" + strconv.Itoa(int(s.Ports[index + 1 % len(s.Ports)]))

		conn, err := grpc.Dial(nextPort, grpc.WithInsecure())
		if err != nil {
			log.Printf("Election: Failed to dial this port: %v", err)
		} else {
			defer conn.Close()
			client := Auction.NewAuctionHouseClient(conn)

			protoListOfPorts := Auction.RingMessage{
				ListOfPorts: listOfPorts,
			}

			client.RingElection(ctx, &protoListOfPorts)
		}
	}

	return &Auction.Void{}, nil
}

func (s *Server) Replicate(ctx context.Context, update *Auction.BidMessage) (*Auction.BidReply, error) {
	s.HighestBid = update.Amount
	s.HighestBidder = *update.User
	return &Auction.BidReply{
		Timestamp:  s.ServerTimestamp,
		ReturnType: 1,
	}, nil
}

func (s *Server) Bid(ctx context.Context, bid *Auction.BidMessage) (*Auction.BidReply, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if time.Now().Before(s.StartTime.Add(s.Duration)) {
		if s.HighestBid < bid.Amount {
			s.HighestBid = bid.Amount
			s.HighestBidder = *bid.User

			log.Printf("Highest bid: %d, by: %s", s.HighestBid, s.HighestBidder.Name)
			return &Auction.BidReply{
				ReturnType: 1,
				Timestamp:  s.ServerTimestamp + 1,
			}, nil

		} else {
			log.Printf("Highest bid: %d, by: %s", s.HighestBid, s.HighestBidder.Name)
			return &Auction.BidReply{
				ReturnType: 2,
				Timestamp:  s.ServerTimestamp + 1,
			}, nil
		}

	}

	return &Auction.BidReply{
		ReturnType: 3,
		Timestamp:  s.ServerTimestamp + 1,
	}, nil
}

func (s *Server) Result(ctx context.Context, msg *Auction.Void) (*Auction.ResultMessage, error) {
	if time.Now().Before(s.StartTime.Add(s.Duration)) {
		return &Auction.ResultMessage{
			Amount:      s.HighestBid,
			User:        &Auction.User{Name: s.HighestBidder.Name},
			Timestamp:   0,
			StillActive: true}, nil
	} else {
		return &Auction.ResultMessage{
			Amount:      s.HighestBid,
			User:        &Auction.User{Name: s.HighestBidder.Name},
			Timestamp:   0,
			StillActive: false}, nil
	}
}

func (s *Server) Broadcast(ctx context.Context, msg *Auction.Message) (*Auction.Void, error) {
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
	return &Auction.Void{}, nil
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
	log.Printf("Auction House: %s has joined the auction!", connect.User.Name)
	s.Broadcast(context.Background(), &joinMessage)

	return <-conn.error
}

func (s *Server) CloseConnection(ctx context.Context, user *Auction.User) (*Auction.Void, error) {

	var deleted *Connection

	for index, conn := range s.Connections {
		if conn.id == user.Name+strconv.Itoa(int(user.Id)) {
			s.Connections = remove(s.Connections, index)
			deleted = conn
		}
	}
	if deleted == nil {
		log.Print("No such connection to close")
		return &Auction.Void{}, nil
	}

	leaveMessage := Auction.Message{
		User:      &Auction.User{Name: "Auction House"},
		Timestamp: s.ServerTimestamp,
	}

	log.Print("__________________________________")
	log.Printf("Auction House: %s has left the auction", user.Name)
	s.Broadcast(context.Background(), &leaveMessage)

	return &Auction.Void{}, nil
}

func remove(slice []*Connection, i int) []*Connection {
	slice[i] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}

func (s *Server) StartAuction(Duration time.Duration) {
	s.Duration = Duration
	a, b, c := s.StartTime.Clock()
	log.Printf("Auctioneer: Auction started at %d:%d:%d. Duration will be: %v seconds", a, b, c, s.Duration.Seconds())

	for {
		time.Sleep(time.Second * 1)

		if math.Abs(float64(time.Now().Second()-s.StartTime.Add(s.Duration).Second())) <= 3 {
			log.Printf("Auctioneer: %v", math.Abs(float64(time.Now().Second()-s.StartTime.Add(s.Duration).Second())))
			if math.Abs(float64(time.Now().Second()-s.StartTime.Add(s.Duration).Second())) <= 0 {
				break
			}
		}
	}
	log.Printf("Auctioneer: Auction closed. Highest bid: %d by %s", s.HighestBid, s.HighestBidder.Name)
}

func main() {
	var connections []*Connection

	portFile, err := os.Open("../ports.txt")
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(portFile)

	var ports []int32

	for scanner.Scan() {
		nextPort, _ := strconv.Atoi(scanner.Text())
		ports = append(ports, int32(nextPort))
	}

	s := &Server{
		Connections:                     connections,
		UnimplementedAuctionHouseServer: Auction.UnimplementedAuctionHouseServer{},
		ServerTimestamp:                 0,
		StartTime:                       time.Now(),
		lock:                            sync.Mutex{},
		Ports:                           ports,
	}

	go s.StartAuction(time.Second * 30)
	s.Pulse()

	// If the file doesn't exist, create it or append to the file. For append functionality : os.O_APPEND
	file, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	//Create multiwriter
	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)

	grpcServer := grpc.NewServer()

	lis, err := net.Listen("tcp", ":" + strconv.Itoa(int(s.Port)))
	if err != nil {
		log.Fatalf("Failed to listen port: %v", err)
	}
	log.Printf("Auction open at: %v", lis.Addr())

	Auction.RegisterAuctionHouseServer(grpcServer, s)
	defer func() {
		lis.Close()
		log.Printf("Server stopped listening")
	}()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve server: %v", err)
	}

}


/*
----------------------------------------------------------------------------------------------
	HELPER METHODS
----------------------------------------------------------------------------------------------
*/

func (s *Server) FindIndex(port int32) int {
	index := -1

	for i := 0; i < len(s.Ports); i++ {
		if s.Ports[i] == port {
			index = i
			break
		}
	}

	return index
}
