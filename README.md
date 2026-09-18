# gnss-rinex-tools-image

Container image for GNSS RINEX QC and helper tools.

Packaged with [ko](https://ko.build/) (no Dockerfile), same pattern as
[waf-log-worker-image](https://github.com/platformfuzz/waf-log-worker-image).

**CLI docs (commands, usage, how to add tools):**
[`cmd/rinex-tools/README.md`](cmd/rinex-tools/README.md)

BNC stream production stays in `bkg-ntrip-client-image`. TEC processing stays in
`tec-processor-image`.

## Quick start

```bash
go test ./...
go run ./cmd/rinex-tools check fixtures/minimal_obs_3.05.rnx

ko build ./cmd/rinex-tools --local

docker pull ghcr.io/platformfuzz/gnss-rinex-tools-image:latest
docker run --rm -v "$PWD:/data" \
  ghcr.io/platformfuzz/gnss-rinex-tools-image:latest \
  check /data/your_file.rnx
```

## CI

Pull requests and pushes run golangci-lint, govulncheck, Go tests, and a local
`ko build`. Pushes to `main` and `v*` tags publish multi-arch images to GHCR
(`latest`, short SHA, and version tag when applicable).
