// Helper: Create software entry
local newSoftware(i) = {
    name: "app" + i,
    version: "1.0." + i,
    url: "",
};

// Settings: Default values:
local defaults = {
    name: "test",
    type: "defaultType",
    softwareCount: 2,
} + vars;

// Output: Device Managed Object
{
    name: "name1",
    value: defaults.name,
    type: defaults.type,
    ["c8y_" + defaults.type]: {},
    c8y_SoftwareList: [ newSoftware(i) for i in std.range(1, defaults.softwareCount)],
}