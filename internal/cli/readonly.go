package cli

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"osf-cli-go/internal/auth"
	"osf-cli-go/internal/osfapi"
)

const osfAPIBaseURL = "https://api.osf.io/v2/"

type readonlyClient interface {
	ListProjects(context.Context) ([]osfapi.Node, error)
	GetNode(context.Context, string) (osfapi.Node, error)
	ListNodeChildren(context.Context, string) ([]osfapi.Node, error)
	ListStorageFiles(context.Context, string) ([]osfapi.StorageFile, error)
}

type defaultReadonlyClient struct {
	api         *osfapi.Client
	bearerToken bool
}

func newDefaultReadonlyClient() readonlyClient {
	return newDefaultReadonlyClientFromSource(auth.EnvSource{})
}

func newDefaultReadonlyClientFromSource(source auth.Source) readonlyClient {
	token, err := auth.LoadToken(source)
	if err != nil {
		token = ""
	}

	api, err := osfapi.New(osfAPIBaseURL, osfapi.WithBearerToken(token))
	if err != nil {
		panic(err)
	}

	return &defaultReadonlyClient{
		api:         api,
		bearerToken: token != "",
	}
}

func (c *defaultReadonlyClient) ListProjects(ctx context.Context) ([]osfapi.Node, error) {
	if !c.bearerToken {
		return nil, auth.MissingTokenError{Env: auth.TokenEnv}
	}

	return c.api.ListCurrentUserProjects(ctx)
}

func (c *defaultReadonlyClient) GetNode(ctx context.Context, id string) (osfapi.Node, error) {
	return c.api.GetNode(ctx, id)
}

func (c *defaultReadonlyClient) ListNodeChildren(ctx context.Context, id string) ([]osfapi.Node, error) {
	return c.api.ListNodeChildren(ctx, id)
}

func (c *defaultReadonlyClient) ListStorageFiles(ctx context.Context, id string) ([]osfapi.StorageFile, error) {
	return c.api.ListStorageFiles(ctx, id)
}

func parseNodeIDOrURL(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", fmt.Errorf("missing node id or url")
	}

	parsed, err := url.Parse(trimmed)
	if err == nil && parsed.Scheme != "" && parsed.Host != "" {
		parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			return "", fmt.Errorf("could not find node id in %q", input)
		}
		for i, part := range parts {
			if part == "nodes" && i+1 < len(parts) {
				return parts[i+1], nil
			}
		}
		if parsed.Host == "osf.io" || strings.HasSuffix(parsed.Host, ".osf.io") {
			return parts[0], nil
		}
		return parts[len(parts)-1], nil
	}

	if strings.Contains(trimmed, "://") {
		return "", fmt.Errorf("could not parse %q as a node url", input)
	}

	return trimmed, nil
}
