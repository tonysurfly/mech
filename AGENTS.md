# AGENTS.md

> See [AGENTS.universal.md](./AGENTS.universal.md) and [AGENTS.go.md](./AGENTS.go.md) for universal conventions.
> Refresh: `make standards`

---

## Overview

`mech` automates Constellix DNS configuration as code, similar to
[octodns](https://github.com/octodns/octodns) and
[terraform](https://www.terraform.io/). It manages DNS records, Sonar health
checks, and GeoProximity locations by reconciling a local YAML configuration
tree against the live state of the Constellix REST API.

---

## Architecture

```
main.go                                 Entry point, calls cmd.Execute()
cmd/
  cmd_root.go                           Root cobra command, --debug/--trace/--version flags, credential check, Execute()
  logger.go                             slog setup for --debug/--trace (LevelTrace, initLogger)
  cmd_dns.go                            `dns discover`/`dns sync` commands
  cmd_sonar.go                          `sonar discover`/`sonar sync` commands
  cmd_geoproximity.go                   `geoproximity discover`/`geoproximity sync` commands
  config.go                             Loads/parses the local YAML config tree into a Config struct
  constellix.go                         API constants, HMAC request signing, DNS v4 response envelope types
  utils.go                              HTTP client helpers (rate-limit retry, pagination), reflection helpers, ANSI colors
  compare.go                            Compare(): diffs an expected (YAML) resource against an active (API) resource
  sync.go                               Sync(): builds delete/update/create sets, renders the report, applies changes
  dns_domain.go                         DNSDomain model + GetDNSDomains()
  dns_domain_record.go                  DNSRecord/ExpectedDNSRecord models + CRUD against the v4 API
  dns_domain_record_value.go            Polymorphic `value` field parsing per record type/mode
  dns_domain_record_geoproximity.go     `geoproximity` field glue (API object <-> int/@name shorthand)
  dns_domain_record_ipfilter.go         `ipfilter` field glue (API object <-> int)
  geoproximities.go                     GeoProximity/ExpectedGeoProximity models + GetGeoProximities (mockable var)
  sonar_http_check.go                   SonarHTTPCheck/ExpectedSonarHTTPCheck models + cached GetSonarHTTPChecks(), GetSonarHTTPCheckStatus()
  sonar_tcp_check.go                    SonarTCPCheck/ExpectedSonarTCPCheck models + GetSonarTCPChecks()
  cache.go                              Package-level Sonar HTTP check cache
```

---

## Key Flows

1. **`mech dns|sonar|geoproximity sync --config <file> [--doit] [--remove]`**
   `getConfig()` parses the local YAML tree -> the matching `GetX()` fetches
   active state from Constellix -> both collections go through
   `toResourceMatcher()` -> `Sync()` diffs them and renders a report. Without
   `--doit` this only plans; with it, `syncChanges()` applies the plan via
   each resource's `SyncResourceCreate`/`SyncResourceUpdate`/`SyncResourceDelete`.
   Deletions require `--remove`.
2. **`mech sonar discover static|runtime -t http|tcp`**
   `static` calls the matching `GetX()` and dumps the result as YAML via
   `writeDiscoveryResult()`. `runtime` (http only) spawns one goroutine per
   check calling `GetSonarHTTPCheckStatus()` and collects results into a
   report table.
3. **Config parsing** (`config.go`)
   The top-level YAML lists sub-config file paths/globs per resource type.
   `readConfigs()` resolves each path (literal file or glob, relative to the
   top-level config's directory). Each resource's custom `UnmarshalYAML`
   records which YAML keys were actually set (`definedFieldsMap`), so
   `Compare()` and `generatePayload()` only touch user-specified fields.
4. **HTTP layer** (`utils.go`, `constellix.go`)
   `makeSimpleAPIRequest()` signs every request with `buildSecurityToken()`
   (HMAC-SHA1) and retries indefinitely on HTTP 429 using the
   `X-Ratelimit-Reset` header (or a 5s fallback). `makev4APIRequest()` wraps
   it for the paginated v4 DNS API, including a workaround for a known
   Constellix pagination bug (see the comment above it).

---

## Build & Run

```bash
make check      # fmt, vet, build, test, lint - run after every change
make build       # linux/darwin amd64/arm64 binaries in bin/
CONSTELLIX_API_KEY=... CONSTELLIX_SECRET_KEY=... ./mech dns sync --config mech.yaml
./mech --help          # works without credentials
./mech --version
```

`go test ./...` does not require `CONSTELLIX_API_KEY`/`CONSTELLIX_SECRET_KEY`
to be set - credential validation happens once, in the root command's
`PersistentPreRunE`, not at package init time.

---

## Configuration

- The top-level YAML lists sub-config file paths/globs per resource type:
  `constellix.sonar.http_checks`, `constellix.sonar.tcp_checks`,
  `constellix.dns.<domain>`, `constellix.geoproximity`. See README.md for the
  full shape.
- `mech sonar discover static -t http` (or `tcp`, or `mech dns discover
  records <domain>`, `mech geoproximity discover`) prints the existing remote
  configuration in the same YAML shape, which is a convenient starting point
  for a local config file.
- Resource references support `@sonar,http:<name>` and `@geoproximity:<name>`
  shorthand in place of a literal Constellix ID; these are resolved against
  the live API while the config is being parsed.

---

## Design Decisions

- `--debug`/`--trace` route only diagnostic logging (`log/slog`, via `L`).
  The program's actual output - report tables, discovered YAML, sync
  summaries - goes through `logger`/stdout unconditionally, since that is the
  tool's primary output, not a diagnostic log.
- Credential validation happens in `rootCmd.PersistentPreRunE`, not in an
  `init()` function, so `--help`, `--version`, and direct calls into the
  `cmd` package (tests, library use) don't require credentials.
- `GetGeoProximities` is a package-level func var (not a plain func)
  specifically so tests can substitute it. The other `GetX()` fetchers are
  tested by spinning up an `httptest.Server` and overriding the matching
  `*RESTAPIBaseURL` package var instead.
- `generatePayload()` sends only the fields the user actually defined in
  YAML (`definedFieldsMap`), because the Constellix API is inconsistent about
  which immutable fields it accepts or rejects in a request body.

---

## Gotchas

- `sonarCheckId`, `geoproximity`, and `ipfilter` fields in YAML can be either
  a literal Constellix ID or an `@sonar,http:<name>` / `@geoproximity:<name>`
  shorthand resolved by calling the live API during config parsing - so even
  a dry-run `sync` (no `--doit`) needs network access and valid credentials
  once the config references any of these by name.
- `sonar discover runtime -t http` spawns one goroutine per check, each
  appending to a shared `go-pretty` `table.Writer`. `table.Writer` isn't
  documented as goroutine-safe; watch for `-race` reports if more concurrent
  writers are added to this path.
- `--trace` writes full request/response bodies (`/tmp/mech.log`, truncated
  each run) - treat that file as sensitive if the DNS/Sonar payloads it
  captures could be.

---

## Known Issues

- `mech sonar discover runtime` only supports `-t http`. There is no runtime
  status endpoint for TCP checks in the Constellix API.
