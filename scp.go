package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func deliverBySCP(cfg Config, folder string) error {
	dststring := cfg.Scp_host + ":"
	if cfg.Scp_path != "" {
		dststring += strings.TrimRight(cfg.Scp_path, "/") + "/"
	}

	if cfg.Scp_user != "" {
		dststring = cfg.Scp_user + "@" + dststring
	}

	scpargs := []string{"-r", folder, dststring}

	if cfg.Scp_port != "" {
		scpargs = append([]string{"-P", cfg.Scp_port}, scpargs...)
	}

	cmd := exec.Command("scp", scpargs...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println(string(out))
	}

	return err
}
