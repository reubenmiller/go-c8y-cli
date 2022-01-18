
## Recording a demo

These instructions show how to make a recording which can be uploaded to [asciinema](https://asciinema.org/) and referenced from the documentation.

The demo scripts are used to run multiple command and simulate typing before then executing the script.

### Recording a demo

1. Execute a prepared demo script

    ```bash
    cd tools/docs/demos
    ./demo.sh ./activitylog_01.sh
    ```

2. Upload the file to asciinema

    ```bash
    asciinema upload activitylog_01.sh.asc
    ```

3. Reference the video from a markdown file

    ```markdown
    import Video from '@site/src/components/video';

    <Video
    videoSrcURL="https://asciinema.org/a/414235/iframe?speed=1.0&autoplay=false&size=small&rows=30"
    videoTitle="Activitylog example"
    width="90%"
    height="600px"
    ></Video>
    ```
