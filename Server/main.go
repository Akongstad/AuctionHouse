package main

import (
	"bufio"
	"context"
	"flag"
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
	MainPort = 5001
)

type STATE int

const (
	RELEASED STATE = iota
	WANTED
	HELD
)

type Connection struct {
	stream Auction.AuctionHouse_OpenConnectionServer
	id     string
	Name   string
	active bool
	error  chan error
}

type Server struct {
	ID            int
	HighestBid    int32
	HighestBidder Auction.User
	StartTime     time.Time
	Duration      time.Duration
	Connections   []*Connection
	PrimePulse    time.Time
	Auction.UnimplementedAuctionHouseServer
	ServerTimestamp   Auction.LamportClock
	lock              sync.Mutex
	Ports             []int32
	Port              int32
	state             STATE
	replyCounter      int
	grpcServer        *grpc.Server
	receivedNewLeader bool
}

func (s *Server) Pulse() {
	for {
		time.Sleep(time.Second * 2)
		if s.GetPort() == MainPort {
			s.ServerTimestamp.Tick()

			go func() {
				portIndex := s.FindIndex(s.Port)
				for i := 0; i < len(s.Ports); i++ {

					if i != int(portIndex) {
						conn, err := grpc.Dial(":"+strconv.Itoa(int(s.Ports[i])), grpc.WithInsecure())

						if err != nil {
							log.Fatalf("Failed to dial this port(Message all): %v", err)
						}

						defer conn.Close()
						client := Auction.NewAuctionHouseClient(conn)
						msg := Auction.Message{
							Timestamp: s.ServerTimestamp.GetTime(),
						}
						client.RegisterPulse(context.Background(), &msg)
					}
				}
			}()
		} else if s.receivedNewLeader {
			//This makes sure, that a port will not request to become leader right after it has received a new leader
			s.receivedNewLeader = false
		} else if time.Since(s.PrimePulse).Seconds() > time.Second.Seconds()*6 && s.state != WANTED && s.state != HELD {

			if len(s.Ports) > 2 {
				s.ServerTimestamp.Tick()
				log.Printf("Missing pulse from prime replica. Last Pulse: %v seconds ago", time.Since(s.PrimePulse))

				log.Printf("Leader election called")

				msg := Auction.RequestMessage{
					ServerId:  int32(s.ID),
					Timestamp: int32(s.ServerTimestamp.GetTime()),
					Port:      s.Port,
				}
				_, err := s.AccessCritical(context.Background(), &msg)
				if err != nil {
					log.Fatalf("Failed to access critical: %v", err)
				}

			} else {
				s.LastLeader()
			}

		}
	}
}

func (s *Server) LastLeader() {
	updatedPort := s.FindIndex(s.Port)
	s.Ports = removePort(s.Ports, int32(updatedPort))
	s.Port = MainPort

	s.grpcServer = grpc.NewServer()

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(int(s.Port)))
	if err != nil {
		log.Fatalf("Failed to listen port: %v", err)
	}

	log.Printf("Auction open at: %v", lis.Addr())

	Auction.RegisterAuctionHouseServer(s.grpcServer, s)
	go func(listener net.Listener) {
		defer func() {
			lis.Close()
			log.Printf("Server stopped listening")
		}()

		if err := s.grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve server: %v", err)
		}
	}(lis)
}

func (s *Server) RegisterPulse(ctx context.Context, msg *Auction.Message) (*Auction.Void, error) {

	s.ServerTimestamp.SyncClocks(msg.GetTimestamp())
	s.PrimePulse = time.Now().UTC()
	return &Auction.Void{}, nil

}

