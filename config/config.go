package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"text/tabwriter"

	"github.com/AlecAivazis/survey/v2"
)

var (
	errConfigNotFound = errors.New("config file not found")
)

type Configuration struct {
	Username   string
	Token      string
	Repository string
	AuthMethod string
	SSHkeyPath string
}

func Get() (Configuration, error) {

	var data Configuration

	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	configPath := usr.HomeDir + "/" + ".config/wgit/config.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return data, errConfigNotFound
	}

	content, _ := ioutil.ReadFile(configPath)
	if err := json.Unmarshal(content, &data); err != nil {
		panic(err)
	}
	return data, nil
}

func getConfigPath() string {

	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return usr.HomeDir + "/" + ".config/wgit/config.json"
}

func CreateNew() Configuration {

	fmt.Println("Answer next few questions to setup config file")
	var qs = []*survey.Question{
		{
			Name:   "username",
			Prompt: &survey.Input{Message: "Please enter your account username:"},
		},
		{
			Name:   "repository",
			Prompt: &survey.Input{Message: "Please enter repository:"},
		},
		{
			Name: "authmethod",
			Prompt: &survey.Select{
				Message: "Please select your favorite authentication method:",
				Options: []string{"SSH", "HTTP", "Access Token"},
			},
		},
		{
			Name:   "token",
			Prompt: &survey.Input{Message: "Please enter your access token:"},
		},
		{
			Name: "sshkeypath",
			Prompt: &survey.Input{
				Message: "Please enter your ssh private key full path:",
				Suggest: func(toComplete string) []string {
					files, _ := filepath.Glob(toComplete + "*")
					return files
				}},
		},
	}

	var data Configuration
	survey.Ask(qs, &data)

	return data
}

func Write(data Configuration) {

	config, _ := json.Marshal(data)
	configPath := getConfigPath()

	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	os.Mkdir(usr.HomeDir+"/.config", os.ModePerm)
	os.Mkdir(usr.HomeDir+"/.config/wgit", os.ModePerm)
	ioutil.WriteFile(configPath, config, os.ModePerm)

}

func Keep() bool {
	keep := true
	prompt := &survey.Confirm{
		Message: "Do you want to use current configuration ?",
		Default: true,
	}
	survey.AskOne(prompt, &keep)
	return keep
}

func Save() bool {
	save := true
	prompt := &survey.Confirm{
		Message: "Do you want to save new configuration ? (empty fields won't be overwritten)",
		Default: true,
	}
	survey.AskOne(prompt, &save)
	return save
}

func Update(oldConfiguration Configuration, newConfiguration Configuration) Configuration {
	if newConfiguration.Username == "" {
		newConfiguration.Username = oldConfiguration.Username
	}

	if newConfiguration.Token == "" {
		newConfiguration.Token = oldConfiguration.Token
	}

	if newConfiguration.AuthMethod == "" {
		newConfiguration.AuthMethod = oldConfiguration.AuthMethod
	}

	if newConfiguration.Repository == "" {
		newConfiguration.Repository = oldConfiguration.Repository
	}

	if newConfiguration.SSHkeyPath == "" {
		newConfiguration.SSHkeyPath = oldConfiguration.SSHkeyPath
	}

	return newConfiguration
}

func Display(config Configuration) {

	writer := tabwriter.NewWriter(os.Stdout, 0, 20, 0, ' ', tabwriter.TabIndent)

	fmt.Println("Current configuration")
	fmt.Fprintln(writer, "Username\t :\t  "+config.Username)
	fmt.Fprintln(writer, "Repository\t :\t  "+config.Repository)
	fmt.Fprintln(writer, "Access Token\t :\t  "+config.Token)
	fmt.Fprintln(writer, "Authentication method\t :\t  "+config.AuthMethod)
	fmt.Fprintln(writer, "SSH private key path\t :\t "+config.SSHkeyPath)

	writer.Flush()
	fmt.Println()
}
