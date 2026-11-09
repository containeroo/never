# HTTP and HTTPS

An HTTP target sends a request and validates the final response status. HTTPS targets use the same `http` target type; the URL scheme selects TLS.

## Flags

| Flag                                          | Type            | Default          | Description                                                                                                                         |
| --------------------------------------------- | --------------- | ---------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `--http.<ID>.name`                            | string          | `<ID>`           | Name shown in logs.                                                                                                                 |
| `--http.<ID>.address`                         | string          | required         | Target URL. Supports a [resolved value](resolved-values.md).                                                                        |
| `--http.<ID>.address-detail`                  | enum            | `origin`         | URL detail shown in logs: `origin`, `path`, `query`, or `full`.                                                                     |
| `--http.<ID>.interval`                        | duration        | `0`              | Delay between requests. Uses `--default-interval` when `0`.                                                                         |
| `--http.<ID>.max-attempts`                    | int             | `--max-attempts` | Per-target attempt limit. `0` means unlimited.                                                                                       |
| `--http.<ID>.backoff`                         | enum            | `linear`         | Retry backoff: `linear` or `exponential`.                                                                                            |
| `--http.<ID>.max-interval`                    | duration        | `0`              | Maximum delay when backoff grows. `0` leaves it uncapped.                                                                            |
| `--http.<ID>.method`                          | enum            | `GET`            | `GET`, `HEAD`, `POST`, `PUT`, `PATCH`, `DELETE`, `CONNECT`, `OPTIONS`, or `TRACE`.                                                   |
| `--http.<ID>.header`                          | string list     | empty            | Header in `KEY=VALUE` format. Repeat for multiple headers. Values support [resolution](resolved-values.md).                          |
| `--http.<ID>.allow-duplicate-headers`         | bool            | `false`          | Preserve repeated headers with the same name.                                                                                        |
| `--http.<ID>.expected-status-codes`           | int list/ranges | `200`            | Accepted statuses, such as `200,204,301-302`.                                                                                        |
| `--http.<ID>.follow-redirects`                | bool            | `true`           | Follow redirects before validating the response.                                                                                     |
| `--http.<ID>.max-redirects`                   | int             | `10`             | Maximum redirects to follow. `0` disables following.                                                                                 |
| `--http.<ID>.skip-tls-verify`                 | bool            | `false`          | Disable TLS certificate verification.                                                                                                |
| `--http.<ID>.timeout`                         | duration        | `2s`             | Timeout for the complete request, including redirects.                                                                               |

## Example

```sh
never \
  --http.api.address=https://api.example.com/healthz \
  --http.api.method=GET \
  --http.api.expected-status-codes=200,204 \
  --http.api.follow-redirects=true \
  --http.api.max-redirects=5 \
  --http.api.header="Authorization=env:API_TOKEN"
```

## Redirects

When redirects are enabled, the checker validates the response after following at most `max-redirects` hops. Set `follow-redirects=false` or `max-redirects=0` to validate the first redirect response instead. For example, a target expecting `301` should disable redirect following.

## Headers

Repeat `--http.<ID>.header` to send multiple headers. Commas inside a value are preserved:

```sh
never \
  --http.api.address=https://api.example.com/healthz \
  --http.api.header="Accept=text/html, application/json" \
  --http.api.header="Cache-Control=no-cache, no-store"
```

Multiple headers in one environment variable are separated by newlines:

```sh
NEVER__HTTP_API_HEADER=$'Accept=text/html, application/json\nCache-Control=no-cache, no-store'
```

### Secret header values

Keep the header name in configuration and resolve its value from an environment variable or a mounted secret file:

```sh
never \
  --http.api.address=https://api.example.com/healthz \
  --http.api.header="Authorization=env:API_AUTHORIZATION" \
  --http.api.header="X-API-Key=file:/run/secrets/api-key"
```

`API_AUTHORIZATION` must contain the complete header value, including `Bearer ` when required. Resolution applies to the entire value after the first `=`; `Bearer env:API_TOKEN` is a literal value and does not resolve the token.

You can also configure the reference entirely through environment variables:

```sh
export NEVER__HTTP_API_ADDRESS=https://api.example.com/healthz
export NEVER__HTTP_API_HEADER='Authorization=env:API_AUTHORIZATION'
never
```

Values also support `json:`, `yaml:`, and `ini:` references; see [resolved values](resolved-values.md). References are resolved once when configuration is parsed, before checks start. A missing variable, unreadable file, or missing key stops startup. Secret changes require restarting Never.

## Address visibility

The default `origin` detail logs only the URL scheme and host. `path` includes the path, `query` includes its query string, and `full` also includes user credentials and fragments. Use `query` and `full` carefully because they can expose secrets. HTTP request errors do not echo the request URL.

## Proxies

The HTTP checker honors Go's standard `HTTP_PROXY`, `HTTPS_PROXY`, and `NO_PROXY` environment variables.
