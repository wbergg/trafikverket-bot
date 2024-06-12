package apipoller

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wbergg/bordershop-bot/message"
	"github.com/wbergg/trafikverket-bot/config"
	"github.com/wbergg/trafikverket-bot/db"
	"github.com/wbergg/trafikverket-bot/helper"
)

type APIRespTMsg struct {
	Response struct {
		Result []struct {
			TrainMessage []struct {
				CountyNo            []int  `json:"CountyNo"`
				Deleted             bool   `json:"Deleted"`
				ExternalDescription string `json:"ExternalDescription"`
				Geometry            struct {
					Sweref99Tm string `json:"SWEREF99TM"`
					Wgs84      string `json:"WGS84"`
				} `json:"Geometry"`
				EventID    string `json:"EventId"`
				Header     string `json:"Header"`
				ReasonCode []struct {
					Code        string `json:"Code"`
					Description string `json:"Description"`
				} `json:"ReasonCode"`
				TrafficImpact []struct {
					IsConfirmed      bool     `json:"IsConfirmed"`
					FromLocation     []string `json:"FromLocation"`
					AffectedLocation []struct {
						LocationSignature       string `json:"LocationSignature"`
						ShouldBeTrafficInformed bool   `json:"ShouldBeTrafficInformed"`
					} `json:"AffectedLocation"`
					ToLocation []string `json:"ToLocation"`
				} `json:"TrafficImpact"`
				StartDateTime                          time.Time `json:"StartDateTime"`
				PrognosticatedEndDateTimeTrafficImpact time.Time `json:"PrognosticatedEndDateTimeTrafficImpact"`
				LastUpdateDateTime                     time.Time `json:"LastUpdateDateTime"`
				ModifiedTime                           time.Time `json:"ModifiedTime"`
			} `json:"TrainMessage"`
		} `json:"RESULT"`
	} `json:"RESPONSE"`
}

var CountyMap = map[int]string{
	1:  "Stockholms län",
	2:  "DEPRECATED",
	3:  "Uppsala län",
	4:  "Södermanlands län",
	5:  "Östergötlands län",
	6:  "Jönköpings län",
	7:  "Kronobergs län",
	8:  "Kalmar län",
	9:  "Gotlands län",
	10: "Blekinge län",
	12: "Skåne län",
	13: "Hallands län",
	14: "Västra Götalands län",
	17: "Värmlands län",
	18: "Örebro län",
	19: "Västmanlands län",
	20: "Dalarnas län",
	21: "Gävleborgs län",
	22: "Västernorrlands län",
	23: "Jämtlands län",
	24: "Västerbottens län",
	25: "Norrbottens län",
}

