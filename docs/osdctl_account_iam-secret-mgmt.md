## osdctl account iam-secret-mgmt

Rotate IAM credentials and secrets

### Synopsis


Util to rotate managed IAM credentials and related secrets for a given cluster account.
Please use with caution!
'Rotate': operations are intended to be somewhat interactive and will provide information 
followed by prompting the user to verify if and/or how to proceed. 
'Describe': operations 
are intended to provide info and status related to the artifacts to be rotated. 

These operations require the following:
	- A valid cluster-id (not an alias)
	- An elevation reason (hint: This will usually involve a Jira ticket ID.)  
	- A local AWS profile for the cluster environment (ie: osd-staging, rhcontrol, etc), if not provided 'default' is used. 
	- A valid OCM token and OCM_CONFIG.  


```
osdctl account iam-secret-mgmt <clusterId> --reason <reason> <options> [flags]
```

### Examples

```

Basic usage involves selecting a set actions to run:
Choose 1 or more IAM users to rotate: --rotate-managed-admin, --rotate-ccs-admin
-or- 
Choose 1 or more describe actions: --describe-keys, --describe-secrets

# Rotate credentials for IAM user "OsdManagedAdmin" (or user provided by --admin-username)
iam-secret-mgmt $CLUSTER_ID --reason "$JIRA_TICKET" --aws-profile rhcontrol  --rotate-managed-admin

# Rotate credentials for special IAM user "OsdCcsAdmin", print secret access key contents to stdout. 
iam-secret-mgmt $CLUSTER_ID --reason "$JIRA_TICKET" --aws-profile rhcontrol  --rotate-ccs-admin --output-keys --verbose 4

# Rotate credentials for both users "OsdManagedAdmin" and then "OsdCcsAdmin"
iam-secret-mgmt $CLUSTER_ID --reason "$JIRA_TICKET" --aws-profile rhcontrol  --rotate-managed-admin --rotate-ccs-admin --output-keys --verbose 4

# Describe credential-request secrets 
iam-secret-mgmt $CLUSTER_ID --reason "$JIRA_TICKET" --aws-profile rhcontrol  --describe-secrets

# Describe AWS Access keys in use by users "OsdManagedAdmin" and "OsdCcsAdmin"
iam-secret-mgmt $CLUSTER_ID --reason "$JIRA_TICKET" --aws-profile rhcontrol  --describe-keys 

Describe credential-request secrets and AWS Access keys in use by users "OsdManagedAdmin" and "OsdCcsAdmin"
iam-secret-mgmt $CLUSTER_ID --reason "$JIRA_TICKET" --aws-profile rhcontrol  --describe-secrets --describe-keys 
```

### Options

```
      --admin-username osdManagedAdmin*   The admin username to use for generating access keys. Must be in the format of osdManagedAdmin*. If not specified, this is inferred from the account CR.
  -p, --aws-profile string                specify AWS profile from local config to use(default='default')
      --describe-keys                     Print AWS AccessKey info for osdManagedAdmin and osdCcsAdmin relevant cred rotation, and exit
      --describe-secrets                  Print AWS CredentialRequests ref'd secrets info relecant to cred rotation, and exit
  -h, --help                              help for iam-secret-mgmt
      --hive-ocm-url string               (optional) OCM environment URL for hive operations. Aliases: 'production', 'staging', 'integration'. If not specified, uses the same OCM environment as the target cluster.
  -r, --reason string                     (Required) The reason for this command, which requires elevation, to be run (usually an OHSS or PD ticket).
      --rotate-ccs-admin                  Rotate osdCcsAdmin user credentials. Interactive. Use caution!
      --rotate-managed-admin              Rotate osdManagedAdmin user credentials. Interactive. Use caution.
      --save-keys                         Save 'newly created' secret access key contents stdout output during execution
  -v, --verbose int                       debug=4, (default)info=3, warn=2, error=1 (default 3)
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

* [osdctl account](osdctl_account.md)	 - AWS Account related utilities

