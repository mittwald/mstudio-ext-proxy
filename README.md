# mittwald mStudio authentication proxy

This repository contains a small HTTP reverse proxy which can be used to quickly add mStudio authentication to an existing web application.

Features:

- Handles the installation state of mittwald mStudio extension instances
- Support for one-click authentication from within the mittwald mStudio

Pending:

- [ ] Support for mittwald mStudio oAuth authentication

## Running

### Using docker

```
$ docker run -p 8000:8000 \
    -e MONGODB_URI=mongodb://... \
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
- `MONGODB_URI` is the URI for a MongoDB connection. Used to store active extension instances and sessions.
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
