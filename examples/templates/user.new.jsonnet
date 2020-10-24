{
    userName: if (std.objectHas(base, "userName")) then
            base.userName
        else
            self.email,
    email: std.asciiLower(base.email),
    password: rand.password,
}