func GetData() APIRespTMsg {
	//Set up logging
	f, err := os.OpenFile("trafikverket-log.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	logrus.SetOutput(f)

	// API query setup
	url := "https://api.trafikinfo.trafikverket.se/v2/data.json"
	query := fmt.Sprintf(`<REQUEST>
	<LOGIN authenticationkey="%s"></LOGIN>
	  <QUERY
		objecttype="TrainMessage"
		schemaversion="1.7">
		<INCLUDE>CountyNo</INCLUDE>
		<INCLUDE>Deleted</INCLUDE>
		<INCLUDE>EndDateTime</INCLUDE>
		<INCLUDE>EventId</INCLUDE>
		<INCLUDE>ExternalDescription</INCLUDE>
		<INCLUDE>Geometry</INCLUDE>
		<INCLUDE>Header</INCLUDE>
		<INCLUDE>LastUpdateDateTime</INCLUDE>
		<INCLUDE>ModifiedTime</INCLUDE>
		<INCLUDE>PrognosticatedEndDateTimeTrafficImpact</INCLUDE>
		<INCLUDE>ReasonCode</INCLUDE>
		<INCLUDE>StartDateTime</INCLUDE>
		<INCLUDE>TrafficImpact</INCLUDE>
	  </QUERY>
	</REQUEST>`, config.Loaded.TrafikverketAPIKey)

	byteQuery := []byte(query)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(byteQuery))
	req.Header.Add("Content-Type", "application/xml")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.141 Safari/537.36")
	if err != nil {
		panic(err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	var data APIRespTMsg

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		panic(err)
	}

	return data
}

func UpdateData(t message.Message, d *db.DBobject, data APIRespTMsg) {
	for _, response := range data.Response.Result {
		for _, result := range response.TrainMessage {
			for _, reason := range result.ReasonCode {
				if reason.Description == "banarbete" || reason.Description == "passkontroll" || reason.Code == "OMÄ02" {
					// Ignore these reasons by continuing
					continue
				} else {
					// Used for debug:
					fmt.Println(result.EventID)
					_, err := d.GetMessagesByPid(result.EventID)
					// If no row with eventID found in the DB
					if err == sql.ErrNoRows {
						err := d.InsertDBTrainMessage(result.EventID, result.CountyNo[0], result.Deleted, result.ExternalDescription, result.Geometry.Sweref99Tm, result.Geometry.Wgs84, result.Header, result.StartDateTime, result.PrognosticatedEndDateTimeTrafficImpact, result.LastUpdateDateTime, result.ModifiedTime)
						if err != nil {
							log.Fatal(err)
						}
						message := ""
						//message = message + "\nhttps://wberg.com/trafikverket-small.png"
						message = message + "\n\n*Ny trafikhändelse!*\n"
						message = message + "\n*Län:* " + CountyMap[result.CountyNo[0]]
						message = message + "\n*Starttid:* " + result.StartDateTime.String()
						message = message + "\n*Prognos klart:* " + result.PrognosticatedEndDateTimeTrafficImpact.String()
						message = message + "\n\n*Orsak:* " + result.Header
						message = message + "\n\n*Information:* " + result.ExternalDescription
						message = message + "\n\n*Google Maps:* https://maps.google.com/maps/place/" + helper.FixCords(result.Geometry.Wgs84)

						// Loop through affected stations and look them up
						var stationNames []string
						var isConfirmed bool
						for _, location := range result.TrafficImpact {
							isConfirmed = location.IsConfirmed
							// If impact is confirmed
							if location.IsConfirmed {
								for _, station := range location.AffectedLocation {
									// Should station be informed?
									if station.ShouldBeTrafficInformed {
										stationNames = append(stationNames, station.LocationSignature)
									}
								}
							}
						}
						stationNames = helper.GetStationNameAll(stationNames)

						// Reduce spam, only post affected stations if <30 and !0
						if len(stationNames) < 30 && len(stationNames) != 0 {
							message = message + "\n\n*Stationer som påverkas:* "
							message = message + strings.Join(stationNames, ", ")
						} else if !isConfirmed {
							message = message + "\n\n*Stationer som påverkas:* Ej confirmed"
						}

						// Send message if no county filter is applied
						if config.Loaded.County[0] == 0 {
							t.SendM(message)
						} else {
							// Create a map
							countySet := make(map[int]struct{})
							for _, c := range config.Loaded.County {
								countySet[c] = struct{}{}
							}

							// Check if any county matches
							for _, resultCounty := range result.CountyNo {
								if _, found := countySet[resultCounty]; found {
									t.SendM(message)
									break
								}
							}
						}

					} else {
						// Future use code to handle updates
						/*
							// Check if item has a new update time
							if result.ModifiedTime.After(dbresp.ModifiedTime) {
								fmt.Printf("Prognostid ")
								fmt.Println(dbresp.PrognosticatedEndDateTimeTrafficImpact, result.PrognosticatedEndDateTimeTrafficImpact)
								fmt.Printf("Last update ")
								fmt.Println(dbresp.LastUpdateDateTime, result.LastUpdateDateTime)

								fmt.Printf("dbresp är: %s", dbresp.ExternalDescription.String)
								fmt.Printf("\nresult är: %s", result.ExternalDescription)
								fmt.Printf("\n")
								// Check change in external description field
								if dbresp.ExternalDescription.String != result.ExternalDescription {
									fmt.Println("update in text")

									// Update DB
									// uppdatera fält
									// ExternalDescription
									// Send update msg to TeleG
									//message = message + "\n*Senast uppdaterad:* " + result.LastUpdateDateTime.String()

								} else {
									fmt.Println("no change")
								}

							}
						*/
						fmt.Println("--------")
					}
				}
			}
		}
	}
}

func DbSetup(d *db.DBobject) {
	data := GetData()

	for _, response := range data.Response.Result {
		for _, result := range response.TrainMessage {
			for _, reason := range result.ReasonCode {
				if reason.Description == "banarbete" || reason.Description == "passkontroll" {
					// Ignore these reasons
					continue
				} else {
					err := d.InsertDBTrainMessage(result.EventID, result.CountyNo[0], result.Deleted, result.ExternalDescription, result.Geometry.Sweref99Tm, result.Geometry.Wgs84, result.Header, result.StartDateTime, result.PrognosticatedEndDateTimeTrafficImpact, result.LastUpdateDateTime, result.ModifiedTime)
					if err != nil {
						panic(err)
					}
				}
			}
		}
	}
}
