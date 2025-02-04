// Package performance implments a test tool that opens a lot of connections to
// an autoupdate server and measures how long it takes to connect and receive
// some data.
//
// The autoupdate service has to be started with the redis backend.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"
)

const (
	// connections tells how many connection should be created to the server.
	connections = 5000

	// The url of the request.
	url = "http://localhost:9012/system/autoupdate?k=" + keyName

	// The addr of redis server.
	redisAddr = "localhost:6379"

	// The redis key where the updated keys should be written to.
	redisTopic = "ModifiedFields"

	// The name of the key that is requested by the clients and updated in
	// redis.
	keyName = "meeting/1/id"
)

func main() {
	keepOpen := flag.Bool("keep-open", false, "Keeps the connections open after the test succeeded.")
	flag.Parse()

	pool := newPool(redisAddr)

	// Create clients.
	clients := make([]*client, connections)
	for i := 0; i < connections; i++ {
		clients[i] = &client{}
	}

	// Connect test
	keys := make(chan string, connections)
	start := time.Now()
	for _, c := range clients {
		go func(c *client) {
			if err := c.connect(context.Background(), keys); err != nil {
				log.Fatalf("Can not connect client: %v", err)
			}
		}(c)
	}
	readClients(connections, keys)
	log.Printf("Connect %d clients took %d milliseconds.", connections, time.Since(start)/time.Millisecond)

	// Update one key
	start = time.Now()
	pool.sendKey(keyName)
	readClients(connections, keys)
	log.Printf("Send and Receive one key took %d milliseconds.", time.Since(start)/time.Millisecond)

	if *keepOpen {
		fmt.Println("Connections are kept open...")

		for {
			readClients(connections, keys)
			log.Println("Connections received data.")
		}
	}
}

// readClients reads count messages from the channel.
func readClients(count int, c <-chan string) {
	for i := 0; i < count; i++ {
		<-c
	}
}
