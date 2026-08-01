package confvision

import (
	"net/http"

	"github.com/gorilla/mux"
)

func CarregarMenu(w http.ResponseWriter, r *http.Request) {
	carregarPagina(w, r, "menu-confvision.html", paginaBase("ConfVision", "/carregar-menu-confvision"))
}

func CarregarCameras(w http.ResponseWriter, r *http.Request) {
	carregarPagina(w, r, "cameras.html", paginaBase("Câmeras", "/carregar-menu-confvision"))
}

func CarregarCamerasNova(w http.ResponseWriter, r *http.Request) {
	d := paginaBase("Nova câmera", "/cameras")
	carregarPagina(w, r, "cameras-form.html", d)
}

func CarregarCamerasEditar(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	d := paginaBase("Editar câmera #"+id, "/cameras")
	d.LinkJs = id
	carregarPagina(w, r, "cameras-form.html", d)
}

func CarregarEventos(w http.ResponseWriter, r *http.Request) {
	carregarPagina(w, r, "eventos.html", paginaBase("Eventos", "/carregar-menu-confvision"))
}

func CarregarAoVivo(w http.ResponseWriter, r *http.Request) {
	d := paginaBase("Ao vivo", "/carregar-menu-confvision")
	carregarPagina(w, r, "ao-vivo.html", d)
}

func CarregarAoVivoCamera(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	d := paginaBase("Ao vivo — câmera "+id, "/cameras")
	d.LinkJs = id
	carregarPagina(w, r, "ao-vivo.html", d)
}

func CarregarGravacoes(w http.ResponseWriter, r *http.Request) {
	carregarPagina(w, r, "gravacoes.html", paginaBase("Gravações", "/carregar-menu-confvision"))
}

func CarregarGravacoesTimeline(w http.ResponseWriter, r *http.Request) {
	carregarPagina(w, r, "gravacoes-timeline.html", paginaBase("Timeline", "/gravacoes"))
}

func CarregarGravacoesDvr(w http.ResponseWriter, r *http.Request) {
	carregarPagina(w, r, "gravacoes-dvr.html", paginaBase("DVR", "/gravacoes"))
}

func CarregarRelatorioArmado(w http.ResponseWriter, r *http.Request) {
	carregarPagina(w, r, "relatorio-armado.html", paginaBase("Armado", "/carregar-menu-confvision"))
}

func CarregarRelatorioLicencas(w http.ResponseWriter, r *http.Request) {
	carregarPagina(w, r, "relatorio-licencas.html", paginaBase("Licenças", "/carregar-menu-confvision"))
}

func CarregarRelatorioFaturas(w http.ResponseWriter, r *http.Request) {
	carregarPagina(w, r, "relatorio-faturas.html", paginaBase("Faturas", "/carregar-menu-confvision"))
}

func CarregarGradeHorario(w http.ResponseWriter, r *http.Request) {
	carregarPagina(w, r, "grade-horario.html", paginaBase("Grade horária", "/carregar-menu-confvision"))
}
