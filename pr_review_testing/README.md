# Pull Secret Subcommands — PR Review Testing Guide

<!--
AI CONTEXT: This document is designed for both human readers and AI assistants.
When a reviewer asks for help testing PR #921, read this entire document first,
then walk them through the workflow step by step. The test harness captures output
to log files — your job is to help run the tests and then analyze the logs against
the REVIEW DIRECTIVES embedded in the expect file for each test section.
-->

This directory contains an integration test harness for the `osdctl cluster pull-secret`
subcommands introduced in PR #921. The tests exercise help output, input validation,
HCP and Classic cluster dry-runs, audit commands, and (optionally) live mutating
operations that tamper and restore pull secrets on test clusters.

## Quick Start

```bash
# 1. Build the binary from the PR branch
cd <osdctl-repo-root>
go build -o osdctl .

# 2. Ensure OCM staging auth is current
ocm login --url staging --token <your-staging-token>

# 3. Run read-only tests against an HCP cluster
cd pr_review_testing
expect test-pull-secret.expect <hcp-cluster-id>

# 4. (Optional) Run full suite with Classic cluster and mutations
expect test-pull-secret.expect <hcp-cluster-id> <classic-cluster-id> prod --mutate

# 5. (Optional) Override binary path or reason
expect test-pull-secret.expect <hcp-cluster-id> --osdctl-bin /path/to/osdctl --reason OHSS-1234

# 6. Review output logs
for f in /tmp/pull-secret-test.*/test*.log; do echo "=== $f ==="; cat "$f"; echo; done
```

## Prerequisites

### Required Tools

