# mittwald mStudio authentication proxy

> [!NOTE]
> This is a very incomplete proof-of-concept implementation; usage in production environments is not recommended, yet. Use at own risk.

This repository contains a small HTTP reverse proxy which can be used to quickly add mStudio authentication to an existing web application.

Features:

- Handles the installation state of mittwald mStudio extension instances
- Support for one-click authentication from within the mittwald mStudio

Pending:

- [ ] Support for mittwald mStudio oAuth authentication
- [ ] Support for other database backends than MongoDB (for better integration into mittwald managed services)

## Running

### Using docker

```
$ docker run -p 8000:8000 \
    -e MITTWALD_EXT_PROXY_MONGODB_URI=mongodb://... \
    -e MITTWALD_EXT_PROXY_SECRET=... \
    -e MITTWALD_EXT_PROXY_STATIC_PASSWORD=supersecret \
    -e MITTWALD_EXT_PROXY_UPSTREAMS='{"/":{"upstreamURL":"http://..."}}' \
    mittwald/mstudio-ext-proxy
```

### Using mittwald container hosting

Write me

### Using Kubernetes

Write me

## Configuration

### Environment variables

The following environment variables can be used to modify this proxy's behaviour:

- `PORT` is the port that the HTTP proxy should listen on. If omitted, this will default to `8000`.
- `MITTWALD_EXT_PROXY_MONGODB_URI` is the URI for a MongoDB connection. Used to store active extension instances and sessions.
- `MITTWALD_EXT_PROXY_SECRET` is the secret used for signing JWTs that are passed to the upstream application. **If omitted, this service will not start**.
- `MITTWALD_EXT_PROXY_STATIC_PASSWORD` defines a static password that can be used to bypass the mStudio authentication by navigating to the `/mstudio/auth/password` endpoint. If this variable is omitted, that endpoint will not be available.
- `MITTWALD_EXT_PROXY_CONTEXT` can be used to enable development mode (by setting it to `dev`). In development, secure cookies are not enforced, and the `/mstudio/auth/fake` endpoint is available.
- `MITTWALD_EXT_PROXY_UPSTREAMS` contains a JSON object with the proxy configuration. See section below for examples.

### Proxy configuration

The proxy configuration (passed in the `MITTWALD_EXT_PROXY_UPSTREAMS` environment variable) is a JSON map that should contain all proxy upstream definitions.

Example:

```json
{
  "/foo": {
    "upstreamURL": "http://foo-service:3000",
    "stripPrefix": "/foo"
  },
  "/": {
    "upstreamURL": "http://bar-service:3030"
  }
}
```

### mStudio marketplace configuration

When registering an extension to the mStudio marketplace using this component, your configuration YAML should look like this:

```yaml
...
externalComponents:
  backend:
    extensionAddedToContext:
      url: https://extension.example/mstudio/webhooks
    extensionInstanceUpdated:
      url: https://extension.example/mstudio/webhooks
    extensionInstanceSecretRotated:
      url: https://extension.example/mstudio/webhooks
    extensionInstanceRemovedFromContext:
      url: https://extension.example/mstudio/webhooks
  frontends:
    index:
      url: https://extension.example/mstudio/auth/oneclick?atrek=:accessTokenRetrievalKey&userId=:userId&instanceID=:extensionInstanceId

```

## Accessing user data in upstream applications

Upstream applications will receive an additional HTTP header `X-Mstudio-User` with an JWT that contains the relevant user information in its claims:

- `sub`: mStudio user ID
- `fname` and `lname`: First and last name
- `email`: email address
- `inst`: information about the extension instance; the subfields `id` identify the extension instance, and `context.id` and `context.kind` the mstudio resource (meaning the organization or project), in which the extension was installed
- `tok`: an mStudio access token, which can be used to access the mStudio API as the accessing user

The JWT is signed with the secret that needs to be specified in `MITTWALD_EXT_PROXY_SECRET`. Your upstream applications need access to this secret in order to verify the JWT for authenticity.
