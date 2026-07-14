package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/edithatogo/osf-cli-go/internal/download"
	"github.com/edithatogo/osf-cli-go/internal/osfapi"
	"github.com/edithatogo/osf-cli-go/internal/output"

	"github.com/spf13/cobra"
)

func int64Ptr(v int64) *int64 {
	return &v
}

type filesDownloadRecord = download.FolderDownloadRecord

type filesDownloadResult struct {
	Mode           string                  `json:"mode"`
	Source         string                  `json:"source"`
	Destination    string                  `json:"destination"`
	ConflictPolicy download.ConflictPolicy `json:"conflictPolicy"`
	Resumed        bool                    `json:"resumed"`
	Checksum       string                  `json:"checksum,omitempty"`
	CheckpointPath string                  `json:"checkpointPath,omitempty"`
	Records        []filesDownloadRecord   `json:"records"`
}

func newFilesDownloadCommand(client readonlyClient) *cobra.Command {
	var fileSource string
	var treeSource string
	var conflictName string

	cmd := &cobra.Command{
		Use:   "download --file <file-id-or-url> | --tree <project-or-component-guid-or-url> <destination>",
		Short: "Download a file or folder tree",
		Long:  "Download a single OSF Storage file or a project/component folder tree with conservative conflict handling.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputMode, err := resolveOutputMode(cmd)
			if err != nil {
				return err
			}

			mode, source, err := resolveFilesDownloadSource(fileSource, treeSource)
			if err != nil {
				return err
			}

			policy, err := download.ParseConflictPolicy(conflictName)
			if err != nil {
				return err
			}

			destination := args[0]
			result, err := runFilesDownload(cmd.Context(), client, mode, source, destination, policy)
			if err != nil {
				return err
			}

			if outputMode == outputModeJSON {
				return output.WriteJSON(cmd.OutOrStdout(), result)
			}
			return writeFilesDownloadSummary(cmd.OutOrStdout(), result)
		},
	}

	cmd.Flags().StringVar(&fileSource, "file", "", "download a single file by file id or download URL")
	cmd.Flags().StringVar(&treeSource, "tree", "", "download a project or component folder tree")
	cmd.Flags().StringVar(&conflictName, "conflict", string(download.ConflictFail), "conflict policy: fail, skip, or overwrite")

	return cmd
}

func resolveFilesDownloadSource(fileSource, treeSource string) (string, string, error) {
	hasFile := strings.TrimSpace(fileSource) != ""
	hasTree := strings.TrimSpace(treeSource) != ""
	switch {
	case hasFile && hasTree:
		return "", "", fmt.Errorf("cannot combine --file with --tree")
	case !hasFile && !hasTree:
		return "", "", fmt.Errorf("either --file or --tree is required")
	case hasFile:
		return "file", strings.TrimSpace(fileSource), nil
	default:
		return "tree", strings.TrimSpace(treeSource), nil
	}
}

func runFilesDownload(ctx context.Context, client readonlyClient, mode, source, destination string, policy download.ConflictPolicy) (filesDownloadResult, error) {
	switch mode {
	case "file":
		return downloadSingleFile(ctx, client, source, destination, policy)
	case "tree":
		return downloadFolderTree(ctx, client, source, destination, policy)
	default:
		return filesDownloadResult{}, fmt.Errorf("unsupported download mode %q", mode)
	}
}

func downloadSingleFile(ctx context.Context, client readonlyClient, source, destination string, policy download.ConflictPolicy) (filesDownloadResult, error) {
	file, downloadURL, err := resolveFileSource(ctx, client, source)
	if err != nil {
		return filesDownloadResult{}, err
	}

	localPath, err := filepath.Abs(destination)
	if err != nil {
		return filesDownloadResult{}, fmt.Errorf("resolve destination %q: %w", destination, err)
	}

	record := filesDownloadRecord{
		RemotePath: source,
		LocalPath:  localPath,
	}
	if file.Attributes.Size > 0 {
		record.Bytes = int64Ptr(file.Attributes.Size)
	}

	expectedChecksum := ""
	if md5 := strings.TrimSpace(file.Attributes.Extra.Hashes.MD5); md5 != "" {
		expectedChecksum = "md5:" + md5
	}
	resume, err := download.ResumeStreamAtomically(func(offset int64) (io.ReadCloser, error) {
		return openDownloadAt(ctx, client, downloadURL, offset)
	}, download.ResumeOptions{
		Destination:      localPath,
		Source:           downloadURL,
		ExpectedSize:     record.Bytes,
		ExpectedChecksum: expectedChecksum,
		Policy:           policy,
	})
	result := filesDownloadResult{
		Mode:           "file",
		Source:         source,
		Destination:    localPath,
		ConflictPolicy: policy,
		Resumed:        resume.Resumed,
		Checksum:       resume.Checksum,
		CheckpointPath: resume.CheckpointPath,
		Records:        []filesDownloadRecord{record},
	}
	if err != nil {
		record.Status = download.FolderDownloadFailed
		record.Error = err.Error()
		result.Records[0] = record
		return result, err
	}
	if resume.Completed {
		record.Status = download.FolderDownloadWritten
		record.Bytes = int64Ptr(resume.Bytes)
	} else {
		record.Status = download.FolderDownloadSkipped
	}
	result.Records[0] = record
	return result, nil
}

