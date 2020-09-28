# CHANGELOG

## Unreleased

## Released

### v1.4.0

#### PSc8y (PowerShell)

**Renamed cmdlets**

* `New-ChildAssetReference` to `Add-AssetToGroup`
* `Remove-ChildAssetReference` to `Remove-AssetFromGroup`
* `New-ChildDeviceReference` to `Add-ChildDeviceToDevice`
* `Remove-ChildDeviceReference` to `Remove-ChildDeviceFromDevice`

**New cmdlets**

* `Add-DeviceToGroup`
* `Add-ChildGroupToGroup`
* `Remove-DeviceFromGroup`

**Changes**

* `New-TestDevice` has changed to require confirmation before creating the device
* Fixed `Get-Measurement` alias `m` in PSc8y

#### c8y (binary)

* Removed duplicate command `c8y devices find` as query functionality is provided by `c8y devices list --query "name eq 'test*'"`

* Added common options and array items processing to the following commands
    * `c8y agents list`
    * `c8y devices list`
    * `c8y devices listDeviceGroups`

* Added common options to (i.e. --outputFile, --pretty)
    * `c8y micrservices create`
    * `c8y micrservices createHostedApplication`

* Added bash profile script to add support for aliases
* Added guide to creating custom bash aliases
* Removed unreferenced commands

* `c8y session list` Renamed `--filter` to `--sessionFilter` to avoid conflict with the global `--filter` option


#### Docker

* Added docker images to make it easier to try out c8y

    * ghcr.io/reubenmiller/c8y-bash
    * ghcr.io/reubenmiller/c8y-zsh
    * ghcr.io/reubenmiller/c8y-pwsh

#### Docs

* Fixed line wrapping within code blocks. Now horizontal scrollbars are show to preserve the line spacing.
* Added github project link

#### Build

* Improved reliability of realtime api tests
* c8y (golang) binaries are now statically linked (using `CGO_ENABLED=0`) to make them more portable
* Added docker images for pwsh, bash and zsh with c8y already configured

### v1.3.0

**PSc8y (PowerShell)**

* Using a relative date/timestamp of "30" no longer causes a panic when using `DateFrom` and `DateTo` parameters

* Removed unsupported query parameters from Remove-MeasurementCollection
    * `valueFragmentType`
    * `valueFragmentSeries`

* Removed deprecated commands
    `Remove-AuditRecordCollection`

* Fixed `New-Microservice` cmdlet and updated the examples. Resolves #14

* `Set-Session` supports improved filtering
    * Search terms are split on whitespace, and each search term is matched individually against the following session properties
        * index
        * filename (basename only)
        * host
        * tenant
        * username

    * The session can be pre-filtered. Use full when you have a lot of sessions

        ```powershell
        # Powershell
        Set-Session "customer", "dev"

        # Example custom function to set a session for a specific tenant group
        # This can be stored in your powershell profile
        Function Set-CustomerXSession { Set-Session "customer" }
        ```

        ```sh
        # Bash
        c8y session list --filter "customer dev"
        ```

**Build Improvements**

* Updated Pester module to v5
* Updated `go-c8y` dependency to 0.8.0
* Enabled ci test runner
* Added parallel powershell tests execution to reduce test execution times
* Updated golang dependencies

### v1.2.0

**PSc8y (PowerShell)**

* Added Powershell documentation to [online docs](https://reubenmiller.github.io/go-c8y-cli/pwsh/cmdlets/Get-Agent/)

* Removed conflicting cmdlets to make it more obvious which one is correct to use
    * Deleted `Invoke-CustomRestMethod`. Use `Invoke-ClientRequest` instead.
    * Deleted `Invoke-CustomRestRequest`. Use `Invoke-ClientRequest` instead.
    * Moved `Invoke-ClientCommand` to a private command as it is only used internally by other PSc8y cmdlets

* Added additional parameters to `Invoke-ClientRequest`
    * `-Headers`. Add a hashtable of key/values which will be added to headers of the REST request
    * `-ContentType`. Add the `Content-Type` header to the REST request.
    * `-Accept`. Add the `Accept` header to the REST request.
    * `-IgnoreAcceptHeader`. Ignore the accept header when sending the request.
