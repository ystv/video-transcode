# video-transcode

A rabbitmq based video transcoder.

Should be used in tandem with [web-api](https://github.com/ystv/web-api). Server mode is for development purposes.

### dependencies

- rabbitmq (both)
- ffmpeg (client)
- s3-like api (client)

## Deploying

`go build cmd/client/main.go`
