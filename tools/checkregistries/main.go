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
}

func main() {
	var server serverMetadata
	mustReadJSON("server.json", &server)

	var submissions directorySubmissions
	mustReadJSON("registry/directory-submissions.json", &submissions)

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

func fail(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
