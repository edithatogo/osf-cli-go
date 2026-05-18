package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/edithatogo/osf-cli-go/internal/osfapi"
	"github.com/edithatogo/osf-cli-go/internal/output"

	"github.com/spf13/cobra"
)

type projectRecord struct {
	ID          string `json:"id"`
	Type        string `json:"type,omitempty"`
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

type addonRecord struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category,omitempty"`
	URL      string `json:"url,omitempty"`
}

func newProjectsCommand(client readonlyClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "List, inspect, and manage OSF projects and components",
		Long:  "List, inspect, and manage OSF projects and components.",
	}
	cmd.AddCommand(newProjectsListCommand(client))
	cmd.AddCommand(newProjectsGetCommand(client))
	cmd.AddCommand(newProjectsCreateCommand(client))
	cmd.AddCommand(newProjectsUpdateCommand(client))
	cmd.AddCommand(newProjectsDeleteCommand(client))
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
		Short: "List, download, and manage OSF Storage files",
		Long:  "List, download, upload, create folders, delete files, and list configured storage add-ons.",
	}
	cmd.AddCommand(newFilesListCommand(client))
	cmd.AddCommand(newFilesDownloadCommand(client))
	cmd.AddCommand(newFilesUploadCommand(client))
	cmd.AddCommand(newFilesMkdirCommand(client))
	cmd.AddCommand(newFilesRmCommand(client))
	cmd.AddCommand(newFilesAddonsCommand(client))
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

func newProjectsCreateCommand(client readonlyClient) *cobra.Command {
	var title string
	var category string
	var description string
	var yes bool
	cmd := &cobra.Command{
		Use:   "create --title <title>",
		Short: "Create an OSF project or component",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			outputMode, err := resolveOutputMode(cmd)
			if err != nil {
				return err
			}
			if strings.TrimSpace(title) == "" {
				return fmt.Errorf("--title flag is required")
			}
			if strings.TrimSpace(category) == "" {
				return fmt.Errorf("--category flag is required")
			}
			if !yes {
				if err := confirmAction(cmd, fmt.Sprintf("Create OSF node %q with category %q? Type yes to continue: ", title, category), "node creation confirmation required", "node creation declined"); err != nil {
					return err
				}
			}
			node, err := client.CreateNode(cmd.Context(), title, category, description)
			if err != nil {
				return err
			}
			return writeProjectRecord(cmd.OutOrStdout(), outputMode, node)
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "node title")
	cmd.Flags().StringVar(&category, "category", "project", "node category")
	cmd.Flags().StringVar(&description, "description", "", "node description")
	cmd.Flags().BoolVar(&yes, "yes", false, "confirm node creation without prompting")
	return cmd
}

func newProjectsUpdateCommand(client readonlyClient) *cobra.Command {
	var title string
	var description string
	var yes bool
	cmd := &cobra.Command{
		Use:   "update <guid-or-url>",
		Short: "Update an OSF node title or description",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputMode, err := resolveOutputMode(cmd)
			if err != nil {
				return err
			}
			if !cmd.Flags().Changed("title") && !cmd.Flags().Changed("description") {
				return fmt.Errorf("at least one of --title or --description is required")
			}
			id, err := parseNodeIDOrURL(args[0])
			if err != nil {
				return err
			}
			current, err := client.GetNode(cmd.Context(), id)
			if err != nil {
				return err
			}
			nextTitle := current.Attributes.Title
			nextDescription := current.Attributes.Description
			if cmd.Flags().Changed("title") {
				nextTitle = title
			}
			if cmd.Flags().Changed("description") {
				nextDescription = description
			}
			if !yes {
				if err := confirmAction(cmd, fmt.Sprintf("Update OSF node %s? Type yes to continue: ", id), "node update confirmation required", "node update declined"); err != nil {
					return err
				}
			}
			node, err := client.UpdateNode(cmd.Context(), id, nextTitle, nextDescription)
			if err != nil {
				return err
			}
			return writeProjectRecord(cmd.OutOrStdout(), outputMode, node)
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "new node title")
	cmd.Flags().StringVar(&description, "description", "", "new node description")
	cmd.Flags().BoolVar(&yes, "yes", false, "confirm node update without prompting")
	return cmd
}

