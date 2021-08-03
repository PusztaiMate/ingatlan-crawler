package crawlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

var DunaHouseBaseUrl string = "https://dh.hu/"

func CreateDunaHouseQueryUrl(c Config) string {
	url := JoinUri(DunaHouseBaseUrl, "elado-ingatlan")

	districts := []string{}
	for _, d := range c.Districts {
		distNumber, err := getNumberForRomanNumeric(d)
		if err != nil {
			log.Println(err)
		}
		districts = append(districts, fmt.Sprintf("budapest-%d.-kerulet", distNumber))
	}
	url = JoinUri(url, c.Type)
	url = JoinUri(url, strings.Join(districts, "+"))
	url = JoinUri(url, "-")
	url = JoinUri(url, fmt.Sprintf("%d-%d-mFt", c.MinPrice, c.MaxPrice))
	url = JoinUri(url, fmt.Sprintf("%d-%d-m2", c.MinSize, c.MaxSize))

	return url
}

type DunaHouseListingPagesExtractor struct {
	maxPageNumber int
}

func (lpe *DunaHouseListingPagesExtractor) Predicate(n *html.Node) bool {
	return isLinkNode(n) && strings.Contains(findHrefAttribute(n), "oldal-")
}

func (lpe *DunaHouseListingPagesExtractor) ProcessNode(n *html.Node) {
	pageNum, err := strconv.Atoi(n.FirstChild.Data)
	if err != nil {
		log.Printf("could not convert %s to int", n.FirstChild.Data)
	}

	if pageNum > lpe.maxPageNumber {
		lpe.maxPageNumber = pageNum
	}
}

func (lpe *DunaHouseListingPagesExtractor) MaxPageNumber() int {
	return lpe.maxPageNumber
}

func (lpe *DunaHouseListingPagesExtractor) NextPageFormat() string {
	return "/oldal-%d"
}

type DunaHouseLinkCollector struct {
	Links []string
}

func (lc *DunaHouseLinkCollector) Predicate(n *html.Node) bool {
	return isLinkNode(n) && (doesClassAttrContainsVal(n, "listEstateWithoutPicOnPicture") || doesClassAttrContainsVal(n, "listEstateWithPicOnPicture"))
}

func (lc *DunaHouseLinkCollector) ProcessNode(n *html.Node) {
	log.Printf("node is %#v\n", n)
	link := findHrefAttribute(n)

	if len(link) == 0 {
		return
	}

	lc.Links = append(lc.Links, link)
}

func (lc *DunaHouseLinkCollector) GetLinks() []string {
	return lc.Links
}

func (lc *DunaHouseLinkCollector) GetNextPageFormatString() string {
	return "/oldal-%d"
}

type DunaHouseGeneralInfoExtractor struct {
	Address, NumOfFloors, Heating, BuiltIn, Condition string
}

func (e *DunaHouseGeneralInfoExtractor) Predicate(n *html.Node) bool {
	return isDivNode(n) && doesClassAttrContainsVal(n, "row table-list-style")
}

func (e *DunaHouseGeneralInfoExtractor) ProcessNode(n *html.Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if doesClassAttrContainsVal(c, "col-xs-6") {
			cc := c.FirstChild
			for cc != nil {
				sib := cc.NextSibling
				if sib == nil {
					break
				}

				switch cc.Data {
				case "Épület állapota belül:":
					e.Condition = sib.Data
				case "Belsö szintek száma:":
					e.NumOfFloors = sib.Data
				case "Fűtés":
					e.Heating = sib.Data
				case "Épült":
					e.BuiltIn = sib.Data
				case "Cím:":
					e.Address = sib.Data
				}

				cc = sib
			}

		}
	}
}

func (e *DunaHouseGeneralInfoExtractor) AddInfoIntoProp(p *PropertyInfo) {
	p.Address = e.Address
	p.NumOfFloors = e.NumOfFloors
	p.Heating = e.Heating
	p.BuiltIn = e.BuiltIn
	p.Condition = e.Condition
}

func getNumberForRomanNumeric(roman string) (int, error) {
	switch strings.ToUpper(roman) {
	case "I":
		return 1, nil
	case "II":
		return 2, nil
	case "III":
		return 3, nil
	case "IV":
		return 4, nil
	case "V":
		return 5, nil
	case "VI":
		return 6, nil
	case "VII":
		return 7, nil
	case "VIII":
		return 8, nil
	case "IX":
		return 9, nil
	case "X":
		return 10, nil
	case "XI":
		return 11, nil
	case "XII":
		return 12, nil
	case "XIII":
		return 13, nil
	case "XIV":
		return 14, nil
	case "XV":
		return 15, nil
	case "XVI":
		return 16, nil
	case "XVII":
		return 17, nil
	case "XVIII":
		return 18, nil
	case "XIX":
		return 19, nil
	case "XX":
		return 20, nil
	case "XXI":
		return 21, nil
	case "XXII":
		return 22, nil
	}
	return 0, fmt.Errorf("could not parse given roman district number: %s", roman)
}