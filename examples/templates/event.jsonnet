local selectType(idx, len) = ["DiskUsage", "RAM", "Network"][std.clamp(idx, 0, len-1)];
{    
    // Measurement (other fields will be added)
    type: "c8y_" + selectType(rand.int % 3, 3),
}