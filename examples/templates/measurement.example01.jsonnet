local seriesPrefix = "series";
local fragmentPrefix = "fragment";

local createSeries() = {
    [seriesPrefix + x]: {value: _.Int(0,32), unit: '°C'},
    for x in std.range(1, 2)
};

local createFragments() = {
  [fragmentPrefix + x]: createSeries(),
  for x in std.range(1,2)
};

{
    type: 'c8y_Temperature',
    
} + createFragments()
