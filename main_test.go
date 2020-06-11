package main

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/bigfatty/avoxi-test/checker"
	mmdb "github.com/oschwald/maxminddb-golang"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func initTestServer() {
	var ipServer = ipCheckerServer{}
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	checker.RegisterIPCheckerServer(s, &ipServer)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestIPChecker(t *testing.T) {
	initTestServer()
	db, err = mmdb.Open(mmdbFile)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Verify()
	if err != nil {
		log.Fatal(err)
	}

	testIP := checker.IP{}
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := checker.NewIPCheckerClient(conn)
	testIP.Ip = "1.2.3.4"
	testIP.Countries = map[string]bool{"au": true, "ca": true}
	resp, err := client.CheckIp(ctx, &testIP)
	if err != nil {
		t.Fatalf("CheckIp failed: %v", err)
	}
	log.Printf("Response: %+v", resp)

}
