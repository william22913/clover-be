---
feature_code: FEAT-001
service: clover
version: 1
spec_version: 1
testcase_version: 1
architecture_version: 1
database_version: 1
implementation_plan_version: 1
last_updated: 2026-11-13
---

# FEAT-001 — Version Lock

Version lock for this feature's helicopter-view overviews (`TRD.md`,
`BackendSchema.md`, `ImplementationPlan.md`), maintained by
`/code-preparation`. Not a source of truth for anything else — `spec.md`,
`testcase.md`, `architecture.md` and `database.md` each carry their own
version/spec_version fields; this file only pins the summaries above to the
design state they were last refreshed against.

- `spec_version: 1` — `spec.md`'s merged version (PR #56); the design and every
  test case are built against it.
- `testcase_version: 1` — `testcase.md`'s merged version (PR #57), which pins
  `spec_version: 1`, matching the spec, and has no open questions.
- `architecture_version: 1` — `architecture.md`'s version, the first design
  pass for this feature (PR #59).
- `database_version: 1` — `database.md`'s version, the first design pass; all
  three tables are new and the single migration batch belongs to this feature
  (PR #59).
- `implementation_plan_version: 1` — the six-part plan in this PR.
- `version: 1` — the feature's code-preparation document set as a whole.
