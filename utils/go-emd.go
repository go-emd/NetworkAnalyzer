package main

import (
	config "./config"

	"github.com/go-emd/emd/log"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"net/http"
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

// Helper func to get project name from path given
func parseProject(path string) string {
	name := ""

	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	i := len(path)-1
	for {
		if path[i] == '/' { return name }
		if i < 0 { return "" }

		name = name + string(path[i])
		i = i-1
	}
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

/*
 *
 * Build node specific leader with "go build"
 *
 */
func buildLeader(path, hostname string) (string, error) {
	out, err := exec.Command("go", "build", "-o", path+"/bin/"+hostname, path+"/"+hostname+".go").Output()

	return string(out), err
}

/*
 *
 * Compiles and builds the distribution leader files.
 *
 */
func compile() {
	path := ""

	if len(os.Args) == 0 || len(os.Args) > 2 {
		log.ERROR.Println("Usage: go-emd compile --path <path to dir containing config.json>")
		log.ERROR.Println("Usage: go-emd compile --help")
		os.Exit(1)
	} else if os.Args[0] == "--help" || os.Args[0] == "-h" || os.Args[0] == "help" {
		log.INFO.Println("Usage: go-emd compile --path <path to dir containing config.json>")
		log.INFO.Println("Usage: go-emd compile --help")
		os.Exit(1)
	} else if os.Args[0] == "--path" || os.Args[0] == "-p" && len(os.Args) == 2 {
		path = os.Args[1]
	} else {
		log.ERROR.Println("Usage: go-emd compile --path <path to dir containing config.json>")
		log.ERROR.Println("Usage: go-emd compile --help")
		os.Exit(1)
	}

	var cfg config.Config
	currentExtPort = 40000
	externalPorts = make(map[string]int)
	config.Process(path+"/config.json", &cfg)

	// Loop through all nodes in config and create
	//   leader files for each, then build them 
	//   placing them into the /leaders/bin dir.
	for _, n := range cfg.Nodes {
		err := createLeader(path+"/leaders", n, cfg.GUI_port, path+"/config.json")
		if err != nil {
			log.ERROR.Println(err)
			os.Exit(1)
		}
		
		log.INFO.Println("Leader "+n.Hostname+" compiled successfully")
		
		_, err = buildLeader(path+"/leaders", n.Hostname)
		if err != nil {
			log.ERROR.Println(err)
			os.Exit(1)
		}
		log.INFO.Println("Leader "+n.Hostname+" built successfully")
	}
	
	log.INFO.Println("Compile successful")
}

/*
 *
 * Cleans and removes leader files and executables.
 *
 */
func clean() {
	path := ""

	if len(os.Args) == 0 || len(os.Args) > 2 {
		log.ERROR.Println("Usage: go-emd clean --path <path to dir containing config.json>")
		log.ERROR.Println("Usage: go-emd clean --help")
		os.Exit(1)
	} else if os.Args[0] == "--help" || os.Args[0] == "-h" || os.Args[0] == "help" {
		log.INFO.Println("Usage: go-emd clean --path <path to dir containing config.json>")
		log.INFO.Println("Usage: go-emd clean --help")
		os.Exit(1)
	} else if os.Args[0] == "--path" || os.Args[0] == "-p" && len(os.Args) == 2 {
		path = os.Args[1]
	} else {
		log.ERROR.Println("Usage: go-emd clean --path <path to dir containing config.json>")
		log.ERROR.Println("Usage: go-emd clean --help")
		os.Exit(1)
	}

	var cfg config.Config
	config.Process(path+"/config.json", &cfg)

	for _, n := range cfg.Nodes {
		log.INFO.Println("Removing "+path+"/leaders/"+n.Hostname+".go")
		err := os.Remove(path+"/leaders/"+n.Hostname+".go")
		if err != nil {
			log.ERROR.Println(err)
			os.Exit(1)
		}

		log.INFO.Println("Removing "+path+"/leaders/bin/"+n.Hostname)
		err = os.Remove(path+"/leaders/bin/"+n.Hostname)
		if err != nil {
			log.ERROR.Println(err)
			os.Exit(1)
		}
	}
	
	log.INFO.Println("Clean successful")
}

/*
 *
 * Rsync must be installed in order to distribute the distribution.
 *
 */
func distribute() {
	path := ""

	if len(os.Args) == 0 || len(os.Args) > 2 {
		log.ERROR.Println("Usage: go-emd distribute --path <path to dir containing config.json>")
		log.ERROR.Println("Usage: go-emd distribute --help")
		os.Exit(1)
	} else if os.Args[0] == "--help" || os.Args[0] == "-h" || os.Args[0] == "help" {
		log.INFO.Println("Usage: go-emd distribute --path <path to dir containing config.json>")
		log.INFO.Println("Usage: go-emd distribute --help")
		os.Exit(1)
	} else if os.Args[0] == "--path" || os.Args[0] == "-p" && len(os.Args) == 2 {
		path = os.Args[1]
		
		// Need to make sure trailing slash is removed.
		if path[len(path)-1] == '/' {
			path = path[:len(path)-1]
		}
	} else {
		log.ERROR.Println("Usage: go-emd distribute --path <path to dir containing config.json>")
		log.ERROR.Println("Usage: go-emd distribute --help")
		os.Exit(1)
	}

	user, err := user.Current()
	if err != nil {
		log.ERROR.Println(err)
		os.Exit(1)
	}

	var cfg config.Config
	config.Process(path+"/config.json", &cfg)

	for _, n := range cfg.Nodes{
		log.INFO.Println("Distributing to "+n.Hostname)
		_, err := exec.Command("rsync", "-a", "-z", path, user.Username+"@"+n.Hostname+":"+os.TempDir()).Output()
		if err != nil {
			log.ERROR.Println(err)
			os.Exit(1)
		}
	}

	log.INFO.Println("Distribute successful")
}

/*
 *
 * Start the distribution given the path to it on each node.
 *
 */
func start() {
	path := ""

	if len(os.Args) == 0 || len(os.Args) > 2 {
		log.ERROR.Println("Usage: go-emd start --path <path to dir containing config.json>")
		log.ERROR.Println("Usage: go-emd start --help")
		os.Exit(1)
	} else if os.Args[0] == "--help" || os.Args[0] == "-h" || os.Args[0] == "help" {
		log.INFO.Println("Usage: go-emd start --path <path to dir containing config.json>")
		log.INFO.Println("Usage: go-emd start --help")
		os.Exit(1)
	} else if os.Args[0] == "--path" || os.Args[0] == "-p" && len(os.Args) == 2 {
		path = os.Args[1]
		
		// Need to make sure trailing slash is removed.
		if path[len(path)-1] == '/' {
			path = path[:len(path)-1]
		}
	} else {
		log.ERROR.Println("Usage: go-emd start --path <path to dir containing config.json>")
		log.ERROR.Println("Usage: go-emd start --help")
		os.Exit(1)
	}

	user, err := user.Current()
	if err != nil {
		log.ERROR.Println(err)
		os.Exit(1)
	}

	projectName := parseProject(path)
	if projectName == "" {
		log.ERROR.Println("Unable to parse distribution name from path.")
		log.ERROR.Println("Usage: go-emd start --path <path to dir containing config.json>")
		log.ERROR.Println("Usage: go-emd start --help")
		os.Exit(1)
	}

	var cfg config.Config
	config.Process(path+"/config.json", &cfg)

	for _, n := range cfg.Nodes{
		log.INFO.Println("Starting leader on "+n.Hostname)
		_, err := exec.Command("ssh", "-n", "-f", user.Username+"@"+n.Hostname, "\"sh -c 'nohup "+os.TempDir()+"/"+projectName+"/leaders/bin/"+n.Hostname+" > /dev/null 2>&1 &'").Output()
		if err != nil {
			log.ERROR.Println(err)
			os.Exit(1)
		}
	}

	log.INFO.Println("Start successful")
}

/*
 *
 * Perform GET request to stop distribution.
 *
 */
func stop() {
	path := ""

	if len(os.Args) == 0 || len(os.Args) > 2 {
		log.ERROR.Println("Usage: go-emd stop --path <path to dir containing config.json>")
		log.ERROR.Println("Usage: go-emd stop --help")
		os.Exit(1)
	} else if os.Args[0] == "--help" || os.Args[0] == "-h" || os.Args[0] == "help" {
		log.INFO.Println("Usage: go-emd stop --path <path to dir containing config.json>")
		log.INFO.Println("Usage: go-emd stop --help")
		os.Exit(1)
	} else if os.Args[0] == "--path" || os.Args[0] == "-p" && len(os.Args) == 2 {
		path = os.Args[1]
		
		// Need to make sure trailing slash is removed.
		if path[len(path)-1] == '/' {
			path = path[:len(path)-1]
		}
	} else {
		log.ERROR.Println("Usage: go-emd stop --path <path to dir containing config.json>")
		log.ERROR.Println("Usage: go-emd stop --help")
		os.Exit(1)
	}
	
	var cfg config.Config
	config.Process(path+"/config.json", &cfg)
	
	log.INFO.Println("Stopping distribution")
	
	for _, n := range cfg.Nodes {
		log.INFO.Println("Stopping node "+n.Hostname)

		// Stop all the workers
		_, err := http.Get("http://"+n.Hostname+":"+cfg.GUI_port+"/stop")
		if err != nil {
			log.ERROR.Println(err)
			os.Exit(1)
		}

		// Stop the leader
		_, err = http.Get("http://"+n.Hostname+":"+cfg.GUI_port+"/stop")
		if err != nil {
			log.ERROR.Println(err)
			os.Exit(1)
		}
	}
	
	log.INFO.Println("Stop successful")
}

/*
 *
 * Perform GET request to get status of distribution.
 *
 */
func status() {
	path := ""

	if len(os.Args) == 0 || len(os.Args) > 2 {
		log.ERROR.Println("Usage: go-emd status --path <path to dir containing config.json>")
		log.ERROR.Println("Usage: go-emd status --help")
		os.Exit(1)
	} else if os.Args[0] == "--help" || os.Args[0] == "-h" || os.Args[0] == "help" {
		log.INFO.Println("Usage: go-emd status --path <path to dir containing config.json>")
		log.INFO.Println("Usage: go-emd status --help")
		os.Exit(1)
	} else if os.Args[0] == "--path" || os.Args[0] == "-p" && len(os.Args) == 2 {
		path = os.Args[1]
		
		// Need to make sure trailing slash is removed.
		if path[len(path)-1] == '/' {
			path = path[:len(path)-1]
		}
	} else {
		log.ERROR.Println("Usage: go-emd status --path <path to dir containing config.json>")
		log.ERROR.Println("Usage: go-emd status --help")
		os.Exit(1)
	}
	
	var cfg config.Config
	config.Process(path+"/config.json", &cfg)
	
	for _, n := range cfg.Nodes {
		log.INFO.Println("Obtaining status of node "+n.Hostname)

		resp, err := http.Get("http://"+n.Hostname+":"+cfg.GUI_port+"/status")
		if err != nil {
			log.ERROR.Println(err)
			os.Exit(1)
		}
		
		log.INFO.Println(resp)
	}

	log.INFO.Println("Status successful")
}

/*
 *
 * Perform GET request to get metrics of distribution.
 *
 */
func metrics() {
	path := ""

	if len(os.Args) == 0 || len(os.Args) > 2 {
		log.ERROR.Println("Usage: go-emd metrics --path <path to dir containing config.json>")
		log.ERROR.Println("Usage: go-emd metrics --help")
		os.Exit(1)
	} else if os.Args[0] == "--help" || os.Args[0] == "-h" || os.Args[0] == "help" {
		log.INFO.Println("Usage: go-emd metrics --path <path to dir containing config.json>")
		log.INFO.Println("Usage: go-emd metrics --help")
		os.Exit(1)
	} else if os.Args[0] == "--path" || os.Args[0] == "-p" && len(os.Args) == 2 {
		path = os.Args[1]
		
		// Need to make sure trailing slash is removed.
		if path[len(path)-1] == '/' {
			path = path[:len(path)-1]
		}
	} else {
		log.ERROR.Println("Usage: go-emd metrics --path <path to dir containing config.json>")
		log.ERROR.Println("Usage: go-emd metrics --help")
		os.Exit(1)
	}
	
	var cfg config.Config
	config.Process(path+"/config.json", &cfg)
	
	for _, n := range cfg.Nodes {
		log.INFO.Println("Obtaining metrics of node "+n.Hostname)

		resp, err := http.Get("http://"+n.Hostname+":"+cfg.GUI_port+"/metrics")
		if err != nil {
			log.ERROR.Println(err)
			os.Exit(1)
		}
		
		log.INFO.Println(resp)
	}

	log.INFO.Println("Metrics successful")
}

/*
 *
 * Create directory structure for new distribution (copy boilerplate 
 *    distribution contents into projectName directory.
 *
 */
func newProject() {
	log.INFO.Println("Creating new distribution")

	// Copy boilerplate/* stuff into the path/name given.
	_, err := exec.Command("git", "clone", "https://github.com/go-emd/boilerplate.git").Output()
	if err != nil {
		log.ERROR.Println(err)
		os.Exit(1)
	}

	log.INFO.Println("New successful")
}

func main() {
	log.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	// Get the command line arguments
	var action string
	if len(os.Args) >= 2 {
		action = os.Args[1]
	} else {
		log.ERROR.Println("Usage: go-emd <action> args {new|compile|clean|distribute|start|stop|status|metrics}")
		os.Exit(1)
	}

	os.Args = os.Args[2:]

	switch action {
	case "new":
		newProject()
	case "compile":
		log.INFO.Println("Performing: compile")
		compile()
	case "clean":
		log.INFO.Println("Performing: clean")
		clean()
	case "distribute":
		log.INFO.Println("Performing: distribute")
		distribute()
	case "start":
		log.INFO.Println("Performing: start")
		start()
	case "stop":
		log.INFO.Println("Performing: stop")
		stop()
	case "status":
		log.INFO.Println("Performing: status")
		status()
	case "metrics":
		log.INFO.Println("Performing: metrics")
		metrics()
	default:
		log.ERROR.Println("Invalid action.")
		log.ERROR.Println("Usage: go-emd <action> args {new|compile|clean|distribute|start|stop|status|metrics}")
		os.Exit(1)
	}
}
