# CHANGELOG

## Unreleased

* Updating tparse library when parsing relative dates. Using a relative date/timestamp of "30" no longer causes a panic.

* Removed unsupported query parameters from Remove-MeasurementCollection
    * `valueFragmentType`
    * `valueFragmentSeries`

* Removed deprecated commands
    `Remove-AuditRecordCollection`

* Updated Pester module to v5

## Released

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
