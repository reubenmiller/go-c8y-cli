{    
    // Measurement (other fields will be added)
    c8y_Weather: {
        temperature: {
            value: _.Int(50,-20),
            unit: "°C",
        },
        barometricPressure: {
            value: _.Float(1000, 1100, 2),
            unit: "Pa",
        },
    },
    type: "c8y_Weather",
}