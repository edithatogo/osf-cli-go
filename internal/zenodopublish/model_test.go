package zenodopublish

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestPlanAcceptsEverySupportedTransition(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	metadata := validMetadata()
	tests := []struct {
		name   string
		state  State
		action Action
		to     State
		scopes []Scope
	}{
		{name: "reserve DOI", state: StateDraft, action: ActionReserveDOI, to: StateDOIReserved, scopes: []Scope{ScopeDepositWrite}},
		{name: "publish draft", state: StateDraft, action: ActionPublish, to: StatePublished, scopes: []Scope{ScopeDepositActions}},
		{name: "publish reserved DOI", state: StateDOIReserved, action: ActionPublish, to: StatePublished, scopes: []Scope{ScopeDepositActions}},
		{name: "new version", state: StatePublished, action: ActionNewVersion, to: StateVersionDraft, scopes: []Scope{ScopeDepositActions}},
		{name: "publish version", state: StateVersionDraft, action: ActionPublish, to: StatePublished, scopes: []Scope{ScopeDepositActions}},
		{name: "discard draft", state: StateDraft, action: ActionDiscard, to: StateDiscarded, scopes: []Scope{ScopeDepositWrite}},
		{name: "discard reserved DOI", state: StateDOIReserved, action: ActionDiscard, to: StateDiscarded, scopes: []Scope{ScopeDepositWrite}},
		{name: "discard version draft", state: StateVersionDraft, action: ActionDiscard, to: StatePublished, scopes: []Scope{ScopeDepositWrite}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := Request{RecordID: "123", State: test.state, Action: test.action, Authorized: true, DryRun: true, Scopes: test.scopes, Metadata: metadata}
			plan, err := BuildPlan(request, now)
			if err != nil {
				t.Fatalf("BuildPlan: %v", err)
			}
			if plan.From != test.state || plan.To != test.to || plan.Action != test.action || plan.RecordID != "123" {
				t.Fatalf("plan = %+v", plan)
			}
		})
	}
}

func TestPlanRejectsEveryOtherStateActionPair(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	valid := map[State]map[Action]bool{
		StateDraft:        {ActionReserveDOI: true, ActionPublish: true, ActionDiscard: true},
		StateDOIReserved:  {ActionPublish: true, ActionDiscard: true},
		StatePublished:    {ActionNewVersion: true},
		StateVersionDraft: {ActionPublish: true, ActionDiscard: true},
	}
	states := []State{StateDraft, StateDOIReserved, StatePublished, StateVersionDraft, StateDiscarded}
	actions := []Action{ActionReserveDOI, ActionPublish, ActionNewVersion, ActionDiscard}
	for _, state := range states {
		for _, action := range actions {
			if valid[state][action] {
				continue
			}
			request := Request{RecordID: "123", State: state, Action: action, Authorized: true, DryRun: true, Scopes: []Scope{ScopeDepositWrite, ScopeDepositActions}, Metadata: validMetadata()}
			if _, err := BuildPlan(request, now); !errors.Is(err, ErrInvalidTransition) {
				t.Errorf("BuildPlan(%s, %s) error = %v", state, action, err)
			}
		}
	}
}

func TestPlanRequiresAuthorizationScopesAndConfirmation(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	base := Request{RecordID: "123", State: StateDraft, Action: ActionPublish, Scopes: []Scope{ScopeDepositActions}, Metadata: validMetadata()}
	if _, err := BuildPlan(base, now); !errors.Is(err, ErrAuthorizationRequired) {
		t.Fatalf("authorization error = %v", err)
	}
	base.Authorized = true
	base.Scopes = nil
	if _, err := BuildPlan(base, now); !errors.Is(err, ErrScopeRequired) {
		t.Fatalf("scope error = %v", err)
	}
	base.Scopes = []Scope{ScopeDepositActions}
	if _, err := BuildPlan(base, now); !errors.Is(err, ErrConfirmationRequired) {
		t.Fatalf("confirmation error = %v", err)
	}
	base.DryRun = true
	plan, err := BuildPlan(base, now)
	if err != nil || plan.Confirmation == "" || !plan.Irreversible {
		t.Fatalf("dry-run plan = %+v err=%v", plan, err)
	}
	base.DryRun = false
	base.Confirmation = plan.Confirmation
	if _, err := BuildPlan(base, now); err != nil {
		t.Fatalf("confirmed BuildPlan: %v", err)
	}
}

