## osdctl cluster replace-pull-secret

Replace a cluster's pull secret with current OCM access token data

### Synopsis

Replace a cluster's pull secret with current OCM access token data.

This updates the pull secret on a ROSA HCP or Classic cluster without performing
an ownership transfer. The pull secret is refreshed using the current cluster
owner's OCM access token.

See documentation prior to executing:
https://github.com/openshift/ops-sop/blob/master/hypershift/knowledge_base/howto/replace-pull-secret.md
https://github.com/openshift/ops-sop/blob/master/v4/howto/transfer_cluster_ownership.md

```
osdctl cluster replace-pull-secret [flags]
```

### Examples

```
  # Replace pull secret on a cluster
  osdctl cluster replace-pull-secret --cluster-id 1kfmyclusterid --reason "OHSS-1234"

  # Dry-run to preview without making changes
  osdctl cluster replace-pull-secret --cluster-id 1kfmyclusterid --reason "OHSS-1234" --dry-run
```

### Options

```
  -C, --cluster-id string   The Internal/External Cluster ID or Cluster Name
  -d, --dry-run             Dry-run - show what would change but do not apply
  -h, --help                help for replace-pull-secret
      --reason string       The reason for this command (usually an OHSS or PD ticket)
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

