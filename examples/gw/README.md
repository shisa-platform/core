# API Gateway Example

This is a sample implementation of an API Gateway using Shisa.  It
written to talk to the other example services to show a complete
working mult-service configuration.

## Running

This service needs to know the addresses of the other services as
environment variables.  The value should be in the form of: `host:port`
where `host` is optional.  If `host` is provided it must be a hostname
or (v4 or v6) IP address.

    export IDP_SERVICE_ADDR=":9601"
    export GOODBYE_SERVICE_ADDR=":9501"
    export HELLO_SERVICE_ADDR=":9401"

See the [parent README.MD](../README.md) for instructions to run
the example services together.

## Endpoints

The API Gateway implements the following RESTful API:

``` yaml
#%RAML 0.8
title: Example API Gateway
version: 1

/api/greeting:
  get:
    queryParameters:
      language:
        description: Language for greeting
        type: string
        enum: ["en-US", "en-GB", "es-ES", "fi", "fr", "ja", "zh-Hans"]
        default: en-US
      name:
        description: The person to greet (defaults to user name)
    responses:
      200:
        body:
          application/json:
            schema:
              $schema: http://json-schema.org/draft-04/schema#
              type: object
              properties:
                greeting:
                  type: string
              required:
                - greeting
            example:
              greeting: "Hello Boss"
/api/farewell:
  get:
    queryParameters:
      name:
        description: The person to bid farewell (defaults to user name)
    responses:
      200:
        body:
          application/json:
            schema:
              $schema: http://json-schema.org/draft-04/schema#
              type: object
              properties:
                farewell:
                  type: string
              required:
                - farewell
            example:
              farewell: "Goodbye Boss"
```
