package vault

import (
	"context"
	"net/url"
	"time"

	vault "github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func GetToken(ctx context.Context, tracer trace.Tracer, vaultAddress *url.URL, roleId string, secretId string) (string, error) {
	ctx, span := tracer.Start(ctx, "vault.GetToken()")
	span.SetAttributes(attribute.String("vaultAddress", vaultAddress.String()))
	defer span.End()

	client, err := vault.New(
		vault.WithAddress(vaultAddress.String()),
		vault.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		return "", err
	}

	resp, err := client.Auth.AppRoleLogin(
		ctx,
		schema.AppRoleLoginRequest{
			RoleId:   roleId,
			SecretId: secretId,
		},
	)
	if err != nil {
		return "", err
	}

	return resp.Auth.ClientToken, nil
}
