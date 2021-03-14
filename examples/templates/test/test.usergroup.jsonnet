# Name: Create randomized test user group
{
    "name": "%s_%03d" % [
        var("prefix", "testgroup"),
        rand.int,
    ]
}