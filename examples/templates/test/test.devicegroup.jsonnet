{
    name: "%s_%03d" % [
        var("prefix", "testdevice"),
        rand.int,
    ],
    type: "c8y_DeviceGroup",
}