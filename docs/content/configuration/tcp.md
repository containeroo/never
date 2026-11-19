# TCP

A TCP target succeeds when it can establish a connection to the configured `host:port` address.

## Flags

| Flag                              | Type     | Default          | Description                                                        |
| --------------------------------- | -------- | ---------------- | ------------------------------------------------------------------ |
| `--tcp.<ID>.name`                 | string   | `<ID>`           | Name shown in logs.                                                |
| `--tcp.<ID>.address`              | string   | required         | Target in `host:port` form. Supports a [resolved value](resolved-values.md). |
| `--tcp.<ID>.interval`             | duration | `0`              | Delay between attempts. Uses `--default-interval` when `0`.        |
| `--tcp.<ID>.max-attempts`         | int      | `--max-attempts` | Per-target attempt limit. `0` means unlimited.                     |
| `--tcp.<ID>.backoff`              | enum     | `linear`         | Retry backoff: `linear` or `exponential`.                          |
| `--tcp.<ID>.max-interval`         | duration | `0`              | Maximum delay when backoff grows. `0` leaves it uncapped.          |
| `--tcp.<ID>.timeout`              | duration | `2s`             | Connection timeout.                                                |

## Example

```sh
never \
  --tcp.database.address=postgres.default.svc.cluster.local:5432 \
  --tcp.database.backoff=exponential \
  --tcp.database.max-interval=30s \
  --tcp.database.timeout=5s
```

## Retry backoff

Set `backoff=exponential` to double the retry delay after each failure, starting with the configured interval. Use `max-interval` to cap the delay. See [retry timing](../behavior.md#retry-timing) for a worked example.
