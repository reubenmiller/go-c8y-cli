
## Piping

### Bash/zsh

#### Piping ids

```sh
echo -e "12345\n22222" | c8y inventory get
```

#### Piping names

```sh
echo -e "device01\ndevice02" | c8y inventory get
```

#### Piping json

Piped json must be flattened first to a stream of objects

```sh
alias jqflat="jq -r '.[]' -c"
c8y inventory find --query="not(has(myTag))" | jqflat | c8y inventory update --data "myTag={}" --dry
```

Or if you are just returning a single item, then you don't need to use the jqflat alias as each object is streamed by itself and is not an actual json array.

```sh
c8y inventory get --id=12345 | c8y inventory update --data "myTag={}" --dry
```

### Pipeing ids

If a json line is piped to

**note**

The json must be on one line (not new line characters). If you you are using a result which returns a list, then the list must be converted to individual values (using jq).

```sh
c8y inventory list --pageSize 15 --select id | jq -r ".[]" -c | c8y devices get --select id --workers 5
```

However if you are getting a list of devices, then you can just pipe the values directly

```sh
echo "1111\n2222" | c8y inventory get | c8y devices get --select id --workers 1
```

### Chaining commands

```
c8y devices create --name "myname" | c8y identity create --type c8y_Serial --template "{ externalId: input.name }"
```

### Reshaping the data

#### Output csv of the inventory list

```sh
c8y inventory list | jq -r ".[] | [.id, .name] | @csv"
```

## Other jq options

Get the length of the results

```sh
c8y inventory list | jq length
```

## Useful jq aliases

```
# Get the total pages. Defaults to -1 if the total pages does not exist in the response
alias jqtotal='jq -r ".statistics?.totalPages // -1"'

# Get the length of the json array
alias jqlength='jq length'

alias jqiter='jq ".[]" -c'
```
