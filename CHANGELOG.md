# Changelog

## [Unreleased]

### Added
- Initial extraction of EDD Core from the EDD Framework toolkit
- Contract schema `edd/contract/v2` — one YAML file per change
- CLI: `validate`, `verify`, `decide`, `close`, `help`
- Decision engine: deterministic merge/close/status computation
- Evidence bundle format `edd/evidence/v2`
- Combine close: union premerge+live bundles for two-phase contracts
- Hash-bound approvals: intent and technical approval invalidate on change
- 27 tests covering schema validation, execution semantics, decision phases, and hashing
