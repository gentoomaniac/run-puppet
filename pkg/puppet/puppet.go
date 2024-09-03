package puppet

import (
	"fmt"
	"os"
	"os/exec"
	"path"
)

func RunPuppetApply(binPath string, manifestPath string, vaultToken string, noop bool, syslog bool) error {
	args := []string{"apply", "--config", path.Join(manifestPath, "puppet.conf"), "-vvvt", path.Join(manifestPath, "manifests/site.pp")}
	if noop {
		args = append(args, "--noop")
	}
	if syslog {
		args = append(args, "-l syslog")
	}
	cmd := exec.Command(binPath, args...)

	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("VAULT_TOKEN=%s", vaultToken))

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
