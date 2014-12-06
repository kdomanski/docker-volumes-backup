package main

import (
	"fmt"
	"log"

	"github.com/samalba/dockerclient"
)

func main() {
	// Init the client
	docker, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	paused, err := pauseRunning(docker)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("Paused %d cointainers.\n", len(paused))

	volumes := getDataVolumes(docker)
	log.Printf("%d data volumes to backup.", len(volumes))

	err = unpause(docker, paused)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("Unpaused %d cointainers.\n", len(paused))
}

func pauseRunning(d *dockerclient.DockerClient) ([]string, error) {
	containers, err := d.ListContainers(false, false, "")
	if err != nil {
		log.Fatal(err)
	}

	errchan := make(chan error)
	idchan := make(chan string)
	var toPause int = 0

	for _, cont := range containers {
		if isRunning(d, &cont) {

			toPause++

			go func(id string) {
				err := d.PauseContainer(id)
				if err != nil {
					errchan <- err
				} else {
					idchan <- id
				}
			}(cont.Id)
		}
	}

	pausedIDs := make([]string, 0, 0)

	for {
		select {
		case id := <-idchan:
			toPause--
			pausedIDs = append(pausedIDs, id)
		case err := <-errchan:
			return nil, err
		default:
		}

		if toPause == 0 {
			return pausedIDs, nil
		}
	}
}

func getDataVolumes(d *dockerclient.DockerClient) []Volume {
	containers, err := d.ListContainers(true, false, "")
	if err != nil {
		log.Fatal(err)
	}

	vols := make([]Volume, 0, 0)

	for _, cont := range containers {
		if isVolumeContainer(&cont) {
			moreVols := getVolumes(d, &cont)
			vols = append(vols, moreVols...)
		}
	}

	return vols
}

func makeBackups(d *dockerclient.DockerClient, cont *dockerclient.Container) []string {
	info, err := d.InspectContainer(cont.Id)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println(info.Volumes)
	return make([]string, 0)
}
