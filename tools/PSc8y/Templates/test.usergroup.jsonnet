# Name: Create randomized test user group
{
    "name": var("prefix", "testgroup") + _.Char(8),
}