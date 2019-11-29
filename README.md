# funnel

```
Funnel is a tool for quickly saving files to AWS S3.

Usage:
  funnel [OPTIONS] [PATHS]

Examples:
funnel --region=us-east-1 --bucket=some-cool-bucket /some/directory

Flags:
  -b, --bucket string   The AWS S3 bucket you want to save files to
  -h, --help            help for funnel
  -r, --region string   The AWS region your S3 bucket is in, eg. "us-east-1"
      --version         version for funnel
  -w, --watch           Whether to watch a path for changes
```