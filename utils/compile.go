package main

import (
	config "./config"

	"emd/log"
	"io/ioutil"
	"os"
	"text/template"
)

// Helper func for template string equality check
func eq(s1, s2 string) bool {
	return s1 == s2
}

// Helper func to assign external ports correctly
var externalPorts map[string]int
var currentExtPort int

func getPort(alias string) int {
	// Check if alias is already assigned
	for k, v := range externalPorts {
		if k == alias {
			return v
		}
	}

	// Add a new alias entry external ports map
	externalPorts[alias] = currentExtPort
	currentExtPort += 1
	return externalPorts[alias]
}

/*
 *
 * Create node specific leader files.
 *
 */
func createLeader(lPath string, node config.NodeConfig, guiPort, cPath string) error {
	type tType struct {
		Node    config.NodeConfig
		GuiPort string
		ConfigPath string
	}

	tmpl := template.New("leader.template")
	tmpl.Funcs(template.FuncMap{"eq": eq, "getPort": getPort})
	tmpl, err := tmpl.ParseFiles(lPath + "/leader.template")
	if err != nil {
		return err
	}

	f, err := os.Create(lPath + "/" + node.Hostname + ".go")
	if err != nil {
		return err
	}
	defer f.Close()

	err = tmpl.Execute(f, tType{node, guiPort, cPath})
	if err != nil {
		return err
	}

	return nil
}

func main() {
	log.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	// Read through config.json to understand
	//   how to create bin/leader go files
	//   for each node in the distribution.
	var cfg config.Config
	currentExtPort = 40000
	externalPorts = make(map[string]int)
	config.Process(os.Args[1]+"/config.json", &cfg)

	// Loop through all nodes in config and create
	//   leader files for each.
	for _, n := range cfg.Nodes {
		err := createLeader(os.Args[1]+"/leaders", n, cfg.GUI_port, os.Args[1]+"/config.json")
		if err != nil {
			log.ERROR.Println(err)
			os.Exit(1)
		}
	}

	log.INFO.Println("Compilation complete.")
}
