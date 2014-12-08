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

	paused, err := pauseRunning(client)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("Paused %d cointainers.\n", len(paused))

	volumes := getDataVolumes(client)
	log.Printf("%d data volumes to backup.", len(volumes))

	err = unpause(client, paused)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("Unpaused %d cointainers.\n", len(paused))
}

func pauseRunning(cli *docker.Client) ([]string, error) {
	containers, err := cli.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		log.Fatal(err)
	}

	errchan := make(chan error)
	idchan := make(chan string)
	var toPause int = 0

	for _, cont := range containers {
		if isRunning(&cont) {

			toPause++

			go func(id string) {
				err := cli.PauseContainer(id)
				if err != nil {
					errchan <- err
				} else {
					idchan <- id
				}
			}(cont.ID)
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

func getDataVolumes(cli *docker.Client) []Volume {
	containers, err := cli.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		log.Fatal(err)
	}

	vols := make([]Volume, 0, 0)

	for _, cont := range containers {
		if isVolumeContainer(&cont) {
			moreVols := getVolumes(cli, &cont)
			vols = append(vols, moreVols...)
		}
	}

	return vols
}
