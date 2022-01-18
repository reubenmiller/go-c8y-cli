module github.com/reubenmiller/go-c8y-cli

require (
	github.com/AlecAivazis/survey/v2 v2.3.2
	github.com/MakeNowJust/heredoc/v2 v2.0.1
	github.com/araddon/dateparse v0.0.0-20210429162001-6b43995a97de
	github.com/charmbracelet/glamour v0.4.0
	github.com/cli/safeexec v1.0.0
	github.com/cpuguy83/go-md2man/v2 v2.0.1
	github.com/fatih/color v1.13.0
	github.com/google/go-jsonnet v0.18.0
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/gorilla/websocket v1.4.2
	github.com/howeyc/gopass v0.0.0-20210920133722-c8aef6fb66ef
	github.com/karrick/tparse/v2 v2.8.2
	github.com/manifoldco/promptui v0.9.0
	github.com/mattn/go-colorable v0.1.12
	github.com/mattn/go-isatty v0.0.14
	github.com/mdp/qrterminal v1.0.1
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/muesli/termenv v0.9.0
	github.com/obeattie/ohmyglob v0.0.0-20150811221449-290764208a0d
	github.com/olekukonko/tablewriter v0.0.5
	github.com/olekukonko/ts v0.0.0-20171002115256-78ecb04241c0
	github.com/pkg/errors v0.9.1
	github.com/pquerna/otp v1.3.0
	github.com/reubenmiller/go-c8y v0.9.1-rc.6
	github.com/santhosh-tekuri/jsonschema/v5 v5.0.0
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/sethvargo/go-password v0.2.0
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.10.1
	github.com/stretchr/testify v1.7.0
	github.com/thedevsaddam/gojsonq v2.3.0+incompatible
	github.com/tidwall/gjson v1.13.0
	github.com/tidwall/pretty v1.2.0
	github.com/tidwall/sjson v1.2.4
	github.com/vbauerster/mpb/v6 v6.0.4
	go.uber.org/zap v1.20.0
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	moul.io/http2curl v1.0.0
)

replace github.com/commander-cli/commander/v2 => github.com/reubenmiller/commander/v2 v2.5.0-alpha3

go 1.16
