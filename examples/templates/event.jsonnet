local selectType(idx) = ["DiskUsage", "RAM", "Network"][std.clamp(idx, 0, 2)];
{    
    // Measurement (other fields will be added)
    type: "c8y_" + selectType(_.Int(3)),
}