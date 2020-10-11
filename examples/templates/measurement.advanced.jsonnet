{
    name: "name1",
    type: var("type", "c8y_Temparature"),
    
    ["c8y_" + var("type", "c8y_Temperature")]: {
        sensor1: {
            value: rand.int,
            unit: "°C",
        },
        barometricPressure: {
            value: rand.float * 100 + 1000,
            unit: "Pa",
        },
    },
}