package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var taskPattern = regexp.MustCompile(`- \[([ x~])\] Task:`)

func main() {
	tracksRoot := filepath.Join("conductor", "tracks")
	entries, err := os.ReadDir(tracksRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "checkreviews: read tracks: %v\n", err)
		os.Exit(1)
	}

	var findings []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		trackDir := filepath.Join(tracksRoot, entry.Name())
		planPath := filepath.Join(trackDir, "plan.md")
		body, err := os.ReadFile(planPath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "checkreviews: read %s: %v\n", filepath.ToSlash(planPath), err)
			os.Exit(1)
		}

		complete, hasTasks := completedPlan(string(body))
		if !hasTasks || !complete {
			continue
		}

		ok, err := hasReviewEvidence(trackDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "checkreviews: inspect %s: %v\n", filepath.ToSlash(trackDir), err)
			os.Exit(1)
		}
		if !ok {
			findings = append(findings, filepath.ToSlash(planPath)+" is complete but has no phase review evidence")
		}
	}

	if len(findings) > 0 {
		fmt.Fprintln(os.Stderr, "completed Conductor tracks missing review evidence:")
		for _, finding := range findings {
			fmt.Fprintln(os.Stderr, "  "+finding)
		}
		os.Exit(1)
	}
}

func completedPlan(plan string) (complete bool, hasTasks bool) {
	matches := taskPattern.FindAllStringSubmatch(plan, -1)
	if len(matches) == 0 {
		return false, false
	}
	for _, match := range matches {
		hasTasks = true
		if match[1] != "x" {
			return false, true
		}
	}
	return true, hasTasks
}

func hasReviewEvidence(trackDir string) (bool, error) {
	entries, err := os.ReadDir(trackDir)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if strings.Contains(name, "review") && strings.HasSuffix(name, ".md") {
			return true, nil
		}
	}
	return false, nil
}
