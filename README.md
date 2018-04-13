# Shisa

[![Circle CI](https://circleci.com/gh/percolate/shisa.svg?style=svg)](https://circleci.com/gh/percolate/shisa)
[![codecov](https://codecov.io/gh/percolate/shisa/branch/master/graph/badge.svg?token=SwfoLAaaS2)](https://codecov.io/gh/percolate/shisa)

## Overview

Percolate's Enterprise API Gateway Platform.

This project provides a platform for building API Gateway services
within your organization.  We understand that each enterprise has
unique challenges to consider when introducing as an API Gateway and
one size does not fit all.  To accommodate that this project allows
you to tailor the functionality to suit your environment.  A minimally
customized service can created quickly and rich options for advanced
configuration offer adaptability for most environments.

The project has these high-level goals:

- Configurability
- Robustness
- Speed

## Architecture

![architecture](doc/diagram/architecture.png)

### Service Discovery

Shisa has built-in support for service discovery and load balancing,
and provides an implementation using [Consul](https://www.consul.io/).

### Distributed Tracing

Shisa has built-in support for [OpenTracing](http://opentracing.io/)
systems such as [Jaeger](https://www.jaegertracing.io/).  By default
tracing is disabled so to have Shisa emit spans a compliant library
must be initialized and set as the global tracer.  Please refer to the
documentation of any compliant implementation for instructions.

## Contributing

To propose a change please open a pull request.  To report a problem
please open an issue.

All of the important build commands are in the [Makefile](Makefile),
please use those recipes.
