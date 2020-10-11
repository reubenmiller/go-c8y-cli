{    
    // Measurement (other fields will be added)
    c8y_Weather: {
        temperature: {
            value: rand.int,
            unit: "°C",
        },
        barometricPressure: {
            value: rand.float * 100 + 1000,
            unit: "Pa",
        },
    },
    type: "c8y_Weather",
}