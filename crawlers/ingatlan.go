package crawlers

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

var IngatlanBaseUrl string = "https://ingatlan.com/"

type IngatlanComLinkCollector struct {
	Links []string
}

func (l *IngatlanComLinkCollector) Predicate(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "a"
}

func (l *IngatlanComLinkCollector) ProcessNode(n *html.Node) {
	link := findHrefAttribute(n)

	if len(link) == 0 || !isNodeListingLink(n) {
		return
	}

	l.Links = append(l.Links, link)
}

func (l *IngatlanComLinkCollector) GetLinks() []string {
	return l.Links
}

type IngatlanComListingPagesExtractor struct {
	maxPageNumber int
}

func (lpe *IngatlanComListingPagesExtractor) Predicate(n *html.Node) bool {
	return isDivNode(n) && doesClassAttrContainsVal(n, "pagination__page-number")
}

func (lpe *IngatlanComListingPagesExtractor) ProcessNode(n *html.Node) {
	pageDesc := n.FirstChild.Data
	if pageDesc == "" {
		return
	}

	secondPart := strings.TrimSpace(strings.Split(pageDesc, "/")[1])
	numAsString := strings.Split(secondPart, " ")[0]

	maxNum, err := strconv.Atoi(numAsString)
	if err != nil {
		log.Printf("could not convert %s to int\n", numAsString)
	}

	lpe.maxPageNumber = maxNum
}

func (lpe *IngatlanComListingPagesExtractor) MaxPageNumber() int {
	return lpe.maxPageNumber
}

func (lpe *IngatlanComListingPagesExtractor) NextPageFormat() string {
	return "?page=%d"
}

type IngatlanComPropertyInfoExtractor struct {
	Condition, BuiltIn, NumOfFloors, Parking, Heating, AirConditioning, ToiletAndBathroom string
}

func (p *IngatlanComPropertyInfoExtractor) Predicate(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "dl"
}

func (p *IngatlanComPropertyInfoExtractor) ProcessNode(n *html.Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		paramName, paramVal, err := getChildsParamNameAndValue(c)
		if err != nil {
			continue
		}

		switch paramName {
		case "Ingatlan állapota":
			p.Condition = paramVal
		case "Építés éve":
			p.BuiltIn = paramVal
		case "Épület szintjei":
			p.NumOfFloors = paramVal
		case "Parkolás":
			p.Parking = paramVal
		case "Fűtés":
			p.Heating = paramVal
		case "Légkondicionáló":
			p.AirConditioning = paramVal
		case "Fürdő és WC":
			p.ToiletAndBathroom = paramVal
		}
	}
}

func (e *IngatlanComPropertyInfoExtractor) AddInfoIntoProp(p *PropertyInfo) {
	p.Condition = e.Condition
	p.BuiltIn = e.BuiltIn
	p.NumOfFloors = e.NumOfFloors
	p.Parking = e.Parking
	p.Heating = e.Heating
	p.AirConditioning = e.AirConditioning
	p.ToiletAndBathroom = e.ToiletAndBathroom
}

type IngatlanComMainInfoExtractor struct {
	HouseArea, LotArea, NumOfRooms int
	Price, PricePerSqrMeter        float64
}

func (m *IngatlanComMainInfoExtractor) Predicate(n *html.Node) bool {
	if n.Type != html.ElementNode || n.Data != "div" {
		return false
	}

	for _, attr := range n.Attr {
		if attr.Key == "class" && attr.Val == "parameters" {
			return true
		}
	}
	return false
}