func (s *Server) ReplicateBackups(ctx context.Context, HighestBid int32, HighestBidder *Auction.User) {

	localWG := new(sync.WaitGroup)

	for i := 0; i < len(s.Ports); i++ {

		currentPort := s.Ports[i]

		if currentPort != s.Port {
			conn, err := grpc.Dial(":"+strconv.Itoa(int(s.Ports[i])), grpc.WithInsecure())

			if err != nil {
				log.Fatalf("Failed to dial this port(Message all): %v", err)
			}

			defer conn.Close()
			client := Auction.NewAuctionHouseClient(conn)

			repMsg := &Auction.ReplicateMessage{
				Amount: HighestBid, User: HighestBidder, Timestamp: s.ServerTimestamp.GetTime(),
			}

			var bidReply *Auction.BidReply
			go func() {
				var error error
				bidReply, error = client.Replicate(ctx, repMsg)
				if error != nil {
					log.Printf("Could not replicate to port. %v", error)
				}
			}()
			localWG.Add(1)

			go func(port int32) {

				defer localWG.Done()
				time.Sleep(time.Second * 2)
				if bidReply == nil {

					s.CallCutOffReplicate(ctx, port)
				}
			}(currentPort)

		}
	}

	localWG.Wait()
}

func (s *Server) CallCutOffReplicate(ctx context.Context, brokenPort int32) {
	for i := 0; i < len(s.Ports); i++ {
		if s.Ports[i] != brokenPort {
			conn, err := grpc.Dial(":"+strconv.Itoa(int(s.Ports[i])), grpc.WithInsecure())

			if err != nil {
				log.Fatalf("Failed to dial this port(Message all): %v", err)
			}

			defer conn.Close()
			client := Auction.NewAuctionHouseClient(conn)

			msg := &Auction.CutOfMessage{
				Port: brokenPort,
			}

			client.CutOfReplicate(ctx, msg)
		}
	}
}

func (s *Server) CutOfReplicate(ctx context.Context, msg *Auction.CutOfMessage) (*Auction.Void, error) {
	brokenPort := msg.Port
	index := s.FindIndex(brokenPort)

	s.Ports = removePort(s.Ports, int32(index))

	return &Auction.Void{}, nil
}

func (s *Server) Replicate(ctx context.Context, update *Auction.ReplicateMessage) (*Auction.BidReply, error) {
	s.HighestBid = update.Amount
	s.HighestBidder = *update.User

	s.ServerTimestamp.SyncClocks(update.Timestamp)

	log.Printf("Replicated: bid:%d, User: %s", s.HighestBid, s.HighestBidder.Name)
	return &Auction.BidReply{
		Timestamp:  s.ServerTimestamp.GetTime(),
		ReturnType: 1,
	}, nil
}

func (s *Server) CallRingElection(ctx context.Context) {

	listOfPorts := make([]int32, 0)
	listOfClocks := make([]uint32, 0)

	listOfPorts = append(listOfPorts, s.Port)
	listOfClocks = append(listOfClocks, s.ServerTimestamp.GetTime())

	index := s.FindIndex(s.Port)

	nextPort := s.FindNextPort(index)

	conn, err := grpc.Dial(nextPort, grpc.WithInsecure())
	if err != nil {
		log.Printf("Election: Failed to dial this port: %v", err)
	} else {
		defer conn.Close()
		client := Auction.NewAuctionHouseClient(conn)

		protoListOfPorts := Auction.PortsAndClocks{
			ListOfPorts:  listOfPorts,
			ListOfClocks: listOfClocks,
		}

		client.RingElection(ctx, &protoListOfPorts)
	}
}

