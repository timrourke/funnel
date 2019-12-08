# funnel

[![CircleCI](https://circleci.com/gh/timrourke/funnel.svg?style=svg)](https://circleci.com/gh/timrourke/funnel) [![codecov](https://codecov.io/gh/timrourke/funnel/branch/master/graph/badge.svg)](https://codecov.io/gh/timrourke/funnel) [![GoDoc](https://godoc.org/github.com/timrourke/funnel/upload?status.svg)](https://godoc.org/github.com/timrourke/funnel/upload)

```
Funnel is a tool for quickly saving files to AWS S3.

Usage:
  funnel [OPTIONS] [PATHS]

Examples:
funnel --region=us-east-1 --bucket=some-cool-bucket /some/directory

Flags:
  -b, --bucket string                   The AWS S3 bucket you want to save files to
      --delete-file-after-upload        Whether to delete the uploaded file after a successful upload
  -h, --help                            help for funnel
  -n, --num-concurrent-uploads int      Number of concurrent uploads (default 10)
  -r, --region string                   The AWS region your S3 bucket is in, eg. "us-east-1"
  -t, --s3-object-key-template string   The layout template to use for defining the key of an uploaded file (default "{{ filePath }}")
      --version                         version for funnel
  -w, --watch                           Whether to watch a path for changes

```

## Installing funnel using go get

1. Install [golang](https://golang.org/)
2. Run `go get github.com/timrourke/funnel`

Assuming you have configured your `$PATH` to include binaries in `"$GOPATH/bin"`,
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
useful in the case you want to also delete files following successful upload, by
using the funnel option `--delete-file-after-upload`. This will allow you to
constantly sync a directory's new contents to an S3 bucket without filling up
your disk.

Paths are polled one second after the last found file was successfully uploaded
in the last polling operation.

## Customizing the keys of uploaded S3 objects

By default, funnel will assume you want to use the path to the local file on
your computer as the S3 object key when uploading files. For example, this file:

```bash
funnel --region=us-east-1 --bucket=my-cool-bucket relative/path/to/text.txt
```

...will be uploaded to S3 with the full path:

```
s3://my-cool-bucket/relative/path/to/text.txt
```

In some cases, you may not want to use the local path of a file on your computer
as the key of the S3 object. More likely, you'll prefer to organize your files
in S3 to suit your specific needs. To achieve this, you can override the default
template for generating S3 keys with your own. Several Go template functions are
available to define static or dynamic keys.

Overriding the default template for key generation requires using the flag
`--s3-object-key-template`, or `-t` for short.

You can learn more about [Go templates here](https://golang.org/pkg/text/template/).

Given the upload command as seen above, the template functions will have the
following effects:

- `-t "{{ absoluteFilePath }}"` -> `/absolute/path/relative/path/to/text.txt`
- `-t "/{{ dateWithFormat \"2006-01-02\" }}/{{ filePath }}"` -> `/2016-01-02/relative/path/to/text.txt`
- `-t "some_new_name{{ fileExtension }}"` -> `/some_new_name.txt`
- `-t "{{ fileName }}"` -> `/text.txt`
- `-t "{{ fileNameWithoutExtension }}"` -> `/text`
- `-t "{{ filePath }}"` -> `/relative/path/to/text.txt`
- `-t "/some/custom/prefix/of/dirs/{{ filePath }}"` -> `/some/custom/prefix/of/dirs/relative/path/to/text.txt`

Note that the date above, `2006-01-02`, is special as far as Go's date format
parsing is concerned. You can learn more about [how Go parses date formats here](https://gobyexample.com/time-formatting-parsing)
and also [here](https://golang.org/pkg/time/#Time.Format).

Finally, note that string arguments passed to Go template functions must be
wrapped in double quotes to be parsed correctly (eg. `dateWithFormat`), but when
calling funnel, you may need to escape those double quotes (see the
`dateWithFormat` example above).
