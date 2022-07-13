package main

import (
	"io/ioutil"
	"log"

	"github.com/cmlight/authdelivery"
)
func main() {
	b, err := ioutil.ReadFile("testdata/basic_schain.json")
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}
	authdelivery.ParseBidRequest(b)
}
