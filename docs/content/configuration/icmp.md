# ICMP

An ICMP target sends echo requests to a hostname or IP address. It is useful when reachability matters independently of an application port.

## Flags

| Flag                               | Type     | Default          | Description                                                               |
| ---------------------------------- | -------- | ---------------- | ------------------------------------------------------------------------- |
| `--icmp.<ID>.name`                 | string   | `<ID>`           | Name shown in logs.                                                       |
| `--icmp.<ID>.address`              | string   | required         | Hostname or IP address without a scheme, path, or port. Supports a [resolved value](resolved-values.md). |
| `--icmp.<ID>.interval`             | duration | `0`              | Delay between attempts. Uses `--default-interval` when `0`.               |
| `--icmp.<ID>.max-attempts`         | int      | `--max-attempts` | Per-target attempt limit. `0` means unlimited.                            |
| `--icmp.<ID>.backoff`              | enum     | `linear`         | Retry backoff: `linear` or `exponential`.                                 |
| `--icmp.<ID>.max-interval`         | duration | `0`              | Maximum delay when backoff grows. `0` leaves it uncapped.                 |
| `--icmp.<ID>.timeout`              | duration | `2s`             | Timeout for ICMP reads and writes.                                        |
| `--icmp.<ID>.read-timeout`         | duration | `0`              | Advanced read override. Uses the target timeout when `0`.                 |
| `--icmp.<ID>.write-timeout`        | duration | `0`              | Advanced write override. Uses the target timeout when `0`.                |

## Example

```sh
never \
  --icmp.gateway.address=10.0.0.1 \
  --icmp.gateway.timeout=3s
```

ICMP requires raw-socket privileges. Containers need the `NET_RAW` capability; HTTP and TCP checks need no extra capability. See [Kubernetes](../kubernetes.md) for a complete example.

## Retry backoff

Set `backoff=exponential` to double the retry delay after each failure, starting with the configured interval. Use `max-interval` to cap the delay. See [retry timing](../behavior.md#retry-timing) for a worked example.
