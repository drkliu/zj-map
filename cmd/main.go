package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	//"github.com/kr/pretty"
	"github.com/kr/pretty"
	"github.com/xuri/excelize/v2"
	"googlemaps.github.io/maps"
)

type DistanceAndTime struct {
	distanceMatrixResponse *maps.DistanceMatrixResponse
	postCode               int
}

func main() {
	t := time.Now()
	//proxy to local to visit google
	proxyUrl, _ := url.Parse("http://127.0.0.1:1080/pac?hash=nRBlD-SPEADIocnyzFIPZA2&secret=rLwfsHWfJePSb3D5Vp4FzWVB9keIEgXbtcGqJamuQlI1")
	myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}

	c, err := maps.NewClient(maps.WithHTTPClient(myClient), maps.WithAPIKey("your api key"))
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}
	destPostCodes := genAuDestPostCodes(2000, 2899, []int{2208})
	distanceAndTimes, err := getDistanceAndTimeByAuPostCodes(c, "Kingsgrove DC NSW 2208, Australia", destPostCodes)
	if err == nil {
		outputDistanceAndTimes(distanceAndTimes)
	}
	pretty.Println("cost time: ", time.Since(t).Seconds())
}
func outputDistanceAndTimes(distanceAndTimes []*DistanceAndTime) {
	f := excelize.NewFile()
	var i = 1
	//Postcodes,KMs,Mins
	f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i), "Postcodes")
	f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i), "KMs")
	f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i), "Mins")
	f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i), "DestinationAddresses")
	i = 2
	style, err := f.NewStyle(`{"fill":{"type":"pattern","color":["#ff0000"],"pattern":1}}`)
	if err != nil {
		fmt.Println(err)
	}

	for index, element := range distanceAndTimes {
		haveErr := false
		i := index + i

		destinationAddress := element.distanceMatrixResponse.DestinationAddresses[0]
		if !strings.Contains(destinationAddress, strconv.Itoa(element.postCode)) {
			haveErr = true
		}
		distanceMatrixElement := element.distanceMatrixResponse.Rows[0].Elements[0]
		cell := fmt.Sprintf("A%d", i)
		f.SetCellValue("Sheet1", cell, element.postCode)
		if haveErr {
			f.SetCellStyle("Sheet1", cell, cell, style)
		}
		cell = fmt.Sprintf("B%d", i)
		f.SetCellValue("Sheet1", cell, distanceMatrixElement.Distance.HumanReadable)
		if haveErr {
			f.SetCellStyle("Sheet1", cell, cell, style)
		}
		cell = fmt.Sprintf("C%d", i)
		f.SetCellValue("Sheet1", cell, math.Round(time.Duration(distanceMatrixElement.Duration).Minutes()))
		if haveErr {
			f.SetCellStyle("Sheet1", cell, cell, style)
		}
		cell = fmt.Sprintf("D%d", i)
		f.SetCellValue("Sheet1", cell, destinationAddress)
		if haveErr {
			f.SetCellStyle("Sheet1", cell, cell, style)
		}
	}
	if err := f.SaveAs("au.xlsx"); err != nil {
		log.Fatal(err)
	}
}
func genAuDestPostCodes(begin int, end int, ignores []int) []int {
	var result []int
	for i := begin; i <= end; i++ {
		if intContains(ignores, i) {
			continue
		}
		result = append(result, i)
	}
	return result
}
func intContains(elements []int, i int) bool {
	for _, element := range elements {
		if i == element {
			return true
		}
	}
	return false
}
func getDistanceAndTimeByAuPostCodes(client *maps.Client, originPostCode string, destPostCodes []int) ([]*DistanceAndTime, error) {

	var result []*DistanceAndTime
	for _, postCode := range destPostCodes {
		r := &maps.DistanceMatrixRequest{
			Origins:      []string{originPostCode},
			Destinations: []string{fmt.Sprintf("NSW %d, Australia", postCode)},
			Mode:         "ModeDriving",
		}

		response, err := client.DistanceMatrix(context.Background(), r)
		if err != nil {
			log.Fatalf("fatal error: %s", err)
			continue
		}

		result = append(result, &DistanceAndTime{distanceMatrixResponse: response, postCode: postCode})
	}
	return result, nil
}