func (s *Server) RingElection(ctx context.Context, msg *Auction.PortsAndClocks) (*Auction.Void, error) {

	listOfPorts := msg.ListOfPorts
	listOfClocks := msg.ListOfClocks

	if listOfPorts[0] == s.Port {
		var highestClock uint32
		var highestPort int32

		for i := 0; i < len(listOfPorts); i++ {
			if listOfClocks[i] > highestClock {
				highestPort = listOfPorts[i]
				highestClock = listOfClocks[i]
			} else if listOfClocks[i] == highestClock && listOfPorts[i] > highestPort {
				highestPort = listOfPorts[i]
				highestClock = listOfClocks[i]
			}
		}

		conn, err := grpc.Dial(":"+strconv.Itoa(int(highestPort)), grpc.WithInsecure())
		if err != nil {
			log.Printf("Election: Failed to dial this port: %v", err)
		} else {

			defer conn.Close()
			client := Auction.NewAuctionHouseClient(conn)

			newPortList, err := client.SelectNewLeader(ctx, &Auction.Void{})
			if err != nil {
				log.Printf("Election: Failed to select new leader: %v", err)
			}

			for i := 0; i < len(newPortList.ListOfPorts); i++ {

				conn, err := grpc.Dial(":"+strconv.Itoa(int(newPortList.ListOfPorts[i])), grpc.WithInsecure())

				if err != nil {
					log.Printf("Election: Failed to dial this port: %v", err)
				} else {
					defer conn.Close()
					client := Auction.NewAuctionHouseClient(conn)

					client.BroadcastNewLeader(ctx, newPortList)
				}

			}
		}
	} else {
		msg.ListOfPorts = append(msg.ListOfPorts, s.Port)

		index := s.FindIndex(s.Port)

		nextPort := s.FindNextPort(index)

		conn, err := grpc.Dial(nextPort, grpc.WithInsecure())
		if err != nil {
			log.Printf("Election: Failed to dial this port: %v", err)
		} else {
			defer conn.Close()
			client := Auction.NewAuctionHouseClient(conn)

			protoListOfPorts := Auction.PortsAndClocks{
				ListOfPorts:  listOfPorts,
				ListOfClocks: listOfClocks,
			}

			client.RingElection(ctx, &protoListOfPorts)
		}
	}

	return &Auction.Void{}, nil
}

func (s *Server) BroadcastNewLeader(ctx context.Context, newPorts *Auction.ElectionPorts) (*Auction.Void, error) {
	log.Println("Received new leader")
	s.receivedNewLeader = true
	s.Ports = newPorts.ListOfPorts
	s.state = RELEASED
	s.replyCounter = 0

	return &Auction.Void{}, nil
}

func (s *Server) SelectNewLeader(ctx context.Context, void *Auction.Void) (*Auction.ElectionPorts, error) {

	updatedPort := s.FindIndex(s.Port)
	s.Ports = removePort(s.Ports, int32(updatedPort))
	s.Port = MainPort

	s.grpcServer = grpc.NewServer()

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(int(s.Port)))
	if err != nil {
		log.Fatalf("Failed to listen port: %v", err)
	}

	log.Printf("Auction open at: %v", lis.Addr())

	Auction.RegisterAuctionHouseServer(s.grpcServer, s)
	go func(listener net.Listener) {
		defer func() {
			lis.Close()
			log.Printf("Server stopped listening")
		}()

		if err := s.grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve server: %v", err)
		}
	}(lis)

	newPortList := Auction.ElectionPorts{
		ListOfPorts: s.Ports,
	}

	return &newPortList, nil
}

/*
----------------------------------------------------------------------------------------------
	CLient API
----------------------------------------------------------------------------------------------
*/

func (s *Server) Bid(ctx context.Context, bid *Auction.BidMessage) (*Auction.BidReply, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ServerTimestamp.SyncClocks(bid.Timestamp)
	if time.Now().UTC().Before(s.StartTime.Add(s.Duration)) {
		if s.HighestBid < bid.Amount {

			s.ReplicateBackups(ctx, bid.Amount, bid.User)

			s.HighestBid = bid.Amount
			s.HighestBidder = *bid.User

			log.Printf("Highest bid: %d, by: %s", s.HighestBid, s.HighestBidder.Name)
			return &Auction.BidReply{
				ReturnType: 1,
				Timestamp:  s.ServerTimestamp.GetTime(),
			}, nil

		} else {
			log.Printf("Highest bid: %d, by: %s", s.HighestBid, s.HighestBidder.Name)
			return &Auction.BidReply{
				ReturnType: 2,
				Timestamp:  s.ServerTimestamp.GetTime(),
			}, nil
		}

	}
	return &Auction.BidReply{
		ReturnType: 3,
		Timestamp:  s.ServerTimestamp.GetTime(),
	}, nil
}

