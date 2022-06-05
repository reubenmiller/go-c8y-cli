local item() = {
    "id": 1,
    "first_name": "Doloritas",
    "last_name": "Clow",
    "email": "dclow0@berkeley.edu",
    "favourite_food": "popcorn",
    "ip_address": "249.218.130.49"
};

{
    "data": [item() for i in std.range(1, 100000)]
}