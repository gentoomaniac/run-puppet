package main

import (
	"github.com/alecthomas/kong"
	"github.com/rs/zerolog/log"

	"github.com/gentoomaniac/gocli"
	"github.com/gentoomaniac/logging"

	"github.com/gentoomaniac/run-puppet/pkg/runner"
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

	BinPath       string `help:"Path to puppet binary" default:"/opt/puppetlabs/bin/puppet"`
	LocalRepoPath string `help:"local path for the checked out puppet manifests" default:"/var/lib/puppet-repo"`
	RemoteRepoUrl string `help:"puppet repository to use" required:""`
	PuppetBranch  string `help:"puppet branch to use" default:"master"`

	Clone bool `help:"do a fresh clone of the manifest repository" default:"true" negatable:""`
	Now   bool `help:"skip the random delay" default:"false"`
	Noop  bool `short:"n" help:"don't apply puppet changes" default:"false"`

	Version gocli.VersionFlag `short:"V" help:"Display version."`
}

func main() {
	ctx := kong.Parse(&cli, kong.UsageOnError(), kong.Vars{
		"version": version,
		"commit":  commit,
		"binName": binName,
		"builtBy": builtBy,
		"date":    date,
	})
	logging.Setup(&cli.LoggingConfig)

	rp := runner.New(
		runner.WithBinPath(cli.BinPath),
		runner.WithLocalRepoPath(cli.LocalRepoPath),
		runner.WithRemoteRepoUrl(cli.RemoteRepoUrl),
		runner.WithPuppetBranch(cli.PuppetBranch),
		runner.WithClone(cli.Clone),
		runner.WithNow(cli.Now),
		runner.WithNoop(cli.Noop),
	)

	if err := rp.Run(); err != nil {
		log.Error().Err(err).Msg("puppet run failed")
		ctx.Exit(1)
	}

	ctx.Exit(0)
}
