package runner

import (
	"context"
	"math/rand/v2"
	"net/url"
	"os"
	"time"

	"github.com/gentoomaniac/run-puppet/pkg/puppet"
	"github.com/gentoomaniac/run-puppet/pkg/vault"
	"go.opentelemetry.io/otel/codes"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	vaultIdByteSize    = 37
	randomDelaySeconds = 5 * 60
)

type RunPuppet struct {
	options runPuppetOptions
}

type runPuppetOptions struct {
	binPath           string
	localRepoPath     string
	remoteRepoUrl     *url.URL
	vaultUrl          *url.URL
	vaultRoleIdFile   *os.File
	vaultSecretIdFile *os.File
	puppetBranch      string
	clone             bool
	now               bool
	noop              bool
	ctx               context.Context
	tracer            trace.Tracer
}

type RunPuppetOption func(o *runPuppetOptions)

func WithBinPath(path string) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.binPath = path
	}
}

func WithLocalRepoPath(path string) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.localRepoPath = path
	}
}

func WithRemoteRepoUrl(url *url.URL) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.remoteRepoUrl = url
	}
}

func WithPuppetBranch(branch string) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.puppetBranch = branch
	}
}

func WithClone(clone bool) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.clone = clone
	}
}

func WithNow(now bool) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.now = now
	}
}

func WithNoop(noop bool) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.noop = noop
	}
}

func WithContext(ctx context.Context) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.ctx = ctx
	}
}

func WithTracer(tracer trace.Tracer) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.tracer = tracer
	}
}

func WithVaultUrl(url *url.URL) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.vaultUrl = url
	}
}

func WithRoleIdFile(file *os.File) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.vaultRoleIdFile = file
	}
}

func WithSecretIdFile(file *os.File) RunPuppetOption {
	return func(o *runPuppetOptions) {
		o.vaultSecretIdFile = file
	}
}

func New(opts ...RunPuppetOption) RunPuppet {
	opt := &runPuppetOptions{
		puppetBranch:  "main",
		localRepoPath: "/var/lib/puppet-repo",
	}

	for _, o := range opts {
		o(opt)
	}

	return RunPuppet{
		options: *opt,
	}
}

func (r *RunPuppet) Run() (int, error) {
	ctx, span := r.options.tracer.Start(r.options.ctx, "runner.Run()")
	defer span.End()

	if !r.options.now {
		delay(ctx, r.options.tracer)
	}

	if r.options.clone {
		err := cloneRepo(ctx, r.options.tracer, r.options.localRepoPath, r.options.remoteRepoUrl, r.options.puppetBranch)
		if err != nil {
			span.SetStatus(codes.Error, "failed cloning repo")
			log.Error().Err(err).Msg("failed cloning puppet repo")
			return -1, err
		}
	}

	vaultToken := getVaultToken(ctx, r.options.tracer, r.options.vaultUrl, r.options.vaultRoleIdFile, r.options.vaultSecretIdFile)

	code, err := puppet.RunPuppetApply(ctx, r.options.tracer, r.options.binPath, r.options.localRepoPath, vaultToken, r.options.noop, !r.options.now)
	if err != nil {
		span.SetStatus(codes.Error, "failed reading vault SecretID")
		log.Error().Err(err).Msg("executing puppet failed")
		return -1, err
	}

	span.SetStatus(codes.Ok, "puppet run finished")

	return code, nil
}

func getVaultToken(ctx context.Context, tracer trace.Tracer, vaultUrl *url.URL, appIdFile *os.File, secretIdFile *os.File) string {
	ctx, span := tracer.Start(ctx, "runner.getVaultToken()")
	span.SetAttributes(attribute.String("vaultAppIdFile", appIdFile.Name()))
	span.SetAttributes(attribute.String("vaultSecretIdFile", secretIdFile.Name()))
	defer span.End()

	appId := make([]byte, vaultIdByteSize)
	_, err := appIdFile.Read(appId)
	if err != nil {
		span.SetStatus(codes.Error, "failed reading vault AppID")
		log.Panic().Err(err).Msg("failed reading vault AppID")
	}
	secretId := make([]byte, vaultIdByteSize)
	_, err = secretIdFile.Read(secretId)
	if err != nil {
		span.SetStatus(codes.Error, "failed reading vault SecretID")
		log.Panic().Err(err).Msg("failed reading vault SecretID")
	}
	vaultToken, err := vault.GetToken(ctx, tracer, vaultUrl, string(appId), string(secretId))
	if err != nil {
		span.SetStatus(codes.Error, "failed getting vault token")
		log.Panic().Err(err).Msg("failed getting vault token")
	}

	span.SetStatus(codes.Ok, "got vault token")

	return vaultToken
}

func delay(ctx context.Context, tracer trace.Tracer) {
	_, span := tracer.Start(ctx, "runner.delay()")
	defer span.End()

	delay := time.Duration(rand.IntN(randomDelaySeconds)) * time.Second
	span.SetAttributes(attribute.Int("delaySeconds", int(delay.Seconds())))
	time.Sleep(delay)
	span.SetStatus(codes.Ok, "delay finished")
}

func cloneRepo(ctx context.Context, tracer trace.Tracer, localPath string, remoteUrl *url.URL, branch string) error {
	_, span := tracer.Start(ctx, "runner.cloneRepo()")
	defer span.End()
	span.SetAttributes(attribute.String("gitRemoteUrl", remoteUrl.String()))
	span.SetAttributes(attribute.String("gitBranch", branch))

	if err := os.RemoveAll(localPath); err != nil {
		span.SetStatus(codes.Error, "failed removing local copy")
		return err
	}

	_, err := git.PlainClone(localPath, false, &git.CloneOptions{
		URL:           remoteUrl.String(),
		Progress:      os.Stdout,
		Depth:         1,
		SingleBranch:  true,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		span.SetStatus(codes.Error, "clone repo failed")
		return err
	}

	span.SetStatus(codes.Ok, "clone repo finished")

	return nil
}
