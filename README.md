[![Circle CI](https://circleci.com/gh/percolate/api-gw.svg?style=svg&circle-token=)](https://circleci.com/gh/percolate/api-gw)
[![codecov.io](https://codecov.io/github/percolate/api-gw/coverage.svg?token=&branch=master)](https://codecov.io/github/percolate/api-gw?branch=master)

## Overview

Zuul is Percolate's API Gateway Service.

## Architecture

![architecture](doc/diagram/architecture.png)

## Contributing

Zuul is automatically installed and running on devolate, to make changes locally you must [replace the managed repository](https://github.com/percolate/devolate#other-repos).  Clone a this repository to your `devolate` directory (next to hotlanta, et al.) and restart the VM.

All of the important build commands are in the [`Makefile`](Makefile), please use those recipes.

## Adding Dependencies

    git remote add -f <name-of-remote> <repo-url>
    git subtree add --squash --prefix=<path-to-vendor-repo> <name-of-remote> master

For example:

    $ git remote add -f raven-go https://github.com/getsentry/raven-go.git
    $ git subtree add --squash --prefix=vendor/src/github.com/getsentry/raven-go raven-go master
