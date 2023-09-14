package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func main() {
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{Views: engine})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{})
	})

	app.Get("/search", func(c *fiber.Ctx) error {
		ticker := c.Query("ticker")
		results := SearchTicker(ticker)

		return c.Render("results", fiber.Map{
			"Results": results,
		})
	})

	app.Listen(":3000")
}

const PoligonPath = "https://api.polygon.io"
const TickerPath = PoligonPath + "/v3/reference/tickers"
const DailyValuesPath = PoligonPath + "/v1/open-close"

var (
	ApiKey = "apiKey=" + os.Getenv("POLYGON_API_KEY")
)

type Stock struct {
	Ticker          string    `json:"ticker"`
	Name            string    `json:"name"`
	Market          string    `json:"market"`
	Locale          string    `json:"locale"`
	PrimaryExchange string    `json:"primary_exchange"`
	Type            string    `json:"type"`
	Active          bool      `json:"active"`
	Currency        string    `json:"currency_name"`
	LastUpdated     time.Time `json:"last_updated_utc"`
}

type SearchResult struct {
	Results []Stock `json:"results"`
}

func SearchTicker(ticker string) []Stock {
	body := Fetch(TickerPath + "?" + ApiKey + "&ticker=" + strings.ToUpper(ticker))
	data := SearchResult{}
	_ = json.Unmarshal(body, &data)
	return data.Results
}

type Values struct {
	Open  float64 `json:"open"`
	Close float64 `json:"close"`
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
}

func GetDailyValues(ticker string) Values {
	yesterday := time.Now().UTC().Add(-(time.Hour * 24)).Format("2006-01-02")
	body := Fetch(DailyValuesPath + "/" + strings.ToUpper(ticker) + "/" + yesterday + "?" + ApiKey)
	data := Values{}
	_ = json.Unmarshal(body, &data)
	return data
}

func Fetch(endpoint string) []byte {
	fmt.Fprintf(os.Stderr, "GET: %q\n", endpoint)

	resp, err := http.Get(endpoint)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fetch error %q: %v\n", endpoint, err)
		return nil
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Fetch Bad Response %q: %s\n", endpoint, resp.Status)
		return nil
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Read error %q: %v\n", endpoint, err)
		return nil
	}

	return data
}
