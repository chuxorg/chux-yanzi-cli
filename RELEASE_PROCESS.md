# Yanzi Release Process

## Overview
Yanzi consists of multiple Go modules:

- chux-yanzi-core (domain primitives)
- chux-yanzi-library (domain + DB + HTTP server)
- chux-yanzi-cli (user-facing binary)
- chux-yanzi-emitter (optional transport client)

Only the CLI repository produces downloadable binaries.

## Versioning Rules
- Semantic versioning required.
- QA releases use: vX.Y.Z-qa
- Production releases use: vX.Y.Z
- Only bump a module’s version if that module’s code changes.
- Versions across modules do NOT need to match.

## Release Order (Bottom-Up)
1. Tag chux-yanzi-core
   - Ensure clean working tree
   - go mod tidy
   - go test ./...
   - git tag -a vX.Y.Z-qa -m "QA release"
   - git push origin vX.Y.Z-qa

2. Tag chux-yanzi-library
   - Update core dependency:
     go get github.com/chuxorg/chux-yanzi-core@vX.Y.Z-qa
   - go mod tidy
   - go test ./...
   - git tag -a vX.Y.Z-qa
   - git push origin vX.Y.Z-qa

3. Tag chux-yanzi-cli
   - Update dependencies:
     go get github.com/chuxorg/chux-yanzi-core@vX.Y.Z-qa
     go get github.com/chuxorg/chux-yanzi-library@vX.Y.Z-qa
   - go mod tidy
   - go test ./...
   - git tag -a vX.Y.Z-qa
   - git push origin vX.Y.Z-qa

## Binary Release
- CLI tag triggers GitHub Action.
- GitHub Action builds platform binaries.
- Artifacts are attached to GitHub release.
- -qa tags are marked prerelease.

## Important Rules
- Never reuse tags.
- Never change module path without bumping version.
- Never rely on pseudo-versions in release.
- Always run go mod tidy before tagging.
- CLI is the only repo that produces artifacts.

## Production Release
Repeat the same process without "-qa".
