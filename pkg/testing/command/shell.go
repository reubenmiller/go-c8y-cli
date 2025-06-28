package command

import "strings"

func ParseShellEnv(in string) map[string]string {
	env := map[string]string{}

	for _, item := range strings.Split(in, "\n") {
		if strings.HasPrefix(item, "export C8Y_") {
			if k, v, ok := strings.Cut(item, "="); ok {
				k = strings.TrimPrefix(k, "export ")
				if (strings.HasPrefix(v, "'") && strings.HasSuffix(v, "'")) || (strings.HasPrefix(v, "\"") && strings.HasSuffix(v, "\"")) {
					env[k] = v[1 : len(v)-1]
				} else {
					env[k] = v
				}
			}
		}
	}
	return env
}
