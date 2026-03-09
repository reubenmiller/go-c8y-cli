
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
    import AsciinemaPlayer from '@site/src/components/AsciinemaPlayer';

    <AsciinemaPlayer
    src="https://asciinema.org/a/414235.cast"
    rows={30}
    preload
    fit="width"
    theme="monokai"
    poster="npt:0:03"
    />
    ```
