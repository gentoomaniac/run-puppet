package puppet

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func RunPuppetApply(ctx context.Context, tracer trace.Tracer, binPath string, manifestPath string, vaultToken string, noop bool, syslog bool) (int, error) {
	_, span := tracer.Start(ctx, "puppet.RunPuppetApply()")
	defer span.End()
	span.SetAttributes(attribute.Bool("noop", noop))

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

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			span.SetStatus(codes.Ok, "executing puppet succeeded with non zero exit code")
			return exitErr.ExitCode(), nil
		}
		span.SetStatus(codes.Error, "parsing error failed")
		return -1, err
	}

	span.SetStatus(codes.Ok, "executing puppet succeeded")
	return 0, nil
}
