package main

import (
	"context"
	"github.com/scrapeless-ai/scrapeless-actor-sdk-go/scrapeless"
	"log"
	"net/http"
	"net/url"

	proxyModel "github.com/scrapeless-ai/scrapeless-actor-sdk-go/scrapeless/proxy"
)

var (
	client *http.Client
)

type RequestParam struct {
	ProductId string `json:"product_id" url:"product_id"`
	Gl        string `json:"gl" url:"gl"`
	Hl        string `json:"hl" url:"hl"`

	Filter  string `json:"filter" url:"filter"`
	OfferId string `json:"offer_id" url:"offer_id"` // v1 not use
}

func main() {
	// new actor
	actor := scrapeless.New(scrapeless.WithProxy(), scrapeless.WithStorage())
	defer actor.Close()
	var param = &RequestParam{}
	if err := actor.Input(param); err != nil {
		log.Fatal(err)
	}
	// get proxy url
	proxy, err := actor.Proxy.Proxy(context.TODO(), proxyModel.ProxyActor{
		Country:         "us",
		SessionDuration: 10,
	})

	if err != nil {
		panic(err)
	}
	parse, err := url.Parse(proxy)
	if err != nil {
		panic(err)
	}
	// init client with proxy
	client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(parse)}}

	data, err := prd.GetProduct(context.TODO(), param)
	if err != nil {
		log.Fatal(err)
	}

	ok, err := actor.Storage.GetDataset().AddItems(context.Background(), []map[string]any{
		{
			"data": data,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(ok)
}
