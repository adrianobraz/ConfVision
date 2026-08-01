package tipos

import "net/http"

type Rota struct {
	Uri      string
	Metodo   string
	Controle func(w http.ResponseWriter, r *http.Request)
	Seguro   bool
	Master   bool
}
