. $PSScriptRoot/imports.ps1

Describe -Name "Expand-Device" {
    BeforeAll {
        $Device = PSc8y\New-TestAgent
    }

    It "Expand device (with object)" {
        $Result = PSc8y\Expand-Device $Device
        $Result.id | Should -BeExactly $Device.id
    }

    It "Expand device (with object) by pipeline" {
        $Result = $Device | PSc8y\Expand-Device
        $Result.id | Should -BeExactly $Device.id
    }

    It "Expand device (with id)" {
        $Result = PSc8y\Expand-Device $Device.id
        $Result.id | Should -BeExactly $Device.id
    }

    It "Expand device (with id) by pipeline" {
        $Result = $Device.id | PSc8y\Expand-Device
        $Result.id | Should -BeExactly $Device.id
    }

    It "Expand device (with name)" {
        $Result = PSc8y\Expand-Device $Device.name
        $Result.id | Should -BeExactly $Device.id
    }

    It "Expand device (with name) by pipeline" {
        $Result = $Device.name | PSc8y\Expand-Device
        $Result.id | Should -BeExactly $Device.id
    }

    It "Expand device from Get-DeviceCollection" {
        $Result = Get-DeviceCollection $Device.name | PSc8y\Expand-Device
        $Result.id | Should -BeExactly $Device.id
    }

    It "Expand device from operation" {
        $operation = PSc8y\New-TestOperation -Device $Device.id
        $Result = $operation | PSc8y\Expand-Device
        $Result.id | Should -BeExactly $Device.id
    }

    It "Expand device from operation" {
        $measurement = PSc8y\New-TestMeasurement -Device $Device.id
        $Result = $measurement | PSc8y\Expand-Device
        $Result.id | Should -BeExactly $Device.id
    }

    It "Expand device called from a function with force" {
        Function Update-MyObject {
            [cmdletbinding()]
            Param(
                [Parameter(
                    Mandatory = $true,
                    Position = 0,
                    ValueFromPipeline = $true,
                    ValueFromPipelineByPropertyName = $true
                )]
                [object[]] $Device,
                [switch] $Force
            )

            foreach ($iDevice in (PSc8y\Expand-Device $Device)) {
                $iDevice
            }
        }
        # Passing id to an object
        $VerboseMessages = $( $Results = $Device.id | Update-MyObject -Force -Verbose ) 4>&1
        $Results.id | Should -Be $Device.id
        $Results.name | Should -Not -Be $Device.name
        [array] $APICalls = $VerboseMessages -like "*Sending request*"
        $APICalls | Should -HaveCount 0

        # Passing an object
        $VerboseMessages = $( $Results = $Device | Update-MyObject -Force -Verbose ) 4>&1
        $Results.id | Should -Be $Device.id
        [array] $APICalls = $VerboseMessages -like "*Sending request*"
        $APICalls | Should -HaveCount 0
    }
    
    It "Expand device called from a function using expand object" {
        Function Update-MyObject {
            [cmdletbinding()]
            Param(
                [Parameter(
                    Mandatory = $true,
                    Position = 0,
                    ValueFromPipeline = $true,
                    ValueFromPipelineByPropertyName = $true
                )]
                [object[]] $Device,
                [switch] $Force
            )

            foreach ($iDevice in (PSc8y\Expand-Device $Device -Fetch)) {
                $iDevice
            }
        }
        # Passing id to an object
        $VerboseMessages = $( $Results = $Device.id | Update-MyObject -Verbose ) 4>&1
        $Results.id | Should -Be $Device.id
        $Results.name | Should -Be $Device.name
        [array] $APICalls = $VerboseMessages -like "*Sending request*"
        $APICalls | Should -HaveCount 1

        # Passing an object (no fetch should be done)
        $VerboseMessages = $( $Results = $Device | Update-MyObject -Force -Verbose ) 4>&1
        $Results.id | Should -Be $Device.id
        $Results.name | Should -Be $Device.name
        [array] $APICalls = $VerboseMessages -like "*Sending request*"
        $APICalls | Should -HaveCount 0
    }

    AfterAll {
        Remove-ManagedObject -Id $Device.id
    }
}
