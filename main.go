package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

func main() {
	config := GetConfig()
	fmt.Println(config)

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

	// set destination for all backups
	dst := "/tmp/"
	t := time.Now().UTC()
	datetime := fmt.Sprintf("%d-%02d-%02d_%02d-%02d-%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	dst += datetime + "/"
	err = os.Mkdir(dst, 0700)
	if err != nil {
		log.Fatalf("Could not create backup destination: %s", dst)
	}

	wipetmp := func() {
		err := os.RemoveAll(dst)
		if err != nil {
			log.Fatalf("Failed to remove archives at %s", dst)
		}
		log.Printf("Removed archives at %s", dst)
	}
	defer wipetmp()

	log.Printf("%d data volumes to backup in %s .", len(volumes), dst)

	for _, vol := range volumes {
		file, err := backupVolume(client, dst, vol)
		if err != nil {
			wipetmp()
			log.Fatal(err.Error())
		}
		log.Printf(" * %s", path.Base(file))
	}

	err = unpauseContainers(client, toPause)
	if err != nil {
		wipetmp()
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
