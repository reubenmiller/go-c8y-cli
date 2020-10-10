local severity(idx, len=4) = ["MAJOR", "CRITICAL", "MINOR", "WARNING"][std.clamp(idx, 0, len-1)];
{    
    // Measurement (other fields will be added)
    severity: severity(rand.int % 4),
}