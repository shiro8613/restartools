package main

import (
	"context"
	"log"
	"os"
	"os/exec"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func main() {
	args := os.Args[1:]
	if len(args) != 2 {
		log.Println("cmd [container_name] [comnnad]")
		os.Exit(0)
	}
	containerName := args[0]
	commandToExecute := args[1]
	if containerName == "" || commandToExecute == "" {
		log.Fatalln("args is null")
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	filterArgs := filters.NewArgs()
	filterArgs.Add("type", "container")
	filterArgs.Add("event", "restart")
	filterArgs.Add("name", containerName)

	log.Printf("Monitoring container '%s' for restart events...\n", containerName)

	msgs, errs := cli.Events(context.Background(), events.ListOptions{
		Filters: filterArgs,
	})

	for {
		select {
		case err := <-errs:
			if err != nil {
				log.Fatalf("Error from event stream: %v\n", err)
			}
		case msg := <-msgs:
			log.Printf("Detected restart event for container '%s' at %d\n", msg.Actor.Attributes["name"], msg.Time)
			
			log.Printf("Executing command: %s\n", commandToExecute)
			cmd := exec.Command("sh", "-c", commandToExecute)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Printf("Error executing command: %v\n", err)
			}
			log.Println("Command execution finished.")
		}
	}
}