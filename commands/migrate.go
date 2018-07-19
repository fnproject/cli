package commands

import (
	"errors"
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
	newFF *common.FuncFileV20180707
}

func MigrateCommand() cli.Command {
	m := &migrateFnCmd{newFF: &common.FuncFileV20180707{}}

	return cli.Command{
		Name:        "migrate",
		Usage:       "Migrate a local func.yaml file to the latest version",
		Category:    "DEVELOPMENT COMMANDS",
		Aliases:     []string{"m"},
		Description: "This command will detect the version of a func.yaml file and update it to match the latest version supported by the Fn CLI. Any old or unsupported attributes will be removed, and any new ones may be added. The current func.yaml will be renamed to func.yaml.bak and a new func.yaml created",
		Action:      m.migrate,
	}
}

func (m *migrateFnCmd) migrate(c *cli.Context) error {
	var err error
	oldFF, err := common.ReadInFuncFile()
	if err != nil {
		return err
	}

	version := common.GetFuncYamlVersion(oldFF)
	if version == common.LatestYamlVersion {
		return errors.New("you have an up to date func.yaml file and do not need to migrate.")
	}

	err = backUpYamlFile(oldFF)
	if err != nil {
		return err
	}

	b, err := m.creatFuncFileBytes(oldFF)
	if err != nil {
		return err
	}

	err = vaidateFuncFileSchema(b)
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

func backUpYamlFile(ff map[string]interface{}) error {
	b, err := yaml.Marshal(ff)
	if err != nil {
		return err
	}

	return writeYamlFile(b, "func.yaml.bak")
}

func (m *migrateFnCmd) decodeFuncFile(oldFF map[string]interface{}) ([]byte, error) {
	_ = mapstructure.Decode(oldFF, &m.newFF)
	return yaml.Marshal(oldFF)
}

func (m *migrateFnCmd) creatFuncFileBytes(oldFF map[string]interface{}) ([]byte, error) {
	b, err := m.decodeFuncFile(oldFF)
	if err != nil {
		return nil, err
	}

	err = vaidateFuncFileSchema(b)
	if err != nil {
		return nil, err
	}

	m.newFF.Schema_version = 20180708
	trig := make([]common.Trigger, 1)

	var trigName, trigSource string

	if oldFF["name"] != nil {
		trigName = oldFF["name"].(string)
		trigSource = "/" + oldFF["name"].(string)
	}

	trigType := "http"

	trig[0] = common.Trigger{
		trigName,
		trigType,
		trigSource,
	}
	m.newFF.Triggers = trig

	return yaml.Marshal(m.newFF)
}

func vaidateFuncFileSchema(b []byte) error {
	jsonB, err := yamltojson.YAMLToJSON(b)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("temp.json", jsonB, config.ReadWritePerms)
	if err != nil {
		return err
	}
	defer os.Remove("temp.json")

	err = common.ValidateFileAgainstSchema("temp.json", common.V20180707Schema)
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
