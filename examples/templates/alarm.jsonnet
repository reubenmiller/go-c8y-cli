local severity(idx) = ["MAJOR", "CRITICAL", "MINOR", "WARNING"][std.clamp(idx, 0, 3)];
{    
    // Measurement (other fields will be added)
    severity: severity(rand.int % 5),
}