package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/homedir"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/docker/registry"
)

// CmdLogin logs in or registers a user to a Docker registry service.
//
// If no server is specified, the user will be logged into or registered to the registry's index server.
//
// Usage: docker login SERVER
func (cli *DockerCli) CmdLogin(args ...string) error {
	cmd := cli.Subcmd("login", "[SERVER]", "Register or log in to a Docker registry server, if no server is\nspecified \""+registry.IndexServerAddress()+"\" is the default.", true)
	cmd.Require(flag.Max, 1)

	var username, password, email string

	cmd.StringVar(&username, []string{"u", "-username"}, "", "Username")
	cmd.StringVar(&password, []string{"p", "-password"}, "", "Password")
	cmd.StringVar(&email, []string{"e", "-email"}, "", "Email")

	cmd.ParseFlags(args, true)

	serverAddress := registry.IndexServerAddress()
	if len(cmd.Args()) > 0 {
		serverAddress = cmd.Arg(0)
	}

	promptDefault := func(prompt string, configDefault string) {
		if configDefault == "" {
			fmt.Fprintf(cli.out, "%s: ", prompt)
		} else {
			fmt.Fprintf(cli.out, "%s (%s): ", prompt, configDefault)
		}
	}

	readInput := func(in io.Reader, out io.Writer) string {
		reader := bufio.NewReader(in)
		line, _, err := reader.ReadLine()
		if err != nil {
			fmt.Fprintln(out, err.Error())
			os.Exit(1)
		}
		return string(line)
	}

	cli.LoadConfigFile()
	authconfig, ok := cli.configFile.Configs[serverAddress]
	if !ok {
		authconfig = registry.AuthConfig{}
	}

	if username == "" {
		promptDefault("Username", authconfig.Username)
		username = readInput(cli.in, cli.out)
		username = strings.Trim(username, " ")
		if username == "" {
			username = authconfig.Username
		}
	}
	// Assume that a different username means they may not want to use
	// the password or email from the config file, so prompt them
	if username != authconfig.Username {
		if password == "" {
			oldState, err := term.SaveState(cli.inFd)
			if err != nil {
				return err
			}
			fmt.Fprintf(cli.out, "Password: ")
			term.DisableEcho(cli.inFd, oldState)

			password = readInput(cli.in, cli.out)
			fmt.Fprint(cli.out, "\n")

			term.RestoreTerminal(cli.inFd, oldState)
			if password == "" {
				return fmt.Errorf("Error : Password Required")
			}
		}

		if email == "" {
			promptDefault("Email", authconfig.Email)
			email = readInput(cli.in, cli.out)
			if email == "" {
				email = authconfig.Email
			}
		}
	} else {
		// However, if they don't override the username use the
		// password or email from the cmd line if specified. IOW, allow
		// then to change/override them.  And if not specified, just
		// use what's in the config file
		if password == "" {
			password = authconfig.Password
		}
		if email == "" {
			email = authconfig.Email
		}
	}
	authconfig.Username = username
	authconfig.Password = password
	authconfig.Email = email
	authconfig.ServerAddress = serverAddress
	cli.configFile.Configs[serverAddress] = authconfig

	stream, statusCode, err := cli.call("POST", "/auth", cli.configFile.Configs[serverAddress], nil)
	if statusCode == 401 {
		delete(cli.configFile.Configs, serverAddress)
		registry.SaveConfig(cli.configFile)
		return err
	}
	if err != nil {
		return err
	}

	var response types.AuthResponse
	if err := json.NewDecoder(stream).Decode(&response); err != nil {
		cli.configFile, _ = registry.LoadConfig(homedir.Get())
		return err
	}

	registry.SaveConfig(cli.configFile)
	fmt.Fprintf(cli.out, "WARNING: login credentials saved in %s.\n", path.Join(homedir.Get(), registry.CONFIGFILE))

	if response.Status != "" {
		fmt.Fprintf(cli.out, "%s\n", response.Status)
	}
	return nil
}
