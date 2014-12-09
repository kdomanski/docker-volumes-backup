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

	// set destination for all backups
	dst := "/tmp/"
	t := time.Now().UTC()
	datetime := fmt.Sprintf("%d-%02d-%02d_%02d-%02d-%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	dst += datetime + "/"
	err := os.Mkdir(dst, 0700)
	if err != nil {
		log.Fatalf("Could not create backup destination: %s", dst)
	}
	defer eraseFolder(dst)

	_, err = archiveVolumes(dst, config.keep_failed_container)
	if err != nil {
		log.Panicf("Failed to archive volumes: %s", err.Error())
	}

	err = deliverToFTP(config, dst)
	if err != nil {
		log.Panic("Failed uploading to FTP: %s", err.Error())
	}
	log.Println("FTP upload successful")
}

func eraseFolder(destination string) error {
	err := os.RemoveAll(destination)
	if err != nil {
		log.Printf("Failed to remove archives at %s", destination)
		return err
	}
	log.Printf("Removed archives at %s", destination)
	return nil
}

func archiveVolumes(destination string, keepFailed bool) ([]string, error) {
	// Init the client
	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		return nil, err
	}

	toPause, err := getRunningContainers(client)
	if err != nil {
		return nil, err
	}

	err = pauseContainers(client, toPause)
	if err != nil {
		return nil, err
	}
	log.Printf("Paused %d cointainers.\n", len(toPause))

	defer func() {
		err = unpauseContainers(client, toPause)
		if err != nil {
			log.Panic(err.Error())
		}
		log.Printf("Unpaused %d cointainers.\n", len(toPause))
	}()

	volumes := getDataVolumes(client)
	log.Printf("%d data volumes to backup in %s .", len(volumes), destination)

	files := make([]string, 0)

	for _, vol := range volumes {
		file, err := backupVolume(client, destination, vol, keepFailed)
		if err != nil {
			return files, err
		}
		files = append(files, file)
		log.Printf(" * %s", path.Base(file))
	}

	return files, nil
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
