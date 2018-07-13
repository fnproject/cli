package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/config"
	yamltojson "github.com/ghodss/yaml"
	"github.com/mitchellh/mapstructure"
	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"
)

type migrateFnCmd struct {
	ff   *common.FuncFile
	ffV2 *common.FuncFileV2
}

func MigrateCommand() cli.Command {
	m := &migrateFnCmd{ff: &common.FuncFile{}, ffV2: &common.FuncFileV2{}}

	return cli.Command{
		Name:        "migrate",
		Usage:       "Migrate a local func.yaml file to the new version",
		Category:    "DEVELOPMENT COMMANDS",
		Aliases:     []string{"m"},
		Description: "Migrate will detect the version of the current func.yaml file, to the new version.",
		Action:      m.migrate,
	}
}

func (m *migrateFnCmd) migrate(c *cli.Context) error {
	return detectFuncYamlVersion()
}

func detectFuncYamlVersion() error {
	wd := common.GetWd()

	fpath, err := common.FindFuncfile(wd)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		return fmt.Errorf("could not open %s for parsing. Error: %v", fpath, err)
	}
	var ff map[string]interface{}
	err = yaml.Unmarshal(b, &ff)

	if _, ok := ff["schema_version"]; !ok {
		return migrateFuncYaml(ff)
	}

	fmt.Println("You have an up to date func file and do not need to migrate.")
	return nil
}

func migrateFuncYaml(ff map[string]interface{}) error {
	b, err := yaml.Marshal(ff)
	if err != nil {
		return err
	}

	err = writeYamlFile(b, "func.yaml.bak")
	if err != nil {
		return err
	}

	return writeNewFile(ff)
}

func writeNewFile(ff map[string]interface{}) error {
	var ffV2 common.FuncFileV2
	mapstructure.Decode(ff, &ffV2)

	b, err := yaml.Marshal(ff)
	if err != nil {
		return err
	}

	err = convertYamlToJson(b)
	if err != nil {
		return err
	}

	ffV2.Schema_version = 20180708
	trig := make([]common.Trigger, 1)
	trig[0] = common.Trigger{
		ff["name"].(string),
		"http",
		"/" + ff["name"].(string),
	}
	ffV2.Triggers = trig

	b, err = yaml.Marshal(ffV2)
	if err != nil {
		return err
	}

	err = convertYamlToJson(b)
	if err != nil {
		return err
	}

	err = writeYamlFile(b, "func.yaml")
	if err != nil {
		return err
	}

	fmt.Println("Successfully migrated func.yaml and created a back up func.yaml.bak")
	return nil
}

func convertYamlToJson(b []byte) error {
	jsonB, err := yamltojson.YAMLToJSON(b)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("temp.json", jsonB, config.ReadWritePerms)
	if err != nil {
		return err
	}
	defer os.Remove("temp.json")

	err = common.ValidateSchema("temp.json")
	if err != nil {
		return err
	}

	return nil
}

func writeYamlFile(b []byte, filename string) error {
	wd := common.GetWd()
	fpath := filepath.Join(wd, filename)

	return ioutil.WriteFile(fpath, b, config.ReadWritePerms)
}
