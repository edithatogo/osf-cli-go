package cli

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"osf-cli-go/internal/auth"
	"osf-cli-go/internal/osfapi"
)

const osfAPIBaseURL = "https://api.osf.io/v2/"

type readonlyClient interface {
	CurrentUser(context.Context) (osfapi.User, error)
	ListProjects(context.Context) ([]osfapi.Node, error)
	GetNode(context.Context, string) (osfapi.Node, error)
	ListNodeChildren(context.Context, string) ([]osfapi.Node, error)
	ListStorageFiles(context.Context, string, ...string) ([]osfapi.StorageFile, error)
	GetStorageFile(context.Context, string) (osfapi.StorageFile, error)
	OpenDownload(context.Context, string) (io.ReadCloser, error)
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

func (c *defaultReadonlyClient) CurrentUser(ctx context.Context) (osfapi.User, error) {
	if !c.bearerToken {
		return osfapi.User{}, auth.MissingTokenError{Env: auth.TokenEnv}
	}

	return c.api.CurrentUser(ctx)
}

func (c *defaultReadonlyClient) GetNode(ctx context.Context, id string) (osfapi.Node, error) {
	return c.api.GetNode(ctx, id)
}

func (c *defaultReadonlyClient) ListNodeChildren(ctx context.Context, id string) ([]osfapi.Node, error) {
	return c.api.ListNodeChildren(ctx, id)
}

func (c *defaultReadonlyClient) ListStorageFiles(ctx context.Context, id string, segments ...string) ([]osfapi.StorageFile, error) {
	return c.api.ListStorageFiles(ctx, id, segments...)
}

func (c *defaultReadonlyClient) GetStorageFile(ctx context.Context, id string) (osfapi.StorageFile, error) {
	return c.api.GetStorageFile(ctx, id)
}

func (c *defaultReadonlyClient) OpenDownload(ctx context.Context, downloadURL string) (io.ReadCloser, error) {
	return c.api.OpenDownload(ctx, downloadURL)
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
		if parsed.Host != "osf.io" && !strings.HasSuffix(parsed.Host, ".osf.io") {
			return "", fmt.Errorf("node url host %q is not an OSF host", parsed.Host)
		}
		for i, part := range parts {
			if part == "nodes" && i+1 < len(parts) {
				return parts[i+1], nil
			}
		}
		return parts[0], nil
	}

	if strings.Contains(trimmed, "://") {
		return "", fmt.Errorf("could not parse %q as a node url", input)
	}

	return trimmed, nil
}
