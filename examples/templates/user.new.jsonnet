{
    userName: self.email,
    email: std.asciiLower(var('email', '')),
    password: rand.password,
}