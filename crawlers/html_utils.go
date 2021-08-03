package crawlers

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"golang.org/x/net/html"
)

type HtmlNodeProcessor interface {
	Predicate(n *html.Node) bool
	ProcessNode(n *html.Node)
}

type LinkExtractor interface {
	HtmlNodeProcessor
	GetLinks() []string
}

type ListingPagesExtractor interface {
	HtmlNodeProcessor
	MaxPageNumber() int
	NextPageFormat() string
}

type PageDataExtractor interface {
	HtmlNodeProcessor
	AddInfoIntoProp(prop *PropertyInfo)
}

func CollectInfoFromPropertyPage(url string, propChan chan<- PropertyInfo, extractors ...PageDataExtractor) {
	resp, err := sendGetRequest(url)
	if err != nil {
		log.Printf("error when opening '%s': '%s'", url, err)
	}
	if resp.StatusCode == 404 {
		log.Printf("could not find page '%s'", url)
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Printf("error parsing response body: '%s'\n", err)
	}

	nodeProcessors := convertPageDataExtractorsToHtmlNodeProcessors(extractors...)
	traverseHtmlTreeAndExecuteExtractors(doc, nodeProcessors...)

	propInfo := PropertyInfo{}
	for _, extractor := range extractors {
		extractor.AddInfoIntoProp(&propInfo)
	}

	propChan <- propInfo
}

func convertPageDataExtractorsToHtmlNodeProcessors(extractors ...PageDataExtractor) []HtmlNodeProcessor {
	processors := make([]HtmlNodeProcessor, len(extractors))
	for i, ex := range extractors {
		processors[i] = ex
	}
	return processors
}
func traverseHtmlTreeAndExtractString(doc *html.Node, extractor HtmlNodeProcessor) {
	var f func(n *html.Node)

	f = func(n *html.Node) {
		if extractor.Predicate(n) {
			extractor.ProcessNode(n)
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
}

func findHrefAttribute(n *html.Node) string {
	for _, attr := range n.Attr {
		if attr.Key == "href" {
			return attr.Val
		}
	}
	return ""
}

func doesClassAttrContainsVal(n *html.Node, val string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			return strings.Contains(attr.Val, val)
		}
	}
	return false
}

func getChildsParamNameAndValue(n *html.Node) (string, string, error) {
	var name, value *html.Node

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if doesClassAttrContainsVal(c, "parameterName") {
			name = c
		} else if doesClassAttrContainsVal(c, "parameterValue") {
			value = c
		}
	}

	if name == nil || value == nil {
		return "", "", errors.New("parameter name or value not found")
	}

	return strings.TrimSpace(name.FirstChild.Data), strings.TrimSpace(value.FirstChild.Data), nil
}

func isNodeTypeOf(n *html.Node, t string) bool {
	return n.Type == html.ElementNode && n.Data == t
}
func isLinkNode(n *html.Node) bool {
	return isNodeTypeOf(n, "a")
}

func isDivNode(n *html.Node) bool {
	return isNodeTypeOf(n, "div")
}

func isH1Node(n *html.Node) bool {
	return isNodeTypeOf(n, "h1")
}

func findParameterValuesClassAmongSiblings(n *html.Node) *html.Node {
	for s := n.NextSibling; s != nil; s = s.NextSibling {
		for _, attr := range s.Attr {
			if attr.Key != "class" {
				continue
			}

			if attr.Val == "parameterValues" {
				return s
			}
		}
	}
	return nil
}

func isPriceHeaderNode(n *html.Node) bool {
	return isLinkNode(n) && n.FirstChild != nil && strings.TrimSpace(n.FirstChild.Data) == "Hitelre van szükséged? Kalkulálj!"
}

func isNodeParameterTitle(n *html.Node) bool {
	return isDivNode(n) && doesClassAttrContainsVal(n, "parameterTitle")
}

func traverseHtmlTreeAndExecuteExtractors(root *html.Node, extractors ...HtmlNodeProcessor) {
	for _, extractor := range extractors {
		traverseHtmlTreeAndExtractString(root, extractor)
	}
}

func CollectPropertyLinksForQuery(url string, le LinkExtractor, lpe ListingPagesExtractor) error {
	err := extractListingPagesInfo(lpe, url)
	if err != nil {
		return err
	}

	urlTemplate := url + lpe.NextPageFormat()
	extractLinksFromListingPages(urlTemplate, lpe, le)

	return nil
}

func extractLinksFromListingPages(urlTemplate string, lpe ListingPagesExtractor, le LinkExtractor) error {
	for i := 1; i <= lpe.MaxPageNumber(); i++ {
		queryUrl := fmt.Sprintf(urlTemplate, i)
		log.Printf("reading from url %s\n", queryUrl)

		resp, err := sendGetRequest(queryUrl)
		if err != nil {
			return fmt.Errorf("error when requesting content from %s: '%s'", queryUrl, err)
		}
		if resp.StatusCode == 404 {
			log.Printf("stopped at page %d\n", i)
		}
		defer resp.Body.Close()

		doc, err := html.Parse(resp.Body)
		if err != nil {
			return err
		}

		traverseHtmlTreeAndExtractString(doc, le)
	}

	return nil
}

func extractListingPagesInfo(lpe ListingPagesExtractor, url string) error {

	resp, err := sendGetRequest(url)
	if err != nil {
		return fmt.Errorf("error when requesting content from %s: '%s'", url, err)
	}
	if resp.StatusCode == 404 {
		log.Printf("could not find page at url: %s\n", url)
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return err
	}

	traverseHtmlTreeAndExecuteExtractors(doc, lpe)

	log.Printf("found %d page(s) of data, starting parsing", lpe.MaxPageNumber())

	return nil
}
