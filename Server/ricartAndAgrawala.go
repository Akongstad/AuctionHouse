package main

/*
import (
	"context"
	"log"
	"strconv"

	"github.com/Akongstad/AuctionHouse/Auction"
	"google.golang.org/grpc"
)

func (s *Server) AccessCritical(ctx context.Context, requestMessage *Auction.RequestMessage) (*Auction.ReplyMessage, error) {
	log.Printf("%d, Stamp: %d Requesting access to leader election", s.ID, s.ServerTimestamp.GetTime())
	log.Printf("----------------------------------------------------------")
	s.state = WANTED
	s.MessageAll(ctx, requestMessage)
	reply := Auction.ReplyMessage{Timestamp: int32(s.ServerTimestamp.GetTime()), ServerId: int32(s.ID)}

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

	// UPDATE TIMESTAMP HER?

	requestID := int(requestMessage.ServerId)
	void := Auction.Void{}

	if s.shouldDefer(requestMessage) {
		log.Printf("%d is not accepting request from: %d", s.ID, requestMessage.ServerId)
		log.Printf("----------------------------------------------------------")
		return &void, nil
	} else {
		log.Printf("%d is sending reply to: %d", s.ID, requestMessage.ServerId)
		log.Printf("----------------------------------------------------------")
		s.SendReply(requestID)
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

func (s *Server) SendReply(index int) {
	port := s.Ports[index]
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

	if s.replyCounter == len(s.Ports)-1 {
		s.EnterCriticalSection()
	}

	//UPDATE TIMESTAMP HER?

	return &Auction.Void{}, nil
}

func (s *Server) EnterCriticalSection() {
	s.state = HELD
	log.Printf("%d entered the critical section", s.ID)
	log.Printf("----------------------------------------------------------")

	//UPDATE TIMESTAMP HER?

	s.CallRingElection(context.Background())

	s.LeaveCriticalSection()
}

func (s *Server) LeaveCriticalSection() {
	log.Printf("%d is leaving the critical section", s.ID)
	log.Printf("----------------------------------------------------------")
	s.state = RELEASED
	//UPDATE TIMESTAMP HER?
	s.replyCounter = 0
}
*/
