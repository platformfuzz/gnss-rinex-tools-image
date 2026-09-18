# rinex-tools

Multi-command CLI for GNSS RINEX helpers. This is the image entrypoint
(packaged with [ko](https://ko.build/) from this directory).

Repo packaging, CI, and GHCR details live in the [root README](../../README.md).

## Commands

| Subcommand | Package | Purpose |
| --- | --- | --- |
| `check` | [`internal/check`](../../internal/check) | Custom quick RINEX 3 OBS sanity check (header, epochs, systems) |

`check` is a custom lightweight QC implemented in this repo. It is not a wrapper
around third-party RINEX processors.

## Usage

Local:

```bash
go run ./cmd/rinex-tools check fixtures/minimal_obs_3.05.rnx
go run ./cmd/rinex-tools --help
```

From the published image (binary is the entrypoint; pass the subcommand):

```bash
docker run --rm -v "$PWD:/data" \
  ghcr.io/platformfuzz/gnss-rinex-tools-image:latest \
  check /data/your_file.rnx
```

### `check` results

- `RESULT: OK` (exit 0): size, epoch count, and GPS observations meet thresholds
- `RESULT: WEAK/CHECK` (exit 1): parsed but below thresholds (common for short samples)
- Fail / missing file: exit 1 or 2

## Adding a tool

Keep one `cmd/rinex-tools` binary. Document each new subcommand in **this** README
(not the root README).

1. Add `internal/<tool>/` with tests.
2. Register a cobra subcommand in [`internal/cli`](../../internal/cli).
3. Add a row to the Commands table above and a short usage section if needed.
