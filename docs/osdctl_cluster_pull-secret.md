## osdctl cluster pull-secret

Diagnose and manage cluster pull secrets

### Synopsis

Subcommands for inspecting and replacing cluster pull secrets.

### Options

```
  -h, --help   help for pull-secret
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
* [osdctl cluster pull-secret replace](osdctl_cluster_pull-secret_replace.md)	 - Refresh a cluster's pull secret from the cluster owner's OCM account
* [osdctl cluster pull-secret snapshot](osdctl_cluster_pull-secret_snapshot.md)	 - Show pull secret status for all clusters owned by an account

