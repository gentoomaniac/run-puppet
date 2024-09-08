package runner

import (
	"context"
	"math/rand/v2"
	"net/url"
	"os"
	"time"

	"github.com/gentoomaniac/run-puppet/pkg/puppet"
	"github.com/gentoomaniac/run-puppet/pkg/vault"

	git "github.com/go-git/go-git/v5"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	vaultIdByteSize = 37
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

func (r *RunPuppet) Run() int {
	log.Debug().Msg("runner")

	ctx, span := r.options.tracer.Start(r.options.ctx, "runner.Run()")
	defer span.End()

	if !r.options.now {
		delay(ctx, r.options.tracer)
	}

	if r.options.clone {
		err := cloneRepo(ctx, r.options.tracer, r.options.localRepoPath, r.options.remoteRepoUrl)
		if err != nil {
			log.Error().Err(err).Msg("failed cloning puppet repo")
			return -1
		}
	}

	vaultToken := getVaultToken(ctx, r.options.tracer, r.options.vaultUrl, r.options.vaultRoleIdFile, r.options.vaultSecretIdFile)

	code, err := puppet.RunPuppetApply(ctx, r.options.tracer, r.options.binPath, r.options.localRepoPath, vaultToken, r.options.noop, !r.options.now)
	if err != nil {
		log.Error().Err(err).Msg("executing puppet failed")
		return -1
	}

	return code
}

func getVaultToken(ctx context.Context, tracer trace.Tracer, vaultUrl *url.URL, appIdFile *os.File, secretIdFile *os.File) string {
	ctx, span := tracer.Start(ctx, "runner.getVaultToken()")
	span.SetAttributes(attribute.String("appIdFile", appIdFile.Name()))
	span.SetAttributes(attribute.String("secretIdFile", secretIdFile.Name()))
	defer span.End()

	appId := make([]byte, vaultIdByteSize)
	_, err := appIdFile.Read(appId)
	if err != nil {
		log.Panic().Err(err).Msg("faild reading vault AppID")
	}
	secretId := make([]byte, vaultIdByteSize)
	_, err = secretIdFile.Read(secretId)
	if err != nil {
		log.Panic().Err(err).Msg("faild reading vault SecretID")
	}
	vaultToken, err := vault.GetToken(ctx, tracer, vaultUrl, string(appId), string(secretId))
	if err != nil {
		log.Panic().Err(err).Msg("faild getting vault token")
	}

	return vaultToken
}

func delay(ctx context.Context, tracer trace.Tracer) {
	_, span := tracer.Start(ctx, "runner.delay()")
	defer span.End()

	delay := time.Duration(rand.IntN(5)) * time.Second
	span.SetAttributes(attribute.Int("secondsDelay", int(delay.Seconds())))
	time.Sleep(delay)
}

func cloneRepo(ctx context.Context, tracer trace.Tracer, localPath string, remoteUrl *url.URL) error {
	_, span := tracer.Start(ctx, "runner.cloneRepo()")
	defer span.End()
	span.SetAttributes(attribute.String("remoteUrl", remoteUrl.String()))

	if err := os.RemoveAll(localPath); err != nil {
		return err
	}

	_, err := git.PlainClone(localPath, false, &git.CloneOptions{
		URL:      remoteUrl.String(),
		Progress: os.Stdout,
	})
	if err != nil {
		return err
	}

	return nil
}
