# Name: Create randomized test device
{
    "name": "%s_%03d" % [
        var("prefix", "testdevice"),
        rand.int,
    ],
}