package main

import (
	"context"
	"net/url"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog/log"

	"github.com/gentoomaniac/run-puppet/pkg/cli"
	"github.com/gentoomaniac/run-puppet/pkg/logging"
	"github.com/gentoomaniac/run-puppet/pkg/runner"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
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

	OtelEndpointUrl *url.URL `help:"OTEl endpoint to send traces to" default:"http://tempo.srv.gentoomaniac.net:4318"`

	Version gocli.VersionFlag `short:"V" help:"Display version."`
}

func main() {
	kctx := kong.Parse(&cli, kong.UsageOnError(), kong.Vars{
		"version": version,
		"commit":  commit,
		"binName": binName,
		"builtBy": builtBy,
		"date":    date,
	})

	logging.Setup(&cli.LoggingConfig)

	ctx := context.Background()

	var exporter trace.SpanExporter
	if cli.Debug {
		log.Debug().Msg("Sending spans to stdout ...")
		exp, err := stdouttrace.New()
		if err != nil {
			log.Fatal().Err(err).Msg("faild creating stdout trace exporter")
		}
		exporter = exp
	} else {
		exp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(cli.OtelEndpointUrl.String()))
		if err != nil {
			log.Error().Err(err).Msg("faild creating http trace exporter")
		}

		exporter = exp
	}
	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter,
			trace.WithBatchTimeout(time.Second)))

	defer tracerProvider.Shutdown(ctx)
	otel.SetTracerProvider(tracerProvider)

	tracer := otel.Tracer("run-puppet")
	ctx, span := tracer.Start(context.Background(), "main")
	defer span.End()

	kctx.Exit = func(code int) {
		hostname, _ := os.Hostname()
		span.SetAttributes(attribute.String("hostname", hostname))
		span.SetAttributes(attribute.Int("exitCode", code))
		defer span.End()
		tracerProvider.Shutdown(ctx)
		os.Exit(code)
	}

	if _, err := os.Stat("/etc/puppet_disable"); err == nil {
		cli.Noop = true
		log.Warn().Msg("`/etc/puppet_disable` found. Running with --noop.")
		span.SetAttributes(attribute.Bool("puppetDisabled", true))
	}

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
