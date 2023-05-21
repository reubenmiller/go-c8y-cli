---
category: Concepts - Extensions - API based commands
title: API based commands
---

import DocCardList from '@theme/DocCardList';

## Overview

:::tip
A general rule of thumb is that try to use API based commands by default as they are more portable as they do not rely on any external executables, only go-c8y-cli.
:::

API based commands are commands which are generated via a YAML specification. Each specification can contain multiple commands and each command corresponds to one HTTP REST request.

API based commands offer a first class integration into go-c8y-cli as the feel much more like other in-built commands as they include; tab completion, documentation, support for in-built data types (e.g. devices, agents, applications etc.). This is possible because under the hood go-c8y-cli is applying the same API processing that it uses internally to generate the golang code for each command, however the main difference is that go-c8y-cli is interpreting the commands at runtime rather than generating static golang code.

<DocCardList />
