package utils

import (
	"log"
	"testing"
)

func TestDocker_CreateContainer(t *testing.T) {

	docker, err := GetDefaultDocker()
	if err != nil {
		log.Printf("get default docker error %v",err)
		return
	}

	id, err := docker.CreateContainer(&CreateContainerReq{
		Image: "hello-world",
	})

	if err != nil {
		log.Printf("create container error %v", err)
		return
	}

	err = docker.StartContainer(id)
	if err != nil {
		log.Printf("Start container error %v", err)
		return 		
	}

	err = docker.WaitContainer(id)
	if err != nil {
		log.Printf("wait container erro %v", err)
	}

	logs, err := docker.LogContainer(id)
	if err != nil {
		log.Printf("error when log container %v", err)
	}

	print(len(logs))
	docker.RemoveContainer(id)
}
