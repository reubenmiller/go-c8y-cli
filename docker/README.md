
## Bash


## PowerShell

Mount your existing sessions

```sh
docker run -v "$env:HOME/.cumulocity:/root/.cumulocity" -it c8y
```

Use a dotenv file

```sh
docker run --env-file=.env -it c8y
```
