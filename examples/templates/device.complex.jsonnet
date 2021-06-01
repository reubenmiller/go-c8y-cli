local newSoftware(i) = {
    name: "app" + i,
    version: "" + i + ".0.0",
    url: "https://myexample.com/packages/" + self.name + "/" + i + ".0.0",
};

local randomType() = ["c8y_Linux", "c8y_MacOS", "c8y_Windows"][_.Int(3)];
local randomCountry() = [{name:"DE",code:49}, {name:"AU",code:"61"}][_.Int(2)];

local paddedIndex = std.format("%03d", input.index);
{
    name: "device" + paddedIndex,
    [var("fragment", "company_Example")]: {},
    type: var("type", randomType()),
    c8y_Hardware: {
        serialNumber: 'XYDA' + paddedIndex,
    },
    agent_Details: {
        country: randomCountry(),
        details: [1, 2, 3],
    },
    c8y_SoftwareList: [newSoftware(i) for i in std.range(1, var("softwareCount", 2))],
}