# Login Tests (`login-tests`)

Integration test command to validate OSDCTL's OCM and backplane client connection functions.

## Purpose

The `login-tests` command exercises and validates OSDCTL's shared utility functions related to OCM and backplane client connections. This provides a way to test these functions without running potentially intrusive production commands that may modify cluster state.

## What It Tests

This test validates the following shared utility functions:

### OCM Connection Functions (`pkg/utils`)
- `CreateConnection()` - Legacy OCM connection creation
- `GetClusterAnyStatus()` - Cluster retrieval from OCM
- `GetOcmConfigFromFilePath()` - OCM config loading from file
- `GetOCMConfigFromEnv()` - OCM config from environment
- `GetOCMSdkConnBuilderFromConfig()` - OCM SDK connection builder
- `CreateConnectionWithUrl()` - OCM connection with custom URL
- `GetHiveCluster()` - Hive cluster discovery
- `GetHiveClusterWithConn()` - Hive cluster with custom connection
- `GetHiveBPClientForCluster()` - Hive backplane client creation

### Kubernetes Client Functions (`pkg/k8s`)
- `New()` - Standard kube client creation
- `NewWithConn()` - Kube client with custom OCM connection
- `NewAsBackplaneClusterAdminWithConn()` - Elevated kube client

### Common Helper Functions (`cmd/common`)
- `GetKubeConfigAndClient()` - Kube config and client without elevation
- `GetKubeConfigAndClientWithConn()` - Kube config and client with custom connection
- `GetKubeConfigAndClient()` (with elevation) - Elevated kube config and client
- `GetKubeConfigAndClientWithConn()` (with elevation) - Elevated client with custom connection

## Requirements

- Valid OCM credentials configured (environment variables or config file)
- Access to a live OSD/ROSA Classic cluster
- Cluster ID of the target cluster
- Backplane access permissions
- (Optional) Separate OCM credentials for Hive cluster if testing cross-environment scenarios

## Usage

### Basic Test
Test connection to a cluster using default OCM environment:

```bash
./osdctl integrationtests login-tests --cluster-id <cluster-id>
```

### Test with Custom Hive OCM URL
Test cross-environment scenario where target cluster is in one OCM environment and Hive is in another:

```bash
./osdctl integrationtests login-tests \
  --cluster-id <cluster-id> \
  --hive-ocm-url production
```

Or with a full URL:

```bash
./osdctl integrationtests login-tests \
  --cluster-id <cluster-id> \
  --hive-ocm-url https://api.openshift.com
```

### Test with Custom Hive OCM Config File
Provide a separate OCM config file for Hive cluster access:

```bash
./osdctl integrationtests login-tests \
  --cluster-id <cluster-id> \
  --hive-ocm-config /path/to/hive-ocm.json
```

### Verbose Output
Get detailed output of each test step:

```bash
./osdctl integrationtests login-tests \
  --cluster-id <cluster-id> \
  --verbose
```

## Configuration

The test can read Hive OCM URL from the osdctl config file (`~/.osdctl/config.yaml`):

```yaml
hive_ocm_url: https://api.openshift.com
# or
hive_ocm_url: production
```

To test with an empty/unset Hive OCM URL, comment out the `hive_ocm_url` line in the config file.

## What the Test Does

The `login-tests` command runs through the following sequence:

1. **Setup Phase**
   - Creates OCM connection using environment variables
   - Fetches target cluster from OCM
   - Builds Hive OCM configuration (if provided)
   - Creates Hive OCM connection (if separate from target)
   - Discovers Hive cluster for the target cluster

2. **Test Phase** - Executes 9 different test scenarios:
   - Test `k8s.New()` - Standard kube client
   - Test `k8s.NewWithConn()` - Kube client with custom OCM connection
   - Test `k8s.NewAsBackplaneClusterAdminWithConn()` - Elevated kube client
   - Test `GetKubeConfigAndClient()` - Non-admin client and clientset
   - Test `GetKubeConfigAndClientWithConn()` - Non-admin client with custom connection
   - Test `GetKubeConfigAndClient()` with elevation - Admin client and clientset
   - Test `GetKubeConfigAndClientWithConn()` with elevation - Admin client with custom connection
   - Test `GetHiveBPClientForCluster()` without elevation - Hive backplane client
   - Test `GetHiveBPClientForCluster()` with elevation - Elevated Hive backplane client

3. **Validation**
   - Each test performs basic operations (list cluster operators, namespaces, pods)
   - Elevated tests attempt to access restricted resources to verify elevation works
   - Hive tests fetch ClusterDeployment to verify access to Hive cluster

## Test Scenarios

### Single Environment (Default)
Target cluster and Hive cluster are in the same OCM environment:
```bash
./osdctl integrationtests login-tests --cluster-id <cluster-id>
```

### Cross-Environment
Target cluster is in one OCM environment (e.g., staging) and Hive is in another (e.g., production):
```bash
# Set OCM environment vars for target cluster (staging)
export OCM_URL=https://api.stage.openshift.com
export OCM_TOKEN=<staging-token>

# Provide production OCM URL for Hive
./osdctl integrationtests login-tests \
  --cluster-id <cluster-id> \
  --hive-ocm-url production
```

## Exit Status

- **0** - All tests passed successfully
- **Non-zero** - One or more tests failed (error details printed to stderr)

## Notes

- Tests require live cluster access and will fail if credentials are invalid
- Some tests require elevated permissions (backplane-admin)
- Tests perform read-only operations by default but may require cluster-admin for elevated scenarios
- Always verify you're targeting the correct cluster and environment before running
- The test creates temporary OCM and kube client connections that are cleaned up on exit
