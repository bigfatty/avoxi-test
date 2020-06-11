package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bigfatty/avoxi-test/checker"
	gin "github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func httpServer() {
	testIP := checker.IP{}
	ctx := context.Background()
	log.Print("dialing")
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("localhost:%s", grpcPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := checker.NewIPCheckerClient(conn)
	g := gin.Default()
	g.POST("/v1/:ip", func(ctx *gin.Context) {
		var countryMap map[string]bool
		var readBytes []byte
		if readBytes, err = ioutil.ReadAll(ctx.Request.Body); err == nil {
			log.Println("readBytes:\n", string(readBytes))
		}

		if err = json.Unmarshal(readBytes, &countryMap); err != nil {
			log.Println("error unmarshalling json:", err)
		}

		log.Print(countryMap)

		testIP.Ip = ctx.Param("ip")
		testIP.Countries = countryMap
		if response, err := client.CheckIp(ctx, &testIP); !response.IsBlacklisted && err == nil {
			ctx.JSON(http.StatusOK, gin.H{
				"result":  "Allowed",
				"country": response.Country,
			})
		} else {
			ctx.JSON(http.StatusForbidden, gin.H{
				"result":  "Not Allowed",
				"country": response.Country,
			})
		}
	})

	if err := g.Run(fmt.Sprintf(":%s", httpPort)); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}

}
