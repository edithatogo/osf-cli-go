# Quality Review Automation

## Objective

Prevent incomplete stub work from being marked done and require an automatic review-fix-continue loop at each phase boundary.

## Acceptance Criteria

- Workflow forbids marking stubbed behavior complete.
- Phase exit requires `$conductor-review`, safe fix application, re-review, and phase evidence.
- CI includes a production-code anti-stub check.
- Agents continue automatically to the next phase unless blocked by credentials, destructive actions, live writes, dependency/license changes, or product-scope ambiguity.
