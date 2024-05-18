---
title: Remote Access
---

import CodeExample from '@site/src/components/CodeExample';

:::tip
The `remoteacccess` subcommand is only available from go-c8y-cli >= 2.41
:::

The [Cumulocity IoT Cloud Remote Access ](https://cumulocity.com/docs/cloud-remote-access/cra-general-aspects/) features allows you to connect to your devices using a number of different protocols like SSH, VNC, Telnet etc. In addition, it also supports a **PASSTHROUGH** mode which allows you to use any TCP based protocol.

Please refer to the [official Cumulocity IoT documentation](https://cumulocity.com/docs/cloud-remote-access/cra-general-aspects/) for more details.

**Common use-cases**

Below are some common use-cases for the Cloud Remote Access Feature:

* Access a local webpage from the device (such as Node-RED, or tools like monit)
* Connect to the device using SSH (via the cloud - local network access is not required)
* Connect a debugger to local control software to troubleshoot complex control logic

**What it is not designed for**

The Cloud Remote Access feature is designed for adhoc connections to provide access to services running locally on a device in a secure manner. For instance, if you need to perform some troubleshooting (e.g. root-cause analysis) on individual devices, in order to develop a fix that can be deployed to the rest of the device fleet using the other Cumulocity IoT device management features like [Firmware](https://cumulocity.com/docs/device-management-application/managing-device-data/#managing-firmware) and [Software Updates](https://cumulocity.com/docs/device-management-application/managing-device-data/#managing-software).

The feature (in my opinion) is not designed for "always-on" connections and a large number of parallel connections. You should definitely NOT be trying to abuse the feature by performing some kind of Ansible deployment to devices via SSH. If you have this use-case, it is highly recommended looking into [thin-edge.io](https://thin-edge.io) which can fulfill all your device management requirements in a more scalable manner.

**Background on the PASSTHROUGH feature**

The **PASSTHROUGH** is relatively new and there doesn't seem to be much documentation out there about it. The feature is insanely useful as it allows you to create secure adhoc connections to your device using existing technologies such as SSH (using the standard ssh client). Cumulocity IoT facilitates the connection between a client on your local machine, and a component on the device, and routes traffic bi-directionally between the two. Since Cumulocity IoT is just passing bytes between the two clients, any TCP based protocol can be used.

There are a few moving parts in this scenario, as there needs to be a client on both sides of the connection; one on your machine, which is interfacing with the local protocol, and a client on the device which routes the traffic between the cloud and a local service (such as the SSH daemon/service, or a local HTTP server).

The good news is that there are existing open source projects which can be used to take advantage of the Remote Access feature; these components are:

* [go-c8y-cli](https://goc8ycli.netlify.app/docs/introduction/) (on your machine)
* [thin-edge.io](https://thin-edge.io) - a Rust based agent that has out-of-the-box support for the Cumulocity IoT Cloud Remote Access feature (on the device)

### Prerequisites

Before you can use this feature you need to have the "Cloud Remote Access" feature enabled in your tenant. Please consult the [official docs](https://cumulocity.com/docs/cloud-remote-access/using-cloud-remote-access/) on how to do this.

#### Granting permission to use Cloud Remote Access

By default, the Cumulocity IoT `ROLE_REMOTE_ACCESS_ADMIN` permission is not assigned to any user group or user. This means that you'll need to add it before you will even be able to add any configurations and even see it in the Cumulocity IoT Device Management application.

One option is to add the permission to the existing "admins" user group, and then assign the "admins" group to your user. You can do this via the Cumulocity IoT Administration app, or using the following command:

<CodeExample>

```sh
c8y userroles addRoleToGroup --group admins --role ROLE_REMOTE_ACCESS_ADMIN
```

</CodeExample>

Alternatively, you can create a new custom user group, and only assign the permission to it. This user group can be used to grant fine grain access to this feature ensure that the user does not inherit any additional permissions that are provided by the "admins" user group. Below shows the cli commands used to create a new user group called "devmgmt-powerusers":

<CodeExample>

```sh
c8y usergroups create --name "devmgmt-powerusers"
c8y userroles addRoleToGroup --group  "devmgmt-powerusers" --role ROLE_REMOTE_ACCESS_ADMIN
```

</CodeExample>

You can add users to the user group to grant them access to this feature using the following command:


<CodeExample>

```sh
c8y userreferences addUserToGroup --group devmgmt-powerusers --user "myuser@example.com"
```

</CodeExample>


### Using ssh config to launch proxy command automatically

The remote access feature works very well with the ssh `ProxyCommand` which provides the most "native" ssh experience as you don't have to manually call the `c8y` command when you want to start your ssh session.

For example, with specific configuration, you can then connect to your device (via Cumulocity IoT) using a plain ssh command.

```sh
ssh my_device
```

To achieve this, you need to create an entry in your ssh client configuration file (e.g. `~/.ssh/config`) which specifies the name of the device you want to connect to and add the `ProxyCommand` line which calls the `c8y remoteaccess server` command.

Below shows the format of the ssh client configuration entry that you need to add:

```text title="file: ~/.ssh/config"
Host <device>
    User <device_username>
    PreferredAuthentications publickey
    IdentityFile <identify_file>
    ServerAliveInterval 120
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
    ProxyCommand c8y remoteaccess server --device %n --listen -
```

Using the above format, if you want to connect to a device called `my_device`, then your ssh client config entry would like this (remember to update your `IdentityFile` to match your private ssh key):

```text title="file: ~/.ssh/config"
Host my_device
    User admin
    PreferredAuthentications publickey
    IdentityFile ~/.ssh/id_rsa
    ServerAliveInterval 120
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
    ProxyCommand c8y remoteaccess server --device %n --listen -
```

Afterwards, when you run the `ssh my_device` command, the ssh client will automatically run the `c8y remoteaccess server` command. The `--listen -` option tells the local proxy server to use Standard Input/Output instead of using a unix socket, or TCP server.

## Examples

### Connect using ssh

This example walks you through connecting to your device using ssh. The ssh connection will go through a local proxy (which will be started automatically by the c8y command), this will trigger an operation which is sent to the device, where thin-edge.io will pass the connection through to the local SSH daemon (service).

Before we can connect to the device with a native ssh session, you will need to ensure that a **PASSTHROUGH** configuration has been added to the device. This configuration tells the Remote Access service which port the traffic should be routed to on the target device.


#### Step 1: Create the passthrough configuration in Cumulocity IoT

You can create the required remote access configuration using the following command:

<CodeExample>

```sh
c8y remoteaccess configurations create-passthrough --device my_device --name native-ssh --port 22
```

</CodeExample>

:::tip
If you want to check if a configuration already exists, then you can also list the existing configuration using.

<CodeExample>

```sh
c8y settings update remoteaccess.sshuser admin
```

</CodeExample>

Afterwards, you don't need to provide the `--user <name>` flag when connecting to the device.
:::


#### Step 2:  Add your public key to the device

Connecting to any machine via ssh needs some authentication to be already setup. Some common SSH authentications mechanisms are as follows:

* Username/Password (not recommended)
* Public Key (via `~/.ssh/authorized_keys`)
* Certificate based authentication

:::info
**c8y remoteaccess** is not responsible for any part of the ssh authentication, this is handled entirely by the ssh client. You might find it helpful to read up on SSH authentication methods and how to configure an ssh-agent on your machine so that you don't have to constantly enter the password to your private ssh key (if you're using Public key authentication etc.).
:::

**Example: Adding your public key to the authorized_keys file**

If you're using Public Key authentication, you can add your public ssh key to the device using the following commands (assuming you have access to the device locally, or it is part of your base image):

```sh
mkdir -p ~/.ssh
chmod 700 ~/.ssh
echo "ssh-ed25519 AAAAC3NzbC1lZDi1NTE5AAAdIDxm81Bbp+bQfMM0jgoMmUD7nXpBCclEq+NqVxUf5D65 myuser@example.com" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
```

:::note
If the above information means nothing to you, then I would recommend reading up on SSH authentication or contacting whomever is responsible for the administration of the device your are trying to connect to.
:::

#### Step 3: Connect to the device via ssh

On your local machine, you can then connect to the device via ssh using the following command, where `--user` is the SSH user on the target device.

<CodeExample>

```sh
c8y remoteaccess connect ssh --device my_device --user admin
```

</CodeExample>

If your fleet of devices always has the same SSH user, then you can configure the default ssh user

<CodeExample>

```sh
c8y settings update remoteaccess connect ssh --device my_device --user admin
```

</CodeExample>

### Running a custom command

```sh
#!/usr/bin/env bash
echo "TARGET: $TARGET"
echo "PORT: $PORT"
```

```sh
c8y remoteaccess run --device my_device -- 
```

### Create local proxy server

In some cases you don't want to launch the services immediately, but install you first want to start the local server, and then allow clients to connect to it.

A local proxy server can be configured to listen on one of the following mediums:

* Unix socket
* TCP server
* Standard Input/Output (stdio)

**Local host with a randomized port**

Use port `0` if you want a randomized/free port. The used port will be printed to the console.

<CodeExample>

```sh
c8y remoteaccess server --listen "127.0.0.1:0"
```

</CodeExample>

Alternatively, you can explicitly set the port using:

<CodeExample>

```sh
c8y remoteaccess server --listen "127.0.0.1:9000"
```

</CodeExample>


**Unix Socket**

Unix sockets can also be used by using the `unix://` prefix in the `listen` flag.

<CodeExample>

```sh
c8y remoteaccess server --listen "unix:///tmp/example.sock"
```

</CodeExample>

Clients can then communicate with the local proxy server instance via given unix socket, `unix:///tmp/example.sock`.

## Migrating from c8ylp

**go-c8y-cli** can now be used to fully replace the previous Python based tool, [c8ylp](https://github.com/SoftwareAG/cumulocity-remote-access-local-proxy).

The **c8y remoteaccess** commands can be used to provide the same functionality, though maybe slightly less configurable (e.g. you can't currently change the buffer size). If you feel that there are some additional settings that should be exposed, then please create an issue on the [Github project](https://github.com/reubenmiller/go-c8y-cli/issues/new).

The following are the main differences between **c8ylp** and the **c8y remoteaccess** commands:

* The `--env-file` flag is not supported as **go-c8y-cli** uses session files/env variables to managed the Cumulocity IoT tenant/credentials

* **c8ylp** only supported referring to a device by its external identity (with the `c8y_Serial` type). The **go-c8y-cli** version uses [reference by name](/docs/concepts/reference-by-name/) which allows supports both a named lookup and providing the explicit external id. The new way is more consistent with existing behaviour where you can also benefit from tab completion, something that is not possible when using external identities. In the future, **go-c8y-cli** will support additional lookup methods, however for now if you need to do a lookup via the external identity, then you can use an alias which can include an identity lookup. See the [migrating from c8ylp](#migrating-from-c8ylp) section for more details.

To help with the migration, the following sections show how to transition from [c8ylp](https://github.com/SoftwareAG/cumulocity-remote-access-local-proxy).


#### Starting local (tcp) proxy server on a random port

**Before**

```sh
c8ylp server "<device>" --env-file .env
```

**After**

```sh
c8y remoteaccess server --device "<device>"
```

### Start interactive ssh session

**Before**

```sh
c8ylp connect ssh "<device>" --ssh-user "<device_username>" --env-file .env
```

**After**

```sh
c8y remoteaccess server --device "<device>" --user "<device_username>"
```


### Starting a local proxy server using a unix socket

**Before**

```sh
c8ylp server <device> --env-file .env --socket-path /tmp/device.socket
ssh -o 'ProxyCommand=socat - UNIX-CLIENT:/tmp/device.socket' <device_username>@localhost
```

**After**

```sh
c8y remoteaccess server --device <device> --listen unix:///tmp/device.socket
ssh -o 'ProxyCommand=socat - UNIX-CLIENT:/tmp/device.socket' <device_username>@localhost
```

### Starting a local proxy using stdio

If you've previously used the ssh client config with the `ProxyCommand` property, you were probably using the `--stdio` mode. This mode is also supported in `c8y remoteaccess`, and you can activate the mode by setting the `--listen` flag to `-`, where `-` means StandardInput.

**Before**

```sh
ProxyCommand c8ylp server %n --stdio --env-file .env
```

**After**

```sh
ProxyCommand c8y remoteaccess server --device %n --listen -
```

### Connect to a device using the external identity

MacOS/Linux/WSL users can use an alias which uses the `c8y identity get` to do the external identity lookup and passes the result to the `--device` flag. To make the one-liner more accessible, it can be assigned to an alias using the following command:

```sh
c8y alias set ssh 'c8y remoteaccess connect ssh --device "$(c8y identity get --name "$1" --select managedObject.id -o csv)" --user $2' --shell
```

The alias expects 2 arguments, the external identity, then the ssh user name. An example showing the alias usage is shown below:

```sh
c8y ssh rpi4-abc632486800 root
```

```sh title="Output"
Starting interactive ssh session with 4550259 (https://example.cumulocity.com)

Warning: Permanently added '[127.0.0.1]:52480' (RSA) to the list of known hosts.
root@rpi4-abc632486800:~# 
```
