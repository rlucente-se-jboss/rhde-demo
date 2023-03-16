package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"time"
)

// raw data format for ADS-B reports
type RawADSBReports struct {
	States [][]any `json:"states"`
}

// a very abbreviated ADS-B report
type AircraftState struct {
    ICAO24       string  `json:"icao24"`
	CallSign     string  `json:"callsign"`
	TimePosition int64   `json:"time_position"`
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
	TrueTrack    float64 `json:"true_track"`
}

type AircraftStates []AircraftState

// for formatting json responses
type AircraftStatesResponse struct {
	States AircraftStates `json:"states"`
}

// sort functions
func (states AircraftStates) Less(i, j int) bool {
	return states[i].TimePosition < states[j].TimePosition
}

func (states AircraftStates) Swap(i, j int) {
	states[i], states[j] = states[j], states[i]
}

func (states AircraftStates) Len() int {
	return len(states)
}

const (
	jsonFilePath = "./data/ads-b-data.json"
)

var fileArg = flag.String("f", jsonFilePath, "JSON file containing ADS-B reports")

func applyTimeCorrection(states AircraftStates) {
	timeCorrection := time.Now().Unix() - states[0].TimePosition

	for i, _ := range states {
		states[i].TimePosition += timeCorrection
	}
}

func readRawADSBReports() RawADSBReports {
	var allRawADSBReports RawADSBReports

	buf, err := ioutil.ReadFile(*fileArg)
	if err != nil {
		// use small set of data when no file available
        buf = []byte("{\"states\":[[\"a7dec2\",\"JIA5316 \",\"United States\",1678803380,1678803380,-77.4538,38.9492,null,true,0.06,182.81,null,null,null,\"7272\",false,0],[\"a2cf43\",\"N280PC  \",\"United States\",1678803634,1678803643,-77.4532,38.9697,null,true,0,267.19,null,null,null,null,false,0],[\"a448d7\",\"UAL1101 \",\"United States\",1678803638,1678803640,-77.4487,38.9464,null,true,1.03,216.56,null,null,null,null,false,0],[\"a2a7c4\",\"00000000\",\"United States\",1678803641,1678803647,-77.4532,38.9695,null,true,0.06,270,null,null,null,\"2012\",true,0]]}")
	}

	err = json.Unmarshal(buf, &allRawADSBReports)
	if err != nil {
		panic(err)
	}

	return allRawADSBReports
}

func sortAllAircraftStates(rawReports RawADSBReports) AircraftStates {
	var sortedAircraftStates AircraftStates

	for _, rawState := range rawReports.States {
		var state AircraftState

        state.ICAO24 = rawState[0].(string)
		state.CallSign = rawState[1].(string)
		state.TimePosition = int64(rawState[3].(float64))
		state.Longitude = rawState[5].(float64)
		state.Latitude = rawState[6].(float64)
		state.TrueTrack = rawState[10].(float64)

		sortedAircraftStates = append(sortedAircraftStates, state)
	}

	sort.Sort(sortedAircraftStates)

	return sortedAircraftStates
}

// JSON report of all current aircraft states
var currentAircraftStates []byte

func main() {
	flag.Parse()

	rawADSBReports := readRawADSBReports()
	timeOrderedStates := sortAllAircraftStates(rawADSBReports)
	applyTimeCorrection(timeOrderedStates)

	// update the aircraft states in the web service response every two seconds
	ticker := time.NewTicker(2 * time.Second)
	go func() {
		fmt.Println("Total aircraft states: ", len(timeOrderedStates))
        aircraftStateMap := make(map[string]AircraftState)

		// this loop repeats every two seconds
		for _ = range ticker.C {
			rightNow := time.Now().Unix()

			for _, state := range timeOrderedStates {
				if state.TimePosition > rightNow {
					break
				}

                aircraftStateMap[state.ICAO24] = state
			}

			var currentStatesResponse AircraftStatesResponse
            for _, state := range aircraftStateMap {
                currentStatesResponse.States = append(currentStatesResponse.States, state)
            }

			marshalledStatesReport, err := json.Marshal(currentStatesResponse)
			if err != nil {
				panic(err)
			}

			currentAircraftStates = marshalledStatesReport
		}
	}()

	http.HandleFunc("/ads-b-states", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(currentAircraftStates)
}
