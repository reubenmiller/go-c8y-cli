// Description: Create custom operation
{
    deviceId: "1234", // Dummy value (this will be overwritten when using with c8y operations create)
    description: "Executing custom operation: " + $.com_CustomOperation.param1,
    com_CustomOperation: {
        param1: var("param1", "do_something"),
    }
}