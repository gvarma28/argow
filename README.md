# argow

A small CLI that bulk-runs ArgoCD operations (sync, restart) across all applications in one or more ArgoCD projects, using the `argocd` CLI under the hood.

## What it does

Given a repo URL and one or more ArgoCD project names, `argow`:

1. Verifies the `argocd` CLI is installed and that you're logged in.
2. Lists every application in the given project(s) that comes from the given repo (`argocd app list --repo=... -p=...`).
3. Shows you the list and asks for confirmation before doing anything.
4. Runs the requested operation(s) — `sync` and/or `restart` — against each application concurrently (10 workers at a time), retrying each operation up to 3 times with backoff before giving up on that application.

`sync` and `restart` are always applied in that order (sync before restart), regardless of the order you pass them in.

## Prerequisites

- Go 1.27+ (to build)
- The [`argocd` CLI](https://argo-cd.readthedocs.io/en/stable/cli_installation/) installed and on your `PATH`
- Already logged in via `argocd login <your-server>`

## Build

```sh
go build -o argow .
```

## Usage

```sh
./argow -repo=<repo-url> -project=<project1>,<project2> -operation=<op1>,<op2>
```

Any flag left empty will be prompted for interactively.

### Flags

| Flag | Description | Example |
|---|---|---|
| `-repo` | GitHub/GitLab repo URL used to filter applications | `https://github.com/org/repo.git` |
| `-project` | Comma-separated ArgoCD project name(s) | `server-prod,server-test` |
| `-operation` | Comma-separated operation(s) to run | `restart,sync` |

Supported operations: `sync`, `restart`.

### Example

```sh
./argow -repo=https://github.com/org/repo.git -project=server-prod -operation=sync,restart
```

This lists every `server-prod` application sourced from that repo, prompts for confirmation, then syncs and restarts each one, printing per-application progress and retrying transient failures.
