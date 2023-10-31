# A Unfolded Circle Remote Two Integration for Denon AV Receiver

![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/splattner/remotetwo-integration-denonavr/main.yaml)
![GitHub](https://img.shields.io/github/license/splattner/remotetwo-integration-denonavr)
![GitHub (Pre-)Release Date](https://img.shields.io/github/release-date-pre/splattner/remotetwo-integration-denonavr)

The project uses the [goucrt](https://github.com/splattner/goucrt) library.

This [Unfolded Circle Remote Two](https://www.unfoldedcircle.com/) integration driver written in Go implements is to control a Denon AV Receiver.

Currently a [`MediaPlayer` entity](https://github.com/unfoldedcircle/core-api/blob/main/doc/entities/entity_media_player.md) and some [`Button` entities](https://github.com/unfoldedcircle/core-api/blob/main/doc/entities/entity_button.md) are implemented.

The Denon AV Receiver is controlled via its http based interface. Optionally you can enable telnet based integration during setup which improves the response speed of the integration. Using telnet provides realtime updates (local push) for many values but each receiver is limited to a single connection. If you enable this setting, no other connection to your device can be made via telnet.

## How to use

```bash
Denon AVR Integration for a Unfolded Circle Remote Two

Usage:
  rtintg-denonavr-amd64 [flags]

Flags:
      --debug                         Enable debug log level
      --disableMDNS                   Disable integration advertisement via mDNS
  -h, --help                          help for rtintg-denonavr-amd64
  -l, --listenPort int                the port this integration is listening for websocket connection from the remote (default 8080)
      --registration                  Enable driver registration on the Remote Two instead of mDNS advertisement
      --registrationPin string        Pin of the RemoteTwo for driver registration
      --registrationUsername string   Username of the RemoteTwo for driver registration (default "web-configurator")
      --remoteTwoIP string            IP Address of your Remote Two instance (disables Remote Two discovery)
      --remoteTwoPort int             Port of your Remote Two instance (disables Remote Two discovery) (default 80)
      --ucconfighome string           Configuration directory to save the user configuration from the driver setup (default "./ucconfig/")
      --websocketPath string          path where this integration is available for websocket connections (default "/ws")

```

### As a Container

You can start the Integration as a container by using the released container images:

Example to start the DenonAVR Integration:

```bash
docker run ghcr.io/splattner/remotetwo-integration-denonavr:v0.2.7
```

To keep the setup data persistet mount a volume to `/app/ucconfig`:

```bash
docker run -v ./localdir:/app/ucconfig ghcr.io/splattner/remotetwo-integration-denonavr:v0.2.7
```

For the mDNS adventisement to work correctly I suggest starting the integration in the `host` network. And you can set your websocket listening port with the environment variable `UC_INTEGRATION_LISTEN_PORT`:

```bash
docker run --net=host -e UC_INTEGRATION_LISTEN_PORT=10000 -v ./localdir:/app/ucconfig ghcr.io/splattner/remotetwo-integration-denonavr:v0.3.7
```

### Configuration

#### Environment Variables

The following environment variables exist in addition to the configuration file:

| Variable                     | Values               |Description |
|------------------------------|----------------------|--------------------------------------------------------------------------------|
| UC_CONFIG_HOME               | _directory path_     | Configuration directory to save the user configuration from the driver setup.<br>Default: `./ucconfig/` |
| UC_DISABLE_MDNS_PUBLISH      | `true` / `false`     | Disables mDNS service advertisement.<br>Default: `false` |
| UC_INTEGRATION_LISTEN_PORT | `int` | The port this integration is listening for websocket connection from the remote.<br> Default: `8080` |
| UC_INTEGRATION_WEBSOCKET_PATH | `string` | Path where this integration is available for websocket connections.<br> Default: `/ws` |
| UC_RT_HOST | `string` | IP Address of your Remote Two instance (disables Remote Two discovery via mDNS for registration) |
| UC_RT_PORT | `int` | Port of your Remote Two instance (disables Remote Two discovery via mDNS for registration) |
| UC_ENABLE_REGISTRATION | `string` | Enable driver registration on the Remote Two instead of mDNS advertisement.<br> Default: `false` |
| UC_REGISTRATION_USERNAME | `string` | Username of the RemoteTwo for driver registration.<br> Default: `web-configurator` |
| UC_REGISTRATION_PIN | `string` | Pin of the RemoteTwo for driver registration |


## How to Build and Run

```bash
# in cmd/rtintg-denonavr
go get -u
go build .
```

### Docker

```bash
docker build -f build/Dockerfile -t  ghcr.io/splattner/remotetwo-integration-denonavr:latest
```

## Verifying

### Checksum

### Checksums

```shell
wget https://github.com/splattner/remotetwo-integration-denonavr/releases/download/v0.2.7/remotetwo-integration-denonavr_0.2.7_checksums.txt
cosign verify-blob \
  --certificate-identity 'https://github.com/splattner/remotetwo-integration-denonavr/.github/workflows/release.yaml@refs/tags/v0.2.7' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  --cert https://github.com/splattner/remotetwo-integration-denonavr/releases/download/v0.2.7/remotetwo-integration-denonavr_0.2.7_checksums.txt.pem \
  --signature https://github.com/splattner/remotetwo-integration-denonavr/releases/download/v0.2.7/remotetwo-integration-denonavr_0.2.7_checksums.txt.sig \
  ./remotetwo-integration-denonavr_0.2.7_checksums.txt
```

You can then download any file you want from the release, and verify it with, for example:

```shell
wget https://github.com/splattner/remotetwo-integration-denonavr/releases/download/v0.2.7/remotetwo-integration-denonavr_0.2.7_linux_amd64.tar.gz.sbom
wget https://github.com/splattner/remotetwo-integration-denonavr/releases/download/v0.2.7/remotetwo-integration-denonavr_0.2.7_linux_amd64.tar.gz
sha256sum --ignore-missing -c remotetwo-integration-denonavr_0.2.7_checksums.txt
```

And both should say "OK".

You can then inspect the `.sbom` file to see the entire dependency tree of the binary.

### Docker image

```shell
cosign verify ghcr.io/splattner/remotetwo-integration-denonavr:v0.2.7 --certificate-identity 'https://github.com/splattner/remotetwo-integration-denonavr/.github/workflows/release.yaml@refs/tags/v0.2.7' --certificate-oidc-issuer 'https://token.actions.githubusercontent.com'
```

## License

This project is licensed under the [**Mozilla Public License 2.0**](https://choosealicense.com/licenses/mpl-2.0/).
See the [LICENSE](LICENSE) file for details.
