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

## Philosophy
- Thin client
- Talks to library over HTTP
- No embedded storage
- No AI logic

## Contributing
Small disciplined scope. Major feature proposals require discussion first.
