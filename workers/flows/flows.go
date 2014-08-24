/*
	Flows monitors flows based on hashing the 
	control data []bytes given and checking if 
	it is equal to the end of flow signature given.  
	This way this package can be used for any flow/
	protocol possible.
*/
package flows

import (
	"hash"
	"hash/adler32"
	"time"
)

// Hashing utility
var (
	hasher hash.Hash32
)

// Flows struct contains the PartialFlows being 
// built in real-time and the FinalFlows that 
// have officially ended without being forced to 
// flush everything.
type Flows struct {
	PartialFlows map[uint32]*Netflow
	FinalFlows []*Netflow
	EndSignature uint32
}

// Creates a new Flows instance allowing for 
// accumulation flow data.
func (f *Flows) New(es []byte) *Flows {
	hasher = adler32.New()

	return &Flows{
		PartialFlows: make(map[uint32]*Netflow),
		FinalFlows: make([]*Netflow)
		EndSignature: hasher.Checksum(es),
	}
}

// Updates the PartialFlows map, adds the current 
// rolling bytes count in the flow and checks to see 
// if this is an end of flow signature, if so then 
// it will transfer this flow entry to the FinalFlows 
// slice and add the current time to the duration field.
func (f *Flows) Update(data []byte, netflow *Netflow) {
	hash := hasher.Checksum(data)
	
	if f.PartialFlows[hash] == nil { // Start of flow
		f.PartialFlows[hash] = netflow
	} else { // Flow already started
		f.PartialFlows[hash].Bytes += netflow.Bytes

		if hash == f.EndSignature { // End of flow
			f.PartialFlows[hash].Duration = time.Now()
			f.FinalFlows = append(f.FinalFlows, f.PartialFlows[hash])
			f.PartialFlows[hash] = nil
		}
	}
}

// Flushes the FinalFlows cache and then clears the appropriate 
// data structures in the Flows struct.  If both is set to true 
// then both FinalFlows and PartialFlows will be returned then 
// cleared.  If both is false then only the FinalFlows will be 
// returned then cleared.
func (f *Flows) Flush(both bool) []*Netflow, map[uint32]*Netflow {
	def func() { // Used to clear the data structures after returning.
		f.FinalFlows = nil
		if both {
			f.PartialFlows = make(map[uint32]*Netflow)
		}
	}

	if both {
		return f.FinalFlows, f.PartialFlows
	} else {
		return f.FinalFlows, nil
	}
}
