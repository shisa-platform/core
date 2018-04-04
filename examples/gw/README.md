# API Gateway Example

This is a sample implementation of an API Gateway using Shisa.  It
written to talk to the other example services to show a complete
working mult-service configuration.

## Running

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
