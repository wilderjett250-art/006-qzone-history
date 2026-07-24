# Project 006: Qzone History Recovery and Local Archive

[中文](README.md) · [Upstream provenance](UPSTREAM.md) · [Original README](README.upstream.md) · [Quick start](quickStart.md)

This is a local recovery tool for personal Qzone backups. It uses QQ's QR-code login flow, collects currently visible moments, the “related to me” activity feed, and board messages, then reconstructs historical records from residual like, comment, view, forward, and message events.

This Project 006 repository is based on commit `666f8dd4e7fb3ad88248f7818e2f95c16f48adb6` of [ZHChen2000/qzone-history](https://github.com/ZHChen2000/qzone-history). The upstream project is licensed under Apache License 2.0. The license, original author attribution, upstream README, and quick-start guide are preserved.

## What it does

- Logs in through QQ's QR-code flow and keeps cookies on the local machine.
- Imports moments and board messages that remain directly visible.
- Deep-scans the “related to me” activity feed using target years and large offsets.
- Reconstructs deleted moments from residual likes, comments, views, and forwards.
- Merges board API results with message traces found in the activity feed.
- Exports complete JSON, raw activity JSON, and an offline HTML timeline.
- Provides a loopback-only Web console for logs, progress, activity count, and earliest date.

## Recovery model

The activity feed is not a complete snapshot of the original moments. It is a collection of event evidence. Even after a moment is deleted, related events may still retain text fragments, participants, timestamps, or images.

The scanner therefore combines several overlapping strategies:

1. Sequential pagination establishes a recent-data baseline.
2. Sparse offset scanning locates older reachable regions.
3. Fine-grained overlap scans fill historically discontinuous ranges.
4. Half-year time windows recover records missed by offset jumps.
5. `set`, `scope`, and legacy `feeds3` variants provide independent coverage.
6. All results flow through shared deduplication before reconstruction.

Offset and year are not linearly related. Deletions, permission changes, and feed gaps shift record positions, so the displayed earliest date is the authoritative completion signal.

## Build from source

The repository's `go.mod` specifies Go `1.25.2`.

```powershell
go mod verify
go test ./...
go vet ./...
go build -ldflags="-H windowsgui -s -w -X qzone-history/version.Version=v0.0.4" -o qzone-history-gui.exe ./cmd/main.go
```

No prebuilt executable is committed. The Windows binary is expected to be built from the auditable source currently checked out.

## Run

1. Start the locally built `qzone-history-gui.exe`.
2. Open `http://127.0.0.1:17890`.
3. Request a QR code and confirm login with the QQ mobile app.
4. Choose a target year and maximum offset.
5. Watch the activity count and earliest date; increase the offset if the scan has not reached the requested year.

See [quickStart.md](quickStart.md) for the full offset table, time estimates, controls, and output layout.

## Local data boundary

The program creates `session.db`, a per-account SQLite database, JSON exports, and an offline HTML viewer next to the executable. These files may contain cookies, account identifiers, and personal Qzone content. They are excluded by `.gitignore` and are not part of this source repository.

The Web console binds only to `127.0.0.1:17890`. The tool is intended only for backing up the user's own data or data for which explicit authorization has been granted.

## Project 006 changes

- Imported a clean source baseline from upstream commit `666f8dd4…`.
- Excluded the upstream prebuilt executable and all local runtime/account data.
- Added explanatory Chinese comments around scan coverage, deduplication, time inference, reconstruction, offsets, and GUI lifecycle.
- Added bilingual documentation and explicit security boundaries.
- Preserved the Apache-2.0 license, upstream attribution, original README, and quick-start guide.

## Validation

- `go mod verify`: passed.
- `go test ./...`: passed.
- `go vet ./...`: passed.
- High-confidence secret scan: no private keys, cloud access keys, or GitHub tokens found.
- Repository audit: no executable, session database, account database, or Qzone export is tracked.

## Upstream and license

- Upstream: [ZHChen2000/qzone-history](https://github.com/ZHChen2000/qzone-history)
- Original author: ZHChen
- Source baseline: `666f8dd4e7fb3ad88248f7818e2f95c16f48adb6`
- License: [Apache License 2.0](LICENSE)

The documentation and explanatory comments in this repository do not alter upstream authorship.
