# RESTful Web Service

This is a sample of a "legacy" RESTful web service that implements the
Goodbye service.  The IdP service is required to be running and
accessible.

## Running

This services needs to know the address of the IdP service as an
environment variable.  The value should be in the form of: `host:port`
where `host` is optional.  If `host` is provided it must be a hostname
or (v4 or v6) IP address.

    export IDP_SERVICE_ADDR=":9601"

See the [parent README.MD](../README.md) for instructions to run
the example services together.

## Endpoints

This example implements the following API:

``` yaml
#%RAML 0.8
title: Goodbye Service
version: 1

/goodbye:
  get:
    queryParameters:
      name:
        description: The person departing (defaults to user name)
    responses:
      200:
        body:
          application/json:
            schema:
              $schema: http://json-schema.org/draft-04/schema#
              type: object
              properties:
                goodbye:
                  type: string
              required:
                - goodbye
            example:
              goodbye: "Boss"
```
