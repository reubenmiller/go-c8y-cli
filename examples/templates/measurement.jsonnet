{    
    // Measurement (other fields will be added)
    c8y_Weather: {
        temperature: {
            value: rand.int,
            unit: "°C",
        },
        barometricPressure: {
            value: rand.float * 100 + 1000,
            value2: if rand.bool then -20 else rand.int * 10,
            value3: rand.int % 5,
            unit: "Pa",
        },
    },
}