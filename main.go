package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
)

type Game struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	ImageURL    string   `json:"imageUrl"`
	GameURL     string   `json:"gameUrl"`
	ReleaseDate string   `json:"release"`
	Tags        []string `json:"tags"`
}

func scrap(w http.ResponseWriter, r *http.Request) {
	c := colly.NewCollector()

	gamesList := make(map[string]Game)

	c.OnHTML("div[class=template-games__listing-category]", func(e *colly.HTMLElement) {
		e.ForEach("article", func(_ int, el *colly.HTMLElement) {
			title := el.ChildText("h1.teaser__title")

			_, ok := gamesList[title]
			if !ok {
				typeGame := el.ChildText(".teaser__game-type")
				trimmedTypeAfter := strings.TrimSuffix(typeGame, " games")
				trimmedType := strings.Split(trimmedTypeAfter, "Game type: ")[1]

				imageURL := el.ChildAttr(".teaser__header-image", "style")
				imageURLAfter := strings.TrimSuffix(imageURL, "');")
				trimmedImageURL := strings.Split(imageURLAfter, "background-image: url('")[1]

				release := el.ChildText(".teaser__launch")
				trimmedRelease := strings.Split(release, "Release: ")[1]

				gameURL := el.ChildAttr(".teaser__book-demo", "data-demo")

				tags := el.ChildTexts(".teaser__header-labels")

				for i := len(tags) - 1; i >= 0; i-- {
					if tags[i] == "" {
						tags = append(tags[:i], tags[i+1:]...)
					}

				}

				descCollector := colly.NewCollector()

				descCollector.OnRequest(func(r *colly.Request) {
					fmt.Println("Visiting", r.URL.String())
				})

				descCollector.OnHTML("body", func(ed *colly.HTMLElement) {

					desc := ed.ChildTexts(".paragraph-styles")
					gameStruct := Game{
						Name:        title,
						Description: desc[0],
						Type:        trimmedType,
						ImageURL:    trimmedImageURL,
						GameURL:     gameURL,
						ReleaseDate: trimmedRelease,
						Tags:        tags,
					}
					gamesList[title] = gameStruct
				})

				descCollector.Visit(el.ChildAttr(".teaser__button", "href"))
			}
		})
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnScraped(func(r *colly.Response) {
		json.NewEncoder(w).Encode(gamesList)
	})

	c.Visit("https://www.onetouch.io/games/")
}

func main() {
	http.HandleFunc("/", scrap)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
