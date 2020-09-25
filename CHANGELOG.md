# CHANGELOG

## Unreleased


## Released

### v1.3.0

**PSc8y (PowerShell)**

* Using a relative date/timestamp of "30" no longer causes a panic when using `DateFrom` and `DateTo` parameters

* Removed unsupported query parameters from Remove-MeasurementCollection
    * `valueFragmentType`
    * `valueFragmentSeries`

* Removed deprecated commands
    `Remove-AuditRecordCollection`

* Fixed `New-Microservice` cmdlet and updated the examples. Resolves #14

**Build Improvements**

* Updated Pester module to v5
* Updated `go-c8y` dependency to 0.8.0
* Enabled ci test runner
* Added parallel powershell tests execution to reduce test execution times

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
