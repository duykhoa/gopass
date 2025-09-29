# GoPass UI

## Introduction
GoPass UI is a modern, cross-platform password manager inspired by the Unix `pass` command. It provides a graphical interface and future API for managing secrets, while remaining fully compatible with the `pass` ecosystem.

## Purpose & Motivation
- **Why pay for password managers?** Your secrets belong to you. The Unix `pass` command is a proven way to store passwords securely using your own GPG key and a git remote for sync and backup.
- **Problems with pass:**
  - No graphical UI, only command-line.
  - No SDK or API for integration with other apps or cloud environments.
- **GoPass UI is not a wrapper for pass:**
  - It is a compatible implementation, allowing you to use both the `pass` CLI and GoPass UI interchangeably.
  - Future plans include a REST API for integration with cloud apps and automation.

## Features
- Pass-compatible password store: secrets encrypted with your GPG key, stored in git.
- Cross-platform desktop UI (Linux, macOS, Windows).
- Add, edit, delete, and view secrets with a simple interface.
- Git sync and remote backup.
- Future: API for cloud and app integration.

## How to Run

Unfortunately I can't get the prebuild version for MacOS and Windows, please build from the source instead.

### Prerequisites
- Go 1.25+
- GPG installed (`gpg --version`)
- Git installed (`git --version`)

### Run the App
```sh
# Clone the repo
$ git clone https://github.com/duykhoa/gopass.git
$ cd gopass

# Run the app (Linux, macOS, Windows)
$ make run
```
Or build for your platform:
```sh
# Linux
$ make build
# Windows
$ GOOS=windows GOARCH=amd64 make build
# macOS
$ GOOS=darwin GOARCH=arm64 make build
```
Binaries will be in the `bin/` directory.

## Screenshots

![screenshot1](/assets/screenshot1.png)

## Roadmap
- REST API for secret management
- Cloud integration
- Mobile support

## License
MIT