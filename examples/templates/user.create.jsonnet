// Description: Create user where the username is the email address (in lowercase)
local domain = var("domain", "@noreply.example");
local email = std.asciiLower(var('email', if std.type(input.value) == 'string' then input.value else input.value.username));
{
    email: if std.length(std.findSubstr("@", email)) > 0 then email else email + domain,
    userName+: self.email,
}