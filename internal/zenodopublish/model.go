package zenodopublish

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/auth"
)

// State identifies a Zenodo deposition's publication lifecycle position.
type State string

const (
	StateDraft        State = "draft"
	StateDOIReserved  State = "doi_reserved"
	StatePublished    State = "published"
	StateVersionDraft State = "version_draft"
	StateDiscarded    State = "discarded"
)

// Action is an explicit Zenodo lifecycle operation.
type Action string

const (
	ActionReserveDOI Action = "reserve_doi"
	ActionPublish    Action = "publish"
	ActionNewVersion Action = "new_version"
	ActionDiscard    Action = "discard"
)

// Scope is a Zenodo OAuth token permission required by a lifecycle action.
type Scope string

const (
	ScopeDepositWrite   Scope = "deposit:write"
	ScopeDepositActions Scope = "deposit:actions"
)

// Access is the publication-time file access policy.
type Access string

const (
	AccessOpen       Access = "open"
	AccessEmbargoed  Access = "embargoed"
	AccessRestricted Access = "restricted"
	AccessClosed     Access = "closed"
)

var (
	ErrInvalidTransition     = errors.New("invalid Zenodo publication transition")
	ErrAuthorizationRequired = errors.New("explicit Zenodo publication authorization is required")
	ErrScopeRequired         = errors.New("required Zenodo token scope is missing")
	ErrConfirmationRequired  = errors.New("exact Zenodo publication confirmation is required")
	ErrInvalidMetadata       = errors.New("invalid Zenodo publication metadata")
	ErrDryRun                = errors.New("zenodo publication dry-run cannot execute")
)

// Creator is the minimum creator identity required for publication.
type Creator struct {
	Name string `json:"name"`
}

// Metadata contains publication policy fields validated before any request.
type Metadata struct {
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	UploadType       string     `json:"uploadType"`
	Creators         []Creator  `json:"creators"`
	Access           Access     `json:"access"`
	License          string     `json:"license,omitempty"`
	EmbargoDate      *time.Time `json:"embargoDate,omitempty"`
	AccessConditions string     `json:"accessConditions,omitempty"`
}

// Request describes the caller's complete lifecycle intent.
type Request struct {
	RecordID     string   `json:"recordId"`
	State        State    `json:"state"`
	Action       Action   `json:"action"`
	Authorized   bool     `json:"authorized"`
	DryRun       bool     `json:"dryRun"`
	Confirmation string   `json:"confirmation,omitempty"`
	Scopes       []Scope  `json:"scopes"`
	Metadata     Metadata `json:"metadata"`
}

// Plan is the validated, non-secret lifecycle decision consumed by execution.
type Plan struct {
	RecordID       string   `json:"recordId"`
	From           State    `json:"from"`
	To             State    `json:"to"`
	Action         Action   `json:"action"`
	RequiredScopes []Scope  `json:"requiredScopes"`
	DryRun         bool     `json:"dryRun"`
	Destructive    bool     `json:"destructive"`
	Irreversible   bool     `json:"irreversible"`
	Confirmation   string   `json:"confirmation,omitempty"`
	Metadata       Metadata `json:"metadata"`
}

type transition struct {
	to           State
	scopes       []Scope
	destructive  bool
	irreversible bool
}

var transitions = map[State]map[Action]transition{
	StateDraft: {
		ActionReserveDOI: {to: StateDOIReserved, scopes: []Scope{ScopeDepositWrite}},
		ActionPublish:    {to: StatePublished, scopes: []Scope{ScopeDepositWrite, ScopeDepositActions}, irreversible: true},
		ActionDiscard:    {to: StateDiscarded, scopes: []Scope{ScopeDepositWrite}, destructive: true},
	},
	StateDOIReserved: {
		ActionPublish: {to: StatePublished, scopes: []Scope{ScopeDepositWrite, ScopeDepositActions}, irreversible: true},
		ActionDiscard: {to: StateDiscarded, scopes: []Scope{ScopeDepositWrite}, destructive: true},
	},
	StatePublished: {
		ActionNewVersion: {to: StateVersionDraft, scopes: []Scope{ScopeDepositActions}},
	},
	StateVersionDraft: {
		ActionPublish: {to: StatePublished, scopes: []Scope{ScopeDepositWrite, ScopeDepositActions}, irreversible: true},
		ActionDiscard: {to: StatePublished, scopes: []Scope{ScopeDepositWrite}, destructive: true},
	},
}

