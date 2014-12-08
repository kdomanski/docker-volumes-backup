package main

import (
	"strings"

	"code.google.com/p/ftp4go"
)

func deliverToFTP(cfg Config, folder string) error {
	cli := ftp4go.NewFTP(0)
	_, err := cli.Connect(cfg.Ftp_host, ftp4go.DefaultFtpPort, cfg.Ftp_proxy)
	if err != nil {
		return err
	}

	defer cli.Quit()

	_, err = cli.Login(cfg.Ftp_user, cfg.Ftp_pass, "")
	if err != nil {
		return err
	}

	if strings.HasSuffix(folder, "/") {
		folder = folder[:len(folder)-1]
	}
	_, err = cli.UploadDirTree(folder, cfg.Ftp_path, 5, nil, nil)
	if err != nil {
		return err
	}

	return nil
}
