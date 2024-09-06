package main

import (
	"context"
	"net/url"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog/log"

	"github.com/gentoomaniac/gocli"
	"github.com/gentoomaniac/logging"
	"github.com/gentoomaniac/run-puppet/pkg/runner"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
)

var (
	version = "unknown"
	commit  = "unknown"
	binName = "run-puppet"
	builtBy = "unknown"
	date    = "unknown"
)

var cli struct {
	logging.LoggingConfig

	BinPath       string   `help:"Path to puppet binary" default:"/opt/puppetlabs/bin/puppet"`
	LocalRepoPath string   `help:"local path for the checked out puppet manifests" default:"/var/lib/puppet-repo"`
	RemoteRepoUrl *url.URL `help:"puppet repository to use" default:"https://github.com/gentoomaniac/puppet.git"`
	VaultUrl      *url.URL `help:"url of the vault instance" default:"https://vault.srv.gentoomaniac.net"`
	RoleIdFile    *os.File `help:"path to the vault approle app id file" default:"/etc/vault_role_id"`
	SecretIdFile  *os.File `help:"path to the vault approle secret id file" default:"/etc/vault_secret_id"`
	PuppetBranch  string   `help:"puppet branch to use" default:"master"`

	Clone bool `help:"do a fresh clone of the manifest repository" default:"true" negatable:""`
	Now   bool `help:"skip the random delay" default:"false"`
	Noop  bool `short:"n" help:"don't apply puppet changes" default:"false"`

	Version gocli.VersionFlag `short:"V" help:"Display version."`
}

func main() {
	if _, err := os.Stat("/etc/puppet_disable"); err == nil {
		cli.Noop = true
		log.Info().Msg("`/etc/puppet_disable` found. Running with --noop.")
	}

	ctx := context.Background()
	tracerProvider := trace.NewTracerProvider()

	if cli.Debug {
		stdoutExporter, err := stdouttrace.New()
		if err != nil {
			panic("Failed get console exporter.")
		}
		tracerProvider = trace.NewTracerProvider(
			trace.WithBatcher(stdoutExporter,
				trace.WithBatchTimeout(time.Second)))
	}

	defer tracerProvider.Shutdown(ctx)
	otel.SetTracerProvider(tracerProvider)

	tracer := otel.Tracer("run-puppet")
	ctx, span := tracer.Start(context.Background(), "main")
	defer span.End()

	kctx := kong.Parse(&cli, kong.UsageOnError(), kong.Vars{
		"version": version,
		"commit":  commit,
		"binName": binName,
		"builtBy": builtBy,
		"date":    date,
	})
	kctx.Exit = func(code int) {
		tracerProvider.Shutdown(ctx)
		defer span.End()
	}

	logging.Setup(&cli.LoggingConfig)

	rp := runner.New(
		runner.WithBinPath(cli.BinPath),
		runner.WithLocalRepoPath(cli.LocalRepoPath),
		runner.WithRemoteRepoUrl(cli.RemoteRepoUrl),
		runner.WithPuppetBranch(cli.PuppetBranch),
		runner.WithClone(cli.Clone),
		runner.WithNow(cli.Now),
		runner.WithNoop(cli.Noop),
		runner.WithContext(ctx),
		runner.WithTracer(tracer),
		runner.WithVaultUrl(cli.VaultUrl),
		runner.WithRoleIdFile(cli.RoleIdFile),
		runner.WithSecretIdFile(cli.SecretIdFile),
	)

	code := rp.Run()
	kctx.Exit(code)
}
