package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type serverMetadata struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Repository  struct {
		URL string `json:"url"`
	} `json:"repository"`
	Packages []struct {
		RegistryType string `json:"registryType"`
		Identifier   string `json:"identifier"`
		Transport    struct {
			Type string `json:"type"`
		} `json:"transport"`
		EnvironmentVariables []struct {
			Name       string `json:"name"`
			Format     string `json:"format"`
			IsSecret   bool   `json:"isSecret"`
			IsRequired bool   `json:"isRequired"`
		} `json:"environmentVariables"`
	} `json:"packages"`
}

type directorySubmissions struct {
	Canonical struct {
		ServerName          string `json:"serverName"`
		Title               string `json:"title"`
		ShortDescription    string `json:"shortDescription"`
		RepositoryURL       string `json:"repositoryUrl"`
		OfficialRegistryURL string `json:"officialRegistryUrl"`
		Version             string `json:"version"`
		Package             string `json:"package"`
		Transport           string `json:"transport"`
	} `json:"canonical"`
	Classification struct {
		Categories []string `json:"categories"`
		Keywords   []string `json:"keywords"`
		ReadOnly   bool     `json:"readOnly"`
		Auth       []string `json:"auth"`
	} `json:"classification"`
	Directories struct {
		Smithery struct {
			Status   string `json:"status"`
			Evidence struct {
				MCPURL string `json:"mcpUrl"`
			} `json:"evidence"`
		} `json:"smithery"`
	} `json:"directories"`
}

type mcpbManifest struct {
	Server struct {
		MCPConfig struct {
			Env map[string]string `json:"env"`
		} `json:"mcp_config"`
	} `json:"server"`
	Tools []struct {
		Name string `json:"name"`
	} `json:"tools"`
	UserConfig map[string]struct {
		Sensitive bool `json:"sensitive"`
	} `json:"user_config"`
}

type glamaMetadata struct {
	Schema      string   `json:"$schema"`
	Maintainers []string `json:"maintainers"`
}

func main() {
	if err := run(); err != nil {
		fail("%v", err)
	}
}