func newProjectsDeleteCommand(client readonlyClient) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:     "delete <guid-or-url>",
		Aliases: []string{"rm"},
		Short:   "Delete an OSF node",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseNodeIDOrURL(args[0])
			if err != nil {
				return err
			}
			if !yes {
				if err := confirmAction(cmd, fmt.Sprintf("Delete OSF node %s? Type yes to continue: ", id), "node deletion confirmation required", "node deletion declined"); err != nil {
					return err
				}
			}
			if err := client.DeleteNode(cmd.Context(), id); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "deleted node %s\n", id)
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "confirm node deletion without prompting")
	return cmd
}

func writeProjectRecord(w io.Writer, outputMode string, node osfapi.Node) error {
	row := projectRecord{
		ID:          node.ID,
		Type:        node.Type,
		Title:       node.Attributes.Title,
		Category:    node.Attributes.Category,
		Description: node.Attributes.Description,
		URL:         node.Links.Self,
	}
	if outputMode == outputModeJSON {
		return output.WriteJSON(w, row)
	}
	return output.WriteTable(w, []string{"ID", "TYPE", "TITLE", "CATEGORY", "URL"}, [][]string{{row.ID, row.Type, row.Title, row.Category, row.URL}})
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
		Use:   "list <project-or-component-guid> [folder-id-or-path]",
		Short: "List OSF Storage files",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputMode, err := resolveOutputMode(cmd)
			if err != nil {
				return err
			}

			id, err := parseNodeIDOrURL(args[0])
			if err != nil {
				return err
			}

			var segments []string
			if len(args) == 2 {
				segments = strings.Split(strings.Trim(args[1], "/"), "/")
			}
			files, err := client.ListStorageFiles(cmd.Context(), id, segments...)
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
	var provider string
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all preprints",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			outputMode, err := resolveOutputMode(cmd)
			if err != nil {
				return err
			}

			nodes, err := client.ListPreprints(cmd.Context(), provider, limit)
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
	cmd.Flags().StringVar(&provider, "provider", "", "filter by preprint provider")
	cmd.Flags().IntVar(&limit, "limit", 20, "maximum records to return; use 0 for all pages")
	return cmd
}

func newSearchCommand(client readonlyClient) *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search OSF projects and components",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputMode, err := resolveOutputMode(cmd)
			if err != nil {
				return err
			}

			results, err := client.SearchOSF(cmd.Context(), args[0], limit)
			if err != nil {
				return err
			}

			rows := make([]projectRecord, 0, len(results))
			for _, n := range results {
				rows = append(rows, projectRecord{
					ID:          n.ID,
					Type:        n.Type,
					Title:       n.Title,
					Category:    n.Category,
					Description: n.Description,
					URL:         n.URL,
				})
			}

			if outputMode == outputModeJSON {
				return output.WriteJSON(cmd.OutOrStdout(), rows)
			}

			tableRows := make([][]string, 0, len(rows))
			for _, row := range rows {
				tableRows = append(tableRows, []string{row.ID, row.Type, row.Title, row.Category, row.URL})
			}
			return output.WriteTable(cmd.OutOrStdout(), []string{"ID", "TYPE", "TITLE", "CATEGORY", "URL"}, tableRows)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "maximum records to return; use 0 for all pages")
	return cmd
}

func newRegistrationsCommand(client readonlyClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registrations",
		Short: "Create and inspect OSF registrations",
		Long:  "Create and inspect OSF registrations.",
	}
	cmd.AddCommand(newRegistrationsCreateCommand(client))
	return cmd
}

