# Name: Create randomized test user
{
    "userName": "%s_%03d" % [
        var("prefix", "testuser"),
        rand.int,
    ],
    "email": std.asciiLower(self.userName + var("domain", "@example.c8y.com")),
    "password": rand.password
}