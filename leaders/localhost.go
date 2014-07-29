package main

import (
	"emd/connector"
	"emd/core"
	"emd/leader"
	"emd/log"
	"emd/worker"
	"io/ioutil"
	"os"

	workers_ "../workers"
)

func main() {
	log.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	// Worker Connections
	
		MGMT_Sniffer := &connector.Local{
			connector.Base{
				core.Core{"MGMT_Sniffer"},
				make(chan interface{}),
			},
		}

		
			
				Sniffer_and_Parser := &connector.Local{
					connector.Base{
						core.Core{"Sniffer_and_Parser"},
						make(chan interface{}, 0),
					},
				}
			
		
	
		MGMT_Parser := &connector.Local{
			connector.Base{
				core.Core{"MGMT_Parser"},
				make(chan interface{}),
			},
		}

		
			
		
			
				Sink_and_Parser := &connector.Local{
					connector.Base{
						core.Core{"Sink_and_Parser"},
						make(chan interface{}, 0),
					},
				}
			
		
	
		MGMT_Sink := &connector.Local{
			connector.Base{
				core.Core{"MGMT_Sink"},
				make(chan interface{}),
			},
		}

		
			
		
	

	// Workers
	workers := []worker.Worker{
		
			workers_.Sniffer{
				worker.Work{
					core.Core{"Sniffer"},
					map[string]connector.Connector{
						"MGMT_Sniffer": MGMT_Sniffer,

						
							"Sniffer_and_Parser": Sniffer_and_Parser,
						
					},
				},
			},
		
			workers_.Parser{
				worker.Work{
					core.Core{"Parser"},
					map[string]connector.Connector{
						"MGMT_Parser": MGMT_Parser,

						
							"Sniffer_and_Parser": Sniffer_and_Parser,
						
							"Sink_and_Parser": Sink_and_Parser,
						
					},
				},
			},
		
			workers_.Sink{
				worker.Work{
					core.Core{"Sink"},
					map[string]connector.Connector{
						"MGMT_Sink": MGMT_Sink,

						
							"Sink_and_Parser": Sink_and_Parser,
						
					},
				},
			},
		
	}

	// Node leader
	nodeLeader := &leader.Lead{
		core.Core{"NodeLeader"},
		"30000",
		"./config.json",
		workers,
		map[string]connector.Connector{
			
				"Sniffer": MGMT_Sniffer,
			
				"Parser": MGMT_Parser,
			
				"Sink": MGMT_Sink,
			
		},
	}

	nodeLeader.Init()
	nodeLeader.Run()
}
