# pingrb cli

Send yourself a push notification from anything that can shell out.

## Install

```sh
brew install ruby-native/tap/pingrb
```

Or build from source:

```sh
go install github.com/ruby-native/pingrb-cli@latest
```

## Usage

```sh
pingrb config <your-token>
pingrb "deploy failed"
pingrb "job done" --body "backfill finished" --url https://example.com/jobs/42
some-long-job && pingrb "$?" --body "done"
```

Get the token from your Custom source on https://pingrb.com (it's the last
segment of the webhook URL).

Config is stored at the platform's standard user config dir
(`~/.config/pingrb` on Linux, `~/Library/Application Support/pingrb` on macOS).

Set `PINGRB_HOST` to point at a non-production instance (defaults to
`https://pingrb.com`).

## Develop

```sh
go test ./...
go build -o pingrb .
./pingrb --help
```

## Release

Tag and push:

```sh
git tag v0.1.0
git push --tags
```

GitHub Actions builds binaries for darwin and linux on amd64 and arm64 via
goreleaser, then publishes them as a GitHub Release.
