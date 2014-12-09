package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
)

type Volume struct {
	container *docker.APIContainers
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

func backupVolume(cli *docker.Client, directory string, v Volume, keepFailed bool) (string, error) {
	filename := v.getBareFilename() + ".tar"
	path := "/backup/" + filename
	hostpath := directory + filename

	user := strconv.Itoa(os.Geteuid())

	//cmd := "tar zcvf " + path + " " + v.path
	//log.Println(cmd)

	vols := make(map[string]struct{})
	vols["/backup/"] = struct{}{}

	binds := []string{
		directory + ":/backup/",
	}

	conf := &docker.Config{
		Cmd:         []string{"tar", "cvf", path, v.path},
		Image:       "busybox",
		User:        user,
		VolumesFrom: v.container.ID,
		Volumes:     vols,
	}
	hostconf := &docker.HostConfig{
		Binds:       binds,
		VolumesFrom: []string{v.container.ID},
	}
	opts := docker.CreateContainerOptions{
		Config:     conf,
		HostConfig: hostconf,
	}

	cont, err := cli.CreateContainer(opts)
	if err != nil {
		return "", err
	}

	if keepFailed == false {
		defer cli.RemoveContainer(docker.RemoveContainerOptions{ID: cont.ID, Force: true})
	}

	err = cli.StartContainer(cont.ID, hostconf)
	if err != nil {
		return "", err
	}

	retcode, err := cli.WaitContainer(cont.ID)
	if err != nil {
		os.Remove(hostpath)
		return "", err
	}

	if retcode != 0 {
		os.Remove(hostpath)
		str := fmt.Sprintf("Failed to create volume backup - tar return code is %d", retcode)
		return "", errors.New(str)
	}

	fileinfo, err := os.Stat(hostpath)
	if err != nil {
		return "", err
	}

	if false == fileinfo.Mode().IsRegular() {
		str := fmt.Sprintf("Failed to create volume backup - created file %s is not regular", hostpath)
		return "", errors.New(str)
	}

	if keepFailed == true {
		cli.RemoveContainer(docker.RemoveContainerOptions{ID: cont.ID, Force: true})
	}

	return hostpath, nil
}

//func createVolumeFilename
