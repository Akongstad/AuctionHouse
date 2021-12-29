package server

import (
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