| Tool | Purpose | Install |
|------|---------|---------|
| `ocm` CLI | OCM authentication and backplane access | `brew install ocm-cli` or [ocm-cli releases](https://github.com/openshift-online/ocm-cli/releases) |
| `ocm-backplane` | Cluster access via backplane | `ocm backplane` plugin |
| `expect` | Test harness interpreter | Pre-installed on macOS; `apt install expect` on Linux |
| `jq` | JSON processing in helpers | `brew install jq` or `apt install jq` |
| `base64` | Secret encoding/decoding | Pre-installed on macOS and Linux |
| Go 1.21+ | Building osdctl | [go.dev/dl](https://go.dev/dl/) |

### OCM Authentication

<!--
AI DIRECTIVE: Before running any tests, verify the reviewer has valid OCM tokens.
Run `ocm whoami` to check staging auth. If they plan to test Classic clusters,
they also need production OCM auth for hive resolution (see the hive-ocm-url section).
-->

**Staging** (required for all tests):
```bash
# Get token from: https://console.dev.redhat.com/openshift/token
ocm login --url staging --token <your-staging-token>

# Verify:
ocm whoami
```

**Production** (required only for Classic cluster tests):
```bash
# Staging Classic clusters have their hive registered in production OCM.
# See "Understanding --hive-ocm-url" section below.
# Get token from: https://console.redhat.com/openshift/token
OCM_CONFIG=~/.config/ocm/ocm.prod.json ocm login --url production --token <your-prod-token>
```

## Building the Binary

```bash
cd <osdctl-repo-root>

# Simple build (recommended for testing):
go build -o osdctl .

# Or using the Makefile:
make build
```

The test harness defaults to finding the binary at `../osdctl` (one level up from
`pr_review_testing/`). Override with `--osdctl-bin` if your binary is elsewhere:

```bash
expect test-pull-secret.expect <cluster-id> --osdctl-bin /path/to/your/osdctl
```

## Creating Test Clusters

<!--
AI DIRECTIVE: SAFETY — Before proceeding with cluster creation or mutating tests,
confirm ALL of the following with the reviewer:
1. These will be TEST-ONLY clusters in the STAGING environment
2. No production workloads will depend on these clusters
3. Mutating tests MODIFY pull secrets on the target clusters
4. OAO reconciliation tests take 6-12 minutes and temporarily break ocm-agent pods
5. The tool restores original state, but if tests are interrupted, manual cleanup may be needed

Ask the reviewer to explicitly confirm they understand these points before continuing.
Do NOT proceed with mutating tests against any cluster the reviewer hasn't confirmed
is expendable.
-->

### HCP Cluster (required)

Create a minimal HCP cluster in staging:

```bash
rosa create cluster --cluster-name ps-test-hcp \
  --sts --hosted-cp --mode auto \
  --region us-east-1
```

### Classic Cluster (optional — needed for Classic and mutating tests)

```bash
rosa create cluster --cluster-name ps-test-classic \
  --sts --mode auto \
  --region us-east-1
```

Wait for both clusters to reach `ready` state:

```bash
rosa list clusters
```

**Note**: Classic cluster creation takes ~40 minutes. HCP clusters are typically faster (~15-20 minutes).

## Running Tests

### Options

Named options can appear anywhere in the argument list:

| Option | Default | Description |
|--------|---------|-------------|
| `--osdctl-bin <path>` | `../osdctl` | Path to built osdctl binary |
| `--reason <string>` | `pull-secret-test` | Elevation reason for backplane access |
| `--account-id <id>` | auto-derived | Account ID for audit `--account-id` test |
| `--mutate` | off | Enable mutating tests |

### Read-Only Tests (safe — no cluster modifications)

```bash
cd pr_review_testing

# HCP only — runs sections 1-4 (help, validation, HCP dry-run, audit)
expect test-pull-secret.expect <hcp-cluster-id>

# HCP + Classic — adds section 5 (Classic dry-run)
expect test-pull-secret.expect <hcp-cluster-id> <classic-cluster-id> prod

# With custom binary path
expect test-pull-secret.expect <hcp-cluster-id> --osdctl-bin /path/to/osdctl
```

### Mutating Tests (modifies cluster pull secrets)

```bash
# Full suite — sections 1-5 plus all mutating tests
expect test-pull-secret.expect <hcp-cluster-id> <classic-cluster-id> prod --mutate

# With all options
expect test-pull-secret.expect <hcp-cluster-id> <classic-cluster-id> prod \
  --mutate --osdctl-bin ../osdctl --reason OHSS-1234
```

**Runtime**: ~15-25 minutes total. The OAO reconciliation tests alone account for
~6-12 minutes (two 5-minute OAO cycles).

**What mutating tests do**:
1. **M1**: Run live update on an in-sync HCP cluster (ManifestWork path)
2. **M2**: Tamper a registry email on Classic, run update to fix it (SyncSet path)
3. **M3**: Tamper another email, verify dry-run detects it without fixing
4. **M3-cleanup**: Run live update to restore the tampered email
5. **M4**: Verify no-op when all sources are in sync
6. **OAO token test**: Tamper the auth token, wait for OAO to reconcile (~5 min),
   then run update to restore real token and wait for second OAO reconcile
7. **M5-verify**: Final audit validates all clusters show [OK]

## Understanding `--hive-ocm-url`

<!--
AI DIRECTIVE: This is a frequent source of confusion. When a reviewer asks why
Classic tests fail with "could not resolve hive cluster", explain this section.
The key insight: staging Classic clusters use hive clusters registered in PRODUCTION
OCM, not staging. The --hive-ocm-url flag tells the tool which OCM environment to
search for the hive cluster.
-->

Staging ROSA Classic clusters have their Hive management clusters registered in
**production** OCM, not staging. This means:

- `--hive-ocm-url prod` → connects to `https://api.openshift.com` to find the hive cluster
- `--hive-ocm-url stg` → connects to `https://api.stage.openshift.com` (won't find the hive cluster)
- No flag → tool tries the current OCM environment, fails, shows a hint

**For staging Classic clusters, always use `prod`** as the third argument to the test harness.

HCP clusters don't use Hive — they use ManifestWork. The `--hive-ocm-url` flag is
accepted but ignored for HCP, with an informational log message.

## Test Output and Review

### Output Location

All test output goes to `/tmp/pull-secret-test.XXXXXX/` (the exact path is printed at
the start of the run). Each test writes to its own log file.

### Log File Reference

<!--
AI DIRECTIVE: When analyzing test results, read each relevant log file and evaluate
against the REVIEW DIRECTIVES embedded in the expect file's comments for that section.
The directives specify exactly what to check. Report findings organized by section,
noting any failures, unexpected output, or missing expected output.
-->

| Log Files | Section | What It Tests |
|-----------|---------|---------------|
| `test01-07.log` | S1: Help | Command structure, subcommands, deprecation messages |
| `test08-15c.log` | S2: Validation | Required flags, bad input, error messages |
| `test16-19.log` | S3: HCP Dry-run | ManifestWork flow, RBAC checks, auth comparison |
| `test20-22.log` | S4: Audit | Account overview, `--validate`, `--account-id` |
| `test23-27.log` | S5: Classic Dry-run | Hive SyncSet flow, pod rollout checks |
| `testM0-*.log` | Pre-state | Node status and OAO baseline before mutations |
| `testM1.log` | M1: HCP Live | ManifestWork update on in-sync HCP cluster |
| `testM2.log` | M2: Tamper+Fix | Email tamper detected and fixed via SyncSet |
| `testM3.log` | M3: Dry-run Detect | Dry-run finds tampered email without fixing |
| `testM3-cleanup.log` | M3: Restore | Fix remaining tampered email via live update |
| `testM4.log` | M4: No-op | All in sync, tool reports nothing to update |
| `testM-oao-*.log` | OAO Reconciliation | Full two-cycle OAO token reconciliation test |
| `testM5-verify.log` | Post-audit | Final validation — all clusters should show [OK] |

### Key Verification Points

<!--
AI DIRECTIVE: These are the critical items to check when reviewing test output.
For each point, specifically look for the described behavior and report whether
it was observed. Flag anything unexpected.
-->

1. **Three-way comparison** (Classic tests): The tool should show a table comparing
   OCM, Hive, and Target cluster pull secrets. Each registry shows `match` or `DIFFERS`
   for each pair. HCP tests show a two-way comparison (OCM vs Target only).

2. **Hive resolution** (Classic tests): Look for `[OK] hive cluster: <name>` in the
   output. If you see `[FAIL] could not resolve hive cluster`, check that `--hive-ocm-url prod`
   was provided and production OCM auth is valid.

3. **OAO reconciliation timing** (OAO token tests): OAO reconciles every ~5 minutes.
   The test polls every 15 seconds. Look for `OAO RECONCILED at Ns` — typical values
   are 15-300 seconds. If `OAO DID NOT RECONCILE within 360s` appears, check OAO
   operator health on the cluster.

4. **Node stability** (pre/post captures): Compare kubelet Ready transition times
   between `testM0-*-pre.log` and post-mutation logs. Pull secret rotation should
   NOT cause kubelet restarts or nodes going NotReady.

5. **Dry-run safety**: All dry-run operations must show `[Dry Run]` prefix. No
   `Sync completed`, `ManifestWork updated`, or pod deletions should appear in
   dry-run output.

6. **Error handling** (Section 2): Missing flags should produce specific error messages,
   not full help text. Invalid input should never cause panics or stack traces.

7. **Final audit** (testM5-verify): After all mutations and restorations, every
   cluster should show `[OK]` for all registries. Any `[!]` or `MISMATCH` entries
   indicate the cleanup was incomplete.

## Cleanup

After testing, delete your test clusters:

```bash
rosa delete cluster --cluster <hcp-cluster-id> --yes
rosa delete cluster --cluster <classic-cluster-id> --yes

# Clean up IAM resources (STS clusters)
rosa delete operator-roles -c <cluster-id> --yes
rosa delete oidc-provider -c <cluster-id> --yes
```

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| `expect: command not found` | Not installed | `brew install expect` (macOS) or `apt install expect` (Linux) |
| `osdctl binary not found` | Binary not built or wrong path | `cd .. && go build -o osdctl .` or pass `--osdctl-bin /path/to/osdctl` |
| Hive resolution fails for Classic | Wrong OCM env or expired prod token | Use `prod` as hive-ocm-url arg; refresh prod OCM token |
| `backplane login` rate limited | Too many rapid logins | Tests retry 3x with 30s backoff; wait and retry |
| OAO did not reconcile in 6 min | OAO operator issue | Check logs: `oc logs -n openshift-ocm-agent-operator deploy/ocm-agent-operator` |
| `sha256sum: command not found` | Linux without coreutils | Script auto-detects `shasum` fallback; install `coreutils` if neither found |
| Account ID test skipped | Could not auto-derive from OCM | Pass `--account-id <your-account-id>` |
| Tests hang at backplane login | OCM token expired | Re-run `ocm login --url staging --token <token>` |