func run() error {
	var server serverMetadata
	if err := readJSON("server.json", &server); err != nil {
		return err
	}

	var submissions directorySubmissions
	if err := readJSON("registry/directory-submissions.json", &submissions); err != nil {
		return err
	}

	var manifest mcpbManifest
	if err := readJSON("packaging/mcpb/manifest.json", &manifest); err != nil {
		return err
	}

	var glama glamaMetadata
	if err := readJSON("glama.json", &glama); err != nil {
		return err
	}
	if err := checkEqual("glama.$schema", glama.Schema, "https://glama.ai/mcp/schemas/server.json"); err != nil {
		return err
	}
	if len(glama.Maintainers) != 1 || glama.Maintainers[0] != "edithatogo" {
		return fmt.Errorf("glama maintainers = %v, want [edithatogo]", glama.Maintainers)
	}

	if err := checkEqual("serverName", submissions.Canonical.ServerName, server.Name); err != nil {
		return err
	}
	if err := checkEqual("title", submissions.Canonical.Title, server.Title); err != nil {
		return err
	}
	if err := checkEqual("shortDescription", submissions.Canonical.ShortDescription, server.Description); err != nil {
		return err
	}
	if err := checkEqual("repositoryUrl", submissions.Canonical.RepositoryURL, server.Repository.URL); err != nil {
		return err
	}
	if err := checkEqual("version", submissions.Canonical.Version, server.Version); err != nil {
		return err
	}
	if len(server.Packages) != 1 {
		return fmt.Errorf("server.json packages count = %d, want 1", len(server.Packages))
	}
	pkg := server.Packages[0]
	if err := checkEqual("package", submissions.Canonical.Package, pkg.Identifier); err != nil {
		return err
	}
	if err := checkEqual("transport", submissions.Canonical.Transport, pkg.Transport.Type); err != nil {
		return err
	}
	if err := checkEqual("officialRegistryUrl", submissions.Canonical.OfficialRegistryURL, officialRegistryURL(server.Name)); err != nil {
		return err
	}
	if len(submissions.Classification.Categories) == 0 {
		return fmt.Errorf("directory categories must not be empty")
	}
	if len(submissions.Classification.Keywords) == 0 {
		return fmt.Errorf("directory keywords must not be empty")
	}
	if !submissions.Classification.ReadOnly {
		return fmt.Errorf("directory classification readOnly must be true")
	}
	if pkg.RegistryType != "oci" {
		return fmt.Errorf("server.json package registryType = %q, want %q", pkg.RegistryType, "oci")
	}

	wantAuth := map[string]bool{}
	for _, name := range pkg.EnvironmentVariables {
		if !name.IsSecret {
			return fmt.Errorf("server.json environment variable %q must be secret", name.Name)
		}
		wantAuth[name.Name] = true
	}
	for _, name := range submissions.Classification.Auth {
		if !wantAuth[name] {
			return fmt.Errorf("directory auth %q not present in server.json environmentVariables", name)
		}
	}
	for name := range wantAuth {
		found := false
		for _, got := range submissions.Classification.Auth {
			if got == name {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("server.json environment variable %q missing from directory auth list", name)
		}
	}

	if err := checkAuthSecrets(server, submissions, manifest); err != nil {
		return err
	}
	if err := checkMCPBToolSchemas(manifest); err != nil {
		return err
	}
	if err := checkSmitheryRoute(submissions); err != nil {
		return err
	}
	if err := checkRegistryReadme(server, pkg, submissions); err != nil {
		return err
	}
	return nil
}

func readJSON(path string, dst any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}

func checkEqual(field, got, want string) error {
	if got != want {
		return fmt.Errorf("%s = %q, want %q", field, got, want)
	}
	return nil
}

func checkAuthSecrets(server serverMetadata, submissions directorySubmissions, manifest mcpbManifest) error {
	for _, env := range server.Packages[0].EnvironmentVariables {
		switch env.Name {
		case "OSF_TOKEN", "OSF_USERNAME", "OSF_PASSWORD":
		default:
			continue
		}
		if _, ok := manifest.Server.MCPConfig.Env[env.Name]; !ok {
			return fmt.Errorf("MCPB mcp_config env missing %q", env.Name)
		}
		configName := configNameForEnv(env.Name)
		config, ok := manifest.UserConfig[configName]
		if !ok {
			return fmt.Errorf("MCPB user_config missing %q for env %q", configName, env.Name)
		}
		if !config.Sensitive {
			return fmt.Errorf("MCPB user_config %q must be sensitive", configName)
		}
	}
	for _, name := range submissions.Classification.Auth {
		switch name {
		case "OSF_TOKEN", "OSF_USERNAME", "OSF_PASSWORD":
		default:
			return fmt.Errorf("unexpected auth field %q in directory submissions", name)
		}
	}
	return nil
}

func configNameForEnv(env string) string {
	switch env {
	case "OSF_TOKEN":
		return "osf_token"
	case "OSF_USERNAME":
		return "osf_username"
	case "OSF_PASSWORD":
		return "osf_password"
	default:
		return ""
	}
}

func checkMCPBToolSchemas(manifest mcpbManifest) error {
	want := []string{
		"osf_whoami",
		"osf_projects_list",
		"osf_project_get",
		"osf_components_list",
		"osf_files_list",
		"osf_contributors_list",
		"osf_search",
		"osf_preprints_list",
		"osf_doi_resolve",
	}
	seen := map[string]bool{}
	for _, tool := range manifest.Tools {
		if !contains(want, tool.Name) {
			return fmt.Errorf("unexpected MCPB tool %q", tool.Name)
		}
		seen[tool.Name] = true
	}
	for _, name := range want {
		if !seen[name] {
			return fmt.Errorf("MCPB manifest missing tool %q", name)
		}
	}
	return nil
}

func checkSmitheryRoute(submissions directorySubmissions) error {
	if submissions.Directories.Smithery.Status != "published" {
		return fmt.Errorf("smithery status = %q, want published", submissions.Directories.Smithery.Status)
	}
	if submissions.Directories.Smithery.Evidence.MCPURL == "" {
		return fmt.Errorf("smithery published status requires mcpUrl evidence")
	}
	return nil
}

func checkRegistryReadme(server serverMetadata, pkg struct {
	RegistryType string `json:"registryType"`
	Identifier   string `json:"identifier"`
	Transport    struct {
		Type string `json:"type"`
	} `json:"transport"`
	EnvironmentVariables []struct {
		Name       string `json:"name"`
		Format     string `json:"format"`
		IsSecret   bool   `json:"isSecret"`
		IsRequired bool   `json:"isRequired"`
	} `json:"environmentVariables"`
}, submissions directorySubmissions) error {
	data, err := os.ReadFile("registry/README.md")
	if err != nil {
		return fmt.Errorf("read registry/README.md: %w", err)
	}
	content := string(data)
	for _, want := range []string{
		server.Name,
		pkg.Identifier,
		"Official MCP Registry",
		"docker build -f Dockerfile.mcp -t " + pkg.Identifier,
		"mcp-publisher publish",
	} {
		if !strings.Contains(content, want) {
			return fmt.Errorf("registry README missing %q", want)
		}
	}
	return nil
}

func officialRegistryURL(serverName string) string {
	return "https://registry.modelcontextprotocol.io/v0.1/servers?search=" + url.QueryEscape(serverName)
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func fail(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
