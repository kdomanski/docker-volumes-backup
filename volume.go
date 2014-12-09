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

	binds := []string{
		directory + ":/backup/",
	}

	err := volumeBusybox(cli, []string{"tar", "cvf", path, v.path}, []string{v.container.ID}, binds, keepFailed)
	if err != nil {
		return "", err
	}

	err = volumeBusybox(cli, []string{"chown", user, path}, nil, binds, keepFailed)
	if err != nil {
		return "", err
	}

	fileinfo, err := os.Stat(hostpath)
	if err != nil {
		return "", err
	}

	if false == fileinfo.Mode().IsRegular() {
		str := fmt.Sprintf("Failed to create volume backup - created file %s is not regular", hostpath)
		return "", errors.New(str)
	}

	return hostpath, nil
}

func volumeBusybox(cli *docker.Client, cmd []string, volumesFrom []string, volumes []string, keepFailed bool) error {

	vols := make(map[string]struct{})

	for _, v := range volumes {
		point := strings.Split(v, ":")[1]
		vols[point] = struct{}{}
	}

	conf := &docker.Config{
		Cmd:     cmd,
		Image:   "busybox",
		Volumes: vols,
	}
	hostconf := &docker.HostConfig{
		Binds:       volumes,
		VolumesFrom: volumesFrom,
	}
	opts := docker.CreateContainerOptions{
		Config:     conf,
		HostConfig: hostconf,
	}

	cont, err := cli.CreateContainer(opts)
	if err != nil {
		return err
	}

	if keepFailed == false {
		defer cli.RemoveContainer(docker.RemoveContainerOptions{ID: cont.ID, Force: true})
	}

	err = cli.StartContainer(cont.ID, hostconf)
	if err != nil {
		return err
	}

	retcode, err := cli.WaitContainer(cont.ID)
	if err != nil {
		return err
	}

	if retcode != 0 {
		str := fmt.Sprintf("Command failed with code %d", retcode)
		return errors.New(str)
	}

	if keepFailed == true {
		cli.RemoveContainer(docker.RemoveContainerOptions{ID: cont.ID, Force: true})
	}

	return nil
}

//func createVolumeFilename
