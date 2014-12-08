package main

import (
	"log"

	docker "github.com/fsouza/go-dockerclient"
)

func main() {
	// Init the client
	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.Fatal(err.Error())
	}

	toPause, err := getRunningContainers(client)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = pauseContainers(client, toPause)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("Paused %d cointainers.\n", len(toPause))

	volumes := getDataVolumes(client)
	log.Printf("%d data volumes to backup.", len(volumes))

	err = unpauseContainers(client, toPause)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("Unpaused %d cointainers.\n", len(toPause))
}

func getRunningContainers(cli *docker.Client) ([]string, error) {
	containers, err := cli.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0)
	for _, cont := range containers {
		if isRunning(cont) {
			ids = append(ids, cont.ID)
		}
	}

	return ids, nil
}

func getDataVolumes(cli *docker.Client) []Volume {
	containers, err := cli.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		log.Fatal(err)
	}

	vols := make([]Volume, 0, 0)

	for _, cont := range containers {
		c := cont
		if isVolumeContainer(&c) {
			moreVols := getVolumes(cli, &c)
			vols = append(vols, moreVols...)
		}
	}

	return vols
}