func newRegistrationsCreateCommand(client readonlyClient) *cobra.Command {
	var schemaID string
	var title string
	var description string
	var yes bool
	cmd := &cobra.Command{
		Use:   "create <node-id>",
		Short: "Create a draft registration for a node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputMode, err := resolveOutputMode(cmd)
			if err != nil {
				return err
			}
			if schemaID == "" {
				return fmt.Errorf("--schema flag is required")
			}
			nodeID := args[0]
			if !yes {
				if err := confirmAction(cmd, fmt.Sprintf("Create registration for node %s with schema %s? Type yes to continue: ", nodeID, schemaID), "registration creation confirmation required", "registration creation declined"); err != nil {
					return err
				}
			}
			registration, err := client.CreateRegistration(cmd.Context(), nodeID, osfapi.RegistrationRequest{
				SchemaID:    schemaID,
				Title:       title,
				Description: description,
			})
			if err != nil {
				return err
			}
			row := projectRecord{
				ID:          registration.ID,
				Type:        registration.Type,
				Title:       registration.Attributes.Title,
				Category:    registration.Attributes.Category,
				Description: registration.Attributes.Description,
				URL:         registration.Links.Self,
			}
			if outputMode == outputModeJSON {
				return output.WriteJSON(cmd.OutOrStdout(), row)
			}
			return output.WriteTable(cmd.OutOrStdout(), []string{"ID", "TYPE", "TITLE", "CATEGORY", "URL"}, [][]string{{row.ID, row.Type, row.Title, row.Category, row.URL}})
		},
	}
	cmd.Flags().StringVar(&schemaID, "schema", "", "registration schema ID")
	cmd.Flags().StringVar(&title, "title", "", "draft registration title")
	cmd.Flags().StringVar(&description, "description", "", "draft registration description")
	cmd.Flags().BoolVar(&yes, "yes", false, "confirm registration creation without prompting")
	return cmd
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
			progress := NewProgressWriter(cmd.ErrOrStderr())
			if err := client.UploadFile(cmd.Context(), providerURL, fileName, io.TeeReader(f, progress), conflict); err != nil {
				return err
			}
			progress.Finish()
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
	var yes bool
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
			if !yes {
				if err := confirmAction(cmd, fmt.Sprintf("Delete %q from node %s? Type yes to continue: ", args[0], nodeID), "delete confirmation required", "delete confirmation declined"); err != nil {
					return err
				}
			}
			if err := client.DeleteFile(cmd.Context(), providerURL, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "deleted %q from node %s\n", args[0], nodeID)
			return nil
		},
	}
	cmd.Flags().StringVar(&nodeID, "node", "", "target node GUID")
	cmd.Flags().BoolVar(&yes, "yes", false, "confirm deletion without prompting")
	_ = cmd.MarkFlagRequired("node")
	return cmd
}

func newFilesAddonsCommand(client readonlyClient) *cobra.Command {
	return &cobra.Command{
		Use:   "addons <node-id>",
		Short: "List configured storage add-ons for a node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputMode, err := resolveOutputMode(cmd)
			if err != nil {
				return err
			}
			addons, err := client.ListNodeAddons(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			rows := make([]addonRecord, 0, len(addons))
			for _, addon := range addons {
				rows = append(rows, addonRecord{
					ID:       addon.ID,
					Name:     addon.Attributes.Title,
					Category: addon.Attributes.Category,
					URL:      addon.Links.Self,
				})
			}
			if outputMode == outputModeJSON {
				return output.WriteJSON(cmd.OutOrStdout(), rows)
			}
			tableRows := make([][]string, 0, len(rows))
			for _, row := range rows {
				tableRows = append(tableRows, []string{row.ID, row.Name, row.Category, row.URL})
			}
			return output.WriteTable(cmd.OutOrStdout(), []string{"ID", "NAME", "CATEGORY", "URL"}, tableRows)
		},
	}
}

func confirmAction(cmd *cobra.Command, prompt, requiredErr, declinedErr string) error {
	_, _ = fmt.Fprint(cmd.ErrOrStderr(), prompt)
	answer, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil && len(answer) == 0 {
		return errors.New(requiredErr)
	}
	if strings.ToLower(strings.TrimSpace(answer)) != "yes" {
		return errors.New(declinedErr)
	}
	return nil
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
