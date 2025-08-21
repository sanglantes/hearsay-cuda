package main

import (
	"context"
	"hearsay/internal/config"
	"hearsay/internal/core"
	"hearsay/internal/storage"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Println("hearsay is starting...")

	configPath := "config.yaml"
	log.Printf("Reading config from %s.\n", configPath)
	err := config.ReadConfig(configPath, true)
	if err != nil {
		log.Fatalln("Failed to load configuration.")
	} else {
		log.Println("Successfully loaded configuration.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	db, err := storage.InitDatabase()
	if err != nil {
		log.Fatalln("Failed DB init.")
	} else {
		log.Println("Passed DB init.")
	}

	defer db.Close()

	/*err = data.ImportLogs(db, "data/logs.txt")
	if err != nil {
		fmt.Println("err returned", err.Error())
	}
	os.Exit(0)*/

	if err = storage.LoadOptIns(db); err != nil {
		log.Fatalf("Failed loading opt-out map: %s\n", err.Error())
	} else {
		log.Println("Passed opt-out loading.")
	}

	serverDisconnect := make(chan struct{})
	go func() {
		core.HearsayConnect(config.Server, config.Channel, ctx, db)
		close(serverDisconnect)
	}()

	select {
	case <-sigs:
		log.Println("Termination signal received. Shutting down...")
		cancel()
		time.Sleep(2 * time.Second)

	case <-serverDisconnect:
		os.Exit(1)
	}

}
