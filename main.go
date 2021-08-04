package main

import (
	"log"
	"sync"

	"github.com/PusztaiMate/ingatlan-crawler/crawlers"
)

func main() {
	config, err := crawlers.ReadJsonConfig("config.json")
	if err != nil {
		log.Fatalf("could not read config file, exiting: %s\n", err)
	}
	log.Printf("Config used: %#v", config)

	dunaHouseUrl := crawlers.CreateDunaHouseQueryUrl(config)
	dhle := crawlers.DunaHouseLinkCollector{}
	dhlpe := crawlers.DunaHouseListingPagesExtractor{}
	crawlers.CollectPropertyLinksForQuery(dunaHouseUrl, &dhle, &dhlpe)

	ingatlanUrl := crawlers.CrateIngatlanQueryUrl(config)
	ile := crawlers.IngatlanComLinkCollector{}
	ilpe := crawlers.IngatlanComListingPagesExtractor{}
	crawlers.CollectPropertyLinksForQuery(ingatlanUrl, &ile, &ilpe)

	dunaHouseLinks := dhle.GetLinks()
	log.Printf("Collected (%d) links from DunaHouse", len(dunaHouseLinks))
	ingatlanLinks := ile.GetLinks()
	log.Printf("Collected (%d) links from ingatlan.com", len(ingatlanLinks))

	propInfos := make(chan crawlers.PropertyInfo, len(dunaHouseLinks)+len(ingatlanLinks))
	var wg sync.WaitGroup

	dhge := crawlers.DunaHouseGeneralInfoExtractor{}
	dhme := crawlers.DunaHouseMainInfoExtractor{}
	for _, l := range dunaHouseLinks {
		linkToProp := crawlers.JoinUri(crawlers.DunaHouseBaseUrl, l)
		wg.Add(1)
		go func() {
			crawlers.CollectInfoFromPropertyPage(linkToProp, propInfos, &dhge, &dhme)
			defer wg.Done()
		}()
	}
	imie := crawlers.IngatlanComMainInfoExtractor{}
	ipie := crawlers.IngatlanComPropertyInfoExtractor{}
	iae := crawlers.IngatlanComAddressExtractor{}
	for _, l := range ingatlanLinks {
		linkToProp := crawlers.JoinUri(crawlers.IngatlanBaseUrl, l)
		wg.Add(1)
		go func() {
			crawlers.CollectInfoFromPropertyPage(linkToProp, propInfos, &imie, &ipie, &iae)
			defer wg.Done()
		}()
	}

	log.Println("Waiting for crawlers to finish collecting info from individual pages.")
	wg.Wait()
	close(propInfos)

	var props []crawlers.PropertyInfo
	for pi := range propInfos {
		if !crawlers.IsPropPresentInList(props, pi) {
			props = append(props, pi)
		}
	}
	log.Println("Finished waiting, starting processing data")

	filename := crawlers.CreateFileNameFromConfig(config, "")
	log.Printf("Collection finished, writing data to '%s'", filename)
	crawlers.WritePropertiesToCsv(filename, props)
	log.Println("Finished!")
}
