# Services

- services are plugins
- services handle n endpoints
- an endpoint is a tuple of (route, policy, handler)
- routes are path expressions with named parameters (routes may not
overlap w/in or between services)
- routes may specify legal query parameters and how to handle unknown
parameters
- the policy object may be customized to control automated behavior
- handlers service requests to routes
- handlers may use a built-in backend or do its own thing

That is all
