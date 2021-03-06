// Udostępnianie danych wydarzeń historycznych z serwisu XVI-wiek.pl
//
// Dokumentacja API
//
// Schemes: http
// BasePath: /api
// Version: 1.0.0
//
// Consumes:
// -
// Produces:
// - application/json
// swagger:meta
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
)

// SourceJSON type
type SourceJSON struct {
	Name string `json:"value"`
	URL  string `json:"url"`
}

// FactJSON type
// swagger:response factsResponse
type FactJSON struct {
	Date     string       `json:"date"`
	Title    string       `json:"title"`
	Content  string       `json:"content"`
	Location string       `json:"location"`
	Geo      string       `json:"geo"`
	People   string       `json:"people"`
	Keywords string       `json:"keywords"`
	Sources  []SourceJSON `json:"sources"`
}

// FactShortJSON type
// swagger:response factsShortResponse
type FactShortJSON struct {
	Date           string `json:"date"`
	ContentTwitter string `json:"content"`
}

func clearField(value string) string {
	value = prepareTextStyle(value, false)
	return value
}

func errorJSON(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	response, _ := json.Marshal(map[string]string{"message": msg})
	w.Write(response)
}

func toStructJSON(data interface{}) []FactJSON {
	factStruct := data.(*[]Fact)

	factsJSON := []FactJSON{}
	factJSON := FactJSON{}
	sourceJSON := SourceJSON{}

	for _, item := range *factStruct {
		factJSON.Date = fmt.Sprintf("%02d-%02d-%04d", item.Day, item.Month, item.Year)
		factJSON.Title = item.Title
		factJSON.Content = prepareTextStyle(item.Content, true)
		factJSON.Location = item.Location
		factJSON.People = item.People
		factJSON.Geo = item.Geo
		factJSON.Keywords = item.Keywords

		for _, itemSource := range item.Sources {
			sourceJSON.Name = itemSource.Value
			sourceJSON.URL = itemSource.URL
			factJSON.Sources = append(factJSON.Sources, sourceJSON)
			factJSON.Content = strings.Replace(factJSON.Content, "["+itemSource.ID+"]", "", -1)
		}

		factsJSON = append(factsJSON, factJSON)
	}

	return factsJSON
}

// toShortStructJSON
func toShortStructJSON(data interface{}) FactShortJSON {
	factStruct := data.(*[]Fact)
	factJSON := FactShortJSON{}

	choice := randomInt(0, len(*factStruct)-1)

	factJSON.Date = fmt.Sprintf("%02d-%02d-%04d", (*factStruct)[choice].Day, (*factStruct)[choice].Month, (*factStruct)[choice].Year)
	factJSON.ContentTwitter = (*factStruct)[choice].ContentTwitter

	return factJSON
}

// factResponseJSON
func factResponseJSON(w http.ResponseWriter, code int, data interface{}) {
	factsJSON := toStructJSON(data)

	response, _ := json.Marshal(factsJSON)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// factShortResponseJSON
func factShortResponseJSON(w http.ResponseWriter, code int, data interface{}) {
	factShortJSON := toShortStructJSON(data)

	response, _ := json.Marshal(factShortJSON)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// swagger:route GET /dzien/{month}/{day} dzien listaWydarzen
// zwraca wydarzenia historyczne dla wskazanego dnia
// responses:
//   200: factsResponse

// apiFactsByDay
func (app *application) apiFactsByDay(w http.ResponseWriter, r *http.Request) {
	month, err := strconv.Atoi(chi.URLParam(r, "month"))
	if err != nil || month < 1 || month > 12 {
		errorJSON(w, 404, "Błędne zapytanie lub brak danych")
		return
	}

	day, err := strconv.Atoi(chi.URLParam(r, "day"))
	if err != nil || day < 1 || day > 31 {
		errorJSON(w, 404, "Błędne zapytanie lub brak danych")
		return
	}

	var isCorrectDate bool = true

	if month == 2 && day > 29 {
		isCorrectDate = false
	}
	if (month == 4 || month == 6 || month == 9 || month == 11) && day > 30 {
		isCorrectDate = false
	}
	if !isCorrectDate {
		errorJSON(w, 404, "Błędne zapytanie lub brak danych")
		return
	}

	name := fmt.Sprintf("%02d-%02d", month, day)
	facts, ok := app.dataCache.Get(name)
	if ok {
		factResponseJSON(w, 200, facts)
	} else {
		errorJSON(w, 404, "Błędne zapytanie lub brak danych")
	}
}

// swagger:route GET /today today listaWydarzen
// zwraca wydarzenia historyczne dla bieżącego dnia
// responses:
//   200: factsResponse

// apiFactsToday
func (app *application) apiFactsToday(w http.ResponseWriter, r *http.Request) {
	name := fmt.Sprintf("%02d-%02d", int(time.Now().Month()), time.Now().Day())
	facts, ok := app.dataCache.Get(name)
	if ok {
		factResponseJSON(w, 200, facts)
	} else {
		errorJSON(w, 404, "Błędne zapytanie lub brak danych")
	}
}

// swagger:route GET /short short listaWydarzen
// zwraca skrócony opis wydarzenia historyczngo dla bieżąceo dnia
// responses:
//   200: factsShortResponse

// apiFactsShort
func (app *application) apiFactsShort(w http.ResponseWriter, r *http.Request) {
	name := fmt.Sprintf("%02d-%02d", int(time.Now().Month()), time.Now().Day())
	facts, ok := app.dataCache.Get(name)
	if ok {
		factShortResponseJSON(w, 200, facts)
	} else {
		errorJSON(w, 404, "Błędne zapytanie lub brak danych")
	}
}
