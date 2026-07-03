package main

import (
	"flag"

	"github.com/Jojojojodr/bonfirec2/pkg/client"
)

func main() {
	port := flag.String("port", "7777", "Server port to connect to")
	address := flag.String("address", "localhost", "Address to connect the client to")
	localPort := flag.String("local-port", "10007", "Local client source port")
	flag.Parse()

	// Start the client
	c := client.NewClient(*port, *address, *localPort)
	c.Connect()
}