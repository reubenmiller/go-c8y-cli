## DEVELOPER NOTES


### Creating a cli demo

Demos can be recorded using [asciinema](https://asciinema.org/)

1. Start the recording

    ```sh
    asciinema rec demo.cast
    ```

2. Stop the record by pressing `ctrl+d` or exit the console `exit`

3. Convert the recording to an animated svg file

    ```sh
    cat demo.cast | svg-term --out demo.svg --window --term iterm2 --profile  "Afterglow.itermcolors"
    ```

4. The SVG can be referenced in markdown

    ````markdown
    <p align="center">
      <img width="1000" src="demo.svg">
    </p>
    ````