func downloadFolderTree(ctx context.Context, client readonlyClient, source, destination string, policy download.ConflictPolicy) (filesDownloadResult, error) {
	nodeID, err := parseNodeIDOrURL(source)
	if err != nil {
		return filesDownloadResult{}, err
	}

	files, err := collectFolderDownloadFiles(ctx, client, nodeID, nil)
	if err != nil {
		return filesDownloadResult{}, err
	}

	plan, err := download.NewFolderDownloadPlan(destination, policy, files)
	if err != nil {
		return filesDownloadResult{}, err
	}

	manifest, err := plan.Execute()
	if err != nil {
		return filesDownloadResult{
			Mode:           "tree",
			Source:         source,
			Destination:    manifest.DestinationRoot,
			ConflictPolicy: manifest.ConflictPolicy,
			Records:        manifest.Records,
		}, err
	}

	return filesDownloadResult{
		Mode:           "tree",
		Source:         source,
		Destination:    manifest.DestinationRoot,
		ConflictPolicy: manifest.ConflictPolicy,
		Records:        manifest.Records,
	}, nil
}

func collectFolderDownloadFiles(ctx context.Context, client readonlyClient, nodeID string, segments []string) ([]download.FolderDownloadFile, error) {
	files, err := client.ListStorageFiles(ctx, nodeID, segments...)
	if err != nil {
		return nil, err
	}

	planned := make([]download.FolderDownloadFile, 0, len(files))
	for _, file := range files {
		name := strings.TrimSpace(file.Attributes.Name)
		if name == "" {
			return nil, fmt.Errorf("storage file %q is missing a name", file.ID)
		}

		if strings.EqualFold(file.Attributes.Kind, "folder") {
			children, err := collectFolderDownloadFiles(ctx, client, nodeID, append(append([]string(nil), segments...), name))
			if err != nil {
				return nil, err
			}
			planned = append(planned, children...)
			continue
		}

		downloadURL := file.DownloadURL()
		if strings.TrimSpace(downloadURL) == "" {
			return nil, fmt.Errorf("storage file %q is missing a download URL", file.ID)
		}

		remotePath := path.Join(append(append([]string(nil), segments...), name)...)
		urlCopy := downloadURL
		expectedChecksum := ""
		if md5 := strings.TrimSpace(file.Attributes.Extra.Hashes.MD5); md5 != "" {
			expectedChecksum = "md5:" + md5
		}
		planned = append(planned, download.FolderDownloadFile{
			RemotePath: remotePath,
			Open: func() (io.ReadCloser, error) {
				return client.OpenDownload(ctx, urlCopy)
			},
			OpenRange: func(offset int64) (io.ReadCloser, error) {
				return openDownloadAt(ctx, client, urlCopy, offset)
			},
			SourceIdentity:   urlCopy,
			ExpectedChecksum: expectedChecksum,
			KnownBytes: func() *int64 {
				if file.Attributes.Size <= 0 {
					return nil
				}
				return int64Ptr(file.Attributes.Size)
			}(),
		})
	}

	return planned, nil
}

func openDownloadAt(ctx context.Context, client readonlyClient, downloadURL string, offset int64) (io.ReadCloser, error) {
	if offset == 0 {
		return client.OpenDownload(ctx, downloadURL)
	}
	src, err := client.OpenDownloadRange(ctx, downloadURL, offset)
	if errors.Is(err, osfapi.ErrRangeUnsupported) {
		return nil, download.ErrRangeUnsupported
	}
	return src, err
}

func resolveFileSource(ctx context.Context, client readonlyClient, source string) (osfapi.StorageFile, string, error) {
	trimmed := strings.TrimSpace(source)
	if trimmed == "" {
		return osfapi.StorageFile{}, "", fmt.Errorf("file source is required")
	}

	if parsed, err := url.Parse(trimmed); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		if strings.Contains(parsed.Path, "/v2/files/") {
			id := strings.Trim(parsed.Path, "/")
			parts := strings.Split(id, "/")
			for i, part := range parts {
				if part == "files" && i+1 < len(parts) {
					file, err := client.GetStorageFile(ctx, parts[i+1])
					if err != nil {
						return osfapi.StorageFile{}, "", err
					}
					return file, file.DownloadURL(), nil
				}
			}
			return osfapi.StorageFile{}, "", fmt.Errorf("could not find file id in %q", source)
		}

		if parsed.Host == "files.osf.io" || strings.Contains(parsed.Path, "/download") || strings.Contains(parsed.RawQuery, "download=1") {
			return osfapi.StorageFile{}, trimmed, nil
		}

		return osfapi.StorageFile{}, "", fmt.Errorf("unsupported file source URL %q", source)
	}

	file, err := client.GetStorageFile(ctx, trimmed)
	if err != nil {
		return osfapi.StorageFile{}, "", err
	}
	if url := file.DownloadURL(); strings.TrimSpace(url) == "" {
		return osfapi.StorageFile{}, "", fmt.Errorf("storage file %q does not include a download URL", file.ID)
	}
	return file, file.DownloadURL(), nil
}

func writeFilesDownloadSummary(w io.Writer, result filesDownloadResult) error {
	written, skipped, failed := 0, 0, 0
	for _, record := range result.Records {
		switch record.Status {
		case download.FolderDownloadWritten:
			written++
		case download.FolderDownloadSkipped:
			skipped++
		case download.FolderDownloadFailed:
			failed++
		}
	}

	rows := [][]string{
		{"Mode", result.Mode},
		{"Source", result.Source},
		{"Destination", result.Destination},
		{"Conflict policy", string(result.ConflictPolicy)},
		{"Written", fmt.Sprintf("%d", written)},
		{"Skipped", fmt.Sprintf("%d", skipped)},
		{"Failed", fmt.Sprintf("%d", failed)},
	}
	for _, record := range result.Records {
		label := record.RemotePath
		if label == "" {
			label = record.LocalPath
		}
		rows = append(rows, []string{string(record.Status), fmt.Sprintf("%s -> %s", label, record.LocalPath)})
	}
	return output.WriteTable(w, []string{"FIELD", "VALUE"}, rows)
}
