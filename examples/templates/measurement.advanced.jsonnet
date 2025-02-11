{
    name: "name1",
    type: var("type", "c8y_Temperature"),
    
    ["c8y_" + var("type", "c8y_Temperature")]: {
        sensor1: {
            value: _.Int(70,-50),
            unit: "°C",
        },
        barometricPressure: {
            value: _.Float(1100,1000,3),
            unit: "Pa",
        },
    },
}