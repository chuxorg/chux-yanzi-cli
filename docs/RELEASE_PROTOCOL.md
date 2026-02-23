# Yanzi Release Protocol

## Branch Model

- All work must occur on feature branches.
- No direct commits to development or master.
- Each phase is completed on a feature branch.
- After phase completion:
  - Create PR → merge into development.
- After QA validation:
  - Create PR → merge development → master.

## Versioning

- Version type (major/minor/patch) is decided by human before release.
- Tags are only created after merge into development (QA).
- Production tags are created manually on master.
- No tag reuse.
- No pseudo-versions.
- No replace directives in go.mod.

## QA Flow

1. Merge feature branch → development.
2. Tag development with semver (major/minor/patch).
3. QA build runs.
4. Validate binaries.
5. Merge development → master.
6. Tag master for production release.

---
