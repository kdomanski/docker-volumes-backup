package main

import (
	"log"
	"strings"

	"github.com/samalba/dockerclient"
)

type Volume struct {
	container *dockerclient.Container
	path      string
}

func (v *Volume) getBareFilename() string {
	str, err := getTrueName(v.container)
	if err != nil {
		log.Fatal(err.Error())
	}

	// remove leading backslash
	str = str[1:]

	// remove the "-volumes" suffix used to identify the container
	// as a data-only container in the first place
	if strings.HasSuffix(str, "-volumes") {
		str = str[0 : len(str)-7]
	}

	// Add the volume path with backslashes replaced by minuses.
	// Accidentally, this also give double minus
	// between the container name and the path.
	str += strings.Replace(v.path, "/", "-", -1)

	return str
}

func backupVolume(cont *dockerclient.DockerClient, containerID, volume string) {

}

//func createVolumeFilename
