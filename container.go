package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/samalba/dockerclient"
)

func isVolumeContainer(cont *dockerclient.Container) bool {
	truename, err := getTrueName(cont)
	if err != nil {
		log.Fatal(err.Error())
	}

	return strings.HasSuffix(truename, "-volumes")
}

func getTrueName(cont *dockerclient.Container) (string, error) {
	for _, name := range cont.Names {
		if strings.Count(name, "/") == 1 {
			return name, nil
		}
	}

	errmsg := fmt.Sprintf("No unaliased names for container %s", cont.Id)
	return "", errors.New(errmsg)
}

func getVolumes(d *dockerclient.DockerClient, cont *dockerclient.Container) []Volume {
	// get ContainerInfo by inspection
	info, err := d.InspectContainer(cont.Id)
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

func isRunning(d *dockerclient.DockerClient, cont *dockerclient.Container) bool {
	up := strings.HasPrefix(cont.Status, "Up ")
	paused := strings.HasSuffix(cont.Status, " (Paused)")
	return up && !paused
}

func unpause(d *dockerclient.DockerClient, containerIDs []string) error {
	errchan := make(chan error)
	okchan := make(chan int)
	var toUnPause int = 0

	for _, cont := range containerIDs {
		toUnPause++

		go func(id string) {
			err := d.UnpauseContainer(id)
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
