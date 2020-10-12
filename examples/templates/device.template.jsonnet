// Helper: Create software entry
local newSoftware(i) = {
    name: "app" + i,
    version: "1.0." + i,
    url: "",
};

// Output: Device Managed Object
{
    name: "name1",
    value: var("name", "defaultName"),
    type: var("type"),
    ["c8y_" + var("type")]: {},
    c8y_SoftwareList: [ mylib.newSoftware(i) for i in std.range(1, var("softwareCount", 1))],
}