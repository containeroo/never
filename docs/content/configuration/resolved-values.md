# Resolved values

Addresses and HTTP header values can be loaded from environment variables or files instead of being supplied literally.

| Prefix  | Source               | Example                                  | Description                                              |
| ------- | -------------------- | ---------------------------------------- | -------------------------------------------------------- |
| `env:`  | Environment variable | `env:API_ADDRESS`                        | Reads the named environment variable.                    |
| `file:` | Key-value file       | `file:/config/app.txt//KeyName`          | Reads a key from a simple key-value file.                 |
| `json:` | JSON file            | `json:/config/app.json//database.host`   | Reads a dot-separated path from JSON.                     |
| `yaml:` | YAML file            | `yaml:/config/app.yaml//server.port`     | Reads a dot-separated path from YAML.                     |
| `ini:`  | INI file             | `ini:/config/app.ini//Section.Key`       | Reads a section and key from INI.                         |
| none    | Literal              | `https://example.com/healthz`            | Uses the supplied value directly.                        |

This resolution mechanism is separate from the `NEVER__...` environment-variable form used to configure flags.

## Examples

Resolve a target address:

```sh
never --http.api.address=env:API_ADDRESS
```

Resolve only a sensitive header value:

```sh
never \
  --http.api.address=https://example.com/healthz \
  --http.api.header="Authorization=env:API_TOKEN"
```

The environment variable in this example must contain the complete authorization value (for example, `Bearer ` followed by the token). Resolution is applied to the whole header value, not to embedded references such as `Bearer env:API_TOKEN`.

For a mounted secret containing just the header value, omit the key selector:

```sh
never \
  --http.api.address=https://example.com/healthz \
  --http.api.header="Authorization=file:/run/secrets/authorization"
```

`file:` without `//KEY` reads the entire file and trims surrounding whitespace, including its trailing newline. With `//KEY`, it reads the named key instead. Header names remain literal.

Resolution happens once at startup. Missing references cause startup to fail; changing the secret requires restarting Never.
