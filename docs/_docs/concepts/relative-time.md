---
layout: default
category: Concepts
title: Relative time/dates
---

### Using relative time/dates

All date and time parameters accept a full date, i.e. `2020-03-31 10:00:00`. Common variantions of datetime formats are accepted, however normally it is very tedious to have to enter in the full date and timestamp when searching for data.

To make life easier, all date and time parameters also accept a relative time format, which will be translated for you to the full datetime.


#### Formats

The following shows some examples of which kinds of relative time formats are accepted.

| Relative Time | Meaning |
|-------|---------|
| 0s | Now |
| -100ms | 100 milliseconds ago |
| -10m | 10 minutes ago |
| 10m | 10 minutes from now |
| -1h | 1 hour ago |
| -1h20min | 1 hour and 20 minutes ago |
| -14d | 14 days ago |
| -6months | 6 months ago |
| -12mo | 12 months ago |
| -2y | 2 years ago |
| -30h+30min | 29 hours and 30 minutes ago |


The relative time and dates is supported in any commands which have the following parameter names.

| PSc8y | c8y |
|-------|---------|
| -DateFrom | `--dateFrom` |
| -DateTo | `--dateTo` |
| -Time | `--time` |

---

#### Example 1: Get a list of FAILED operations in the last day

##### Bash

Get a list of FAILED operations in the last day

```sh
c8y operations list --dateFrom "-1d" --status FAILED
```

##### Powershell

```powershell
Get-OperationCollection -DateFrom "-1d" -Status FAILED
```
