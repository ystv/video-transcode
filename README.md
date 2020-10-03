# video-transcode

A rabbitmq based video transcoder.

Should be used in tandem with [web-api](https://github.com/ystv/web-api). Server mode is for development purposes.

### dependencies

- rabbitmq (both)
- ffmpeg (client)
- s3-like api (client)

## Deploying

You'll need to first set the environment variables. Either through system's environment or through a `.env` or `.env.local` file.

`go build cmd/client/main.go`

### Environment variables

- `VT_AMQP_ENDPOINT` - AMQP 0.9.1 compatible broker (i.e. rabbitmq)
- `VT_CDN_ENDPOINT` - S3 compatible API (i.e. minio, s3, ceph)
- `VT_CDN_ACCESSKEYID`
- `VT_CDN_SECRETACCESSKEY`