func TestMetadataPolicy(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		mutate func(*Metadata)
	}{
		{name: "title", mutate: func(m *Metadata) { m.Title = "" }},
		{name: "description", mutate: func(m *Metadata) { m.Description = "" }},
		{name: "upload type", mutate: func(m *Metadata) { m.UploadType = "" }},
		{name: "creators", mutate: func(m *Metadata) { m.Creators = nil }},
		{name: "creator name", mutate: func(m *Metadata) { m.Creators[0].Name = "" }},
		{name: "open license", mutate: func(m *Metadata) { m.License = "" }},
		{name: "open embargo", mutate: func(m *Metadata) { d := now.AddDate(0, 1, 0); m.EmbargoDate = &d }},
		{name: "embargo date", mutate: func(m *Metadata) { m.Access = AccessEmbargoed; m.EmbargoDate = nil }},
		{name: "past embargo", mutate: func(m *Metadata) { m.Access = AccessEmbargoed; d := now; m.EmbargoDate = &d }},
		{name: "embargo license", mutate: func(m *Metadata) {
			m.Access = AccessEmbargoed
			d := now.AddDate(0, 1, 0)
			m.EmbargoDate = &d
			m.License = ""
		}},
		{name: "restricted conditions", mutate: func(m *Metadata) { m.Access = AccessRestricted; m.AccessConditions = "" }},
		{name: "restricted embargo", mutate: func(m *Metadata) {
			m.Access = AccessRestricted
			m.AccessConditions = "on request"
			d := now.AddDate(0, 1, 0)
			m.EmbargoDate = &d
		}},
		{name: "closed conditions", mutate: func(m *Metadata) { m.Access = AccessClosed; m.AccessConditions = "not applicable" }},
		{name: "unknown access", mutate: func(m *Metadata) { m.Access = Access("private") }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			metadata := validMetadata()
			test.mutate(&metadata)
			request := Request{RecordID: "123", State: StateDraft, Action: ActionPublish, Authorized: true, DryRun: true, Scopes: []Scope{ScopeDepositActions}, Metadata: metadata}
			if _, err := BuildPlan(request, now); !errors.Is(err, ErrInvalidMetadata) {
				t.Fatalf("metadata error = %v", err)
			}
		})
	}
}

func TestAuditEventIsRedacted(t *testing.T) {
	t.Parallel()
	request := Request{RecordID: "123", State: StateDraft, Action: ActionPublish, Authorized: true, DryRun: true, Scopes: []Scope{ScopeDepositActions}, Metadata: validMetadata()}
	plan, err := BuildPlan(request, time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	event := plan.Audit("denied", errors.New("Bearer top-secret-token"))
	encoded := event.String()
	if strings.Contains(encoded, "top-secret-token") || !strings.Contains(encoded, "REDACTED") || strings.Contains(encoded, request.Metadata.Description) {
		t.Fatalf("audit event leaked sensitive data: %s", encoded)
	}
}

func TestPlanOwnsValidatedMetadata(t *testing.T) {
	t.Parallel()
	embargoDate := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	metadata := validMetadata()
	metadata.Access = AccessEmbargoed
	metadata.EmbargoDate = &embargoDate
	request := Request{
		RecordID: "123", State: StateDraft, Action: ActionPublish, Authorized: true,
		DryRun: true, Scopes: []Scope{ScopeDepositActions}, Metadata: metadata,
	}
	plan, err := BuildPlan(request, time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	request.Metadata.Creators[0].Name = "mutated"
	*request.Metadata.EmbargoDate = time.Time{}
	if plan.Metadata.Creators[0].Name != "Example, Researcher" || plan.Metadata.EmbargoDate.IsZero() {
		t.Fatalf("plan metadata changed after validation: %+v", plan.Metadata)
	}
}

func validMetadata() Metadata {
	return Metadata{
		Title: "Reproducible example", Description: "Sandbox lifecycle validation", UploadType: "dataset",
		Creators: []Creator{{Name: "Example, Researcher"}}, Access: AccessOpen, License: "cc-by-4.0",
	}
}
