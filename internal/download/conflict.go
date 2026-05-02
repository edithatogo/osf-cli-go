package download

import "fmt"

// ConflictPolicy controls what happens when the local destination already exists.
type ConflictPolicy string

const (
	// ConflictFail stops the download when the destination exists.
	ConflictFail ConflictPolicy = "fail"
	// ConflictSkip leaves the destination untouched and reports a skipped write.
	ConflictSkip ConflictPolicy = "skip"
	// ConflictOverwrite replaces the existing destination after a successful stream.
	ConflictOverwrite ConflictPolicy = "overwrite"
)

// ParseConflictPolicy converts text into a validated conflict policy.
func ParseConflictPolicy(value string) (ConflictPolicy, error) {
	policy := ConflictPolicy(value)
	if err := policy.Validate(); err != nil {
		return "", err
	}
	return policy, nil
}

// Validate reports whether the policy is one of the supported values.
func (p ConflictPolicy) Validate() error {
	switch p {
	case ConflictFail, ConflictSkip, ConflictOverwrite:
		return nil
	default:
		return fmt.Errorf("unsupported conflict policy %q", string(p))
	}
}
