package tiposV4

import "net/http"

type Rota struct {
	URI    string
	Metodo string
	Funcao func(http.ResponseWriter, *http.Request)
	Seguro bool
}
