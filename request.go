package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	prd = NewProduct()
)

type Product struct {
}

func NewProduct() *Product {
	return &Product{}
}

func (p *Product) GetProduct(ctx context.Context, param *RequestParam) (string, error) {
	var (
		urlStr = fmt.Sprintf("https://www.google.com/shopping/product/%s?gl=%s&hl=%s&prds=pid:%s&sourceid=chrome&ie=UTF-8", param.ProductId, param.Gl, param.Hl, param.ProductId)
	)
	if param.Filter != "" {
		urlStr = fmt.Sprintf("https://www.google.com/shopping/product/%s?gl=%s&hl=%s&prds=pid:%s,%s&sourceid=chrome&ie=UTF-8", param.ProductId, param.Gl, param.Hl, param.ProductId, param.Filter)
	}
	res, err := getData(ctx, urlStr)
	if err != nil {
		log.Println(err)
		return "", err
	}
	document, _ := goquery.NewDocumentFromReader(strings.NewReader(res))

	var productInfo ProductInfo
	specsResults := SpecsResults(document)
	productInfo.ProductResults = ProductResults(document, param.ProductId)
	productInfo.SellersResults.OnlineSellers = SellersResults(document)
	productInfo.ReviewsResults.Reviews = ReviewsResultsReviews(document)
	productInfo.ReviewsResults.Filters = ReviewsResultsFilters(document)
	productInfo.ReviewsResults.Ratings = ReviewsResultsRatings(document)
	productInfo.SpecsResults = &specsResults
	resultBytes, _ := json.Marshal(productInfo)
	return string(resultBytes), nil
}
func getData(ctx context.Context, url string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/jpeg,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")
	request.Header.Set("Cache-Control", "no-cache")
	request.Header.Set("Pragma", "no-cache")
	request.Header.Set("Priority", "u=0, i")
	request.Header.Set("Sec-Ch-Ua", `"Not A(Brand";v="8", "Chromium";v="132", "Google Chrome";v="132"`)
	request.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	request.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	request.Header.Set("Sec-Fetch-Dest", "document")
	request.Header.Set("Sec-Fetch-Mode", "navigate")
	request.Header.Set("Sec-Fetch-Site", "none")
	request.Header.Set("Sec-Fetch-User", "?1")
	request.Header.Set("Upgrade-Insecure-Requests", "1")
	do, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer do.Body.Close()
	body, err := io.ReadAll(do.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

type ProductInfo struct {
	ProductResults ProductResultsInfo `json:"product_results"`
	SellersResults struct {
		OnlineSellers []OnlineSeller `json:"online_sellers"`
	} `json:"sellers_results"`
	SpecsResults   *SpecsResultsInfo `json:"specs_results,omitempty"`
	ReviewsResults struct {
		Ratings []Rating `json:"ratings"`
		Filters []Filter `json:"filters"`
		Reviews []Review `json:"reviews"`
	}
}

type ProductResultsInfo struct {
	ProductId string   `json:"product_id"`
	Title     string   `json:"title"`
	Price     []string `json:"price"`
	//Conditions    []string              `json:"conditions,omitempty"`
	TypicalPrices *TypicalPrices              `json:"typical_prices,omitempty"`
	Reviews       float64                     `json:"reviews"`
	Rating        float64                     `json:"rating"`
	Extensions    []string                    `json:"extensions"`
	Description   string                      `json:"description"`
	Media         []ProductResultsMedia       `json:"media"`
	Sizes         any                         `json:"sizes,omitempty"`
	Highlight     []string                    `json:"highlight"`
	Variations    map[string][]map[string]any `json:"variations,omitempty"`
}

type TypicalPrices struct {
	Low       string `json:"low"`
	High      string `json:"high"`
	ShowPrice string `json:"show_price"`
}

type ProductResultsMedia struct {
	Type string `json:"type"`
	Link string `json:"link"`
}

type SizeInfo struct {
	Link      string `json:"link"`
	ProductId string `json:"product_id"`
}

func ProductResults(dc *goquery.Document, productId string) ProductResultsInfo {
	var (
		productResultsInfo ProductResultsInfo
	)
	productResultsInfo.ProductId = productId
	title := dc.Find("div[class='LDQll']").Children().Text()
	productResultsInfo.Title = title

	price := []string{}
	dc.Find("div[class='pspo-fade']").Each(func(i int, selection *goquery.Selection) {
		price = append(price, selection.Children().Eq(0).Children().Eq(0).Children().Eq(0).Find("span[class='g9WBQb']").Text())
	})
	productResultsInfo.Price = price
	typicalPricesLow := dc.Find("div[jscontroller='G0u7Ld']").Children().Eq(1).Children().Eq(2).Children().Eq(1).Children().Eq(0).Text()
	typicalPricesHigh := dc.Find("div[jscontroller='G0u7Ld']").Children().Eq(1).Children().Eq(2).Children().Eq(1).Children().Eq(1).Text()
	typicalPricesShownPrice := dc.Find("div[jscontroller='G0u7Ld']").Children().Eq(1).Children().Eq(0).Find("a").Children().Eq(0).Text()
	if typicalPricesShownPrice != "" {
		productResultsInfo.TypicalPrices = &TypicalPrices{
			Low:       typicalPricesLow,
			High:      typicalPricesHigh,
			ShowPrice: typicalPricesShownPrice,
		}
	}

	reviewsNode := dc.Find("section[id='reviews']").Children().Eq(1).Children().Eq(0).Children().Eq(0)
	ratingStr := reviewsNode.Children().Eq(0).Children().Eq(0).Children().Eq(0).Text()
	reviewsStr := reviewsNode.Children().Eq(0).Children().Eq(0).Children().Eq(2).Text()
	rating, _ := strconv.ParseFloat(ratingStr, 64)
	productResultsInfo.Rating = rating
	reviewsStr = strings.Split(reviewsStr, " ")[0]
	reviewsStr = strings.Replace(reviewsStr, ",", "", -1)
	reviews, _ := strconv.ParseFloat(reviewsStr, 64)
	productResultsInfo.Reviews = reviews

	extensions := []string{}
	dc.Find("section[class='KS6Dwf']").Children().Eq(1).Find("span[class='OA4wid']").Each(func(i int, selection *goquery.Selection) {
		extensions = append(extensions, selection.Text())
	})
	productResultsInfo.Extensions = extensions
	description := dc.Find("section[class='KS6Dwf']").Children().Eq(1).Children().Eq(0).Children().Eq(0).
		Children().Eq(1).Text()
	productResultsInfo.Description = description
	var productResultsMedia []ProductResultsMedia
	dc.Find("div[jscontroller='bjweU']").Children().Eq(0).Children().Eq(0).Children().Eq(0).Children().Eq(0).Find("img").Each(func(i int, selection *goquery.Selection) {
		link, _ := selection.Attr("src")
		productResultsMedia = append(productResultsMedia, ProductResultsMedia{
			Type: "image",
			Link: link,
		})
	})
	productResultsInfo.Media = productResultsMedia
	//size
	var sizes = make(map[string]any)
	// default choose size
	firstSize := dc.Find("div[jscontroller='LkP0Fb']").Children().Eq(1).Children().Eq(0).Find("li").Text()
	// other size
	dc.Find("div[jscontroller='LkP0Fb']").Children().Eq(1).Children().Eq(0).Find("a").Each(func(i int, selection *goquery.Selection) {
		link, _ := selection.Attr("href")
		split := strings.Split(link, "?")
		if len(split) != 0 {
			pidArray := strings.Split(split[0], "/")
			pId := pidArray[len(pidArray)-1]
			sizes[selection.Text()] = SizeInfo{
				Link:      fmt.Sprintf("https://www.google.com%s", link),
				ProductId: pId,
			}

		}
	})
	if firstSize != "" {
		sizes[firstSize] = SizeInfo{
			Link:      fmt.Sprintf("https://www.google.com/shopping/product/%s?gl=us&hl=en&sourceid=chrome&ie=UTF-8", productId),
			ProductId: productId,
		}
	}
	productResultsInfo.Sizes = sizes
	dc.Find("uiRekc sh-dvp__variant-picker").Children().Each(func(i int, selection *goquery.Selection) {

	})
	// highlights
	var highlights []string
	dc.Find("div[jscontroller='q1x7of']").Find("li").Each(func(i int, selection *goquery.Selection) {
		highlights = append(highlights, selection.Text())
	})
	varResults := VarResults(dc)
	productResultsInfo.Variations = varResults
	productResultsInfo.Highlight = highlights
	return productResultsInfo
}

type OnlineSeller struct {
	Position         int                 `json:"position"`
	Name             string              `json:"name"`
	PaymentMethods   string              `json:"payment_methods"`
	Link             string              `json:"link"`
	DirectLink       string              `json:"direct_link"`
	DetailsAndOffers []map[string]string `json:"details_and_offers"`
	BasePrice        string              `json:"base_price"`
	AdditionalPrice  AdditionalPrice     `json:"additional_price"`
	Badge            string              `json:"badge,omitempty"`
	TotalPrice       string              `json:"total_price"`
}
type AdditionalPrice struct {
	Shipping string `json:"shipping"`
	Tax      string `json:"tax"`
}

func SellersResults(dc *goquery.Document) []OnlineSeller {
	var (
		onlineSellers []OnlineSeller
	)
	dc.Find("tr[jscontroller='d5bMlb']").Each(func(i int, selection *goquery.Selection) {
		var name string
		selection.Children().Eq(0).Children().Eq(0).Children().Eq(0).Contents().Each(func(i int, selection *goquery.Selection) {
			if goquery.NodeName(selection) == "#text" {
				name = selection.Text()
			}
		})
		paymentMethods := selection.Children().Eq(0).Children().Eq(0).Children().Eq(1).Text()

		selection.Children().Eq(0).Children().Eq(0).Children()
		link, _ := selection.Children().Eq(0).Children().Eq(0).Children().Eq(0).Attr("href")
		detailsAndOffersText := selection.Children().Eq(1).Text()
		basePrice := selection.Children().Eq(2).Find("span[class='g9WBQb fObmGc']").Text()
		additionalPriceShipping := selection.Children().Eq(3).Children().Eq(0).Children().Eq(1).Children().Eq(1).Children().Eq(0).Children().Eq(1).Children().Eq(1).Text()
		additionalPriceTax := selection.Children().Eq(3).Children().Eq(0).Children().Eq(1).Children().Eq(1).Children().Eq(0).Children().Eq(2).Children().Eq(1).Text()
		totalPrice := selection.Children().Eq(3).Children().Eq(0).Children().Eq(1).Children().Eq(1).Children().Eq(0).Children().Eq(4).Children().Eq(1).Text()
		badge := selection.Children().Eq(2).Children().Eq(0).Find("span[class='XhDkmd']").Text()
		link = fmt.Sprintf("https://www.google.com/%s", link)
		parse, _ := url.Parse(link)
		directLink := parse.Query().Get("q")
		onlineSellers = append(onlineSellers, OnlineSeller{
			Position:       i + 1,
			Name:           name,
			PaymentMethods: paymentMethods,
			Link:           link,
			DirectLink:     directLink,
			DetailsAndOffers: []map[string]string{
				{
					"text": detailsAndOffersText,
				},
			},
			BasePrice: basePrice,
			AdditionalPrice: AdditionalPrice{
				Shipping: additionalPriceShipping,
				Tax:      additionalPriceTax,
			},
			Badge:      badge,
			TotalPrice: totalPrice,
		})

	})
	return onlineSellers
}

type SpecsResultsInfo struct {
	General map[string]string `json:"general,omitempty"`
}

func SpecsResults(dc *goquery.Document) SpecsResultsInfo {
	var (
		specsResultsInfo SpecsResultsInfo
	)
	dc.Find("section[id='specs']").Find("div[class='AspnF'] tbody").Children().Each(func(i int, selection *goquery.Selection) {
		if _, ok := selection.Attr("class"); !ok {
			key := selection.Children().Eq(0).Text()
			key = strings.Replace(key, " ", "_", -1)
			key = strings.ToLower(key)
			val := selection.Children().Eq(1).Text()
			if specsResultsInfo.General == nil {
				specsResultsInfo.General = make(map[string]string)
			}
			specsResultsInfo.General[key] = val
		}
	})
	return specsResultsInfo
}

type Rating struct {
	Stars  int `json:"stars"`
	Amount int `json:"amount"`
}
type Filter struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

func ReviewsResultsRatings(dc *goquery.Document) []Rating {
	var (
		ratings []Rating
	)
	dc.Find("div[class='sh-rov__hist-container']").Children().Each(func(i int, selection *goquery.Selection) {
		amountStr := selection.Find("div[class='vL3wxf']").Text()
		amountStr = strings.Replace(amountStr, ",", "", -1)
		amountStr = strings.Split(amountStr, " ")[0]
		amount, _ := strconv.Atoi(amountStr)
		starsStr := selection.Find("div[class='rOdmxf']").Text()
		starsStr = strings.Replace(starsStr, ",", "", -1)
		starsStr = strings.Split(starsStr, " ")[0]
		stars, _ := strconv.Atoi(starsStr)
		ratings = append(ratings, Rating{
			Stars:  stars,
			Amount: amount,
		})
	})
	return ratings
}

func ReviewsResultsFilters(dc *goquery.Document) []Filter {
	var (
		filters []Filter
	)
	dc.Find("div[class='QPborb']").Children().Each(func(i int, selection *goquery.Selection) {
		label := selection.Children().Eq(0).Children().Eq(0).Children().Eq(0).Text()
		countStr := selection.Children().Eq(0).Children().Eq(0).Children().Eq(1).Find("span").Text()
		count, _ := strconv.Atoi(countStr)
		filters = append(filters, Filter{
			Label: label,
			Count: count,
		})
	})
	return filters
}

type Review struct {
	Position int    `json:"position"`
	Title    string `json:"title,omitempty"`
	Date     string `json:"date"`
	Rating   int    `json:"rating,omitempty"`
	Source   string `json:"source"`
	Content  string `json:"content"`
}

func ReviewsResultsReviews(dc *goquery.Document) []Review {
	var (
		review []Review
	)
	dc.Find("div[class='XBANlb']").Each(func(i int, selection *goquery.Selection) {
		title := selection.Find("div[class='P3O8Ne less-spaced']").Text()
		ratingStr, _ := selection.Find("div[class='less-spaced OP1Nkd nMkOOb']").Children().Eq(0).Attr("aria-label")
		split := strings.Split(ratingStr, " ")
		var rating int
		if len(split) > 0 {
			rating, _ = strconv.Atoi(split[0])
		}
		date := selection.Find("div[class='less-spaced OP1Nkd nMkOOb']").Text()
		content := selection.Find("div[class='g1lvWe']").Text()
		source := selection.Find("div[class='sPPcBf']").Text()
		review = append(review, Review{
			Position: i + 1,
			Title:    title,
			Date:     date,
			Rating:   rating,
			Source:   source,
			Content:  content,
		})
	})
	return review
}

func VarResults(dc *goquery.Document) map[string][]map[string]any {
	var (
		choose = make(map[string]any)
		colors = make([]map[string]any, 0)
	)
	data := CapacityAndConnectivity(dc)
	dc.Find("div[class='sh-dc__scroller']").Children().Eq(0).Find("div[class='a8P5sc']").Each(func(i int, selection *goquery.Selection) {
		if i == 0 {
			src, _ := selection.Find("img[class='TL92Hc']").Attr("src")
			choose["thumbnail"] = src
			choose["selected"] = true
			colors = append(colors, choose)
			return
		}
		color := make(map[string]any)

		thumbnail, _ := selection.Find("img").Attr("src")
		link, _ := selection.Find("a").Attr("href")
		name, _ := selection.Find("a").Attr("aria-label")
		color["name"] = name
		color["thumbnail"] = thumbnail
		split := strings.Split(link, "?")
		pidArray := strings.Split(split[0], "/")
		if len(pidArray) > 0 {
			pid := pidArray[len(pidArray)-1]
			color["product_id"] = pid
		}
		link = fmt.Sprintf("https://www.google.com%s", link)
		color["link"] = link
		colors = append(colors, color)
	})
	dc.Find("div[class='eyAVeb']").Contents().Each(func(i int, selection *goquery.Selection) {
		if goquery.NodeName(selection) == "#text" {
			data[strings.ToLower(strings.Split(selection.Text(), ":")[0])] = colors
		}
	})
	return data
}

func CapacityAndConnectivity(dc *goquery.Document) map[string][]map[string]any {
	var (
		all = make(map[string][]map[string]any)
	)
	dc.Find("div[jscontroller='LkP0Fb']").Each(func(i int, selection *goquery.Selection) {
		key := selection.Find("button label").Text()
		selectedName := selection.Find("ul li").Text()
		data := make([]map[string]any, 0)
		data = append(data, map[string]any{
			"name":     selectedName,
			"selected": true,
		})
		selection.Find("ul a").Each(func(i int, s *goquery.Selection) {
			resp := make(map[string]any)
			link, _ := s.Attr("href")
			name := s.Text()
			resp["name"] = name
			split := strings.Split(link, "?")
			pidArray := strings.Split(split[0], "/")
			if len(pidArray) > 0 {
				pid := pidArray[len(pidArray)-1]
				resp["product_id"] = pid
			}
			link = fmt.Sprintf("https://www.google.com%s", link)
			resp["link"] = link
			data = append(data, resp)
		})
		all[strings.ToLower(key)] = data
	})
	return all
}
