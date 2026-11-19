# Runtime behavior

## Concurrent checks

N.E.V.E.R. starts all configured targets concurrently. A successful target stops checking while unsuccessful targets continue retrying. The process exits with status `0` only after every target succeeds.

## Attempts and failure

The global `--max-attempts` setting applies to targets without an override. A per-target `max-attempts` takes precedence. The value `0` means unlimited attempts.

If any target exhausts its attempts, N.E.V.E.R. cancels the remaining checks and exits with status `1`.

## Retry timing

Each target uses its own interval. An interval of `0` inherits `--default-interval`, which defaults to two seconds.

The default `backoff=linear` keeps the delay constant. Set `backoff=exponential` on an HTTP, TCP, or ICMP target to increase the delay between retries:

```sh
never \
  --tcp.database.address=database.example.com:5432 \
  --tcp.database.interval=2s \
  --tcp.database.backoff=exponential \
  --tcp.database.max-interval=30s
```

The first check starts immediately. After the first failed attempt, N.E.V.E.R. waits for the base interval; each subsequent failure doubles that delay. In this example, the waits are:

| Failed attempt | Wait before the next attempt |
| -------------- | ---------------------------- |
| 1              | 2 seconds                    |
| 2              | 4 seconds                    |
| 3              | 8 seconds                    |
| 4              | 16 seconds                   |
| 5 and later    | 30 seconds                   |

The delay after failed attempt `n` is `interval × 2^(n − 1)`, capped by `max-interval` when it is greater than zero. A `max-interval` of `0` (the default) leaves growth uncapped. The cap also applies to the first retry if it is smaller than the base interval. No random jitter is added.

These delays start after each failed check finishes, so the time spent performing the check is additional. Each target calculates its own delay independently. A successful target stops checking; a target that reaches its `max-attempts` limit fails without waiting for another retry.

## HTTP responses

HTTP checks validate the configured expected status codes. Redirects are followed by default and the final response is validated. Redirect following and the maximum number of hops are independently configurable. The request timeout covers the full operation, including redirects.

## Cancellation

Process cancellation stops active network operations through their request contexts. This includes HTTP requests, TCP connection attempts, ICMP operations, and retry waits.
