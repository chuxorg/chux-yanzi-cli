# Yanzi CLI

## What It Is
Command-line interface for interacting with Yanzi Library.

## Install
For the simplest install, see:
https://github.com/chuxorg/yanzi

(Development users can build from source with Go.)

## Quick Start
```sh
yanzi capture --author "Ada" --prompt-file prompt.txt --response-file response.txt
yanzi verify <id>
yanzi chain <id>
```

## Commands
- `capture`: Create a new intent record via the library API.
- `verify`: Verify an intent by id.
- `chain`: Print an intent chain by id.

## Runtime Modes
Yanzi CLI supports two runtime modes selected via `~/.yanzi/config.yaml`.

- Local mode (default): Embedded SQLite storage on the same machine.
- HTTP mode (optional): Remote library server over HTTP.

If the config file is missing, the CLI defaults to local mode. If `mode: http` is set and `base_url` is unreachable, the command fails. There is no automatic fallback.

| Use Case | Mode |
|----------|------|
| Personal dev, single machine | local |
| Team server or shared ledger | http |
| Offline capture | local |
| Remote verification | http |

Example config file: `~/.yanzi/config.yaml`

Example local:
```yaml
mode: local
db_path: ~/.yanzi/yanzi.db
```

Example http:
```yaml
mode: http
base_url: http://localhost:8080
```

## Philosophy
- Thin client
- Talks to library over HTTP
- No embedded storage
- No AI logic

## Contributing
Small disciplined scope. Major feature proposals require discussion first.
