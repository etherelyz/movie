package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

type Movie struct {
	ID    int
	Title string
}

type MovieInfo struct {
	Title         string
	Description   string
	SimilarTitles []string
}

func index(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

func searchSuggestion(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if title == "" {
		return
	}

	db, err := sql.Open("sqlite", "db.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := "SELECT id, title FROM movies_movie WHERE LOWER(title) LIKE LOWER('%' || ? || '%')"
	rows, err := db.Query(query, title)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var movieSuggestions []Movie
	count := 0
	for rows.Next() {
		if count >= 5 {
			break
		}

		var movie Movie
		if err = rows.Scan(&movie.ID, &movie.Title); err != nil {
			log.Fatal(err)
		}
		movieSuggestions = append(movieSuggestions, movie)

		count++
	}

	if len(movieSuggestions) == 0 {
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/search-suggestion.html"))
	tmpl.Execute(w, movieSuggestions)
}

func getMovieInfo(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("movie_id")

	db, err := sql.Open("sqlite", "db.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var title string
	var description string
	query := "SELECT title, description FROM movies_movie WHERE id = ?"
	err = db.QueryRow(query, id).Scan(&title, &description)
	if err != nil {
		log.Fatal(err)
	}

	query = "SELECT m2.title FROM movies_similarmovie AS s JOIN movies_movie AS m2 ON s.movie_2_id = m2.id WHERE s.movie_1_id = ?"
	rows, err := db.Query(query, id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var similarTitles []string
	for rows.Next() {
		var similarTitle string
		if err = rows.Scan(&similarTitle); err != nil {
			log.Fatal(err)
		}
		similarTitles = append(similarTitles, similarTitle)
	}

	data := MovieInfo{
		Title:         title,
		Description:   description,
		SimilarTitles: similarTitles,
	}

	tmpl := template.Must(template.ParseFiles("templates/get-movie-info.html"))
	tmpl.Execute(w, data)
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", index)
	http.HandleFunc("/search_suggestion/", searchSuggestion)
	http.HandleFunc("/movie_info/", getMovieInfo)

	log.Fatal(http.ListenAndServe("localhost:6969", nil))
}
