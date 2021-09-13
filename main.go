package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"wgit/auth"
	"wgit/config"
	"wgit/utils"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
)

type Directory struct {
	Parent  int
	Index   int
	Name    string
	Content []string
}

var (
	Back = "../"
)

func banner() {
	wgit := `
                            $$\   $$\     
                            \__|  $$ |    
    $$\  $$\  $$\  $$$$$$\  $$\ $$$$$$\   
    $$ | $$ | $$ |$$  __$$\ $$ |\_$$  _|  
    $$ | $$ | $$ |$$ /  $$ |$$ |  $$ |    
    $$ | $$ | $$ |$$ |  $$ |$$ |  $$ |$$\ 
    \$$$$$\$$$$  |\$$$$$$$ |$$ |  \$$$$  |
     \_____\____/  \____$$ |\__|   \____/ 
                  $$\   $$ |              
                  \$$$$$$  |              
                   \______/               
    

            made with ❤️  by Atsika


`

	fmt.Print(wgit)
}

func Walk(bfs billy.Filesystem, directory string, parent int, tree *[]Directory) {
	files, err := bfs.ReadDir(directory)
	if err != nil {
		panic(err)
	}

	var d = Directory{
		Parent: parent,
		Index:  len(*tree),
		Name:   directory,
	}

	(*tree) = append((*tree), d)

	for i := range files {
		fullPath := ""
		if directory == "/" {
			fullPath = directory + files[i].Name()
		} else {
			fullPath = directory + "/" + files[i].Name()
		}
		base := filepath.Base(fullPath)
		if files[i].IsDir() {
			(*tree)[d.Index].Content = append((*tree)[d.Index].Content, base+"/")
			Walk(bfs, fullPath, d.Index, tree)
		} else {
			(*tree)[d.Index].Content = append((*tree)[d.Index].Content, base)
		}
	}
}

func GetFullPath(element string, tree []Directory, index int) string {
	fullpath := ""
	if index != 0 {
		fullpath = tree[index].Name + "/" + element
	} else {
		fullpath = tree[index].Name + element
	}
	return fullpath
}

func AddSelectedFile(f []string, s string) []string {
	f = append(f, s)
	return f
}

func DelSelectedFile(f []string, s string) []string {
	i := utils.Find(f, s)
	if i != -1 {
		f = append(f[:i], f[i+1:]...)
	}
	return f
}

func FindDirectory(tree []Directory, s string) int {
	for i := range tree {
		if tree[i].Name == s {
			return tree[i].Index
		}
	}

	return -1
}

func ListFiles(files []string) {
	for i := range files {
		base := filepath.Base(files[i])
		fmt.Println("•", base)
	}
	if len(files) == 0 {
		fmt.Println("No files selected")
	}
	utils.Pause()
}

func WriteFiles(bfs billy.Filesystem, files []string) {
	downloaded := 0
	for i := range files {
		file, err := bfs.Open(files[i])
		if err != nil {
			panic(err)
		}
		content, _ := ioutil.ReadAll(file)
		base := filepath.Base(file.Name())
		if _, err := os.Stat(base); os.IsNotExist(err) {
			ioutil.WriteFile(base, content, os.ModePerm)
			downloaded++
		} else {
			overwrite := false
			prompt := &survey.Confirm{
				Message: "File '" + base + "' already exists. Do you want to overwrite it ?",
				Default: false,
			}
			survey.AskOne(prompt, &overwrite)
			if overwrite {
				ioutil.WriteFile(base, content, os.ModePerm)
				downloaded++
			}
		}
	}
	if downloaded > 0 {
		fmt.Println("⬇️  Files downloaded successfully")
	} else if len(files) == 0 {
		fmt.Println("No files selected")
	} else if downloaded == 0 {
		fmt.Println("No files downloaded")
	}
	utils.Pause()
}

func Selector(tree []Directory, files []string, bfs billy.Filesystem) {
	utils.Flush()
	i := 0
	for {
		selection := ""
		if i != 0 && !utils.Contains(tree[i].Content, Back) {
			tree[i].Content = utils.Prepend(tree[i].Content, Back)
		}
		prompt := &survey.Select{
			Message: "Current directory is: " + tree[i].Name + "\nSelect desired files:",
			Options: tree[i].Content,
		}
		ret := survey.AskOne(prompt, &selection, survey.WithIcons(func(icons *survey.IconSet) {
			icons.Question.Text = "📍"
		}))
		if ret != nil {
			if ret == terminal.InterruptErr {
				HandleInterrupt(files, bfs)
				continue
			}
		}

		if selection == Back && i > 0 {
			i = tree[i].Parent
			utils.Flush()
			continue
		}

		selectionPath := GetFullPath(selection, tree, i)
		if strings.HasSuffix(selection, "/") {
			tmp := FindDirectory(tree, selectionPath[:len(selectionPath)-1])
			if tmp != -1 {
				i = tmp
			}
			utils.Flush()
			continue
		}

		if utils.Contains(files, selectionPath) {
			files = DelSelectedFile(files, selectionPath)
			utils.Flush()
			fmt.Println("🚮 File '" + selection + "' removed")
		} else {
			files = AddSelectedFile(files, selectionPath)
			utils.Flush()
			fmt.Println("✅ File '" + selection + "' added")
		}
	}
}

func HandleInterrupt(files []string, bfs billy.Filesystem) {
	utils.Flush()
	action := ""
	prompt := &survey.Select{
		Message: "Select an action:",
		Options: []string{"Back", "List selected files", "Get selected files", "Exit"},
	}
	survey.AskOne(prompt, &action)

	switch action {
	case "Back":
		break
	case "List selected files":
		ListFiles(files)
	case "Get selected files":
		WriteFiles(bfs, files)
	case "Exit":
		os.Exit(0)
	}
	utils.Flush()
}

func main() {

	utils.Flush()
	banner()

	var configuration config.Configuration

	configuration, err := config.Get()
	if err != nil {
		fmt.Println(err)
		configuration = config.CreateNew()
		config.Write(configuration)
	}

	config.Display(configuration)

	if !config.Keep() {
		oldConfiguration := configuration
		fmt.Println("⚠️  Empty fields won't be overwritten")
		configuration = config.CreateNew()
		configuration = config.Update(oldConfiguration, configuration)
		if config.Save() {
			config.Write(configuration)
		}
	}

	storer := memory.NewStorage()
	bfs := memfs.New()
	authentication := auth.Authenticate(configuration.AuthMethod, configuration.Username, configuration.Token, configuration.SSHkeyPath)

	_, err = git.Clone(storer, bfs, &git.CloneOptions{
		URL:  configuration.Repository,
		Auth: authentication,
	})
	if err != nil {
		panic(err)
	}

	tree := make([]Directory, 0)
	Walk(bfs, "/", -1, &tree)

	files := make([]string, 0)
	Selector(tree, files, bfs)
}
