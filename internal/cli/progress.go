package cli

import (
	"fmt"
	"io"
	"sync"
)

type ProgressWriter struct {
	mu      sync.Mutex
	out     io.Writer
	enabled bool
	written int64
}

func NewProgressWriter(w io.Writer) *ProgressWriter {
	_, enabled := w.(interface{ Fd() uintptr })
	return &ProgressWriter{out: w, enabled: enabled}
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.mu.Lock()
	pw.written += int64(n)
	pw.mu.Unlock()
	if pw.enabled {
		pw.mu.Lock()
		_, _ = fmt.Fprintf(pw.out, "\r  downloaded %d bytes", pw.written)
		pw.mu.Unlock()
	}
	return n, nil
}

func (pw *ProgressWriter) Finish() {
	if pw.enabled {
		pw.mu.Lock()
		_, _ = fmt.Fprintf(pw.out, "\r  downloaded %d bytes - done\n", pw.written)
		pw.mu.Unlock()
	}
}
