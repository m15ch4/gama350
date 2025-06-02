package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"micze.io/gama350/api"
	"micze.io/gama350/config"
	"micze.io/gama350/influx"
	"micze.io/gama350/logger"
	"micze.io/gama350/mqtt"
)

func main() {
	config.LoadEnv()
	logger.InitLogger()

	influxService := influx.NewService()
	defer influxService.Close()

	mqttClient := mqtt.NewClient(influxService)
	mqttClient.Start()
	defer mqttClient.Stop()

	apiServer := api.NewAPIServer(influxService.Client())
	apiServer.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down gracefully...")
}
