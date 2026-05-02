package cli

import (
	"strconv"

	"osf-cli-go/internal/output"

	"github.com/spf13/cobra"
)

type projectRecord struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Category    string `json:"category,omitempty"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
}

type fileRecord struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Size        int64  `json:"size,omitempty"`
	DownloadURL string `json:"download_url,omitempty"`
}

func newProjectsCommand(client readonlyClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "List and inspect OSF projects and components",
		Long:  "List and inspect OSF projects and components.",
	}
	cmd.AddCommand(newProjectsListCommand(client))
	cmd.AddCommand(newProjectsGetCommand(client))
	return cmd
}

func newComponentsCommand(client readonlyClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "components",
		Short: "List project components",
		Long:  "List project components.",
	}
	cmd.AddCommand(newComponentsListCommand(client))
	return cmd
}

func newFilesCommand(client readonlyClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "files",
		Short: "List OSF Storage files",
		Long:  "List OSF Storage files.",
	}
	cmd.AddCommand(newFilesListCommand(client))
	return cmd
}

func newProjectsListCommand(client readonlyClient) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List accessible projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			outputMode, err := resolveOutputMode(cmd)
			if err != nil {
				return err
			}

			projects, err := client.ListProjects(cmd.Context())
			if err != nil {
				return err
			}

			rows := make([]projectRecord, 0, len(projects))
			for _, project := range projects {
				rows = append(rows, projectRecord{
					ID:          project.ID,
					Title:       project.Attributes.Title,
					Category:    project.Attributes.Category,
					Description: project.Attributes.Description,
					URL:         project.Links.Self,
				})
			}

			if outputMode == outputModeJSON {
				return output.WriteJSON(cmd.OutOrStdout(), rows)
			}

			tableRows := make([][]string, 0, len(rows))
			for _, row := range rows {
				tableRows = append(tableRows, []string{row.ID, row.Title, row.Category, row.URL})
			}
			return output.WriteTable(cmd.OutOrStdout(), []string{"ID", "TITLE", "CATEGORY", "URL"}, tableRows)
		},
	}
}

func newProjectsGetCommand(client readonlyClient) *cobra.Command {
	return &cobra.Command{
		Use:   "get <guid-or-url>",
		Short: "Show one project",
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

			node, err := client.GetNode(cmd.Context(), id)
			if err != nil {
				return err
			}

			record := projectRecord{
				ID:          node.ID,
				Title:       node.Attributes.Title,
				Category:    node.Attributes.Category,
				Description: node.Attributes.Description,
				URL:         node.Links.Self,
			}

			if outputMode == outputModeJSON {
				return output.WriteJSON(cmd.OutOrStdout(), record)
			}

			rows := [][]string{
				{"ID", record.ID},
				{"Title", record.Title},
				{"Category", record.Category},
				{"Description", record.Description},
				{"URL", record.URL},
			}
			return output.WriteTable(cmd.OutOrStdout(), []string{"FIELD", "VALUE"}, rows)
		},
	}
}

func newComponentsListCommand(client readonlyClient) *cobra.Command {
	return &cobra.Command{
		Use:   "list <project-guid-or-url>",
		Short: "List child components",
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

			nodes, err := client.ListNodeChildren(cmd.Context(), id)
			if err != nil {
				return err
			}

			rows := make([]projectRecord, 0, len(nodes))
			for _, node := range nodes {
				rows = append(rows, projectRecord{
					ID:          node.ID,
					Title:       node.Attributes.Title,
					Category:    node.Attributes.Category,
					Description: node.Attributes.Description,
					URL:         node.Links.Self,
				})
			}

			if outputMode == outputModeJSON {
				return output.WriteJSON(cmd.OutOrStdout(), rows)
			}

			tableRows := make([][]string, 0, len(rows))
			for _, row := range rows {
				tableRows = append(tableRows, []string{row.ID, row.Title, row.Category, row.URL})
			}
			return output.WriteTable(cmd.OutOrStdout(), []string{"ID", "TITLE", "CATEGORY", "URL"}, tableRows)
		},
	}
}

func newFilesListCommand(client readonlyClient) *cobra.Command {
	return &cobra.Command{
		Use:   "list <project-or-component-guid>",
		Short: "List OSF Storage files",
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

			files, err := client.ListStorageFiles(cmd.Context(), id)
			if err != nil {
				return err
			}

			rows := make([]fileRecord, 0, len(files))
			for _, file := range files {
				rows = append(rows, fileRecord{
					ID:          file.ID,
					Name:        file.Attributes.Name,
					Kind:        file.Attributes.Kind,
					Size:        file.Attributes.Size,
					DownloadURL: file.DownloadURL(),
				})
			}

			if outputMode == outputModeJSON {
				return output.WriteJSON(cmd.OutOrStdout(), rows)
			}

			tableRows := make([][]string, 0, len(rows))
			for _, row := range rows {
				tableRows = append(tableRows, []string{row.ID, row.Name, row.Kind, formatInt64(row.Size), row.DownloadURL})
			}
			return output.WriteTable(cmd.OutOrStdout(), []string{"ID", "NAME", "KIND", "SIZE", "DOWNLOAD_URL"}, tableRows)
		},
	}
}

func formatInt64(v int64) string {
	if v == 0 {
		return ""
	}
	return strconv.FormatInt(v, 10)
}
