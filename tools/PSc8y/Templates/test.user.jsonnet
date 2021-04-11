# Name: Create randomized test user
{
    "userName": var("prefix", "testuser") + _.Char(8),
    "email": std.asciiLower(self.userName + var("domain", "@example.c8y.com")),
    "password": rand.password
}