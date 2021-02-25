# video-transcode

A rabbitmq and websocket based video transcoder.

Server provides a HTTP API to allow job creation without needing
to introduce rabbitmq to an application's stack. Clients will
work with additional or even custom server instances as long as
it meets the client's input requirements.

Client will watch the job queue and pickup jobs and attempt to
complete them. They will also connect to a server instance in
order to provide job statistics and cancellation support.

### dependencies

- rabbitmq (both)
- ffmpeg (client)
- s3-compatible api (client)

## Deploying

You'll need to first set the environment variables. Either through system's environment or through a `.env` or `.env.local` file.

### Client

`go build cmd/client/main.go`

### Server

`go build cmd/server/main.go`

### Environment variables

- `VT_AMQP_ENDPOINT` - AMQP 0.9.1 compatible broker (i.e. rabbitmq)
- `VT_CDN_ENDPOINT` - S3 compatible API (i.e. minio, s3, ceph)
- `VT_CDN_ACCESSKEYID`
- `VT_CDN_SECRETACCESSKEY`

## Developing as a dependency

Rough outline of useful information to integrate with VT.

### Server

Has the following endpoints:
* `/` Version page
* `/task/{name} [POST]` Create a new job
* `/task/vod [POST]` Create a Video on Demand job, uses the CDN
* `/task/raw [POST]` Creates a barebones FFmpeg job
* `/ws` WS connection for workers

### Message Queue

Exchange: `encode`
Queue name: `vod` / `live`
Accepted object:
```
// Task represents a task to transcode for VOD or raw
// Essentially just basic inputs to ffmpeg
Task struct {
	ID      string `json:"id"`      // Task UUID
	Args    string `json:"args"`    // Global arguments
	SrcArgs string `json:"srcArgs"` // Input file options
	SrcURL  string `json:"srcURL"`  // Location of source file on CDN
	DstArgs string `json:"dstArgs"` // Output file options
	DstURL  string `json:"dstURL"`  // Destination of finished encode on CDN
}
```