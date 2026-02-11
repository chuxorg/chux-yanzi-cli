# Stability

## Command Stability
- `capture`, `verify`, and `chain` are stable and expected to remain available.
- Command behavior should remain consistent across minor and patch releases.

## Flag Compatibility
- Existing flags are backward compatible within a major version.
- Flags may be deprecated before removal.
- Removals or breaking changes happen only in a new major release.

## Versioning
- The CLI follows semantic versioning: `MAJOR.MINOR.PATCH`.
- `MAJOR` for breaking changes, `MINOR` for backward-compatible features, `PATCH` for fixes and docs.
