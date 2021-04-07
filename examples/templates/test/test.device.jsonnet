# Name: Create randomized test device
{
    "name": var("prefix", "testdevice") + _.Char(8),
}