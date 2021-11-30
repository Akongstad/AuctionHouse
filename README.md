# AuctionHouse
Implemented by kong, carbr, nihj and garb
To run the program do the following:
- Open 3 terminal windows
- Terminal 1 Enter : "go run Server/main.go -I 0
- Terminal 2 Enter : "go run Server/main.go -I 1
- Terminal 3 Enter : "go run Server/main.go -I 2
This will start an auction with a default duration of 200 second. To change the duration add the additional terminal flag 
"-D (seconds)"

Open an arbitrary amount of terminal windows for the clients. Use the folowing format:
- Terminal Enter: "go run Client/main.go -U Username"
Client can now bid
- Terminal Enter: "bid (amount)"
Client can access the state of the auction
- Terminal Enter: "result"

If a single server replica crashes the system will survive.