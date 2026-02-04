# Google Cloud Spot Price History

Build a SQLite database of historical GCE spot (and on-demand) prices by replaying the [google-cloud-pricing-cost-calculator](https://github.com/Cyclenerd/google-cloud-pricing-cost-calculator) repository’s `pricing.yml` history. Query by region and machine type to inspect how prices changed over time.

## Overview

The script will clone the [google-cloud-pricing-cost-calculator](https://github.com/Cyclenerd/google-cloud-pricing-cost-calculator) repo and extract every version of `pricing.yml`. The tool will parse and ingest the changes to SQLite database, which you can query directly or via simple API
## Prerequisites

- **Go 1.21+** (project uses Go 1.24; 1.21+ should work for building and running)
- **Git** (for cloning the pricing repo and for `extract_git_history.sh`)
- Optional: **gofmt**, **golint** in PATH for `make fmt` and `make lint`

## Quick start

Build both binaries (dataprocessing and API):

```bash
make build
```

Run the full pipeline: build, clone the pricing repo, extract history, and import into SQLite at `/tmp/history.sqlite3`:

```bash
make run
```

Start the web API (after you have a database):

```bash
./bin/api -dbpath /tmp/history.sqlite3
```

Then open http://localhost:8080 (or the port shown in the logs) to use the web UI.

## How it works

- **dataprocessing** reads a directory of YAML files (one per revision of `pricing.yml`), parses GCE pricing, and inserts rows into SQLite. Each row is a (machine_type, region, on-demand price, spot price, timestamp).
- **API** serves the same data over HTTP and renders simple HTML pages for regions, machine types, and price history.
- **extract_git_history.sh** is used by `make collect-pricing-data`: it must be run from inside the cloned pricing repo and takes a single path (e.g. `pricing.yml`). It writes one file per git revision under `/tmp/pricing-data` with names like `2023-05-08.064247.abc1234`.

## Usage examples

### Build and run the pipeline

```bash
make build
make collect-pricing-data
./bin/dataprocessing -data /tmp/pricing-data -dbpath ./history.sqlite3
```

Or in one step (writes DB to `/tmp/history.sqlite3`):

```bash
make run
```

### Custom data and database paths

```bash
./bin/dataprocessing -data /path/to/yaml-files -dbpath /path/to/history.sqlite3 -batch 5000
```

### Run the API with your database

```bash
./bin/api -dbpath ./history.sqlite3
```

Default is `dbpath=db.sqlite3` if you omit the flag.

### Query price history with SQL

```bash
sqlite3 /tmp/history.sqlite3
```

```sql
-- Spot and on-demand price history for one machine type in one region
SELECT * FROM pricing_history
WHERE region_name = 'europe-west1' AND machine_type = 't2d-standard-4'
ORDER BY updated ASC;
```

Example result:

```
9117|t2d-standard-4|europe-west1|0.185892|0.049464|1683528167|2023-05-08 06:42:47+00:00
13666|t2d-standard-4|europe-west1|0.185892|0.049464|1683634028|2023-05-09 12:08:08+00:00
...
1216842|t2d-standard-4|europe-west1|0.185892|0.041112|1747495189|2025-05-17 15:19:49+00:00
1229344|t2d-standard-4|europe-west1|0.185892|0.041112|1747886484|2025-05-22 04:01:24+00:00
```

## Development

- **Validate and build:** `make` or `make all` (runs test, vet, fmt, lint, then build)
- **Run tests:** `make test`
- **Format code:** `gofmt -w .` (CI checks with `make fmt`)
- **List Make targets:** `make help`

Binaries are produced in `bin/`: `bin/dataprocessing` and `bin/api`.

## Possible improvements

- Better Web UI to visualize price changes over time
- Automated SQLite build in GitHub Actions when the pricing repo changes

## Acknowledgements

- [Cyclenerd](https://github.com/Cyclenerd) for the [google-cloud-pricing-cost-calculator](https://github.com/Cyclenerd/google-cloud-pricing-cost-calculator) project and the maintained `pricing.yml` history.
