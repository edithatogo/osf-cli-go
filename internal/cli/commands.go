package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/edithatogo/osf-cli-go/internal/output"

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

func newOpenCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "open <guid-or-url>",
		Short: "Open an OSF node in the default browser",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseNodeIDOrURL(args[0])
			if err != nil {
				return err
			}
			url := "https://osf.io/" + id + "/"
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Opening %s ...\n", url)

			var openCmd string
			var openArgs []string
			switch runtime.GOOS {
			case "darwin":
				openCmd = "open"
				openArgs = []string{url}
			case "windows":
				openCmd = "cmd"
				openArgs = []string{"/c", "start", url}
			default:
				openCmd = "xdg-open"
				openArgs = []string{url}
			}
			return exec.Command(openCmd, openArgs...).Start()
		},
	}
}

func newFilesCommand(client readonlyClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "files",
		Short: "List and download OSF Storage files",
		Long:  "List and download OSF Storage files.",
	}
	cmd.AddCommand(newFilesListCommand(client))
	cmd.AddCommand(newFilesDownloadCommand(client))
	cmd.AddCommand(newFilesUploadCommand(client))
	cmd.AddCommand(newFilesMkdirCommand(client))
	cmd.AddCommand(newFilesRmCommand(client))
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

func newPreprintsCommand(client readonlyClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preprints",
		Short: "List OSF preprints",
		Long:  "List OSF preprints.",
	}
	cmd.AddCommand(newPreprintsListCommand(client))
	return cmd
}

func newPreprintsListCommand(client readonlyClient) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all preprints",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			outputMode, err := resolveOutputMode(cmd)
			if err != nil {
				return err
			}

			nodes, err := client.ListPreprints(cmd.Context())
			if err != nil {
				return err
			}

			rows := make([]projectRecord, 0, len(nodes))
			for _, n := range nodes {
				rows = append(rows, projectRecord{
					ID:          n.ID,
					Title:       n.Attributes.Title,
					Category:    n.Attributes.Category,
					Description: n.Attributes.Description,
					URL:         n.Links.Self,
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

func newSearchCommand(client readonlyClient) *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search OSF projects and components",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputMode, err := resolveOutputMode(cmd)
			if err != nil {
				return err
			}

			nodes, err := client.SearchOSF(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			rows := make([]projectRecord, 0, len(nodes))
			for _, n := range nodes {
				rows = append(rows, projectRecord{
					ID:          n.ID,
					Title:       n.Attributes.Title,
					Category:    n.Attributes.Category,
					Description: n.Attributes.Description,
					URL:         n.Links.Self,
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

func formatInt64(v int64) string {
	if v == 0 {
		return ""
	}
	return strconv.FormatInt(v, 10)
}

func newFilesUploadCommand(client readonlyClient) *cobra.Command {
	var nodeID string
	var conflict string
	cmd := &cobra.Command{
		Use:   "upload --node <guid> <local-path>",
		Short: "Upload a file to OSF Storage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if nodeID == "" {
				return fmt.Errorf("--node flag is required")
			}
			localPath := args[0]
			fileName := filepath.Base(localPath)
			f, err := os.Open(localPath)
			if err != nil {
				return fmt.Errorf("open file %q: %w", localPath, err)
			}
			defer func() { _ = f.Close() }()
			providerURL, err := client.GetNodeFilesProvider(cmd.Context(), nodeID)
			if err != nil {
				return fmt.Errorf("get files provider: %w", err)
			}
			if conflict == "" {
				conflict = "fail"
			}
			if err := client.UploadFile(cmd.Context(), providerURL, fileName, f, conflict); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "uploaded %s to node %s\n", fileName, nodeID)
			return nil
		},
	}
	cmd.Flags().StringVar(&nodeID, "node", "", "target node GUID")
	cmd.Flags().StringVar(&conflict, "conflict", "fail", "conflict policy: fail or overwrite")
	return cmd
}

func newFilesMkdirCommand(client readonlyClient) *cobra.Command {
	var nodeID string
	cmd := &cobra.Command{
		Use:   "mkdir --node <guid> <folder-name>",
		Short: "Create a folder in OSF Storage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if nodeID == "" {
				return fmt.Errorf("--node flag is required")
			}
			providerURL, err := client.GetNodeFilesProvider(cmd.Context(), nodeID)
			if err != nil {
				return fmt.Errorf("get files provider: %w", err)
			}
			if err := client.CreateFolder(cmd.Context(), providerURL, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "created folder %q in node %s\n", args[0], nodeID)
			return nil
		},
	}
	cmd.Flags().StringVar(&nodeID, "node", "", "target node GUID")
	_ = cmd.MarkFlagRequired("node")
	return cmd
}

func newFilesRmCommand(client readonlyClient) *cobra.Command {
	var nodeID string
	cmd := &cobra.Command{
		Use:   "rm --node <guid> <file-name>",
		Short: "Delete a file from OSF Storage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if nodeID == "" {
				return fmt.Errorf("--node flag is required")
			}
			providerURL, err := client.GetNodeFilesProvider(cmd.Context(), nodeID)
			if err != nil {
				return fmt.Errorf("get files provider: %w", err)
			}
			if err := client.DeleteFile(cmd.Context(), providerURL, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "deleted %q from node %s\n", args[0], nodeID)
			return nil
		},
	}
	cmd.Flags().StringVar(&nodeID, "node", "", "target node GUID")
	_ = cmd.MarkFlagRequired("node")
	return cmd
}

func newWhoamiCommand(client readonlyClient) *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the active authenticated OSF account (alias for auth whoami)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return newAuthWhoamiCommand(client).RunE(cmd, args)
		},
	}
}
