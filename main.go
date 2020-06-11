package main

import (
	"context"
	"log"
	"net"

	"github.com/bigfatty/avoxi-test/checker"
	mmdb "github.com/oschwald/maxminddb-golang"
	"google.golang.org/grpc"
)

const (
	mmdbFile = "GeoLite2-Country.mmdb"
)

var (
	err error
	db  *mmdb.Reader
)

type ipCheckerServer struct{}

// CheckIp checks whether the IP address is in one of the countries
func (l *ipCheckerServer) CheckIp(ctx context.Context, ipMesg *checker.IP) (resp *checker.Response, err error) {
	isAllowed := false
	resp = &checker.Response{}
	resp.IsBlacklisted = true

	log.Printf("CheckIP: %#v", ipMesg)
	if isAllowed, err = lookup(ipMesg, resp); isAllowed && err == nil {
		log.Println("Not blacklisted: ", err)
		resp.IsBlacklisted = false
		return
	}
	log.Println("Not allowed: ", err)
	log.Print(ipMesg.Countries)
	return
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	db, err = mmdb.Open(mmdbFile)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Verify()
	if err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", "localhost:8081")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var ipServer = ipCheckerServer{}
	// Create an array of gRPC options with the credentials
	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)
	checker.RegisterIPCheckerServer(grpcServer, &ipServer)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	httpServer()

}

func testClient() {
	testIP := checker.IP{}
	ctx := context.Background()
	log.Print("dialing")
	conn, err := grpc.DialContext(ctx, "localhost:8081", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	log.Print("done dialing")
	client := checker.NewIPCheckerClient(conn)
	testIP.Ip = "1.2.3.4"
	testIP.Countries = map[string]bool{"au": true, "ca": true}
	resp, err := client.CheckIp(ctx, &testIP)
	if err != nil {
		log.Fatalf("CheckIp failed: %v", err)
	}
	log.Printf("Response: %+v", resp)
}
