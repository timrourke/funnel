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

## Installing funnel using go get

1. Install [golang](https://golang.org/)
2. Run `go get github.com/timrourke/funnel`

Assuming you have configured your `$PATH` to include binaries in "$GOPATH/bin",
you should now be able to run `funnel` from anywhere on your system.

## Building funnel from source

1. Install [golang](https://golang.org/) and [dep](https://golang.github.io/dep/)
2. Clone this repository to within your [GOPATH](https://github.com/golang/go/wiki/GOPATH)
3. Run `dep ensure` from within the repository to install its dependencies
4. Run `go build .` to compile the `funnel` binary
5. Move the `funnel` binary to somewhere on your `$PATH` and upload some stuff

## Setting the AWS region

`funnel` will respect the environment variable `AWS_DEFAULT_REGION` if one is
set. Otherwise, pass the AWS region as a CLI flag.

## Watching paths for changes

If you want to continually poll a directory for new files and upload them, you
can use the `--watch` flag to enable this behavior. This may be especially
useful in the case you want to also delete files following successful upload.
This will allow you to constantly sync a directory's new contents to an S3
bucket without filling up your disk.
