package cli

import (
	"fmt"

	"github.com/edithatogo/osf-cli-go/internal/output"
	"github.com/spf13/cobra"
)

// ExportData holds a full snapshot of a node's data.
type ExportData struct {
	Node               projectRecord   `json:"node"`
	Contributors       []projectRecord `json:"contributors,omitempty"`
	Files              []fileRecord    `json:"files,omitempty"`
	Components         []projectRecord `json:"components,omitempty"`
}

func newExportCommand(client readonlyClient) *cobra.Command {
	return &cobra.Command{
		Use:   "export <guid-or-url>",
		Short: "Export a node snapshot with metadata, contributors, files, and components",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputMode, err := resolveOutputMode(cmd)
			if err != nil {
				return err
			}

			id, err := parseNodeIDOrURL(args[0])
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			data := ExportData{}

			node, err := client.GetNode(ctx, id)
			if err != nil {
				return fmt.Errorf("get node: %w", err)
			}
			data.Node = projectRecord{
				ID:          node.ID,
				Title:       node.Attributes.Title,
				Category:    node.Attributes.Category,
				Description: node.Attributes.Description,
				URL:         node.Links.Self,
			}

			contributors, err := client.ListNodeContributors(ctx, id)
			if err == nil {
				data.Contributors = make([]projectRecord, 0, len(contributors))
				for _, c := range contributors {
					data.Contributors = append(data.Contributors, projectRecord{
						ID:    c.ID,
						Title: c.Attributes.FullName,
					})
				}
			}

			children, err := client.ListNodeChildren(ctx, id)
			if err == nil {
				data.Components = make([]projectRecord, 0, len(children))
				for _, child := range children {
					data.Components = append(data.Components, projectRecord{
						ID:          child.ID,
						Title:       child.Attributes.Title,
						Category:    child.Attributes.Category,
						Description: child.Attributes.Description,
						URL:         child.Links.Self,
					})
				}
			}

			storageFiles, err := client.ListStorageFiles(ctx, id)
			if err == nil {
				data.Files = make([]fileRecord, 0, len(storageFiles))
				for _, f := range storageFiles {
					data.Files = append(data.Files, fileRecord{
						ID:          f.ID,
						Name:        f.Attributes.Name,
						Kind:        f.Attributes.Kind,
						Size:        f.Attributes.Size,
						DownloadURL: f.DownloadURL(),
					})
				}
			}

			if outputMode == outputModeJSON {
				return output.WriteJSON(cmd.OutOrStdout(), data)
			}

			rows := [][]string{
				{"ID", data.Node.ID},
				{"Title", data.Node.Title},
				{"Category", data.Node.Category},
				{"Description", data.Node.Description},
				{"URL", data.Node.URL},
				{"Contributors", fmt.Sprintf("%d", len(data.Contributors))},
				{"Files", fmt.Sprintf("%d", len(data.Files))},
				{"Components", fmt.Sprintf("%d", len(data.Components))},
			}
			return output.WriteTable(cmd.OutOrStdout(), []string{"FIELD", "VALUE"}, rows)
		},
	}
}
