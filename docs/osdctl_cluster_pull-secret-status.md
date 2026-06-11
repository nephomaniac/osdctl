## osdctl cluster pull-secret-status

Show pull secret status for all clusters owned by an account

### Synopsis

Show pull secret status for all clusters sharing the same OCM account.

Given any cluster ID, resolves the owner account and lists all clusters
owned by that account. Compares cluster creation dates against the account's
registry credential update timestamps to flag clusters that may have stale
pull secrets.

Use --check to drill into a specific cluster and compare its pull secret
against the current OCM access token auth entries.

```
osdctl cluster pull-secret-status [flags]
```

### Examples

```
  # Overview of all clusters for the account that owns this cluster
  osdctl cluster pull-secret-status -C 1kfmyclusterid --reason "OHSS-1234"

  # Full validation of a specific cluster's pull secret
  osdctl cluster pull-secret-status -C 1kfmyclusterid --reason "OHSS-1234" --check 2abcothercluster
```

### Options

```
      --check string        Drill into a specific cluster ID for full pull secret validation
  -C, --cluster-id string   Any cluster owned by the account (used to resolve the owner)
  -h, --help                help for pull-secret-status
      --reason string       Elevation reason for cluster connections
```

### Options inherited from parent commands

```
      --as string                        Username to impersonate for the operation. User could be a regular user or a service account in a namespace.
      --cluster string                   The name of the kubeconfig cluster to use
      --context string                   The name of the kubeconfig context to use
      --insecure-skip-tls-verify         If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string                Path to the kubeconfig file to use for CLI requests.
  -o, --output string                    Valid formats are ['', 'json', 'yaml', 'env']
      --request-timeout string           The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                    The address and port of the Kubernetes API server
      --skip-aws-proxy-check aws_proxy   Don't use the configured aws_proxy value
  -S, --skip-version-check               skip checking to see if this is the most recent release
```

### SEE ALSO

* [osdctl cluster](osdctl_cluster.md)	 - Provides information for a specified cluster

