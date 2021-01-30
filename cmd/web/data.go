package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"gopkg.in/yaml.v2"
)

// Source type
type Source struct {
	ID      string `yaml:"id"`
	Value   string `yaml:"value"`
	URLName string `yaml:"urlName"`
	URL     string `yaml:"url"`
}

// Fact type
type Fact struct {
	ID             string `yaml:"id"`
	Day            int    `yaml:"day"`
	Month          int    `yaml:"month"`
	Year           int    `yaml:"year"`
	Title          string `yaml:"title"`
	Content        string `yaml:"content"`
	ContentHTML    template.HTML
	ContentTwitter string `yaml:"contentTwitter"`
	Location       string `yaml:"location"`
	Geo            string `yaml:"geo"`
	GeoHTML        template.HTML
	People         string `yaml:"people"`
	Keywords       string `yaml:"keywords"`
	Image          string `yaml:"image"`
	ImageInfo      string `yaml:"imageInfo"`
	ImageHTML      template.HTML
	Sources        []Source `yaml:"sources"`
}

// Quote type
type Quote struct {
	Content string `yaml:"content"`
	Source  string `yaml:"source"`
}

// Book type
type Book struct {
	Author      string `yaml:"author"`
	Title       string `yaml:"title"`
	Year        string `yaml:"year"`
	Pubhause    string `yaml:"pubhause"`
	Where       string `yaml:"where"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	ISBN        string `yaml:"ISBN"`
	URL         string `yaml:"URL"`
	URLHTML     template.HTML
	URLName     string `yaml:"URLName"`
	Image       string `yaml:"image"`
	ImageHTML   template.HTML
	Pages       int `yaml:"pages"`
}

// YearFact type
type YearFact struct {
	Date      string
	DateMonth string
	Title     string
	URLHTML   template.HTML
}

// FactsByYear slice
var FactsByYear map[int][]YearFact

// PeopleFact type
type PeopleFact struct {
	Date      string
	DateMonth string
	Title     string
	URLHTML   template.HTML
}

// FactsByPeople slice
var FactsByPeople map[string][]PeopleFact

// DayFactTable map
var DayFactTable map[string]bool

func createDataCache() *cache.Cache {
	c := cache.New(5*time.Minute, 10*time.Minute)
	return c
}

// readFact func
func readFact(filename string) (*[]Fact, error) {
	var result []Fact
	var fact Fact

	fileBuf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(fileBuf)
	yamlDec := yaml.NewDecoder(r)

	for yamlDec.Decode(&fact) == nil {
		fact.ContentHTML = template.HTML(prepareFactHTML(fact.Content, fact.ID, fact.Sources))
		fact.ImageHTML = template.HTML(prepareImageHTML(fact.Image, fact.ImageInfo))
		if fact.Geo != "" {
			fact.GeoHTML = template.HTML(prepareGeoHTML(fact.Geo))
		}

		// uzupełnienie indeksu lat FactsByYear
		tmpYear := &YearFact{}
		tmpYear.Date = fmt.Sprintf("%04d-%02d-%02d", fact.Year, fact.Month, fact.Day)
		tmpYear.DateMonth = fmt.Sprintf("%d %s", fact.Day, monthName[fact.Month])
		tmpYear.Title = fact.Title
		tmpYear.URLHTML = template.HTML(prepareFactLinkHTML(fact.Month, fact.Day, fact.ID))
		if facts, ok := FactsByYear[fact.Year]; ok {
			facts = append(facts, *tmpYear)
			FactsByYear[fact.Year] = facts
		} else {
			facts := make([]YearFact, 0)
			facts = append(facts, *tmpYear)
			FactsByYear[fact.Year] = facts
		} // FactsByYear

		// uzupełnienie indeksu postaci FactsByPeople
		if fact.People != "" {
			tmpPeople := &PeopleFact{}
			tmpPeople.Date = fmt.Sprintf("%04d-%02d-%02d", fact.Year, fact.Month, fact.Day)
			tmpPeople.DateMonth = fmt.Sprintf("%d %s %d", fact.Day, monthName[fact.Month], fact.Year)
			tmpPeople.Title = fact.Title
			tmpPeople.URLHTML = template.HTML(prepareFactLinkHTML(fact.Month, fact.Day, fact.ID))
			persons := strings.Split(fact.People, ";")
			for _, person := range persons {
				person = strings.TrimSpace(person)
				if facts, ok := FactsByPeople[person]; ok {
					facts = append(facts, *tmpPeople)
					FactsByPeople[person] = facts
				} else {
					facts := make([]PeopleFact, 0)
					facts = append(facts, *tmpPeople)
					FactsByPeople[person] = facts
				}
			}
		} // FactsByPeople

		result = append(result, fact)
	}

	return &result, nil
}

// readQuote func
func readQuote() (*[]Quote, error) {
	var result []Quote
	var quote Quote

	filename, _ := filepath.Abs(dirExecutable + "/data/quotes.yaml")

	fileBuf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(fileBuf)
	yamlDec := yaml.NewDecoder(r)

	for yamlDec.Decode(&quote) == nil {
		result = append(result, quote)
	}

	return &result, nil
}

// readBook func
func readBook() (*[]Book, error) {
	var result []Book
	var book Book

	filename := dirExecutable + "/data/books.yaml"

	fileBuf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(fileBuf)
	yamlDec := yaml.NewDecoder(r)

	for yamlDec.Decode(&book) == nil {
		book.ImageHTML = template.HTML(prepareBookHTML(book.Image))
		book.URLHTML = template.HTML(prepareBookURLHTML(book.URL, book.URLName))
		result = append(result, book)
	}

	return &result, nil
}

// loadData - wczytuje podczas startu serwera dane do struktur w pamięci operacyjnej
func (app *application) loadData(path string) error {
	// wydarzenia
	fmt.Println("Wczytywanie bazy wydarzeń...")

	// mapa z listą dni - czy dla danego dnia istnieją wydarzenia w bazie
	DayFactTable = make(map[string]bool)

	// mapa dla indeksu lat
	FactsByYear = make(map[int][]YearFact)
	// mapa dla indeksu postaci
	FactsByPeople = make(map[string][]PeopleFact)

	dataFiles, _ := filepath.Glob(filepath.Join(path, "*-*.yaml"))
	for _, tFile := range dataFiles {
		name := filenameWithoutExtension(filepath.Base(tFile))
		facts, err := readFact(tFile)
		if err != nil {
			return err
		}
		numberOfFacts += len(*facts)
		DayFactTable[name] = true
		app.dataCache.Add(name, facts, cache.NoExpiration)
	}

	// sortowanie wydarzeń historycznych dla postaci
	for person, facts := range FactsByPeople {
		sort.Slice(facts, func(i, j int) bool {
			return facts[i].Date < facts[j].Date
		})
		FactsByPeople[person] = facts
	}

	// cytaty
	fmt.Println("Wczytywanie bazy cytatów...")

	quotes, err := readQuote()
	if err != nil {
		return err
	}
	app.dataCache.Add("quotes", quotes, cache.NoExpiration)

	// książki
	fmt.Println("Wczytywanie bazy książek...")

	books, err := readBook()
	if err != nil {
		return err
	}
	app.dataCache.Add("books", books, cache.NoExpiration)

	return nil
}

/* dayFact - funkcja zwraca fragment html z linkiem jeżeli dla danego dnia są wydarzenia
   historyczne w bazie, lub sam numer dnia (o szarym kolorze) jeżeli ich nie ma.
   Wykorzystywana w kalendarzu.
*/
func dayFact(month int, day int) template.HTML {
	name := fmt.Sprintf("%02d-%02d", month, day)

	if DayFactTable[name] {
		result := fmt.Sprintf(`<a href="/dzien/%d/%d">%d</a>`, month, day, day)
		return template.HTML(result)
	}

	return template.HTML(fmt.Sprintf(`<span style="color: DarkGrey;">%d</span>`, day))
}
