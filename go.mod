module github.com/reubenmiller/go-c8y-cli

require (
	github.com/araddon/dateparse v0.0.0-20190329160016-74dc0e29b01f
	github.com/fatih/color v1.9.0
	github.com/google/go-jsonnet v0.16.0
	github.com/gorilla/websocket v1.4.2
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c
	github.com/karrick/tparse/v2 v2.8.2
	github.com/manifoldco/promptui v0.3.2
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nicksnyder/go-i18n v1.10.0 // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pkg/errors v0.9.1
	github.com/reubenmiller/go-c8y v0.8.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/thedevsaddam/gojsonq v2.3.0+incompatible
	github.com/tidwall/gjson v1.2.1
	github.com/tidwall/pretty v1.0.0
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f // indirect
)

replace github.com/manifoldco/promptui => github.com/reubenmiller/promptui v0.3.3-0.20191108135340-17a79c13fae0

go 1.13
