module github.com/reubenmiller/go-c8y-cli

require (
	github.com/araddon/dateparse v0.0.0-20190329160016-74dc0e29b01f
	github.com/fatih/color v1.7.0
	github.com/google/go-jsonnet v0.14.0
	github.com/gorilla/websocket v1.4.0
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c
	github.com/karrick/tparse/v2 v2.7.1
	github.com/manifoldco/promptui v0.3.2
	github.com/mitchellh/go-homedir v1.1.0
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pkg/errors v0.8.1
	github.com/reubenmiller/go-c8y v0.7.1-0.20200308212728-9922a1a9ddcb
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.5.0
	github.com/thedevsaddam/gojsonq v2.3.0+incompatible
	github.com/tidwall/gjson v1.2.1
	github.com/tidwall/pretty v1.0.0
)

replace github.com/manifoldco/promptui => github.com/reubenmiller/promptui v0.3.3-0.20191108135340-17a79c13fae0

go 1.13
