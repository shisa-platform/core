# RESTful Web Service

This is a sample of a "legacy" RESTful web service that implements the
Goodbye service.  The IdP service is required to be running and
accessible.

## Running

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
