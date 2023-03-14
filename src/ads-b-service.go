// TODO THIS ISN'T EVEN HALF GOOD AND IT NEEDS A LOT OF WORK
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

type AutoGenerated struct {
	MetroTrackingData []MetroTrackingData `json:"MetroTrackingData"`
}

type BusPositions struct {
	VehicleID     string  `json:"VehicleID"`
	Lat           float64 `json:"Lat"`
	Lon           float64 `json:"Lon"`
	Deviation     int     `json:"Deviation"`
	DateTime      string  `json:"DateTime"`
	TripID        string  `json:"TripID"`
	RouteID       string  `json:"RouteID"`
	DirectionNum  int     `json:"DirectionNum"`
	DirectionText string  `json:"DirectionText"`
	TripHeadsign  string  `json:"TripHeadsign"`
	TripStartTime string  `json:"TripStartTime"`
	TripEndTime   string  `json:"TripEndTime"`
}

type MetroTrackingData struct {
	BusPositions []BusPositions `json:"BusPositions"`
}

type AllBusPositions []BusPositions

func (abp AllBusPositions) Less(i, j int) bool { return abp[i].DateTime < abp[j].DateTime }
func (abp AllBusPositions) Swap(i, j int)      { abp[i], abp[j] = abp[j], abp[i] }
func (abp AllBusPositions) Len() int           { return len(abp) }

const (
	jsonFilePath = "./data/busses.json"
)

var fileArg = flag.String("f", jsonFilePath, "JSON file containing bus position data")

func adjustTime(value *string, correction time.Duration) {
	*value = parse(*value).Add(correction).Format(time.RFC3339Nano)
}

func applyTimeCorrection(busPositions AllBusPositions) {
	timeCorrection := time.Since(parse(busPositions[0].DateTime))

	for i, _ := range busPositions {
		adjustTime(&busPositions[i].DateTime, timeCorrection)
		adjustTime(&busPositions[i].TripStartTime, timeCorrection)
		adjustTime(&busPositions[i].TripEndTime, timeCorrection)
	}
}

func parse(timeValue string) time.Time {
	if !strings.HasSuffix(timeValue, "Z") {
		timeValue += "Z"
	}

	t, err := time.Parse(time.RFC3339Nano, timeValue)
	if err != nil {
		panic(err)
	}
	return t
}

func readAllJSONBusReports() AutoGenerated {
	var busJSONData AutoGenerated

	buf, err := ioutil.ReadFile(*fileArg)
	if err != nil {
		// use basic seed data when no file available
		buf = []byte("{\"MetroTrackingData\":[{\"BusPositions\":[{\"VehicleID\":\"7366\",\"Lat\":39.086246,\"Lon\":-76.93998,\"Deviation\":0,\"DateTime\":\"2017-08-16T13:22:05\",\"TripID\":\"921035840\",\"RouteID\":\"Y7\",\"DirectionNum\":1,\"DirectionText\":\"SOUTH\",\"TripHeadsign\":\"SILVER SPRING STATION\",\"TripStartTime\":\"2017-08-16T08:41:00\",\"TripEndTime\":\"2017-08-16T09:40:00\"}]},{\"BusPositions\":[{\"VehicleID\":\"7366\",\"Lat\":39.086246,\"Lon\":-76.93998,\"Deviation\":0,\"DateTime\":\"2017-08-16T13:22:15\",\"TripID\":\"921035840\",\"RouteID\":\"Y7\",\"DirectionNum\":1,\"DirectionText\":\"SOUTH\",\"TripHeadsign\":\"SILVER SPRING STATION\",\"TripStartTime\":\"2017-08-16T08:41:00\",\"TripEndTime\":\"2017-08-16T09:40:00\"}]}]}")
	}

	err = json.Unmarshal(buf, &busJSONData)
	if err != nil {
		panic(err)
	}

	return busJSONData
}

func reverseBusPositions(busPositions AllBusPositions) {
	sort.Sort(sort.Reverse(busPositions))

	t0 := parse(busPositions[0].DateTime)

	for i, bus := range busPositions[1:] {
		correction := 2 * t0.Sub(parse(bus.DateTime))
		j := i + 1

		adjustTime(&busPositions[j].DateTime, correction)
		adjustTime(&busPositions[j].TripStartTime, correction)
		adjustTime(&busPositions[j].TripEndTime, correction)
	}
}

func sortAllBusPositions(allBusData AutoGenerated) AllBusPositions {
	var sortedBusPositions AllBusPositions

	for _, busses := range allBusData.MetroTrackingData {
		for _, bus := range busses.BusPositions {
			sortedBusPositions = append(sortedBusPositions, bus)
		}
	}

	sort.Sort(sortedBusPositions)

	return sortedBusPositions
}

// JSON report of all current bus positions
var currentBusPositions []byte

func main() {
	flag.Parse()

	allBusData := readAllJSONBusReports()
	busPosTimeOrdered := sortAllBusPositions(allBusData)
	applyTimeCorrection(busPosTimeOrdered)

	ticker := time.NewTicker(2 * time.Second)
	go func() {
		busPosMap := make(map[string]BusPositions)
		var currentPosIndex int

		fmt.Println("Total bus positions: ", len(busPosTimeOrdered))

		// this loop repeats every two seconds
		for _ = range ticker.C {
			for _, bus := range busPosTimeOrdered[currentPosIndex:] {
				busReportTime := parse(bus.DateTime)
				if busReportTime.After(time.Now()) {
					break
				}

				busPosMap[bus.VehicleID] = bus
				fmt.Println("Updated position report for bus: ", bus.VehicleID, ", at time: ", bus.DateTime, ", currentPosIndex: ", currentPosIndex)

				currentPosIndex++
			}

			var latestBusReport MetroTrackingData
			for _, bus := range busPosMap {
				latestBusReport.BusPositions = append(latestBusReport.BusPositions, bus)
			}

			marshalledPosReport, err := json.Marshal(latestBusReport)
			if err != nil {
				panic(err)
			}

			currentBusPositions = marshalledPosReport

			if currentPosIndex == len(busPosTimeOrdered) {
				currentPosIndex = 0

				reverseBusPositions(busPosTimeOrdered)
				applyTimeCorrection(busPosTimeOrdered)

				fmt.Println("Reverse the bus position reports")
			}
		}
	}()

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(currentBusPositions)
}
