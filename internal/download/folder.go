package download

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// FolderDownloadStatus describes the outcome for a folder download entry.
type FolderDownloadStatus string

const (
	// FolderDownloadWritten means the file was written successfully.
	FolderDownloadWritten FolderDownloadStatus = "written"
	// FolderDownloadSkipped means the local destination already existed and was left untouched.
	FolderDownloadSkipped FolderDownloadStatus = "skipped"
	// FolderDownloadFailed means the file could not be read or written.
	FolderDownloadFailed FolderDownloadStatus = "failed"
)

// FolderDownloadFile describes one remote file in a folder-tree download.
//
// The file content can be supplied either as a reader or an opener. If the
// reader also implements io.Closer, it is closed after the write completes.
type FolderDownloadFile struct {
	RemotePath       string
	Reader           io.Reader
	Open             func() (io.ReadCloser, error)
	OpenRange        func(offset int64) (io.ReadCloser, error)
	SourceIdentity   string
	ExpectedChecksum string
	KnownBytes       *int64
}

type plannedFolderDownloadFile struct {
	remotePath string
	localPath  string
	reader     io.Reader
	open       func() (io.ReadCloser, error)
	openRange  func(int64) (io.ReadCloser, error)
	source     string
	checksum   string
	knownBytes *int64
}

// FolderDownloadRecord is one entry in a folder download manifest.
type FolderDownloadRecord struct {
	RemotePath string               `json:"remotePath"`
	LocalPath  string               `json:"localPath"`
	Bytes      *int64               `json:"bytes,omitempty"`
	Resumed    bool                 `json:"resumed,omitempty"`
	Checksum   string               `json:"checksum,omitempty"`
	Checkpoint string               `json:"checkpointPath,omitempty"`
	Status     FolderDownloadStatus `json:"status"`
	Error      string               `json:"error,omitempty"`
}

// FolderDownloadManifest summarizes the result of a folder-tree download.
type FolderDownloadManifest struct {
	DestinationRoot string                 `json:"destinationRoot"`
	ConflictPolicy  ConflictPolicy         `json:"conflictPolicy"`
	Records         []FolderDownloadRecord `json:"records"`
}

// FolderDownloadPlan validates and resolves folder-tree download entries.
type FolderDownloadPlan struct {
	destinationRoot string
	policy          ConflictPolicy
	files           []plannedFolderDownloadFile
}

// NewFolderDownloadPlan builds a download plan for a folder tree.
func NewFolderDownloadPlan(destRoot string, policy ConflictPolicy, files []FolderDownloadFile) (*FolderDownloadPlan, error) {
	if err := policy.Validate(); err != nil {
		return nil, err
	}

	normalizedRoot, err := NormalizeDestination(destRoot)
	if err != nil {
		return nil, err
	}

	planned := make([]plannedFolderDownloadFile, 0, len(files))
	for _, file := range files {
		if file.Reader == nil && file.Open == nil && file.OpenRange == nil {
			return nil, fmt.Errorf("folder file %q requires a reader or opener", file.RemotePath)
		}
		if file.OpenRange != nil && strings.TrimSpace(file.SourceIdentity) == "" {
			return nil, fmt.Errorf("folder file %q requires a source identity for resumable downloads", file.RemotePath)
		}

		remotePath, err := NormalizeRemotePath(file.RemotePath)
		if err != nil {
			return nil, fmt.Errorf("plan folder file %q: %w", file.RemotePath, err)
		}

		localPath := filepath.Join(normalizedRoot, filepath.FromSlash(remotePath))
		ok, err := withinBase(normalizedRoot, localPath)
		if err != nil {
			return nil, fmt.Errorf("plan folder file %q: %w", file.RemotePath, err)
		}
		if !ok {
			return nil, errPathTraversal
		}

		planned = append(planned, plannedFolderDownloadFile{
			remotePath: remotePath,
			localPath:  localPath,
			reader:     file.Reader,
			open:       file.Open,
			openRange:  file.OpenRange,
			source:     file.SourceIdentity,
			checksum:   file.ExpectedChecksum,
			knownBytes: file.KnownBytes,
		})
	}

	return &FolderDownloadPlan{
		destinationRoot: normalizedRoot,
		policy:          policy,
		files:           planned,
	}, nil
}

// Execute downloads the planned folder tree and returns a manifest of the result.
func (p *FolderDownloadPlan) Execute() (FolderDownloadManifest, error) {
	if p == nil {
		return FolderDownloadManifest{}, fmt.Errorf("folder download plan is required")
	}

	manifest := FolderDownloadManifest{
		DestinationRoot: p.destinationRoot,
		ConflictPolicy:  p.policy,
		Records:         make([]FolderDownloadRecord, 0, len(p.files)),
	}

	for _, file := range p.files {
		record := FolderDownloadRecord{
			RemotePath: file.remotePath,
			LocalPath:  file.localPath,
		}

		var written bool
		var bytes int64
		var resumed bool
		var checksum string
		var checkpointPath string
		var writeErr error
		if file.openRange != nil {
			resume, resumeErr := ResumeStreamAtomically(file.openRange, ResumeOptions{Destination: file.localPath, Source: file.source, ExpectedSize: file.knownBytes, ExpectedChecksum: file.checksum, Policy: p.policy})
			written = resume.Completed
			bytes = resume.Bytes
			resumed = resume.Resumed
			checksum = resume.Checksum
			checkpointPath = resume.CheckpointPath
			writeErr = resumeErr
		} else {
			src, closeFn, openErr := file.openReader()
			if openErr != nil {
				writeErr = openErr
			} else {
				counting := &countingReader{reader: src}
				written, writeErr = WriteStreamAtomically(file.localPath, counting, 0o644, p.policy)
				bytes = counting.n
				if closeFn != nil {
					if closeErr := closeFn(); writeErr == nil && closeErr != nil {
						writeErr = fmt.Errorf("close download source: %w", closeErr)
					}
				}
			}
		}
		if writeErr != nil {
			record.Status = FolderDownloadFailed
			record.Error = writeErr.Error()
			if file.knownBytes != nil {
				record.Bytes = file.knownBytes
			}
			record.Resumed = resumed
			record.Checksum = checksum
			record.Checkpoint = checkpointPath
			manifest.Records = append(manifest.Records, record)
			return manifest, writeErr
		}

		if written {
			record.Status = FolderDownloadWritten
			record.Bytes = int64Ptr(bytes)
			record.Resumed = resumed
			record.Checksum = checksum
		} else {
			record.Status = FolderDownloadSkipped
			record.Bytes = file.knownBytes
		}
		manifest.Records = append(manifest.Records, record)
	}

	return manifest, nil
}

func (f plannedFolderDownloadFile) openReader() (io.ReadCloser, func() error, error) {
	if f.open != nil {
		rc, err := f.open()
		if err != nil {
			if rc != nil {
				_ = rc.Close()
			}
			return nil, nil, err
		}
		if rc == nil {
			return nil, nil, fmt.Errorf("folder file %q opener returned nil reader", f.remotePath)
		}
		return rc, rc.Close, nil
	}

	if f.reader != nil {
		if rc, ok := f.reader.(io.ReadCloser); ok {
			return rc, rc.Close, nil
		}
		return io.NopCloser(f.reader), nil, nil
	}

	return nil, nil, fmt.Errorf("folder file %q requires a reader or opener", f.remotePath)
}

type countingReader struct {
	reader io.Reader
	n      int64
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.n += int64(n)
	return n, err
}

func int64Ptr(v int64) *int64 {
	return &v
}
