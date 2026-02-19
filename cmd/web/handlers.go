package main

import "net/http"

func (app *application) VirtualCardHandler(w http.ResponseWriter, r *http.Request) {
	app.infoLog.Println("VirtualCardHandler called")
}
