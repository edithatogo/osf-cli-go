package cli

import (
	"fmt"
	"strings"

	"github.com/edithatogo/osf-cli-go/internal/osfapi"
	"github.com/edithatogo/osf-cli-go/internal/output"
	"github.com/spf13/cobra"
)

type validationFinding struct {
	Rule     string `json:"rule"`
	Status   string `json:"status"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type validationReport struct {
	NodeID   string              `json:"node_id"`
	Profile  string              `json:"profile"`
	Valid    bool                `json:"valid"`
	Findings []validationFinding `json:"findings"`
}

func newValidateCommand(client readonlyClient) *cobra.Command {
	var profile string
	cmd := &cobra.Command{
		Use:   "validate <guid-or-url>",
		Short: "Check OSF metadata against a deterministic research profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputMode, err := resolveOutputMode(cmd)
			if err != nil {
				return err
			}
			profile = strings.ToLower(strings.TrimSpace(profile))
			if profile != "research-output" && profile != "preregistration" {
				return fmt.Errorf("unsupported validation profile %q (want research-output or preregistration)", profile)
			}

			id, err := parseNodeIDOrURL(args[0])
			if err != nil {
				return err
			}
			node, err := client.GetNode(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("get node: %w", err)
			}
			contributors, err := client.ListNodeContributors(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("list contributors: %w", err)
			}
			files, err := client.ListStorageFiles(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("list storage files: %w", err)
			}

			report := buildValidationReport(node, len(contributors), len(files), profile)
			if outputMode == outputModeJSON {
				return output.WriteJSON(cmd.OutOrStdout(), report)
			}

			rows := make([][]string, 0, len(report.Findings))
			for _, finding := range report.Findings {
				rows = append(rows, []string{finding.Rule, finding.Status, finding.Severity, finding.Message})
			}
			return output.WriteTable(cmd.OutOrStdout(), []string{"RULE", "STATUS", "SEVERITY", "MESSAGE"}, rows)
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "research-output", "validation profile: research-output or preregistration")
	return cmd
}

func buildValidationReport(node osfapi.Node, contributorCount, fileCount int, profile string) validationReport {
	report := validationReport{NodeID: node.ID, Profile: profile, Valid: true, Findings: make([]validationFinding, 0, 5)}
	addFinding := func(rule, status, severity, message string) {
		report.Findings = append(report.Findings, validationFinding{Rule: rule, Status: status, Severity: severity, Message: message})
		if status == "fail" {
			report.Valid = false
		}
	}

	if strings.TrimSpace(node.Attributes.Title) == "" {
		addFinding("node.title_present", "fail", "error", "OSF node title is required")
	} else {
		addFinding("node.title_present", "pass", "error", "OSF node has a title")
	}
	if strings.TrimSpace(node.Attributes.Description) == "" {
		addFinding("node.description_present", "warn", "warning", "OSF node has no description")
	} else {
		addFinding("node.description_present", "pass", "warning", "OSF node has a description")
	}
	if contributorCount == 0 {
		addFinding("node.contributors_present", "warn", "warning", "OSF node has no contributors")
	} else {
		addFinding("node.contributors_present", "pass", "warning", fmt.Sprintf("OSF node has %d contributor(s)", contributorCount))
	}

	if profile == "research-output" {
		if fileCount == 0 {
			addFinding("research_output.storage_present", "warn", "warning", "OSF node has no storage files")
		} else {
			addFinding("research_output.storage_present", "pass", "warning", fmt.Sprintf("OSF node has %d storage entr(y/ies)", fileCount))
		}
	} else if strings.EqualFold(node.Attributes.Category, "registration") {
		addFinding("preregistration.category", "pass", "error", "OSF node category is registration")
	} else {
		addFinding("preregistration.category", "fail", "error", "OSF node category is not registration")
	}

	return report
}
