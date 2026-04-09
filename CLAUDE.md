# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Mackerel plugin that executes PromQL queries against a Prometheus server and outputs results in Mackerel's metric format (tab-separated: key, value, timestamp).

## Build & Test Commands

```bash
make                    # Build binary
make test               # Run tests (go test -race ./...)
go test -run TestName ./lib/  # Run a single test
```

## Release

Releases are handled by GoReleaser via GitHub Actions (`.goreleaser.yml`). Push a `v*` tag to trigger a release build for linux/darwin (amd64/arm64).

## Architecture

- `main.go` — CLI entry point, parses flags and creates `promq.Plugin`
- `lib/prometheus.go` — Core logic: `Plugin` struct, Prometheus API query (`fetch`), metric key formatting (`formatKey`), output (`Run`)
- `lib/prometheus_test.go` — Unit tests for metric formatting and key templating

The plugin queries Prometheus via `prometheus/client_golang` API, receives vector results, formats metric keys using a `{label}` template syntax (e.g., `promq.{job}.{instance}`), and prints tab-separated output. Non-alphanumeric characters in label values are normalized to `_`.

## Key Behaviors

- `-emit-zero`: When enabled and query returns no results, emits a single metric with value `0` using the format string as-is
- Label values in metric keys are sanitized: anything matching `[^\w-]` becomes `_`
- Unmatched `{label}` placeholders become `_`
