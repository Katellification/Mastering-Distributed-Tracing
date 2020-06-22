package main

import (
	"log"
	"net/http"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/Katellification/Mastering-Distributed-Tracing/Chapter04/go/exercise3a-tania/people"
	"github.com/Katellification/Mastering-Distributed-Tracing/Chapter04/go/lib/tracing"
)

var repo *people.Repository

// var tracer opentracing.Tracer

func main() {
	repo = people.NewRepository()
	defer repo.Close()

	tracer, closer := tracing.Init("go-3a-hello")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	http.HandleFunc("/sayHello/", handleSayHello)

	log.Print("Listening on http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleSayHello(w http.ResponseWriter, r *http.Request) {
	span := opentracing.GlobalTracer().StartSpan("say-hello")
	defer span.Finish()

	name := strings.TrimPrefix(r.URL.Path, "/sayHello/")
	greeting, err := SayHello(name, span)
	if err != nil {
		span.SetTag("error", true)
		span.LogFields(otlog.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	span.SetTag("response", greeting)
	w.Write([]byte(greeting))
}

// SayHello creates a greeting for the named person.
func SayHello(name string, span opentracing.Span) (string, error) {
	person, err := repo.GetPerson(name, span)
	if err != nil {
		return "", err
	}

	span.LogKV(
		"name", person.Name,
		"title", person.Title,
		"description", person.Description,
	)

	return FormatGreeting(
		person.Name,
		person.Title,
		person.Description,
		span,
	), nil
}

// FormatGreeting combines information about a person into a greeting string.
func FormatGreeting(name, title, description string, span opentracing.Span) string {
	span = opentracing.GlobalTracer().StartSpan(
		"format-greeting",
		opentracing.ChildOf(span.Context()),
	)
	defer span.Finish()
	response := "Hello, "
	if title != "" {
		response += title + " "
	}
	response += name + "!"
	if description != "" {
		response += " " + description
	}
	return response
}
