package main

import (
	"bufio"
	"context"
	"flag"
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
	Connections   []*Connection //altså bruger vi overhovedet det her lort længere
	PrimePulse    time.Time
	Auction.UnimplementedAuctionHouseServer
	ServerTimestamp Auction.LamportClock
	lock            sync.Mutex
	Ports           []int32
	Port            int32
	state           STATE
	replyCounter    int
}

func (s *Server) Pulse() {
	for {
		time.Sleep(time.Second * 10)
		if s.Port == MainPort {
			s.ServerTimestamp.Tick()
			log.Printf("Prime replica: Pulse(%v)", s.ServerTimestamp.GetTime())
			go func() {
				for i := 0; i < len(s.Ports); i++ {
					if i != int(s.ID) {
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
		} else if math.Abs(float64(time.Now().Second()-s.PrimePulse.Second())) > 19 {
			s.ServerTimestamp.Tick()
			log.Printf("Missing pulse from prime replica. Last Pulse: %v seconds ago(%d)", math.Abs(float64(time.Now().Second()-s.PrimePulse.Second())), s.ServerTimestamp.GetTime())

			log.Printf("Leader election called")

			msg := Auction.RequestMessage{
				ServerId:  int32(s.ID),
				Timestamp: int32(s.ServerTimestamp.GetTime()),
			}

			s.AccessCritical(context.Background(), &msg)
			log.Printf("Fuck Rasmus")
		}
	}
}

func (s *Server) RegisterPulse(ctx context.Context, msg *Auction.Message) (*Auction.Void, error) {
	s.ServerTimestamp.SyncClocks(msg.GetTimestamp())
	log.Printf("Received pulse from prime replica(%d)", s.ServerTimestamp.GetTime())
	s.PrimePulse = time.Now()
	return &Auction.Void{}, nil

}

func (s *Server) ReplicateBackups(ctx context.Context, HighestBid int32, HighestBidder string) { //kan være vi skal return en samlet ack ye
	for i := 0; i < len(s.Ports); i++ {
		if i != int(s.ID) {
			conn, err := grpc.Dial(":"+strconv.Itoa(int(s.Ports[i])), grpc.WithInsecure())

			if err != nil {
				log.Fatalf("Failed to dial this port(Message all): %v", err)
			}

			defer conn.Close()
			client := Auction.NewAuctionHouseClient(conn)

			repMsg := &Auction.ReplicateMessage{Amount: s.HighestBid, User: &s.HighestBidder, Timestamp: s.ServerTimestamp.GetTime(),
				AuctionStart: s.StartTime.String()}
			client.Replicate(ctx, repMsg)
		}
	}
}

func (s *Server) Replicate(ctx context.Context, update *Auction.ReplicateMessage) (*Auction.BidReply, error) {
	s.HighestBid = update.Amount
	s.HighestBidder = *update.User
	Auctionstart, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", update.AuctionStart)
	if err != nil {
		log.Fatalf("Failed to parse starttime on replicate")
	}
	s.ServerTimestamp.Tick() //vi skal vel netop increment timestamp her, så vi til sidste kan tjekke hvilket replica som er nyest og så sync
	s.StartTime = Auctionstart
	return &Auction.BidReply{
		Timestamp:  s.ServerTimestamp.GetTime(),
		ReturnType: 1,
	}, nil
}

func (s *Server) CallRingElection(ctx context.Context) {

	listOfPorts := make([]int32, 0)

	listOfPorts = append(listOfPorts, s.Port)

	index := s.FindIndex(s.Port)

	nextPort := s.FindNextPort(index)

	conn, err := grpc.Dial(nextPort, grpc.WithInsecure())
	if err != nil {
		log.Printf("Election: Failed to dial this port: %v", err)
	} else {
		defer conn.Close()
		client := Auction.NewAuctionHouseClient(conn)

		protoListOfPorts := Auction.ElectionPorts{
			ListOfPorts: listOfPorts,
		}

		client.RingElection(ctx, &protoListOfPorts)
	}
}

func (s *Server) RingElection(ctx context.Context, msg *Auction.ElectionPorts) (*Auction.Void, error) {

	listOfPorts := msg.ListOfPorts

	if listOfPorts[0] == s.Port {
		//TODO: Tjek også lige lamport timestamps
		var highestPort int32
		for i := 0; i < len(listOfPorts); i++ {
			if listOfPorts[i] > highestPort {
				highestPort = listOfPorts[i]
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
				log.Printf("Election: Failed to select new leader", err)
			}

			indexOfMainPort := s.FindIndex(MainPort)

			//Kald alle andre ports med den nye liste
			for i := 0; i < len(newPortList.ListOfPorts); i++ {

				if i != indexOfMainPort {
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
		}
	} else {
		msg.ListOfPorts = append(msg.ListOfPorts, s.Port)

		//Call RingElection på alle andre
		index := s.FindIndex(s.Port)

		nextPort := s.FindNextPort(index)

		conn, err := grpc.Dial(nextPort, grpc.WithInsecure())
		if err != nil {
			log.Printf("Election: Failed to dial this port: %v", err)
		} else {
			defer conn.Close()
			client := Auction.NewAuctionHouseClient(conn)

			protoListOfPorts := Auction.ElectionPorts{
				ListOfPorts: listOfPorts,
			}

			client.RingElection(ctx, &protoListOfPorts)
		}
	}

	return &Auction.Void{}, nil
}

func (s *Server) BroadcastNewLeader(ctx context.Context, newPorts *Auction.ElectionPorts) (*Auction.Void, error) {
	s.Ports = newPorts.ListOfPorts
	return &Auction.Void{}, nil
}

func (s *Server) SelectNewLeader(ctx context.Context, void *Auction.Void) (*Auction.ElectionPorts, error) {

	updatedPort := s.FindIndex(s.Port)
	s.Ports = removePort(s.Ports, int32(updatedPort))
	s.Port = MainPort

	// TODO Skal man slette der hvor den lyttede før

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(int(s.Port)))
	if err != nil {
		log.Fatalf("Failed to listen port: %v", err)
	}
	log.Printf("New leader at: %v", lis.Addr())

	defer func() {
		lis.Close()
		log.Printf("Server stopped listening")
	}()

	//if err := grpcServer.Serve(lis); err != nil {
	//	log.Fatalf("Failed to serve server: %v", err)
	//}

	newPortList := Auction.ElectionPorts{
		ListOfPorts: s.Ports,
	}

	return &newPortList, nil
}

func (s *Server) Bid(ctx context.Context, bid *Auction.BidMessage) (*Auction.BidReply, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ServerTimestamp.SyncClocks(bid.Timestamp)
	if time.Now().Before(s.StartTime.Add(s.Duration)) {
		if s.HighestBid < bid.Amount {
			s.HighestBid = bid.Amount
			s.HighestBidder = *bid.User

			s.ReplicateBackups(ctx, s.HighestBid, s.HighestBidder.Name) //her kommer der en opdatering i state, så vi vil gerne replicate

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
	s.ServerTimestamp.SyncClocks(uint32(bid.Timestamp)) //hvorfor sync igen?
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

func (s *Server) OpenConnection(connect *Auction.Connect, stream Auction.AuctionHouse_OpenConnectionServer) error {
	conn := &Connection{
		stream: stream,
		active: true,
		id:     connect.User.Name + strconv.Itoa(int(connect.User.Id)),
		Name:   connect.User.Name,
		error:  make(chan error),
	}

	s.Connections = append(s.Connections, conn)
	s.ServerTimestamp.Tick()
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
	s.ServerTimestamp.Tick()
	leaveMessage := Auction.Message{
		User:      &Auction.User{Name: "Auction House"},
		Timestamp: s.ServerTimestamp.GetTime(),
	}

	log.Print("__________________________________")
	log.Printf("Auction House: %s has left the auction", msg.User.Name)
	s.Broadcast(context.Background(), &leaveMessage)

	return &Auction.Void{}, nil
}

func main() {
	id := flag.Int("I", -1, "id")
	flag.Parse()
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
		ID:                              *id,
		Connections:                     connections,
		UnimplementedAuctionHouseServer: Auction.UnimplementedAuctionHouseServer{},
		ServerTimestamp:                 Auction.LamportClock{},
		StartTime:                       time.Now(),
		Duration:                        time.Second * 100,
		lock:                            sync.Mutex{},
		Ports:                           ports,
		Port:                            ports[*id],
	}

	//go s.StartAuction(time.Second * 100)

	go s.Pulse()

	// If the file doesn't exist, create it or append to the file. For append functionality : os.O_APPEND
	filename := strconv.Itoa(s.ID) + "logs.txt"
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	//Create multiwriter
	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)

	grpcServer := grpc.NewServer()

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(int(s.Port)))
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

func (s *Server) FindNextPort(index int) string {

	nextPort := s.Ports[(index+1)%len(s.Ports)]
	if nextPort == MainPort {
		nextPort = s.Ports[(index+2)%len(s.Ports)]
	}

	return ":" + strconv.Itoa(int(nextPort))
}

func (s *Server) listenToNew(port int32) {
	grpcServer := grpc.NewServer()

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(int(port)))
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

func removeConnection(slice []*Connection, i int) []*Connection {
	slice[i] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}

func removePort(s []int32, i int32) []int32 {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

/*
func (s *Server) StartAuction(Duration time.Duration) {
	s.Duration = Duration
	a, b, c := s.StartTime.Clock()
	log.Printf("Auctioneer: Auction started at %d:%d:%d. On: %d. Duration will be: %v seconds", a, b, c, s.Port, s.Duration.Seconds())

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
*/
