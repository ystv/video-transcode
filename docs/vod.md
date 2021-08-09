# VOD task

`POST` to `/task/video/vod` with a body object of:

```
{
    "srcURL":"$FILE_TO_BE_TRANSCODED",
    "dstArgs":"$FFMPEG_ARGS_APPLIED_IN_THE_OUTPUT_SECTION",
    "dstURL":"$DESTINATION"
}
```
