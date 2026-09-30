package handler

import (
	"net/http"
	"os"
	"strings"

	"apifunction/auth"
	"apifunction/config"
	"apifunction/service/credito"
	"apifunction/service/tarifa"
)

func (h *Handler) ServicoTarifaListar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	idCentral := strings.TrimSpace(r.URL.Query().Get("id_central"))
	if idCentral == "" {
		idCentral = sess.IDCentralCatalogo
	}
	lista, modo, err := tarifa.ListarCentral(idCentral)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": lista, "modo_preco": modo})
}

func (h *Handler) ServicoTarifaSalvar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	var t tarifa.Tarifa
	if err := decodeJSON(r, &t); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := tarifa.SalvarCentral(sess, t); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) ServicoTarifaRepListar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	idRep := strings.TrimSpace(r.URL.Query().Get("id_representante"))
	if idRep == "" {
		idRep = sess.IDRepresentante
	}
	lista, err := tarifa.ListarRep(idRep)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": lista})
}

func (h *Handler) ServicoTarifaRepSalvar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	var t tarifa.TarifaRep
	if err := decodeJSON(r, &t); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := tarifa.SalvarRep(sess, t); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) CreditoRecargaSolicitar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	var in credito.SolicitarInput
	if err := decodeJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	workerKey := strings.TrimSpace(r.Header.Get("X-Worker-Key"))
	if workerKey != "" && workerKey == config.WorkerSecret {
		in.SolicitadoPor = "FRA"
		res, err := credito.SolicitarRecarga(auth.SessaoAdm{
			UsuarioAdm: auth.UsuarioAdm{IDUsuario: "FRANQUEADO"},
			UserTipo:   "FRA",
		}, in)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": res})
		return
	}

	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	res, err := credito.SolicitarRecarga(sess, in)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": res})
}

func (h *Handler) CreditoRecargaConfirmar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if _, err := auth.RequireEscopo(r); err != nil {
		if strings.TrimSpace(r.Header.Get("X-Worker-Key")) != config.WorkerSecret {
			writeErr(w, http.StatusUnauthorized, err.Error())
			return
		}
	}
	var req struct {
		IDFaturaContabil int    `json:"id_fatura_contabil"`
		IDRecarga        int    `json:"id_recarga"`
		AdminUsuario     string `json:"admin_usuario"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	admin := strings.TrimSpace(req.AdminUsuario)
	if admin == "" {
		admin = "admConfmonit"
	}
	var err error
	if req.IDFaturaContabil > 0 {
		err = credito.ConfirmarRecarga(req.IDFaturaContabil, admin)
	} else if req.IDRecarga > 0 {
		err = credito.ConfirmarRecargaPorIDRecarga(req.IDRecarga, admin)
	} else {
		writeErr(w, http.StatusBadRequest, "id_fatura_contabil ou id_recarga obrigatorio")
		return
	}
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) CreditoRecargaListar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	workerKey := strings.TrimSpace(r.Header.Get("X-Worker-Key"))
	if workerKey == "" || workerKey != config.WorkerSecret {
		if _, err := auth.RequireEscopo(r); err != nil {
			writeErr(w, http.StatusUnauthorized, err.Error())
			return
		}
	}
	lista, err := credito.ListarRecargas(idFra, 100)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": lista})
}

func (h *Handler) ServicoTarifaEfetiva(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	workerKey := strings.TrimSpace(r.Header.Get("X-Worker-Key"))
	if workerKey == "" || workerKey != config.WorkerSecret {
		if _, err := auth.RequireEscopo(r); err != nil {
			writeErr(w, http.StatusUnauthorized, err.Error())
			return
		}
	}
	if idFra == "" {
		writeErr(w, http.StatusBadRequest, "id_franqueado obrigatorio")
		return
	}
	lista, err := tarifa.ListarEfetivas(idFra)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": lista})
}

func (h *Handler) ServicoTarifaSeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	sess, err := auth.RequireEscopo(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	idCentral := sess.IDCentralCatalogo
	if err := tarifa.SeedTarifasCentral(idCentral); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func workerOK(r *http.Request) bool {
	return strings.TrimSpace(r.Header.Get("X-Worker-Key")) == strings.TrimSpace(os.Getenv("WORKER_SECRET"))
}

func (h *Handler) CreditoDebitarUso(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	workerKey := strings.TrimSpace(r.Header.Get("X-Worker-Key"))
	if workerKey == "" || workerKey != config.WorkerSecret {
		writeErr(w, http.StatusUnauthorized, "worker key invalida")
		return
	}
	var in credito.DebitarUsoInput
	if err := decodeJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	res, err := credito.DebitarUso(in)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": res})
}
