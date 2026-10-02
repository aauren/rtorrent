# rtorrent

[![Go Reference](https://pkg.go.dev/badge/github.com/aauren/rtorrent/rtorrent.svg)](https://pkg.go.dev/github.com/aauren/rtorrent/rtorrent)
[![ci](https://github.com/aauren/rtorrent/actions/workflows/ci.yml/badge.svg)](https://github.com/aauren/rtorrent/actions/workflows/ci.yml)
[![golangci-lint](https://github.com/aauren/rtorrent/actions/workflows/lint.yml/badge.svg)](https://github.com/aauren/rtorrent/actions/workflows/lint.yml)

Package `rtorrent` implements a client for rTorrent's XML-RPC interface. MIT Licensed.

## Current State

This began as a fork of [mdlayher/rtorrent](https://github.com/mdlayher/rtorrent), which covers the
global transfer counters and the download list. Since then I've added per-download detail lookups,
the basic download commands (start, stop, close, erase, hash check, and moving the directory),
and file, peer, and tracker services.

It's still pretty limited next to rTorrent's full command reference, and I tend to add commands as
I need them rather than aiming for coverage. The API isn't settled either, so you'll want to pin a
commit if you depend on it.

## Install

```bash
go get github.com/aauren/rtorrent
```

## Usage

Point the client at your rTorrent instance's XML-RPC endpoint, then hang the services off of it:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aauren/rtorrent/rtorrent"
)

func main() {
	c, err := rtorrent.New("http://127.0.0.1:8080/RPC2", nil)
	if err != nil {
		log.Fatalf("connecting to rTorrent: %v", err)
	}
	defer c.Close()

	rate, err := c.DownloadRate()
	if err != nil {
		log.Fatalf("fetching download rate: %v", err)
	}
	fmt.Printf("downloading at %d B/s\n", rate)

	ds := &rtorrent.DownloadService{C: c}
	hashes, err := ds.Active()
	if err != nil {
		log.Fatalf("listing active downloads: %v", err)
	}

	ts := &rtorrent.TrackerService{C: c}
	for _, hash := range hashes {
		name, err := ds.BaseFilename(hash)
		if err != nil {
			log.Fatalf("fetching name for %s: %v", hash, err)
		}

		trackers, err := ts.TrackerWithDetails(context.Background(), rtorrent.NewTrackerNoIndex(hash),
			[]rtorrent.TrackerField{rtorrent.FieldURL, rtorrent.FieldSuccessCounter})
		if err != nil {
			log.Fatalf("fetching trackers for %s: %v", name, err)
		}

		for _, tracker := range trackers {
			fmt.Printf("%s: %s\n", name, tracker.GetFieldValueAsString(rtorrent.FieldURL))
		}
	}
}
```

`AllTrackerFields()` returns every field a tracker can be asked for, and the full API is documented
on [pkg.go.dev](https://pkg.go.dev/github.com/aauren/rtorrent/rtorrent).

`DownloadService.DownloadDetails()` fetches any set of `d.*` commands for one info-hash in a single
`system.multicall`, which is a lot cheaper than one call per field once you're polling.

`SystemService` covers rTorrent as a whole (its version, `directory.default`, and the global
throttles), `DownloadService.LoadRaw()` adds a download from the bytes of a `.torrent` file, and
`DownloadService.SetPriority()` and `FileService.SetPriority()` set download and file priorities.
For anything the services don't wrap, `Client.Call()` sends any command as-is, something like the
following:

```go
// Global commands still take an empty target as their first argument
if _, err := c.Call("throttle.max_uploads.global.set", "", 50); err != nil {
	log.Fatalf("setting the upload slots: %v", err)
}
```

## Development

The make targets run inside Docker by default, so pass `BUILD_IN_DOCKER=false` if you'd rather use
your local toolchain:

- **`make all`** - runs `lint`, `test`, and `build`, which is what CI checks
- **`make test`** - runs the tests with the race detector, shuffling, and coverage
- **`make lint`** - runs `golangci-lint`
- **`make genmoqs`** - regenerates the gomock mocks, pinned by the `tool` directive in `go.mod`
- **`make gofmt-fix`** - applies `goimports` and `gofmt -s`

PR's are always welcome! Thanks Matt for the original package that this is built on.
