package main

import (
	"fmt"
	"goxfer/server/cmd/db"
	"goxfer/server/cmd/server"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Welcome to goXfer! Starting Server...")

	flowChan := make(chan os.Signal, 1)
	signal.Notify(flowChan, syscall.SIGINT, syscall.SIGTERM)

	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
		return
	}

	dataStore, err := db.NewDataStore()
	if err != nil {
		fmt.Println(err)
		log.Println(err)
		return
	}

	err = server.InitHTTPServer(dataStore)
	if err != nil {
		fmt.Println(err)
		return
	}

	<-flowChan

	fmt.Println("shutting down")

	if err := db.Close(dataStore); err != nil {
		fmt.Println(err)
	}
}
