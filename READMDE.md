# AWS Nuke

AWS Nuke is trageted in cleanup up a whole AWS account periodically. This might be useful for regullarly nuking all objects in an AWS account

## State

Supported resources:

* S3 Buckets


## Usage


```
# List affected resources
./aws-nuke s3
INFO[0000] s3 buckets to delete: []                      app=aws-nuke

# Gonna nuke 'em all
./aws-nuke s3 --force-destroy
DEBU[0001] ignoring bucket                               app=aws-nuke bucket=my-important-store module=s3
```
