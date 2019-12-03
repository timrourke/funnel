# funnel

[![CircleCI](https://circleci.com/gh/timrourke/funnel.svg?style=svg)](https://circleci.com/gh/timrourke/funnel) [![codecov](https://codecov.io/gh/timrourke/funnel/branch/master/graph/badge.svg)](https://codecov.io/gh/timrourke/funnel)

```
Funnel is a tool for quickly saving files to AWS S3.

Usage:
  funnel [OPTIONS] [PATHS]

Examples:
funnel --region=us-east-1 --bucket=some-cool-bucket /some/directory

Flags:
  -b, --bucket string                The AWS S3 bucket you want to save files to
      --delete-file-after-upload     Whether to delete the uploaded file after a successful upload
  -h, --help                         help for funnel
  -n, --num-concurrent-uploads int   Number of concurrent uploads (default 10)
  -r, --region string                The AWS region your S3 bucket is in, eg. "us-east-1"
      --version                      version for funnel
  -w, --watch                        Whether to watch a path for changes
```
