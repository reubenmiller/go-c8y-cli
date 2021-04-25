local severity(idx, len=4) = ["MAJOR", "CRITICAL", "MINOR", "WARNING"][std.clamp(idx, 0, len-1)];
local text() = ["Too hot", "Too cold", "Disconnected", "Unknown error"][_.Int(4)];
local type() = ["c8y_temperature", "c8y_sensor"][_.Int(2)];

{    
    // Measurement (other fields will be added)
    severity: severity(rand.int % 4),
    text: text(),
    type: type() + (rand.int % 4),
}