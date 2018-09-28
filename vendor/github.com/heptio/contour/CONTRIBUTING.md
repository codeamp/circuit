# Contributing

Thanks for taking the time to join our community and start contributing. 
These guidelines will help you get started with the Contour project.
Please note that we require [DCO sign off](#dco-sign-off).  

## Building from source

This section describes how to build Contour from source.

### Prerequisites

1. *Install Go*

    Contour requires [Go 1.9][1] or later.
    We also assume that you're familiar with Go's [`GOPATH` workspace][3] convention, and have the appropriate environment variables set.

2. *Install `dep`*

    Contour uses [`dep`][2] for dependency management.
   `dep` is a fast moving project so even if you have installed it previously, it's a good idea to update to the latest version using the `go get -u` flag.

    ```
    go get -u github.com/golang/dep/cmd/dep
    ```

### Fetch the source

Contour uses [`dep`][2] for dependency management, but to reduce the size of the repository, does not include a copy of its dependencies.
This might change in the future, but for now use the following command to fetch the source for Contour and its dependencies.

```
go get -d github.com/heptio/contour
cd $GOPATH/src/github.com/heptio/contour
dep ensure -vendor-only
```

Go is very particular when it comes to the location of the source code in your `$GOPATH`.
The easiest way to make the `go` tool happy is to rename Heptio's remote location to something else, and substitute your fork for `origin`.
For example, to set `origin` to your fork, run this command substituting your GitHub username where appropriate.

```
git remote rename origin upstream
git remote add origin git@github.com:davecheney/contour.git
```

This ensures that the source code on disk remains at `$GOPATH/src/github.com/heptio/contour` while the remote repository is configured for your fork.

The remainder of this document assumes your terminal's working directory is `$GOPATH/src/github.com/heptio/contour`.

### Building

To build Contour, run:

```
go build ./cmd/contour
```

This assumes your working directory is set to `$GOPATH/src/github.com/heptio/contour`.
If you're somewhere else in the file system you can instead run:

```
go build github.com/heptio/contour/cmd/contour
```

This produces a `contour` binary in your current working directory.

_TIP_: You may prefer to use `go install` rather than `go build` to cache build artifacts and reduce future compile times.
In this case the binary is placed in `$GOPATH/bin/contour`.

### Running the unit tests

Once you have Contour building, you can run all the unit tests for the project:

```
go test ./...
```

This assumes your working directory is set to `$GOPATH/src/github.com/heptio/contour`. 
If you're working from a different directory, you can instead run:

```
go test github.com/heptio/contour/...
```

To run the tests for a single package, change to package directory and run:

```
go test .
```

_TIP_: If you are running the tests often, you can run `go test -i github.com/heptio/contour/...` occasionally to reduce test compilation times.

## Contribution workflow

This section describes the process for contributing a bug fix or new feature.
It follows from the previous section, so if you haven't set up your Go workspace and built Contour from source, do that first.

### Before you submit a pull request

This project operates according to the _talk, then code_ rule.
If you plan to submit a pull request for anything more than a typo or obvious bug fix, first you _should_ [raise an issue][6] to discuss your proposal, before submitting any code.

### Pre commit CI

Before a change is submitted it should pass all the pre commit CI jobs.
If there are unrelated test failures the change can be merged so long as a reference to an issue that tracks the test failures is provided.

Once a change lands in master it will be built and available at this tag, `gcr.io/heptio-images/contour:master`.
You can read more about the available contour images in the [tagging][7] document.

### Build an image

To build an image of your change using Contour's `Dockerfile`, run these commands (replacing the repository host and tag with your own):

```
docker build -t docker.io/davecheney/contour:latest .
docker push docker.io/davecheney/contour:latest
```

### Verify your change

To verify your change by deploying the image you built, take one of the [deployment manifests][7], edit it to point to your new image, and deploy to your Kubernetes cluster.

## DCO Sign off

All authors to the project retain copyright to their work. However, to ensure
that they are only submitting work that they have rights to, we are requiring
everyone to acknowledge this by signing their work.

Any copyright notices in this repository should specify the authors as "The
project authors".

To sign your work, just add a line like this at the end of your commit message:

```
Signed-off-by: Joe Beda <joe@heptio.com>
```

This can easily be done with the `--signoff` option to `git commit`.

By doing this you state that you can certify the following (from https://developercertificate.org/):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
1 Letterman Drive
Suite D4700
San Francisco, CA, 94129

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.


Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

[1]: https://golang.org/dl/
[2]: https://github.com/golang/dep
[3]: https://golang.org/doc/code.html
[4]: https://golang.org/pkg/testing/
[5]: https://developercertificate.org/
[6]: https://github.com/heptio/contour/issues
[6]: docs/tagging.md
[7]: docs/deploy-options.md
