package main

import (
	"flag"

	"github.com/Jojojojodr/bonfirec2/pkg/client"
)

func main() {
	port := flag.String("port", "7777", "Port to run the client on")
	address := flag.String("address", "localhost", "Address to connect the client to")
	flag.Parse()

	// Start the client
	c := client.NewClient(*port, *address)
	c.Connect()
}