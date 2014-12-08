package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
)

func isVolumeContainer(cont *docker.APIContainers) bool {
	truename, err := getTrueName(cont)
	if err != nil {
		log.Fatal(err.Error())
	}

	return strings.HasSuffix(truename, "-volumes")
}

func getTrueName(cont *docker.APIContainers) (string, error) {
	for _, name := range cont.Names {
		if strings.Count(name, "/") == 1 {
			return name, nil
		}
	}

	errmsg := fmt.Sprintf("No unaliased names for container %s", cont.ID)
	return "", errors.New(errmsg)
}

func getVolumes(cli *docker.Client, cont *docker.APIContainers) []Volume {
	// get ContainerInfo by inspection
	info, err := cli.InspectContainer(cont.ID)
	if err != nil {
		log.Fatal(err.Error())
	}

	// iterate over volume map to get keys into an array
	vols := make([]Volume, 0, len(info.Volumes))
	for k, _ := range info.Volumes {
		v := Volume{container: cont, path: k}
		vols = append(vols, v)
	}

	return vols
}

func isRunning(cont docker.APIContainers) bool {
	up := strings.HasPrefix(cont.Status, "Up ")
	paused := strings.HasSuffix(cont.Status, " (Paused)")
	return up && !paused
}

func unpauseContainers(cli *docker.Client, containerIDs []string) error {
	errchan := make(chan error)
	okchan := make(chan int)
	var toUnPause int = 0

	for _, cont := range containerIDs {
		toUnPause++

		go func(id string) {
			err := cli.UnpauseContainer(id)
			if err != nil {
				errchan <- err
			} else {
				okchan <- 1
			}
		}(cont)
	}

	for {
		select {
		case _ = <-okchan:
			toUnPause--
		case err := <-errchan:
			return err
		default:
		}

		if toUnPause == 0 {
			return nil
		}
	}
}

func pauseContainers(cli *docker.Client, containerIDs []string) error {
	errchan := make(chan error)
	okchan := make(chan int)
	var toPause int = 0

	for _, cont := range containerIDs {
		toPause++

		go func(id string) {
			err := cli.PauseContainer(id)
			if err != nil {
				errchan <- err
			} else {
				okchan <- 1
			}
		}(cont)
	}

	for {
		select {
		case _ = <-okchan:
			toPause--
		case err := <-errchan:
			return err
		default:
		}

		if toPause == 0 {
			return nil
		}
	}
}
