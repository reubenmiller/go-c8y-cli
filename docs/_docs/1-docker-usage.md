---
layout: default
category: Installation - Docker
title: Installation
---

The c8y tool is also available in the following docker images

* ghcr.io/reubenmiller/c8y-shell
* ghcr.io/reubenmiller/c8y-pwsh

**Note:** All of these images are linux containers (not Windows!), but can be called from any x64 based operating system

The docker images can be configured to re-use your existing c8y sessions on your machine, or use one a session configured via environment variables.

The following sections will details both methods. It is assumed that you have docker already installed on your machine, and you have permissions to run `docker` commands (or you have sudo right).

### Using a c8y docker image

The session persistence is achieved by mounting a docker volume to your host operating system. Any sessions that are created in the docker container, will be stored on your operation system in the mounted folder.

It is recommended that you run these commands from your home folder, so that the `~/.cumulocity` folder will be used to store the sessions, as this is the default folder that is used should you install the c8y cli tool on your host machine later on. That way you will you be able to continue using 


#### zsh/bash/fish

```sh
cd ~
docker run -it -v $PWD/.cumulocity:/sessions --rm ghcr.io/reubenmiller/c8y-shell:latest
```

**Note**

ZSH will start by default, however you can load another shell manuall using `bash` or `fish`.

#### PowerShell (pwsh)

```sh
cd ~
docker run -it -v $PWD/.cumulocity:/sessions --rm ghcr.io/reubenmiller/c8y-pwsh:latest
```

### Using a c8y docker image with environment variables

You can provide the session information via environment variables, deviced in a dotenv (.env) file.

1. Create a dotenv file with the following formatting

    *session.env*

    ```sh
    C8Y_HOST=https://example.cumulocity.eu-latest.com
    C8Y_TENANT=t12345
    C8Y_USER=myuser@example.com
    C8Y_PASSWORD=my4s3curep4assword
    ```

    **Note:** Please use LF line endings, and utf8 encoding (without a BOM).

2. Start the docker container

    ```sh
    docker run --rm -it --env-file=session.env ghcr.io/reubenmiller/c8y-shell:latest
    ```

    The `--env-file` argument will direct docker to map the file contents to environment variables within the container.

### Re-using an existing c8y session in docker

If you have already activated a c8y session on a command console, you can re-use the current session by simply passing the environment variables to the docker image. This can be useful if you want to try out the same session that you have loaded but in an other environment (i.e. using zsh).


1. Set a c8y session on your console

    **Shell**

    ```sh
    set-session
    ```

    **PowerShell**

    ```sh
    Set-Session
    ```

    **Note**

    The set-session helpers will create the following environment variables for use by other tools, in this case it will be docker.

    All you have to do to load the value

    * C8Y_HOST
    * C8Y_TENANT
    * C8Y_USER
    * C8Y_PASSWORD

2. Check if the environment variables have been set

    **Shell**

    ```sh
    echo $C8Y_HOST
    echo $C8Y_USER
    ```

    **PowerShell**

    ```sh
    echo $env:C8Y_HOST
    echo $env:C8Y_USER
    ```

3. Create a new container re-using the session

    **Shell**

    ```sh
    docker run -it --rm \
        -e C8Y_HOST=$C8Y_HOST \
        -e C8Y_TENANT=$C8Y_TENANT \
        -e C8Y_USER=$C8Y_USER \
        -e C8Y_PASSWORD=$C8Y_PASSWORD \
        ghcr.io/reubenmiller/c8y-shell:latest
    ```

    **PowerShell**
    
    ```powershell
    docker run -it --rm `
        -e C8Y_HOST=$env:C8Y_HOST `
        -e C8Y_TENANT=$env:C8Y_TENANT `
        -e C8Y_USER=$env:C8Y_USER `
        -e C8Y_PASSWORD=$env:C8Y_PASSWORD `
        ghcr.io/reubenmiller/c8y-shell:latest
    ```

**Note**

You have to execute a docker pull if you want to re-check if there is a newer image available (i.e. also tagged with latest). 

You can also specify the version that you want to try out by replacing `latest` with the version number, i.e. `2.0.0`.

```sh
# update to the latest image
docker pull ghcr.io/reubenmiller/c8y-shell:latest

# use a known version
docker pull ghcr.io/reubenmiller/c8y-shell:2.0.0
```
