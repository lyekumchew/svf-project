package main

import (
	"log"
	"svf-project/config"
	"svf-project/database"
	"svf-project/minio"
	"svf-project/nsq"
	"svf-project/routes"
)

func main() {
	if err := config.Load("config/config.yaml"); err != nil {
		log.Println("Failed to load configuration")
	}

	db, err := database.Init()
	if err != nil {
		log.Println("Failed to init database", err)
		return
	}
	defer db.Close()

	_, err = minio.Init()
	if err != nil {
		log.Println("Failed to init minio client")
		return
	}

	// start consumer
	go nsq.Start()

	router := routes.Init()
	err = router.Run(config.Get().Server.Addr)
	if err != nil {
		log.Println("Failed to run router")
		return
	}
}
