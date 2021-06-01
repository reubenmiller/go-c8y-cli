# Name: Create randomized test device
{
    "name": var("prefix", "testsmartgroup") + _.Char(8),
    "c8y_DeviceQueryString": ""
}