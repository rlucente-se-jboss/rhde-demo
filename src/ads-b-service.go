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

const (
    portNumber = "8888"
	jsonFilePath = "./data/ads-b-data.json"
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

func convertFromRawReports(rawReports RawADSBReports) AircraftStates {
	var convertedAircraftStates AircraftStates

	for _, rawState := range rawReports.States {
		var state AircraftState

		state.ICAO24 = rawState[0].(string)
		state.CallSign = rawState[1].(string)
		state.TimePosition = int64(rawState[3].(float64))
		state.Longitude = rawState[5].(float64)
		state.Latitude = rawState[6].(float64)
		state.TrueTrack = rawState[10].(float64)

		convertedAircraftStates = append(convertedAircraftStates, state)
	}

	return convertedAircraftStates
}

func interpolateAircraftStates(aircraftStates AircraftStates) AircraftStates {
	var interpolatedStates AircraftStates

	statesMap := make(map[string]AircraftStates)

	// map each aircraft to its array of states
	for _, state := range aircraftStates {
		statesMap[state.ICAO24] = append(statesMap[state.ICAO24], state)
	}

	for key := range statesMap {
		interpolatedStates = append(interpolatedStates, statesMap[key][0])

		if len(statesMap[key]) > 1 {
			for i := 1; i < len(statesMap[key]); i++ {
				oldState := statesMap[key][i-1]
				currentState := statesMap[key][i]

				numSamples := currentState.TimePosition - oldState.TimePosition
				lonDelta := (currentState.Longitude - oldState.Longitude) / float64(numSamples)
				latDelta := (currentState.Latitude - oldState.Latitude) / float64(numSamples)

				trkDelta := (currentState.TrueTrack - oldState.TrueTrack)
				if trkDelta < -180.0 {
					trkDelta += 360.0
				}
				if trkDelta > 180.0 {
					trkDelta -= 360.0
				}
				trkDelta /= float64(numSamples)

				var j int64
				for j = 1; j <= numSamples; j++ {
					var newState AircraftState

					newState.ICAO24 = oldState.ICAO24
					newState.CallSign = oldState.CallSign
					newState.TimePosition = oldState.TimePosition + j
					newState.Longitude = oldState.Longitude + float64(j)*lonDelta
					newState.Latitude = oldState.Latitude + float64(j)*latDelta

					newTrack := oldState.TrueTrack + float64(j)*trkDelta
					if newTrack < 0 {
						newTrack += 360
					}
					if newTrack > 359 {
						newTrack -= 360
					}
					newState.TrueTrack = newTrack

					interpolatedStates = append(interpolatedStates, newState)
				}
			}
		}
	}

	sort.Sort(interpolatedStates)
	return interpolatedStates
}

// JSON report of all current aircraft states
var currentAircraftStates []byte

func main() {
	flag.Parse()

	rawADSBReports := readRawADSBReports()
	aircraftStates := convertFromRawReports(rawADSBReports)
	sort.Sort(aircraftStates)
	aircraftStates = interpolateAircraftStates(aircraftStates)
	applyTimeCorrection(aircraftStates)

	// update the aircraft states in the web service response every two seconds
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		lenStates := len(aircraftStates)
		fmt.Println("Total aircraft states: ", lenStates)

		aircraftStateMap := make(map[string]AircraftState)

		// this loop repeats every two seconds
		for _ = range ticker.C {
			rightNow := time.Now().Unix()

			// if past the end, repeat the dataset by adjusting the times
			if rightNow > aircraftStates[lenStates-1].TimePosition {
				applyTimeCorrection(aircraftStates)
			}

			for _, state := range aircraftStates {
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
	log.Fatal(http.ListenAndServe(":" + portNumber, nil))
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func handler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	w.Header().Set("Content-Type", "application/json")
	w.Write(currentAircraftStates)
}
