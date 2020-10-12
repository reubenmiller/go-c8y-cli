// Device
local randomType() = ["c8y_Linux", "c8y_MacOS", "c8y_Windows"][rand.int % 3];

{
    name: var("name", "defaultName"),
    type: randomType(),
    c8y_IsDevice: {},
}