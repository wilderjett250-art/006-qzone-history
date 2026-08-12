# 006 Qzone History Recovery Tool

> With user authorization, organizes traceable Qzone activity history and exports it to local files.

## Problem

The current page may show only part of the history, while deleted posts are difficult to trace and archive offline.

## Demo

![Tool interface](docs/images/gui-overview.png)

Launch the desktop app, scan the QR code, choose a year and range, then export JSON or offline HTML.

Login, scanning, deduplication, reconstruction, and export are joined into one Windows workflow.

## Highlights

- A Windows EXE that ordinary users can launch by double-clicking.
- Deep activity scanning with time-range targeting.
- Aggregates and deduplicates likes, comments, views, and repost events.
- Exports both JSON and offline HTML.

## Tech

`Go · SQLite · Local Web UI · HTML · QQ Space interfaces`

## Reproduce from ZIP

1. Download and fully extract the ZIP from Releases.
2. Double-click `qzone-history-gui.exe`.
3. Click “Get/Refresh QR code” and confirm with your own QQ account.
4. Choose a year and scan range, then open the JSON/HTML files in the output directory.

**Expected result:** After these steps, you should see the project's page, window, device output, or test result.

## Scope and Safety

This is a localized delivery based on an open-source project, not a claim of original authorship; process only Qzone data you are authorized to access and save.

## Contact

Open to technical exchange.
