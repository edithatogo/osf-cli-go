package main

import (
	"encoding/json"
	"fmt"
	"os"
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
		Identifier string `json:"identifier"`
		Transport  struct {
			Type string `json:"type"`
		} `json:"transport"`
		EnvironmentVariables []struct {
			Name string `json:"name"`
		} `json:"environmentVariables"`
	} `json:"packages"`
}

type directorySubmissions struct {
	Canonical struct {
		ServerName       string `json:"serverName"`
		Title            string `json:"title"`
		ShortDescription string `json:"shortDescription"`
		RepositoryURL    string `json:"repositoryUrl"`
		Version          string `json:"version"`
		Package          string `json:"package"`
		Transport        string `json:"transport"`
	} `json:"canonical"`
	Classification struct {
		Auth []string `json:"auth"`
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
		Name        string `json:"name"`
		InputSchema struct {
			Type       string `json:"type"`
			Properties map[string]struct {
				Type string `json:"type"`
			} `json:"properties"`
			Required []string `json:"required"`
		} `json:"inputSchema"`
	} `json:"tools"`
	UserConfig map[string]struct {
		Sensitive bool `json:"sensitive"`
	} `json:"user_config"`
}

func main() {
	var server serverMetadata
	mustReadJSON("server.json", &server)

	var submissions directorySubmissions
	mustReadJSON("registry/directory-submissions.json", &submissions)

	var manifest mcpbManifest
	mustReadJSON("packaging/mcpb/manifest.json", &manifest)

	check("serverName", submissions.Canonical.ServerName, server.Name)
	check("title", submissions.Canonical.Title, server.Title)
	check("shortDescription", submissions.Canonical.ShortDescription, server.Description)
	check("repositoryUrl", submissions.Canonical.RepositoryURL, server.Repository.URL)
	check("version", submissions.Canonical.Version, server.Version)

	if len(server.Packages) != 1 {
		fail("server.json packages count = %d, want 1", len(server.Packages))
	}
	pkg := server.Packages[0]
	check("package", submissions.Canonical.Package, pkg.Identifier)
	check("transport", submissions.Canonical.Transport, pkg.Transport.Type)

	wantAuth := map[string]bool{}
	for _, name := range pkg.EnvironmentVariables {
		wantAuth[name.Name] = true
	}
	for _, name := range submissions.Classification.Auth {
		if !wantAuth[name] {
			fail("directory auth %q not present in server.json environmentVariables", name)
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
			fail("server.json environment variable %q missing from directory auth list", name)
		}
	}

	checkAuthSecrets(server, submissions, manifest)
	checkMCPBToolSchemas(manifest)
	checkSmitheryRoute(submissions)
}

func mustReadJSON(path string, dst any) {
	data, err := os.ReadFile(path)
	if err != nil {
		fail("read %s: %v", path, err)
	}
	if err := json.Unmarshal(data, dst); err != nil {
		fail("parse %s: %v", path, err)
	}
}

func check(field, got, want string) {
	if got != want {
		fail("%s = %q, want %q", field, got, want)
	}
}

func checkAuthSecrets(server serverMetadata, submissions directorySubmissions, manifest mcpbManifest) {
	for _, env := range server.Packages[0].EnvironmentVariables {
		switch env.Name {
		case "OSF_TOKEN", "OSF_USERNAME", "OSF_PASSWORD":
		default:
			continue
		}
		if _, ok := manifest.Server.MCPConfig.Env[env.Name]; !ok {
			fail("MCPB mcp_config env missing %q", env.Name)
		}
		configName := configNameForEnv(env.Name)
		config, ok := manifest.UserConfig[configName]
		if !ok {
			fail("MCPB user_config missing %q for env %q", configName, env.Name)
		}
		if !config.Sensitive {
			fail("MCPB user_config %q must be sensitive", configName)
		}
	}
	for _, name := range submissions.Classification.Auth {
		switch name {
		case "OSF_TOKEN", "OSF_USERNAME", "OSF_PASSWORD":
		default:
			fail("unexpected auth field %q in directory submissions", name)
		}
	}
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

func checkMCPBToolSchemas(manifest mcpbManifest) {
	want := map[string]struct {
		properties []string
		required   []string
	}{
		"osf_whoami":            {properties: nil, required: nil},
		"osf_projects_list":     {properties: nil, required: nil},
		"osf_project_get":       {properties: []string{"id"}, required: []string{"id"}},
		"osf_components_list":   {properties: []string{"id"}, required: []string{"id"}},
		"osf_files_list":        {properties: []string{"id", "path"}, required: []string{"id"}},
		"osf_contributors_list": {properties: []string{"id"}, required: []string{"id"}},
	}
	seen := map[string]bool{}
	for _, tool := range manifest.Tools {
		spec, ok := want[tool.Name]
		if !ok {
			fail("unexpected MCPB tool %q", tool.Name)
		}
		seen[tool.Name] = true
		check("tool "+tool.Name+" inputSchema.type", tool.InputSchema.Type, "object")
		for _, property := range spec.properties {
			field, ok := tool.InputSchema.Properties[property]
			if !ok {
				fail("tool %s inputSchema missing property %q", tool.Name, property)
			}
			check("tool "+tool.Name+" property "+property+" type", field.Type, "string")
		}
		for property := range tool.InputSchema.Properties {
			if !contains(spec.properties, property) {
				fail("tool %s inputSchema has unexpected property %q", tool.Name, property)
			}
		}
		for _, required := range spec.required {
			if !contains(tool.InputSchema.Required, required) {
				fail("tool %s inputSchema missing required %q", tool.Name, required)
			}
		}
	}
	for name := range want {
		if !seen[name] {
			fail("MCPB manifest missing tool %q", name)
		}
	}
}

func checkSmitheryRoute(submissions directorySubmissions) {
	if submissions.Directories.Smithery.Status != "published" {
		fail("Smithery status = %q, want published", submissions.Directories.Smithery.Status)
	}
	if submissions.Directories.Smithery.Evidence.MCPURL == "" {
		fail("Smithery published status requires mcpUrl evidence")
	}
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
