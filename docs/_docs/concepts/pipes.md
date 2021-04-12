
## Piping

### Shell

#### Piping ids

```sh
echo -e "12345\n22222" | c8y inventory get
```

#### Piping names

```sh
echo -e "device01\ndevice02" | c8y inventory get
```

#### Create devices

```sh
seq -f device_%04g 5 | c8y devices create --output table --select id,name,type --template "{ other_props: {}, type: 'ci_Test'  }"
```

**Output**

```sh
| id                    | name                       | type                   |
|-----------------------|----------------------------|------------------------|
| 481033                | device_0001                | ci_Test                |
| 481034                | device_0002                | ci_Test                |
| 480956                | device_0003                | ci_Test                |
| 481035                | device_0004                | ci_Test                |
| 480860                | device_0005                | ci_Test                |
```

Control the start and stop index

```
seq -f device_%04g 6 10 | c8y devices create --output table --select id,name,type --template "{ other_props: {}, type: 'ci_Test'  }"
```

**Response**

```
| id                    | name                       | type                   |
|-----------------------|----------------------------|------------------------|
| 481036                | device_0006                | ci_Test                |
| 480957                | device_0007                | ci_Test                |
| 481037                | device_0008                | ci_Test                |
| 480861                | device_0009                | ci_Test                |
| 480862                | device_0010                | ci_Test                |
```

#### Piping json

c8y also accepts piped json input (each line must represent a json object, this is automatically handled by c8y when piping to other c8y commands)


```sh
c8y inventory find --query="not(has(myTag))" | c8y inventory update --data "myTag={}" --dry
```

Or if you are just returning a single item

```sh
c8y inventory get --id=12345 | c8y inventory update --data "myTag={}" --dry
```

### Chaining commands

```
c8y devices create --name "myname" | c8y identity create --type c8y_Serial --template "{ externalId: input.value.name }"
```

#### Piping inventory items and us multiple workers to speed up the retrieval

```sh
c8y inventory list --pageSize 15 | c8y devices get --workers 5
```

However if you are getting a list of devices, then you can just pipe the values directly

```sh
echo "1111\n2222" | c8y inventory get | c8y devices get --workers 1
```
