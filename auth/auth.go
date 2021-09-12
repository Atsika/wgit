package auth

import (
	"github.com/AlecAivazis/survey/v2"

	SSH "golang.org/x/crypto/ssh"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

func Authenticate(method string, username string, token string, SSHkeyPath string) transport.AuthMethod {
	var auth transport.AuthMethod

	switch method {
	case "SSH":
		auth = SshAuthGit(SSHkeyPath)
	case "HTTP":
		auth = HttpAuthGit(username)
	case "Access Token":
		auth = TokenAuthGit(username, token)
	default:
		auth = HttpAuthGit(username)
	}

	return auth
}

func HttpAuthGit(username string) *http.BasicAuth {

	if username == "" {
		prompt := &survey.Input{
			Message: "Please enter your account username:",
		}
		survey.AskOne(prompt, &username)
	}

	password := ""
	prompt := &survey.Password{
		Message: "Please enter your account password:",
	}
	survey.AskOne(prompt, &password)

	auth := &http.BasicAuth{
		Username: username,
		Password: password,
	}

	return auth
}

func SshAuthGit(SSHkeyPath string) *ssh.PublicKeys {
	password := ""
	prompt := &survey.Password{
		Message: "Please enter your SSH key password:",
	}
	survey.AskOne(prompt, &password)

	if SSHkeyPath == "" {
		prompt := &survey.Input{
			Message: "Please enter your ssh private key path:",
		}
		survey.AskOne(prompt, &SSHkeyPath)
	}

	publicKeys, err := ssh.NewPublicKeysFromFile("git", SSHkeyPath, password)
	if err != nil {
		panic(err)
	}

	publicKeys.HostKeyCallback = SSH.InsecureIgnoreHostKey()
	return publicKeys
}

func TokenAuthGit(username string, token string) *http.BasicAuth {

	if username == "" {
		username = "git"
	}

	if token == "" {
		prompt := &survey.Password{
			Message: "Please enter your access token:",
		}
		survey.AskOne(prompt, &token)
	}

	auth := &http.BasicAuth{
		Username: username,
		Password: token,
	}

	return auth
}
