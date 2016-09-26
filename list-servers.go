package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/urfave/cli"
)

type Server struct {
	Name        string `json:"name"`
	Environment string `json:"environment"`
	Tags        Tags   `json:"tags"`
}

type Servers []Server

type Tags []string

func getServers(configFile string) Servers {
	raw, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var servers []Server
	json.Unmarshal(raw, &servers)
	return servers
}

func filterByEnvironment(servers Servers, environment string) Servers {
	filtered := servers[:0]
	for _, server := range servers {
		if server.Environment == environment {
			filtered = append(filtered, server)
		}
	}
	return filtered
}

func contains(tags []string, expectedTag string) bool {
	for _, tag := range tags {
		if tag == expectedTag {
			return true
		}
	}

	return false
}

func filterByTag(servers Servers, tag string) Servers {
	filtered := servers[:0]
	for _, server := range servers {
		if strings.HasPrefix(tag, "!") {
			if !contains(server.Tags, strings.TrimPrefix(tag, "!")) {
				filtered = append(filtered, server)
			}
		} else {
			if contains(server.Tags, tag) {
				filtered = append(filtered, server)
			}
		}
	}
	return filtered
}

func listNamesOutput(servers Servers) {
	var buffer bytes.Buffer
	for index, server := range servers {
		buffer.WriteString(server.Name)
		if index < (len(servers) - 1) {
			buffer.WriteString(",")
		}
	}
	fmt.Println(buffer.String())
}

func printJSONOutput(servers Servers) {
	serversJSON, err := json.MarshalIndent(servers, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(serversJSON))
}

func columnarOutput(servers Servers) {
	for _, server := range servers {
		fmt.Printf("%s %s %s\n", server.Environment, server.Name, server.Tags)
	}
}

func formatList(servers Servers, format string) {
	if format == "names" {
		listNamesOutput(servers)
	} else if format == "json" {
		printJSONOutput(servers)
	} else {
		columnarOutput(servers)
	}
}

func filterServers(servers Servers, environment string, tags []string) Servers {
	if environment != "" {
		servers = filterByEnvironment(servers, environment)
	}
	if tags != nil && len(tags) != 0 {
		for _, tag := range tags {
			for _, splitTag := range strings.Split(tag, ",") {
				servers = filterByTag(servers, splitTag)
			}
		}
	}
	return servers
}

func main() {
	var configFile string
	var environment string
	var tags []string
	var format string

	app := cli.NewApp()
	app.Usage = "List and filter servers"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config, c",
			Value:       os.Getenv("HOME") + "/.dcr/servers.json",
			Usage:       "Load configuration from `FILE`",
			Destination: &configFile,
		},
		cli.StringFlag{
			Name:        "env, e",
			Usage:       "Filter by environment",
			Destination: &environment,
		},
		cli.StringSliceFlag{
			Name:  "tags, t",
			Usage: "Filter by tags",
		},
		cli.StringFlag{
			Name:        "format, f",
			Usage:       "Output format",
			Destination: &format,
		},
	}
	app.Action = func(c *cli.Context) error {
		tags = c.StringSlice("tags")
		servers := getServers(configFile)
		servers = filterServers(servers, environment, tags)
		formatList(servers, format)
		fmt.Println("")
		return nil
	}

	app.Run(os.Args)
}
