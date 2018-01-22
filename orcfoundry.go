package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/alimac/orc"
	"github.com/gorilla/mux"
)

var templates map[string]*template.Template

// Compile view templates
func init() {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}
	templates["index"] = template.Must(template.ParseFiles("templates/index.html",
		"templates/base.html"))
	templates["add"] = template.Must(template.ParseFiles("templates/add.html",
		"templates/base.html"))
	templates["edit"] = template.Must(template.ParseFiles("templates/edit.html",
		"templates/base.html"))
	templates["view"] = template.Must(template.ParseFiles("templates/view.html",
		"templates/base.html"))
}

// Render templates for the given name, template definition and data object
func renderTemplate(w http.ResponseWriter, name string, template string, viewModel interface{}) {
	// Ensure template exists in the map
	tmpl, _ := templates[name]

	err := tmpl.ExecuteTemplate(w, template, viewModel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// getOrc
func (a *App) getOrc(w http.ResponseWriter, r *http.Request) {
	// read value from route variable
	vars := mux.Vars(r)
	key := vars["id"]
	status, items := getItems(key)

	if status == http.StatusOK {
		if key != "" {
			var orc Orc
			json.Unmarshal(items, &orc)
			renderTemplate(w, "view", "base", OrcModel{orc, key})
		} else {
			var orcs []Orc
			json.Unmarshal(items, &orcs)
			renderTemplate(w, "index", "base", orcs)
		}
	} else {
		http.Error(w, "Could not find the Orc to view", status)
	}
}

// addOrc
func (a *App) addOrc(w http.ResponseWriter, r *http.Request) {
	var viewModel OrcModel
	viewModel = OrcModel{Orc{orc.Forge("name"), orc.Forge("greeting"),
		orc.Forge("weapon"), time.Now()}, "0"}

	renderTemplate(w, "add", "base", viewModel)
}

// saveOrc
func (a *App) saveOrc(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	orc := Orc{
		r.PostFormValue("name"),
		r.PostFormValue("greeting"),
		r.PostFormValue("weapon"),
		time.Now(),
	}

	createItem(orc)
	http.Redirect(w, r, "/", 302)
}

// editOrc
func editOrc(w http.ResponseWriter, r *http.Request) {
	var viewModel OrcModel

	// read value from route variable
	vars := mux.Vars(r)
	key := vars["id"]

	if orc, ok := orcStore[key]; ok {
		viewModel = OrcModel{orc, key}
		renderTemplate(w, "edit", "base", viewModel)
	} else {
		http.Error(w, "Could not find the Orc to edit", http.StatusBadRequest)
	}
}

// updateOrc
func updateOrc(w http.ResponseWriter, r *http.Request) {
	// Read values from route variable
	vars := mux.Vars(r)
	key := vars["id"]
	var orcToUpdate Orc
	if orc, ok := orcStore[key]; ok {
		r.ParseForm()
		orcToUpdate.Name = r.PostFormValue("name")
		orcToUpdate.Greeting = r.PostFormValue("greeting")
		orcToUpdate.Weapon = r.PostFormValue("weapon")
		orcToUpdate.CreatedOn = orc.CreatedOn

		// delete existing item and add the updated item
		delete(orcStore, key)
		orcStore[key] = orcToUpdate
		http.Redirect(w, r, "/", 302)
	} else {
		http.Error(w, "Could not find the Orc to update", http.StatusBadRequest)
	}
}

// deleteOrc is a handler for "/orcs/delete/{id}" which deletes an item from the store
func deleteOrc(w http.ResponseWriter, r *http.Request) {
	// read value from the route Variable
	vars := mux.Vars(r)
	key := vars["id"]
	// Remove from the Store
	if _, ok := orcStore[key]; ok {
		// delete existing item
		delete(orcStore, key)
		http.Redirect(w, r, "/", 302)
	} else {
		http.Error(w, "Could not find the Orc to delete", http.StatusBadRequest)
	}
}

func main() {
	port := os.Getenv("PORT")
	host := ""

	if port == "" {
		port = "8080"
		host = "127.0.0.1"
	}

	app := App{}
	app.Initialize()
	app.Run(":"+port, host)
}
