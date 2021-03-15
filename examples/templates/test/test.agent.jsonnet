# Name: Create randomized test agent
{
    "name": "%s_%03d" % [
        var("prefix", "testdevice"),
        rand.int,
    ],
}