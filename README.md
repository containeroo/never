# N.E.V.E.R. - Network Endpoint Validation with Endless Retries

[![Go Report Card](https://goreportcard.com/badge/github.com/containeroo/never?style=flat-square)](https://goreportcard.com/report/github.com/containeroo/never)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/containeroo/never)
[![Release](https://img.shields.io/github/release/containeroo/never.svg?style=flat-square)](https://github.com/containeroo/never/releases/latest)
[![GitHub tag](https://img.shields.io/github/tag/containeroo/never.svg?style=flat-square)](https://github.com/containeroo/never/releases/latest)
![Tests](https://github.com/containeroo/never/actions/workflows/tests.yml/badge.svg)
[![Build](https://github.com/containeroo/never/actions/workflows/release.yml/badge.svg)](https://github.com/containeroo/never/actions/workflows/release.yml)
[![license](https://img.shields.io/github/license/containeroo/never.svg?style=flat-square)](LICENSE)

---

> **N.E.V.E.R.** (Network Endpoint Validation with Endless Retries) is a lightweight Go application that obsessively checks whether a `TCP`, `HTTP`, or `ICMP` target is reachable.
> It loops endlessly until the target responds — or until it’s killed.

Designed to run as a **Kubernetes `initContainer`**, `N.E.V.E.R.` ensures your service dependencies are fully up before anything else gets a chance to boot.

## Features

- Continuously retries until the target responds — no backoff, no shame.
- Supports multiple concurrent targets, each with its own config.
- Configurable entirely via command-line arguments.
- Exits with `0` the moment everything is ready.
- Exits with `1` if any target exceeds `--max-attempts`.

Whether you're waiting on a `port`, `ping`, or a `200 OK`, `N.E.V.E.R.` backs down never.

## Command-Line Flags

`never` accepts the following command-line flags:

### Common Flags

| Flag                 | Type     | Default | Description                                                         |
| -------------------- | -------- | ------- | ------------------------------------------------------------------- |
| `--default-interval` | duration | `2s`    | Default interval between checks. Can be overridden for each target. |
| `--max-attempts`     | int      | `-1`    | Maximum attempts before giving up. Use `-1` to retry endlessly.     |
| `--version`          | bool     | `false` | Show version and exit.                                              |
| `--help`, `-h`       | bool     | `false` | Show help.                                                          |

### Target Flags

`never` accepts "dynamic" flags that can be defined in the startup arguments.
Use the `--<TYPE>.<IDENTIFIER>.<PROPERTY>=<VALUE>` format to define targets.
Types are: `http`, `icmp` or `tcp`.

#### HTTP-Flags

- **`--http.<IDENTIFIER>.name`** = `string`
  The name of the target. If not specified, it uses the `<IDENTIFIER>` as the name.

- **`--http.<IDENTIFIER>.address`** = `string`
  The target's address.
  **Resolvable:** See [Resolving Variables](#resolving-variables) below.

- **`--http.<IDENTIFIER>.interval`** = `duration`
  The interval between HTTP requests (e.g., `1s`). Overwrites the global `--default-interval`.

- **`--http.<IDENTIFIER>.max-attempts`** = `int`
  Maximum attempts before giving up. Defaults to `--max-attempts` when unset or `0`.

- **`--http.<IDENTIFIER>.method`** = `string`
  The HTTP method to use (e.g., `GET`, `POST`). Defaults to `GET`.

- **`--http.<IDENTIFIER>.header`** = `string`
  A HTTP header in `key=value` format. Can be specified multiple times.
  **Example:** `Authorization=Bearer token`
  **Resolvable:** See [Resolving Variables](#resolving-variables) below.

- **`--http.<IDENTIFIER>.allow-duplicate-headers`** = `bool`
  Allow duplicate headers. Defaults to `false`.

- **`--http.<IDENTIFIER>.expected-status-codes`** = `string`
  A comma-separated list of expected HTTP status codes or ranges (e.g., `200,301-302`). Defaults to `200`.

- **`--http.<IDENTIFIER>.skip-tls-verify`** = `bool`
  Whether to skip TLS verification. Defaults to `false`.

- **`--http.<IDENTIFIER>.timeout`** = `duration`
  The timeout for the HTTP request (e.g., `5s`). Defaults to `1s`.

#### ICMP Flags

- **`--icmp.<IDENTIFIER>.name`** = `string`
  The name of the target. If not specified, it uses the `<IDENTIFIER>` as the name.

- **`--icmp.<IDENTIFIER>.address`** = `string`
  The target's address.
  **Resolvable:** See [Resolving Variables](#resolving-variables) below.

- **`--icmp.<IDENTIFIER>.interval`** = `duration`
  The interval between ICMP requests (e.g., `1s`). Overwrites the global `--default-interval`.

- **`--icmp.<IDENTIFIER>.max-attempts`** = `int`
  Maximum attempts before giving up. Defaults to `--max-attempts` when unset or `0`.

- **`--icmp.<IDENTIFIER>.read-timeout`** = `duration`
  The read timeout for the ICMP connection (e.g., `1s`). Defaults to `1s`.

- **`--icmp.<IDENTIFIER>.write-timeout`** = `duration`
  The write timeout for the ICMP connection (e.g., `1s`).Defaults to `1s`.

#### TCP Flags

- **`--tcp.<IDENTIFIER>.name`** = `string`
  The name of the target. If not specified, it uses the `<IDENTIFIER>` as the name.

- **`--tcp.<IDENTIFIER>.address`** = `string`
  The target's address.
  **Resolvable:** See [Resolving Variables](#resolving-variables) below.

- **`--tcp.<IDENTIFIER>.interval`** = `duration`
  The interval between TCP requests (e.g., `1s`). Overwrites the global `--default-interval`.

- **`--tcp.<IDENTIFIER>.max-attempts`** = `int`
  Maximum attempts before giving up. Defaults to `--max-attempts` when unset or `0`.

#### Resolving variables

Each `address` field can be resolved using `environment variables`, `files`, `JSON`, `YAML`, and `INI` files.

- `env`: – Resolves environment variables.
  Example: `env:PATH` returns the value of the `PATH` environment variable.
- `file`: – Resolves values from a simple key-value file.
  Example: `file:/config/app.txt//KeyName` returns the value associated with `KeyName` in `app.txt`.
- `json`: – Resolves values from a JSON file. Supports dot notation for nested keys.
  Example: `json:/config/app.json//database.host` returns `host` field under `database` in `app.json`. It is also possible to indexing into arrays (e.g., `json:/config/app.json//servers.0.host`).
- `yaml`: – Resolves values from a YAML file. Supports dot notation for nested keys.
  Example: `yaml:/config/app.yaml//server.port` returns `port` under `server` in `app.yaml`.It is also possible to indexing into arrays (e.g., `yaml:/config/app.yaml//servers.0.host`).
- `ini`: – Resolves values from an INI file. Can specify a section and key, or just a key in the default section.
  Example: `ini:/config/app.ini//Section.Key` returns the value of `Key` under `Section`.
- No prefix – Returns the value as-is, unchanged.

HTTP headers values can also be resolved using the same mechanism, (from a environment variable `--http.<IDENTIFIER>.header="header=env:SECRET_HEADER"` or from a file `--http.<IDENTIFIER>.header="header=file:PATH_TO_FILE"`).

### Examples

#### Define an HTTP Target

```sh
never \
  --http.web.address=http://example.com:80 \
  --http.web.method=GET \
  --http.web.expected-status-codes=200,204 \
  --http.web.header="Authorization=Bearer token" \
  --http.web.header="Content-Type=application/json" \
  --http.web.skip-tls-verify=false \
  --default-interval=5s
```

#### Define Multiple Targets (HTTP and TCP) Running in Parallel

```sh
never \
  --http.web.address=http://example.com:80 \
  --tcp.db.address=localhost:5432 \
  --default-interval=10s
```

#### Notes

**Proxy Settings**: Proxy configurations (`HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY`) are managed via environment variables.

## Behavior Flowchart

### TCP Check

<details>
  <summary>Click here to see the flowchart</summary>

```mermaid
graph TD;
    classDef noFill fill:none;
    classDef violet stroke:#9775fa;
    classDef green stroke:#2f9e44;
    classDef error stroke:#fa5252;
    classDef transparent stroke:none,font-size:20px;

    subgraph MainFlow[ ]
        direction TB
        start((Start)) --> attemptConnect[Attempt to connect to <font color=orange>TARGET_ADDRESS</font>];
        class start violet;

        subgraph RetryLoop[Retry Loop]
            subgraph InnerLoop[ ]
                direction TB
                attemptConnect -->|Connection successful| targetReady[Target is ready];
                attemptConnect -->|Connection failed| waitRetry[Wait for retry <font color=orange>CHECK_INTERVAL</font>];
                waitRetry --> attemptConnect;
            end
        end

        targetReady --> processEnd((End));
        class processEnd violet;
        waitRetry --> processEnd;
    end

    programTerminated[Program terminated or canceled] --> processEnd;
    class programTerminated error;

    class start,attemptConnect,targetReady,waitRetry,processEnd,programTerminated,MainFlow,RetryLoop noFill;
    class MainFlow,RetryLoop transparent;
```

</details>

## Permissions

**Only** when using `ICMP` checks in Kubernetes, it's important to ensure that the container has the necessary permissions to send ICMP packets. It is necessary to add the `CAP_NET_RAW` capability to the container's security context.

Example:

```yaml
- name: wait-for-host
  image: ghcr.io/containeroo/never:latest
  env:
    - name: TARGET_ADDRESS
      value: icmp://hostname.domain.com
  securityContext:
    readOnlyRootFilesystem: true
    allowPrivilegeEscalation: false
    capabilities:
      add: ["CAP_NET_RAW"]
```

For `TCP` and `HTTP` checks, the container does not require any additional permissions.

### HTTP Check

<details>
  <summary>Click here to see the flowchart</summary>

```mermaid
flowchart TD;
    direction TB
    classDef noFill fill:none;
    classDef violet stroke:#9775fa;
    classDef green stroke:#2f9e44;
    classDef error stroke:#fa5252;
    classDef decision stroke:#1971c2;
    classDef transparent stroke:none,font-size:20px;

    subgraph MainFlow[ ]
        direction TB
        processStart((Start)) --> createRequest[Create HTTP request for <font color=orange>TARGET_ADDRESS</font>];
        class start processStart;

        createRequest --> addHeaders[Add headers from <font color=orange>HTTP_HEADERS</font>];
        addHeaders --> addSkipTLS[Add skip TLS verify if <font color=orange>HTTP_SKIP_TLS_VERIFY</font> is set];
        addSkipTLS --> sendRequest[Send HTTP request];

        subgraph RetryLoop[Retry Loop]
            subgraph InnerLoop[ ]
                direction TB
                sendRequest --> checkTimeout{Answers within <font color=orange>DIAL_TIMEOUT</font>?};
                class checkTimeout decision;
                checkTimeout -->|Yes| checkStatusCode[Check response status code <font color=orange>HTTP_EXPECTED_STATUS_CODES</font>];
                checkStatusCode --> statusMatch{Matches?};
                class statusMatch decision;

                statusMatch -->|Yes| targetReady[Target is ready];
                class targetReady success;

                statusMatch -->|No| targetNotReady[Target is not ready];
                class targetNotReady error;
                targetNotReady --> waitRetry[Wait for retry <font color=orange>CHECK_INTERVAL</font>];
                waitRetry --> sendRequest;
            end
        end

    targetReady --> processEnd((End));
    class processEnd violet;
    end

    programTerminated[Program terminated or canceled] --> processEnd;
    class programTerminated error;

class processStart,createRequest,addHeaders,addSkipTLS,sendRequest,checkTimeout,checkStatusCode,statusMatch,targetReady,targetNotReady,waitRetry,programTerminated,processEnd,MainFlow,RetryLoop noFill;
class MainFlow,RetryLoop transparent;
```

</details>

### ICMP Check

<details>
  <summary>Click here to see the flowchart</summary>

```mermaid
flowchart TD;
    direction TB
    classDef noFill fill:none;
    classDef violet stroke:#9775fa;
    classDef green stroke:#2f9e44;
    classDef error stroke:#fa5252;
    classDef decision stroke:#1971c2;
    classDef transparent stroke:none,font-size:20px;

    subgraph MainFlow[ ]
        direction TB
        processStart((Start)) --> createRequest[Create ICMP request for <font color=orange>TARGET_ADDRESS</font>];
        class start processStart;

        createRequest --> sendRequest[Send ICMP request];

        subgraph RetryLoop[Retry Loop]
            subgraph InnerLoop[ ]
                direction TB
                sendRequest --> checkTimeout{Answers within timeouts <font color=orange>DIAL_TIMEOUT</font>/<font color=orange>ICMP_READ_TIMEOUT</font>?};

                checkTimeout -->|Connection successful| targetReady[Target is ready];
                checkTimeout -->|Connection failed| waitRetry[Wait for retry <font color=orange>CHECK_INTERVAL</font>];
                waitRetry --> sendRequest;

            end
        end

    targetReady --> processEnd((End));
    class processEnd violet;
    end

    programTerminated[Program terminated or canceled] --> processEnd;
    class programTerminated error;

class processStart,createRequest,sendRequest,checkTimeout,targetReady,waitRetry,processEnd,MainFlow,RetryLoop noFill;
class MainFlow,RetryLoop transparent;
```

</details>

## Kubernetes initContainer Configuration

Configure your Kubernetes deployment to use this init container:

```yaml
initContainers:
  - name: wait-for-vm
    image: ghcr.io/containeroo/never:latest
    args:
      - --icmp.vm.address=hostname.domain.tld
    securityContext: # icmp requires CAP_NET_RAW
      readOnlyRootFilesystem: true
      allowPrivilegeEscalation: false
      capabilities:
        add: ["CAP_NET_RAW"]
  - name: wait-for-it
    image: ghcr.io/containeroo/never:latest
    args:
      - --target.postgres.address=postgres.default.svc.cluster.local:9000/healthz # use healthz endpoint to check if postgres is ready
      - --target.postgres.method=POST
      - --target.postgres.header=Authorization=env:BEARER_TOKEN
      - --target.postgres.expected-status-codes=200,202
      - --target.redis.name=redis
      - --target.redis.address=redis.default.svc.cluster.local:6437
      - --tcp.vaultkey.address=valkey.default.svc.cluster.local:6379
      - --tcp.vaultkey.interval=5s
      - --tcp.vaultkey.timeout=5s
    envFrom:
      - secretRef:
          name: bearer-token
```

## License

This project is licensed under the Apache License. See the [LICENSE](LICENSE) file for details.
