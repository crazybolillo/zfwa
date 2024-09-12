# Zitadel Forward Auth
This is a simple HTTP service meant to be used with API Gateways like Traefik to perform Forward Authentication
using [Zitadel](https://zitadel.com/) as a token provider. For more information on how tokens are verified see
[Zitadel's documentation](https://zitadel.com/docs/guides/solution-scenarios/frontend-calling-backend-API).

## Usage
Available as a binary or docker container, it relies solely on environment variables for configuration:

### `ZITADEL_HOST`
This is the host that will be used to perform
[token introspection](https://zitadel.com/docs/guides/integrate/token-introspection) requests, for example: 
`https://accounts.mydomain.com`. If running Zitadel locally it would be something like `http://localhost:8080`.

### `CLIENT_ID`, `CLIENT_SECRET`
Basic authentication credentials to authorize to Zitadel and perform token introspection. At the moment
JWT is not supported.

### `VERIFY_TENANT`
This is optional and enabled by default. The token's metadata will be verified to ensure there is a `metadata` key that matches
the request's `X-Tenant-Id` HTTP header value. If the header is missing, the request will be rejected.

## Responses
If verification is successful, HTTP 204 OK will be returned. In all other cases HTTP 401 is returned. Successful
responses include the `X-Auth-User` HTTP header, specifying the username associated with the token.
