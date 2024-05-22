package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/wbergg/trafikverket-bot/config"
)

type APIRespTstation struct {
	Response struct {
		Result []struct {
			TrainStation []struct {
				AdvertisedLocationName string `json:"AdvertisedLocationName"`
				LocationSignature      string `json:"LocationSignature"`
			} `json:"TrainStation"`
		} `json:"RESULT"`
	} `json:"RESPONSE"`
}

func FixCords(wsg84 string) string {
	// Remove unessecary chars from the input string
	cords := strings.TrimPrefix(strings.TrimSuffix(wsg84, ")"), "POINT (")
	splitCords := strings.Split(cords, " ")
	// Return opposite direction of cords
	return splitCords[1] + "," + splitCords[0]
}

func GetStationName(shortname string) string {
	url := "https://api.trafikinfo.trafikverket.se/v2/data.json"
	query := fmt.Sprintf(`<REQUEST>
	<LOGIN authenticationkey="%s"></LOGIN>
	  <QUERY
		objecttype="TrainStation"
		schemaversion="1.4">
		<INCLUDE>AdvertisedLocationName</INCLUDE>
		<INCLUDE>LocationSignature</INCLUDE>
		 <FILTER>
			<EQ name="LocationSignature" value="%s" />
		 </FILTER>
	  </QUERY>
	</REQUEST>`, config.Loaded.TrafikverketAPIKey, shortname)

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

	var data APIRespTstation
	var stationname string

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		panic(err)
	}
	for _, result := range data.Response.Result {
		for _, name := range result.TrainStation {
			stationname = name.AdvertisedLocationName
		}
	}

	return stationname
}
