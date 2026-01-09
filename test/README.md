# OSDCTL Integration Tests

This directory contains integration test utilities for OSDCTL.

## Purpose

The **primary purpose** of integration tests in this directory is to test OSDCTL's packages and shared utility functions as an alternative to running potentially intrusive production commands. Rather than executing production OSDCTL commands that may modify cluster state or require elevated permissions, these test utilities provide a controlled way to exercise and validate the underlying functions in packages like `pkg/utils`, `pkg/k8s`, and `cmd/common`.

Integration tests allow you to:
- Verify utility functions work correctly without running intrusive production commands
- Test connection logic and shared functions in isolation
- Validate multi-environment configurations safely
- Debug and troubleshoot package functionality against live environments

## Test Organization

Each integration test is implemented as a separate Go file (e.g., `loginTests.go`) and provides one or more test commands. Each test file **should have a corresponding markdown file** (e.g., `loginTests.md`) that documents:
- The purpose of the test
- What functions/packages it validates
- Requirements and prerequisites
- Usage examples and command flags
- Expected behavior and exit codes

This pattern allows contributors to easily understand what each test does and how to use it.

## Available Integration Tests

The following integration tests are currently available:

- **[login-tests](integration/loginTests.md)** - Validates OCM and backplane client connection functions

## Building with Integration Tests

Integration tests are **excluded** from default builds to keep the production binary lean. To include them, use the `integrationtest` build tag:

```bash
go build -tags integrationtest -o osdctl .
```

Or using make (if configured):

```bash
make build TAGS=integrationtest
```

## Discovering Available Tests

After building with integration tests enabled, discover available test commands:

```bash
# List all integration test commands
./osdctl -S integrationtests -h

# Get help for a specific test command
./osdctl -S integrationtests <test-name> -h
```

**Note:** The `-S` flag skips the version check for faster execution.

## Running Integration Tests

Integration tests typically require:
- Valid credentials for the target environment (OCM, AWS, etc.)
- Access to live clusters or services
- Appropriate permissions for the operations being tested

Always verify you're targeting the correct environment before running integration tests.

Example:
```bash
./osdctl integrationtests <test-name> [flags]
```

Refer to each test's individual documentation for specific usage instructions.

## Directory Structure

```
test/
├── README.md              # This file
└── integration/           # Integration test implementations
    ├── cmd.go            # Command registration (with build tag)
    ├── cmd_stub.go       # Stub for builds without integration tests
    ├── loginTests.go     # Login/connection validation tests
    └── loginTests.md     # Documentation for login-tests command
```

## Adding New Integration Tests

To add a new integration test:

1. **Create the test file** in `test/integration/` with the `//go:build integrationtest` build tag:
   ```go
   //go:build integrationtest
   // +build integrationtest

   package integration
   ```

2. **Implement your test** as one or more cobra commands

3. **Register the command** in `cmd.go` by adding it to `NewCmdIntegrationTests()`

4. **Create documentation** in a corresponding `.md` file (e.g., `mytest.go` → `mytest.md`) that includes:
   - Purpose and what it tests
   - Requirements
   - Usage examples
   - Expected behavior

5. **Update this README** to link to your new test's documentation in the "Available Integration Tests" section

## Build Tags

- `integrationtest` - Includes all integration test commands in the build
- Default (no tag) - Excludes integration tests from the build

## Important Notes

- Integration tests are **excluded** from default builds
- These tests may modify cluster state or require elevated permissions
- Always verify you're targeting the correct environment before running tests
- Some tests require specific environment variables (e.g., `OCM_CONFIG`, `OCM_URL`)
- Tests are designed to validate shared utility functions, not to perform production operations
