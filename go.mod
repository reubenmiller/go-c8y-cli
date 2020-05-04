module github.com/reubenmiller/go-c8y-cli

require (
	github.com/araddon/dateparse v0.0.0-20190329160016-74dc0e29b01f
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575
	github.com/fatih/color v1.7.0
	github.com/google/go-jsonnet v0.14.0
	github.com/google/go-querystring v1.0.0
	github.com/gorilla/websocket v1.4.0
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c
	github.com/jeremywohl/flatten v1.0.1 // indirect
	github.com/karrick/tparse/v2 v2.7.1
	github.com/labstack/echo v3.3.10+incompatible
	github.com/manifoldco/promptui v0.3.2
	github.com/mattn/go-isatty v0.0.8 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nicksnyder/go-i18n v1.10.0 // indirect
	github.com/obeattie/ohmyglob v0.0.0-20150811221449-290764208a0d
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.3
	github.com/reubenmiller/go-c8y v0.7.1-0.20200308212728-9922a1a9ddcb
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.5.0
	github.com/thedevsaddam/gojsonq v2.3.0+incompatible
	github.com/tidwall/gjson v1.2.1
	github.com/tidwall/pretty v1.0.0
	go.uber.org/atomic v1.4.0
	go.uber.org/multierr v1.1.0
	go.uber.org/zap v1.10.0
	golang.org/x/crypto v0.0.0-20191029031824-8986dd9e96cf // indirect
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f // indirect
	golang.org/x/sys v0.0.0-20190712062909-fae7ac547cb7 // indirect
	golang.org/x/text v0.3.2 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/robfig/cron.v2 v2.0.0-20150107220207-be2e0b0deed5
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637
	moul.io/http2curl v1.0.0
)

replace github.com/manifoldco/promptui => github.com/reubenmiller/promptui v0.3.3-0.20191108135340-17a79c13fae0

go 1.13
