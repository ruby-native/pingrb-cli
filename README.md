# pingrb cli

Send yourself a push notification from anything that can shell out.

## Install

Coming soon via Homebrew tap.

For now, build from source:

```sh
go install github.com/ruby-native/pingrb-cli@latest
```

## Usage

```sh
pingrb config https://pingrb.com/webhooks/custom/<your-token>
pingrb "deploy failed"
pingrb "job done" --body "backfill finished" --url https://example.com/jobs/42
some-long-job && pingrb "$?" --body "done"
```

The webhook URL is the Custom source URL from your account at https://pingrb.com.

Config is stored at the platform's standard user config dir
(`~/.config/pingrb` on Linux, `~/Library/Application Support/pingrb` on macOS).

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
