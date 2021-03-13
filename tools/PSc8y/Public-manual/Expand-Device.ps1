Function Expand-Device {
<#
.SYNOPSIS
Expand a list of devices replacing any ids or names with the actual device object.

.DESCRIPTION
The list of devices will be expanded to include the full device representation by fetching
the data from Cumulocity.

.NOTES
If the function calling the Expand-Device has a "Force" parameter and it is set to True, then Expand-Device will not fetch the device managed object
from the server. Instead it will return an object with only the id and name set (and the name will be set to [id={}]). This is to save the
number of calls to the server as usually the ID is the item you need to use in subsequent calls.

If the given object is already an device object, then it is added with no additional lookup

The following cases describe when the managed object is fetched from the server and when not.

Cases when the managed object IS fetched from the server
* Calling function does not have -Force on its function, and the user does not use it. OR
* OR User provides input a string which does not only contain digits
* OR User sets the -Fetch parameter on Expand-Device

Cases when the managed object IS NOT fetched from the server
* User passes an ID like object to Expand-Device
* AND -Force is not used on the calling function
* AND user does not use -Fetch when calling Expand-Device


.OUTPUTS
# Without fetch
[pscustomobject]@{
    id = "1234"
    name = "[id=1234]"
}

# With fetch
[pscustomobject]@{
    id = "1234"
    name = "mydevice"
}

.PARAMETER InputObject
List of ids, names or device objects

.EXAMPLE
Expand-Device "mydevice"

Retrieve the device objects by name or id

.EXAMPLE
Get-DeviceCollection *test* | Expand-Device

Get all the device object (with app in their name). Note the Expand cmdlet won't do much here except for returning the input objects.

.EXAMPLE
Get-DeviceCollection *test* | Expand-Device

Get all the device object (with app in their name). Note the Expand cmdlet won't do much here except for returning the input objects.

.EXAMPLE
"12345", "mydevice" | Expand-Device -Fetch

Expand the devices and always fetch device managed object if an object is not provided via the pipeline

#>
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory=$true,
            ValueFromPipeline=$true,
            Position=0
        )]
        [object[]] $InputObject,

        # Fetch the full managed object if only the id or name is provided.
        [switch] $Fetch
    )

    Begin {
        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }
    }

    Process {
        $FetchFullDevice = $Fetch
        $callstack = Get-PSCallStack | Select-Object -Skip 1 -First 1
        if ($null -ne $callstack.InvocationInfo.MyCommand.Parameters -and $null -ne $callstack.InvocationInfo.BoundParameters) {
            $CanPromptUser = $callstack.InvocationInfo.MyCommand.Parameters.ContainsKey("Force")
            $ForceEnabled = (($callstack.InvocationInfo.BoundParameters["Force"] -eq $true) -or ($ConfirmPreference -match "None"))

            if (!$Fetch) {
                $FetchFullDevice = $CanPromptUser -and !$ForceEnabled
            }
        }
        

        [array] $AllDevices = foreach ($iDevice in $InputObject)
        {
            if ($iDevice.deviceId) {
                # operation
                [PSCustomObject]@{
                    id = $iDevice.deviceId
                    name = $iDevice.deviceName
                }
            } elseif ($iDevice.source.id) {
                # alarms/events/measurements etc.
                [PSCustomObject]@{
                    id = $iDevice.source.id
                    name = $iDevice.source.name
                }
            } elseif ($iDevice.id) {
                $iDevice
            } else {
                if ($iDevice -match "^\d+$") {
                    
                    if ($WhatIfPreference) {
                        # Fake the reponse of the managed object
                        [PSCustomObject]@{
                            id = $iDevice
                            # Dummy value
                            name = "name of $iDevice"
                        }
                    } else {
                        if ($FetchFullDevice) {
                            Get-ManagedObject -Id $iDevice -WhatIf:$false
                        } else {
                            [PSCustomObject]@{
                                id = $iDevice
                                # Dummy value
                                name = "[id=$iDevice]"
                            }
                        }
                    }
                } else {
                    Get-DeviceCollection -Name $iDevice -WhatIf:$false
                }
            }
        }

        $AllDevices
    }
}