func (m *IngatlanComMainInfoExtractor) ProcessNode(n *html.Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		fc := c.FirstChild

		if isPriceHeaderNode(fc) {
			priceNode := findParameterValuesClassAmongSiblings(fc)

			price, err := extractPriceFromNode(priceNode)
			if err != nil {
				log.Println(err)
				continue
			}

			m.Price = price
		}
		if isNodeParameterTitle(fc) {
			if fc.FirstChild == nil {
				break
			}
			paramName := strings.TrimSpace(fc.FirstChild.Data)
			valueNode := findParameterValuesClassAmongSiblings(fc)
			if valueNode == nil {
				log.Printf("did not found value node for '%s'", paramName)
				break
			}

			val, err := extractIntValueFromNode(valueNode)
			if err != nil {
				log.Printf("could not extract value from node: %s\n", err)
			}

			switch paramName {
			case "Alapterület":
				m.HouseArea = val
			case "Telekterület":
				m.LotArea = val
			case "Szobák":
				m.NumOfRooms = val
			}

		}
	}

	m.PricePerSqrMeter = (m.Price / float64(m.HouseArea)) * 1000000.0 // converting it to millionHUF -> HUF
}

func (m *IngatlanComMainInfoExtractor) AddInfoIntoProp(p *PropertyInfo) {
	p.HouseArea = m.HouseArea
	p.LotArea = m.LotArea
	p.NumOfRooms = m.NumOfRooms
	p.Price = m.Price
	p.PricePerSqrMeter = m.PricePerSqrMeter
}

type IngatlanComAddressExtractor struct {
	Address string
}

func (a *IngatlanComAddressExtractor) Predicate(n *html.Node) bool {
	if !isH1Node(n) {
		return false
	}

	if doesClassAttrContainsVal(n, "address") {
		return true
	}
	return false
}

func (a *IngatlanComAddressExtractor) ProcessNode(n *html.Node) {
	textNode := n.FirstChild
	if textNode == nil {
		return
	}

	a.Address = strings.TrimSpace(textNode.Data)
}

func (a *IngatlanComAddressExtractor) AddInfoIntoProp(p *PropertyInfo) {
	p.Address = a.Address
}

func CrateIngatlanQueryUrl(c Config) string {
	url := JoinUri(IngatlanBaseUrl, "lista/elado")
	// lakas/haz
	url += fmt.Sprintf("+%s+%d-%d-m2+%d-%d-mFt", c.Type, c.MinSize, c.MaxSize, c.MinPrice, c.MaxPrice)
	for _, dist := range c.Districts {
		url += fmt.Sprintf("+%s-ker", dist)
	}
	return url
}

func extractIntValueFromNode(n *html.Node) (int, error) {
	if n.FirstChild == nil {
		return 0, errors.New("unknown format, expected node does not exists")
	}

	if n.FirstChild.Data != "span" {
		return 0, fmt.Errorf("unknown format, expected 'span' node, got %s", n.FirstChild.Data)
	}

	valAsString := n.FirstChild.FirstChild.Data

	val, err := extractInValueFromString(valAsString)
	if err != nil {
		return 0, err
	}

	return val, nil
}

func extractInValueFromString(s string) (int, error) {
	// cases: '43 m2', '4'

	s = strings.Split(s, " ")[0]

	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	return val, nil
}

func extractPriceFromNode(priceNode *html.Node) (float64, error) {
	if priceNode == nil {
		return 0.0, errors.New("node containing the price not found")
	}
	if priceNode.FirstChild == nil || priceNode.FirstChild.Data != "span" {
		return 0.0, errors.New("node containing the price not found")
	}

	// div > span > text
	priceAsString := priceNode.FirstChild.FirstChild.Data

	price, err := extractPriceFromString(priceAsString)
	if err != nil {
		return 0.0, err
	}

	return price, nil
}

func extractPriceFromString(s string) (float64, error) {
	priceAsString := strings.Split(s, " ")[0]
	priceAsString = strings.Replace(priceAsString, ",", ".", 1)

	price, err := strconv.ParseFloat(priceAsString, 64)
	if err != nil {
		return 0.0, errors.New("could not convert %s to float")
	}

	return price, nil
}

func copyInfoFromMainExtractorToProp(m *IngatlanComMainInfoExtractor, p *PropertyInfo) {
	p.Price = m.Price
	p.LotArea = m.LotArea
	p.HouseArea = m.HouseArea
	p.NumOfRooms = m.NumOfRooms
	p.PricePerSqrMeter = m.PricePerSqrMeter
}

func isNodeListingLink(n *html.Node) bool {
	return doesClassAttrContainsVal(n, "listing__link")
}
