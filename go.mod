module github.com/fnproject/cli

require (
	github.com/coreos/go-semver v0.3.0
	github.com/fatih/color v0.0.0-20170926111411-5df930a27be2
	github.com/fnproject/fn_go v0.8.6
	github.com/ghodss/yaml v1.0.0
	github.com/giantswarm/semver-bump v0.0.0-20140912095342-88e6c9f2fe39
	github.com/go-openapi/runtime v0.19.23
	github.com/jmoiron/jsonq v0.0.0-20150511023944-e874b168d07e
	github.com/juju/errgo v0.0.0-20140925100237-08cceb5d0b53 // indirect
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.3
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.3.2
	github.com/oracle/oci-go-sdk/v48 v48.0.0
	github.com/spf13/viper v1.6.2
	github.com/urfave/cli v1.20.0
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v0.0.0-20180618132009-1d523034197f
	golang.org/x/sys v0.0.0-20220804214406-8e32c043e418
	gopkg.in/yaml.v2 v2.3.0
)

replace (
	github.com/fnproject/fn_go v0.8.6 => /Users/sunny/Functions/fn_go
)

go 1.14