// BuildPlan validates an action completely without performing network I/O.
func BuildPlan(request Request, now time.Time) (Plan, error) {
	recordID := strings.TrimSpace(request.RecordID)
	if recordID == "" {
		return Plan{}, fmt.Errorf("%w: record id is required", ErrInvalidTransition)
	}
	stateTransitions, ok := transitions[request.State]
	if !ok {
		return Plan{}, fmt.Errorf("%w: state %q is terminal or unknown", ErrInvalidTransition, request.State)
	}
	decision, ok := stateTransitions[request.Action]
	if !ok {
		return Plan{}, fmt.Errorf("%w: action %q is not allowed from %q", ErrInvalidTransition, request.Action, request.State)
	}
	if !request.Authorized {
		return Plan{}, ErrAuthorizationRequired
	}
	for _, required := range decision.scopes {
		if !hasScope(request.Scopes, required) {
			return Plan{}, fmt.Errorf("%w: %s", ErrScopeRequired, required)
		}
	}
	if request.Action == ActionPublish || request.Action == ActionReserveDOI {
		if err := request.Metadata.validate(now); err != nil {
			return Plan{}, err
		}
	}
	challenge := confirmationChallenge(request.Action, recordID, decision.to)
	if (decision.destructive || decision.irreversible) && !request.DryRun && request.Confirmation != challenge {
		return Plan{}, fmt.Errorf("%w: run a dry-run and supply %q", ErrConfirmationRequired, challenge)
	}
	return Plan{
		RecordID: recordID, From: request.State, To: decision.to, Action: request.Action,
		RequiredScopes: append([]Scope(nil), decision.scopes...), DryRun: request.DryRun, Destructive: decision.destructive,
		Irreversible: decision.irreversible, Confirmation: challenge, Metadata: request.Metadata.clone(),
	}, nil
}

func (metadata Metadata) clone() Metadata {
	result := metadata
	result.Creators = append([]Creator(nil), metadata.Creators...)
	if metadata.EmbargoDate != nil {
		embargoDate := *metadata.EmbargoDate
		result.EmbargoDate = &embargoDate
	}
	return result
}

func hasScope(scopes []Scope, required Scope) bool {
	for _, scope := range scopes {
		if scope == required {
			return true
		}
	}
	return false
}

func confirmationChallenge(action Action, recordID string, target State) string {
	return fmt.Sprintf("zenodo:%s:%s:%s", action, recordID, target)
}

func (metadata Metadata) validate(now time.Time) error {
	if strings.TrimSpace(metadata.Title) == "" || strings.TrimSpace(metadata.Description) == "" || strings.TrimSpace(metadata.UploadType) == "" {
		return fmt.Errorf("%w: title, description, and upload type are required", ErrInvalidMetadata)
	}
	if len(metadata.Creators) == 0 {
		return fmt.Errorf("%w: at least one creator is required", ErrInvalidMetadata)
	}
	for _, creator := range metadata.Creators {
		if strings.TrimSpace(creator.Name) == "" {
			return fmt.Errorf("%w: every creator requires a name", ErrInvalidMetadata)
		}
	}
	switch metadata.Access {
	case AccessOpen:
		if strings.TrimSpace(metadata.License) == "" {
			return fmt.Errorf("%w: open access requires a license", ErrInvalidMetadata)
		}
		if metadata.EmbargoDate != nil || strings.TrimSpace(metadata.AccessConditions) != "" {
			return fmt.Errorf("%w: open access cannot set embargo or access conditions", ErrInvalidMetadata)
		}
	case AccessEmbargoed:
		if strings.TrimSpace(metadata.License) == "" || metadata.EmbargoDate == nil || !metadata.EmbargoDate.After(now) {
			return fmt.Errorf("%w: embargoed access requires a license and future embargo date", ErrInvalidMetadata)
		}
		if strings.TrimSpace(metadata.AccessConditions) != "" {
			return fmt.Errorf("%w: embargoed access cannot set restricted-access conditions", ErrInvalidMetadata)
		}
	case AccessRestricted:
		if strings.TrimSpace(metadata.AccessConditions) == "" || metadata.EmbargoDate != nil {
			return fmt.Errorf("%w: restricted access requires conditions and cannot set an embargo date", ErrInvalidMetadata)
		}
	case AccessClosed:
		if metadata.EmbargoDate != nil || strings.TrimSpace(metadata.AccessConditions) != "" {
			return fmt.Errorf("%w: closed access cannot set embargo or access conditions", ErrInvalidMetadata)
		}
	default:
		return fmt.Errorf("%w: access %q is unsupported", ErrInvalidMetadata, metadata.Access)
	}
	return nil
}

// AuditEvent is a deliberately small, redacted lifecycle evidence record.
type AuditEvent struct {
	RecordID     string `json:"recordId"`
	From         State  `json:"from"`
	To           State  `json:"to"`
	Action       Action `json:"action"`
	Outcome      string `json:"outcome"`
	DryRun       bool   `json:"dryRun"`
	Destructive  bool   `json:"destructive"`
	Irreversible bool   `json:"irreversible"`
	Error        string `json:"error,omitempty"`
}

// Audit creates a non-secret event. Metadata and token scopes are omitted.
func (plan Plan) Audit(outcome string, err error) AuditEvent {
	event := AuditEvent{
		RecordID: plan.RecordID, From: plan.From, To: plan.To, Action: plan.Action,
		Outcome: strings.TrimSpace(outcome), DryRun: plan.DryRun,
		Destructive: plan.Destructive, Irreversible: plan.Irreversible,
	}
	if err != nil {
		event.Error = auth.Redact(err.Error())
	}
	return event
}

// String serializes an audit event without exposing metadata or credentials.
func (event AuditEvent) String() string {
	encoded, err := json.Marshal(event)
	if err != nil {
		return `{"outcome":"audit-encoding-failed"}`
	}
	return string(encoded)
}