func (s *Server) Result(ctx context.Context, msg *Auction.Void) (*Auction.ResultMessage, error) {
	s.ServerTimestamp.Tick()
	if time.Now().Before(s.StartTime.Add(s.Duration)) {
		return &Auction.ResultMessage{
			Amount:      s.HighestBid,
			User:        &Auction.User{Name: s.HighestBidder.Name},
			Timestamp:   s.ServerTimestamp.GetTime(),
			StillActive: true}, nil
	} else {
		return &Auction.ResultMessage{
			Amount:      s.HighestBid,
			User:        &Auction.User{Name: s.HighestBidder.Name},
			Timestamp:   s.ServerTimestamp.GetTime(),
			StillActive: false}, nil
	}
}

/*
----------------------------------------------------------------------------------------------
	Client Connection and communitation
----------------------------------------------------------------------------------------------
*/
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
		Timestamp: s.ServerTimestamp.GetTime(),
	}
	log.Print("__________________________________")
	log.Printf("Auction House: %s has joined the auction!", connect.User.Name)
	s.Broadcast(context.Background(), &joinMessage)

	return <-conn.error
}

func (s *Server) CloseConnection(ctx context.Context, msg *Auction.Message) (*Auction.Void, error) {

	var deleted *Connection

	for index, conn := range s.Connections {
		if conn.id == msg.User.Name+strconv.Itoa(int(msg.User.Id)) {
			s.Connections = removeConnection(s.Connections, index)
			deleted = conn
		}
	}
	if deleted == nil {
		log.Print("No such connection to close")
		return &Auction.Void{}, nil
	}
	leaveMessage := Auction.Message{
		User:      &Auction.User{Name: "Auction House"},
		Timestamp: s.ServerTimestamp.GetTime(),
	}

	log.Print("__________________________________")
	log.Printf("Auction House: %s has left the auction", msg.User.Name)
	s.Broadcast(context.Background(), &leaveMessage)

	return &Auction.Void{}, nil
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

/*
----------------------------------------------------------------------------------------------
	Main
----------------------------------------------------------------------------------------------
*/

func main() {
	id := flag.Int("I", -1, "id")
	AuctionTime := flag.Int("D", 100, "seconds")
	flag.Parse()

	connections := make([]*Connection, 0)

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
		ID:                              *id,
		Connections:                     connections,
		UnimplementedAuctionHouseServer: Auction.UnimplementedAuctionHouseServer{},
		ServerTimestamp:                 Auction.LamportClock{},
		lock:                            sync.Mutex{},
		Ports:                           ports,
		Port:                            ports[*id],
		PrimePulse:                      time.Now(),
		state:                           RELEASED,
		grpcServer:                      grpc.NewServer(),
	}
	go s.StartAuction(time.Second * time.Duration(*AuctionTime))
	go s.Pulse()

	filename := "Server" + strconv.Itoa(s.ID) + "logs.txt"
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(int(s.Port)))
	if err != nil {
		log.Fatalf("Failed to listen port: %v", err)
	}
	log.Printf("Auction open at: %v", lis.Addr())

	Auction.RegisterAuctionHouseServer(s.grpcServer, s)
	defer func() {
		lis.Close()
		log.Printf("Server stopped listening")
	}()

	if err := s.grpcServer.Serve(lis); err != nil {
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

func (s *Server) FindNextPort(index int) string {

	nextPort := s.Ports[(index+1)%len(s.Ports)]
	if nextPort == MainPort {
		nextPort = s.Ports[(index+2)%len(s.Ports)]
	}

	return ":" + strconv.Itoa(int(nextPort))
}

func removeConnection(slice []*Connection, i int) []*Connection {
	slice[i] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}

func removePort(s []int32, i int32) []int32 {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func (s *Server) StartAuction(Duration time.Duration) {

	if s.Port != MainPort {
		conn, err := grpc.Dial(":"+strconv.Itoa(int(MainPort)), grpc.WithInsecure())
		if err != nil {
			log.Printf("Election: Failed to dial this port: %v", err)
		} else {
			defer conn.Close()

			client := Auction.NewAuctionHouseClient(conn)
			msg, err := client.PromptTimeAndDuration(context.Background(), &Auction.Void{})
			if err != nil {
				log.Fatalf("Failed to get time and duration from prime")
			}
			Auctionstart, err := time.Parse("2006-01-02 15:04:05 -0700 MST", msg.AuctionStart)
			if err != nil {
				log.Fatalf("Failed to parse starttime on replicate")
			}
			s.StartTime = Auctionstart

			s.Duration = time.Duration(time.Second * time.Duration(msg.Duration))
		}
	} else {
		s.StartTime = time.Now().UTC()
		s.Duration = Duration
	}

	log.Printf("Auctioneer: Auction started at %d. Duration will be: %v seconds", s.Port, s.Duration.Seconds())

	for {
		time.Sleep(time.Second * 1)
		if time.Now().UTC().After(s.StartTime.Add(s.Duration)) {
			break
		}
	}
	log.Printf("Auctioneer: Auction closed. Highest bid: %d by %s", s.HighestBid, s.HighestBidder.Name)

}

func (s *Server) PromptTimeAndDuration(ctx context.Context, void *Auction.Void) (*Auction.TimeMessage, error) {
	msg := Auction.TimeMessage{
		AuctionStart: s.StartTime.String(),
		Duration:     int32(s.Duration.Seconds()),
	}

	return &msg, nil
}

/*
	RICART AND AGRAWALA
*/

func (s *Server) AccessCritical(ctx context.Context, requestMessage *Auction.RequestMessage) (*Auction.ReplyMessage, error) {

	s.state = WANTED
	s.MessageAll(ctx, requestMessage)
	reply := Auction.ReplyMessage{Timestamp: int32(s.ServerTimestamp.GetTime()), ServerId: int32(s.ID), Port: s.Port}

	return &reply, nil
}

func (s *Server) MessageAll(ctx context.Context, msg *Auction.RequestMessage) error {
	mainPortIndex := s.FindIndex(MainPort)
	ownPortIndex := s.FindIndex(s.Port)

	for i := 0; i < len(s.Ports); i++ {

		if i != mainPortIndex && i != ownPortIndex {
			conn, err := grpc.Dial(":"+strconv.Itoa(int(s.Ports[i])), grpc.WithInsecure())

			if err != nil {
				log.Fatalf("Failed to dial this port(Message all): %v", err)
			}
			defer conn.Close()
			client := Auction.NewAuctionHouseClient(conn)

			client.ReceiveRequest(ctx, msg)
		}
	}
	return nil
}

func (s *Server) ReceiveRequest(ctx context.Context, requestMessage *Auction.RequestMessage) (*Auction.Void, error) {

	if !s.shouldDefer(requestMessage) {
		s.SendReply(requestMessage.Port)
		log.Println("Send reply to", requestMessage.Port)
	}

	return &Auction.Void{}, nil
}

func (s *Server) shouldDefer(requestMessage *Auction.RequestMessage) bool {
	if s.state == HELD {
		return true
	}

	if s.state != WANTED {
		return false
	}

	if int32(s.ServerTimestamp.GetTime()) < requestMessage.Timestamp {
		return true
	}

	if int32(s.ServerTimestamp.GetTime()) == requestMessage.Timestamp && s.ID < int(requestMessage.ServerId) {
		return true
	}

	return false
}

func (s *Server) SendReply(port int32) {

	conn, err := grpc.Dial(":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to port: %v", err)
	}
	defer conn.Close()
	client := Auction.NewAuctionHouseClient(conn)
	client.ReceiveReply(context.Background(), &Auction.ReplyMessage{})
}

func (s *Server) ReceiveReply(ctx context.Context, replyMessage *Auction.ReplyMessage) (*Auction.Void, error) {

	s.replyCounter++

	log.Println("Received reply from a port")

	if s.replyCounter == len(s.Ports)-2 {
		s.EnterCriticalSection()
	}
	s.ServerTimestamp.SyncClocks(uint32(replyMessage.Timestamp))

	return &Auction.Void{}, nil
}

func (s *Server) EnterCriticalSection() {
	log.Println("Entered the critical section")
	s.state = HELD
	s.ServerTimestamp.Tick()

	s.CallRingElection(context.Background())

	s.ServerTimestamp.Tick()
}

func (s *Server) GetPort() int32 {
	return s.Port
}
