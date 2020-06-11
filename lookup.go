package main

import (
	"log"
	"net"
	"strings"

	"github.com/bigfatty/avoxi-test/checker"
)

// lookup looks up the ip address in the mmdb.  Then compares the country found with the whitelist
// to see if it's in there
// the whitelistMap keys are the ISO country code i.e. US, AU, CN
// 46.40.128.15 -> Syria
// 64.233.185.138 -> US
func lookup(ipMesg *checker.IP, resp *checker.Response) (isAllowed bool, err error) {
	ipString := ipMesg.GetIp()
	whitelistMap := ipMesg.Countries

	var record struct {
		Country struct {
			ISOCode string            `maxminddb:"iso_code"`
			Names   map[string]string `maxminddb:"names"`
		} `maxminddb:"country"`
	}
	ip, isValid := validateIP(ipString)
	if isValid == false {
		return
	}

	err = db.Lookup(ip, &record)
	if err != nil {
		log.Fatal(err)
	}
	resp.Country = record.Country.Names["en"]

	log.Printf("%+#v", record)
	if _, ok := whitelistMap[record.Country.ISOCode]; ok {
		return true, err
	}
	return
}

// validateIP returnds whether the ip is valid ipV4
func validateIP(ipString string) (ip net.IP, isValid bool) {
	if len(strings.Split(ipString, ".")) != 4 {
		return ip, false
	}
	return net.ParseIP(ipString), true
}
