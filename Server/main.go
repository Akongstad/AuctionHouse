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

var wait *sync.WaitGroup

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
	queue           customQueue
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
		} else if time.Since(s.PrimePulse).Seconds() > time.Second.Seconds()*19 && s.state != WANTED && s.state != HELD {
			//s.ServerTimestamp.Tick()
			log.Printf("Missing pulse from prime replica. Last Pulse: %v seconds ago", time.Now().Sub(s.PrimePulse))

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

	localWG := new(sync.WaitGroup)

	for i := 0; i < len(s.Ports); i++ {
		if s.Ports[i] != s.Port {
			conn, err := grpc.Dial(":"+strconv.Itoa(int(s.Ports[i])), grpc.WithInsecure())

			if err != nil {
				log.Fatalf("Failed to dial this port(Message all): %v", err)
			}

			defer conn.Close()
			client := Auction.NewAuctionHouseClient(conn)

			repMsg := &Auction.ReplicateMessage{
				Amount: s.HighestBid, User: &s.HighestBidder, Timestamp: s.ServerTimestamp.GetTime(),
				AuctionStart: s.StartTime.String(),
			}

			var bidReply *Auction.BidReply
			go func() {
				bidReply, err = client.Replicate(ctx, repMsg)
				if err != nil {
					log.Fatalf("Could not replicate data to port :%d. Reply from replica: %v", s.Ports[i], bidReply.ReturnType)
				}
			}()
			localWG.Add(1)
			go func() {
				defer localWG.Done()
				time.Sleep(time.Second * 2)
				if bidReply == nil {
					//Der er intet svar. Gør et eller andet
					s.CallCutOfReplicate(ctx, s.Ports[i])
				}
			}()

		}
	}

	localWG.Wait()
}

func (s *Server) CallCutOfReplicate(ctx context.Context, brokenPort int32) {
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
	Auctionstart, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", update.AuctionStart)
	if err != nil {
		log.Fatalf("Failed to parse starttime on replicate")
	}
	s.ServerTimestamp.SyncClocks(update.Timestamp)
	s.StartTime = Auctionstart
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
			} else if listOfPorts[i] > highestPort {
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

			//jeg vil gerne ha en at replicateBackups returner en boolean, så vi kan give det korrekte returnType tilbage til client.
			s.ReplicateBackups(ctx, s.HighestBid, s.HighestBidder.Name)
			//så sætter vi ogs kun main replicas state til den nye, hvis det er gået igennem alle replicas? for at ha atomicity.
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

	queue := &customQueue{queue: make([]int, 0)}

	s := &Server{
		ID:                              *id,
		Connections:                     connections,
		UnimplementedAuctionHouseServer: Auction.UnimplementedAuctionHouseServer{},
		ServerTimestamp:                 Auction.LamportClock{},
		lock:                            sync.Mutex{},
		StartTime:                       time.Now(),
		Duration:                        time.Second*0,
		Ports:                           ports,
		Port:                            ports[*id],
		PrimePulse:                      time.Now(),
		state:                           RELEASED,
		queue:                           *queue,
	}

	go s.StartAuction(time.Second*120)
	go s.Pulse()

	// If the file doesn't exist, create it or append to the file. For append functionality : os.O_APPEND
	filename := "Server" + strconv.Itoa(s.ID) + "logs.txt"
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

func (s *Server) StartAuction(Duration time.Duration) {
	s.StartTime = time.Now()
	s.Duration = Duration
	a, b, c := s.StartTime.Clock()
	log.Printf("Auctioneer: Auction started at %d:%d:%d. On: %d. Duration will be: %v seconds", a, b, c, s.Port, s.Duration.Seconds())
	s.ReplicateBackups(context.Background(), 0, "No bids")
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

func (s *Server) AccessCritical(ctx context.Context, requestMessage *Auction.RequestMessage) (*Auction.ReplyMessage, error) {
	log.Printf("%d, Stamp: %d Requesting access to leader election", s.ID, s.ServerTimestamp.GetTime())
	log.Printf("----------------------------------------------------------")
	s.state = WANTED
	s.MessageAll(ctx, requestMessage)
	reply := Auction.ReplyMessage{Timestamp: int32(s.ServerTimestamp.GetTime()), ServerId: int32(s.ID), Port: s.Port}

	return &reply, nil
}

func (s *Server) MessageAll(ctx context.Context, msg *Auction.RequestMessage) error {
	for i := 0; i < len(s.Ports); i++ {
		if i != s.ID {
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
	log.Printf("%d received request from: %d", s.ID, requestMessage.ServerId)
	log.Printf("----------------------------------------------------------")

	s.ServerTimestamp.SyncClocks(uint32(requestMessage.Timestamp))

	void := Auction.Void{}

	if s.shouldDefer(requestMessage) {
		log.Printf("%d is not accepting request from: %d", s.ID, requestMessage.ServerId)
		log.Printf("----------------------------------------------------------")
		s.queue.Enqueue(int(requestMessage.Port))
		return &void, nil
	} else {
		log.Printf("%d is sending reply to: %d", s.ID, requestMessage.ServerId)
		log.Printf("----------------------------------------------------------")
		s.SendReply(requestMessage.Port)
		return &void, nil
	}
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

	if s.replyCounter == len(s.Ports)-2 {
		s.EnterCriticalSection()
	}
	s.ServerTimestamp.SyncClocks(uint32(replyMessage.Timestamp))

	return &Auction.Void{}, nil
}

func (s *Server) EnterCriticalSection() {
	s.state = HELD
	s.ServerTimestamp.Tick()
	log.Printf("%d entered the critical section(%v)", s.ID, s.ServerTimestamp.GetTime())
	log.Printf("----------------------------------------------------------")

	//UPDATE TIMESTAMP HER?

	s.CallRingElection(context.Background())

	s.LeaveCriticalSection()
}

func (s *Server) LeaveCriticalSection() {
	s.ServerTimestamp.Tick()
	log.Printf("%d is leaving the critical section(%v)", s.ID, s.ServerTimestamp.GetTime())
	log.Printf("----------------------------------------------------------")
	s.state = RELEASED
	//UPDATE TIMESTAMP HER?
	s.replyCounter = 0

	for !s.queue.Empty() {
		index := s.queue.Front()
		s.queue.Dequeue()
		log.Printf("%d: reply to defered request", s.ID)
		log.Printf("----------------------------------------------------------")
		s.CallSetStateReleased(index)
	}
}

func (s *Server) CallSetStateReleased(index int) {

	port := s.FindIndex(int32(index))
	conn, err := grpc.Dial(":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to port: %v", err)
	}
	defer conn.Close()
	client := Auction.NewAuctionHouseClient(conn)
	client.SetStateReleased(context.Background(), &Auction.Void{})
}

func (s *Server) SetStateReleased(ctx context.Context, void *Auction.Void) (*Auction.Void, error) {

	s.state = RELEASED

	return &Auction.Void{}, nil
}

/*
QUEUE
*/

type customQueue struct {
	queue []int
	lock  sync.RWMutex
}

func (c *customQueue) Enqueue(name int) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.queue = append(c.queue, name)
}

func (c *customQueue) Dequeue() {
	if len(c.queue) > 0 {
		c.lock.Lock()
		defer c.lock.Unlock()
		c.queue = c.queue[1:]
	}
}

func (c *customQueue) Front() int {
	if len(c.queue) > 0 {
		c.lock.Lock()
		defer c.lock.Unlock()
		return c.queue[0]
	}
	return -1
}

func (c *customQueue) Size() int {
	return len(c.queue)
}

func (c *customQueue) Empty() bool {
	return len(c.queue) == 0
}
