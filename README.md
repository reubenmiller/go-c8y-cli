# go-c8y-cli

![build](https://github.com/reubenmiller/go-c8y-cli/workflows/build/badge.svg)



Unofficial Cumulocity REST Command Line Interface for both PowerShell and *nix (standalone binary).

Compatible with

* Linux (amd64)
* MacOS (amd64)
* Windows (amd64)

## Installation

### PowerShell Module [(PSc8y)](https://www.powershellgallery.com/packages/PSc8y)

```powershell
Install-Module PSc8y -AllowClobber -AllowPrerelease -Scope CurrentUser
Import-Module PSc8y
```

**Note**

Please consult the docs if you are having trouble installing it.

* [bash](https://reubenmiller.github.io/go-c8y-cli/docs/1-bash-installation/)
* [PowerShell](https://reubenmiller.github.io/go-c8y-cli/docs/1-powershell-installation/)


## Documentation

See the [documenation website](https://reubenmiller.github.io/go-c8y-cli/) for instructions on how to install and use it.

## Contributing

1. Fork the project, then clone it

    ```sh
    git clone https://github.com/reubenmiller/go-c8y-cli.git
    ```

2. Open the project in Microsoft VS Code (using Dev Containers - this requires Docker!)

3. Edit a `.yml` specification in `api/spec/yaml/`

4. Build the project using

    ```sh
    make build
    ```

5. Try out the newly built module

    **PowerShell**

    ```powershell
    Import-Module ./tools/PSc8y/dist/PSc8y -Force
    ```

    **Bash**

    ```powershell
    chmod +x ./output/c8y.*

    ./output/c8y.linux
    ```

### Building documentation

1. Build the go and powershell code

    ```sh
    make build
    ```

2. Build the powershell markdown documentation

    ```sh
    make docs-powershell
    ```

3. Start a webserver to preview the documentation locally

    ```sh
    make gh_pages
    ```

4. View the documentation in the [brower](http:/localhost:4000)
