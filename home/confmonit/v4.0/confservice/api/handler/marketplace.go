package handler

import (
	"database/sql"
	"net/http"
	"strings"

	"confservice/db"
	"confservice/internal/auth"

	"github.com/google/uuid"
)

func (h *Handler) MarketplaceCategorias(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	rows, err := db.Conn.Query(`
		SELECT id, slug, nome, icone, ordem FROM cs_categoria
		WHERE ativo = 1 ORDER BY ordem, nome`)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	defer rows.Close()
	list := []map[string]any{}
	for rows.Next() {
		var id, slug, nome, icone string
		var ordem int
		if err := rows.Scan(&id, &slug, &nome, &icone, &ordem); err != nil {
			continue
		}
		list = append(list, map[string]any{
			"id": id, "slug": slug, "nome": nome, "icone": icone, "ordem": ordem,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": list})
}

func (h *Handler) MarketplaceFabricantes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	rows, err := db.Conn.Query(`
		SELECT id, nome, slug, descricao, logo_url, comissao_disponivel_pct
		FROM cs_fabricante WHERE ativo = 1 ORDER BY nome`)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	defer rows.Close()
	list := []map[string]any{}
	for rows.Next() {
		var id, nome, slug string
		var desc, logo sql.NullString
		var comissao float64
		if err := rows.Scan(&id, &nome, &slug, &desc, &logo, &comissao); err != nil {
			continue
		}
		list = append(list, map[string]any{
			"id": id, "nome": nome, "slug": slug,
			"descricao": desc.String, "logoUrl": logo.String,
			"comissaoDisponivelPct": comissao,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": list})
}

func (h *Handler) MarketplacePrestadores(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	cidade := strings.TrimSpace(r.URL.Query().Get("cidade"))
	cat := strings.TrimSpace(r.URL.Query().Get("categoria"))

	sqlStr := `
		SELECT DISTINCT p.id, p.tipo, p.nome, p.nome_fantasia, p.cidade, p.uf,
		       p.regiao_atendimento, p.experiencia_anos, p.comissao_plataforma_pct,
		       p.disponibilidade, p.foto_url, p.bio
		FROM cs_prestador p`
	args := []any{}
	where := []string{"p.ativo = 1"}

	if cat != "" {
		sqlStr += `
		INNER JOIN cs_prestador_categoria pc ON pc.id_prestador = p.id
		INNER JOIN cs_categoria c ON c.id = pc.id_categoria`
		where = append(where, "(c.slug = ? OR c.id = ?)")
		args = append(args, cat, cat)
	}
	if cidade != "" {
		where = append(where, "p.cidade LIKE ?")
		args = append(args, "%"+cidade+"%")
	}
	if q != "" {
		where = append(where, "(p.nome LIKE ? OR p.nome_fantasia LIKE ? OR p.bio LIKE ? OR p.regiao_atendimento LIKE ?)")
		like := "%" + q + "%"
		args = append(args, like, like, like, like)
	}
	sqlStr += " WHERE " + strings.Join(where, " AND ") + " ORDER BY p.nome LIMIT 100"

	rows, err := db.Conn.Query(sqlStr, args...)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	defer rows.Close()

	list := []map[string]any{}
	ids := []string{}
	for rows.Next() {
		var id, tipo, nome string
		var fantasia, regiao, disp, foto, bio sql.NullString
		var cidadeV, uf string
		var exp int
		var comissao float64
		if err := rows.Scan(&id, &tipo, &nome, &fantasia, &cidadeV, &uf,
			&regiao, &exp, &comissao, &disp, &foto, &bio); err != nil {
			continue
		}
		ids = append(ids, id)
		list = append(list, map[string]any{
			"id": id, "tipo": tipo, "nome": nome,
			"nomeFantasia": fantasia.String, "cidade": cidadeV, "uf": uf,
			"regiaoAtendimento": regiao.String, "experienciaAnos": exp,
			"comissaoPlataformaPct": comissao, "disponibilidade": disp.String,
			"fotoUrl": foto.String, "bio": bio.String,
			"categorias": []any{}, "fabricantes": []any{},
		})
	}

	catsBy := map[string][]map[string]any{}
	fabsBy := map[string][]map[string]any{}
	if len(ids) > 0 {
		catsBy = carregarCategoriasPrestadores(ids)
		fabsBy = carregarFabricantesPrestadores(ids)
	}
	for i := range list {
		id := list[i]["id"].(string)
		if c, ok := catsBy[id]; ok {
			list[i]["categorias"] = c
		}
		if f, ok := fabsBy[id]; ok {
			list[i]["fabricantes"] = f
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": list})
}

func (h *Handler) MarketplacePrestadorDetalhe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeErr(w, http.StatusBadRequest, "id obrigatorio")
		return
	}
	var tipo, nome string
	var fantasia, cnpj, tel, regiao, bio, disp, foto sql.NullString
	var email, cidade, uf string
	var exp int
	var comissao float64
	err := db.Conn.QueryRow(`
		SELECT tipo, nome, nome_fantasia, cnpj, email, telefone, cidade, uf,
		       regiao_atendimento, bio, experiencia_anos, comissao_plataforma_pct,
		       disponibilidade, foto_url
		FROM cs_prestador WHERE id = ? AND ativo = 1`, id).
		Scan(&tipo, &nome, &fantasia, &cnpj, &email, &tel, &cidade, &uf,
			&regiao, &bio, &exp, &comissao, &disp, &foto)
	if err == sql.ErrNoRows {
		writeErr(w, http.StatusNotFound, "prestador nao encontrado")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	cats := carregarCategoriasPrestadores([]string{id})[id]
	fabs := carregarFabricantesPrestadores([]string{id})[id]
	if cats == nil {
		cats = []map[string]any{}
	}
	if fabs == nil {
		fabs = []map[string]any{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true,
		"prestador": map[string]any{
			"id": id, "tipo": tipo, "nome": nome, "nomeFantasia": fantasia.String,
			"cnpj": cnpj.String, "email": email, "telefone": tel.String,
			"cidade": cidade, "uf": uf, "regiaoAtendimento": regiao.String,
			"bio": bio.String, "experienciaAnos": exp,
			"comissaoPlataformaPct": comissao, "disponibilidade": disp.String,
			"fotoUrl": foto.String, "categorias": cats, "fabricantes": fabs,
		},
	})
}

func (h *Handler) MarketplacePrestadorRegistrar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	var in struct {
		Tipo                   string   `json:"tipo"`
		Nome                   string   `json:"nome"`
		NomeFantasia           string   `json:"nomeFantasia"`
		CNPJ                   string   `json:"cnpj"`
		Email                  string   `json:"email"`
		Telefone               string   `json:"telefone"`
		Senha                  string   `json:"senha"`
		Cidade                 string   `json:"cidade"`
		UF                     string   `json:"uf"`
		RegiaoAtendimento      string   `json:"regiaoAtendimento"`
		Bio                    string   `json:"bio"`
		ExperienciaAnos        int      `json:"experienciaAnos"`
		ComissaoPlataformaPct  float64  `json:"comissaoPlataformaPct"`
		Disponibilidade        string   `json:"disponibilidade"`
		Categorias             []string `json:"categorias"`
		Fabricantes            []string `json:"fabricantes"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.Nome = strings.TrimSpace(in.Nome)
	in.Tipo = strings.ToUpper(strings.TrimSpace(in.Tipo))
	in.UF = strings.ToUpper(strings.TrimSpace(in.UF))
	if in.Tipo != "EMPRESA" {
		in.Tipo = "PROFISSIONAL"
	}
	if in.Email == "" || in.Nome == "" || len(in.Senha) < 6 {
		writeErr(w, http.StatusBadRequest, "nome, email e senha(>=6) obrigatorios")
		return
	}
	if in.ComissaoPlataformaPct <= 0 {
		in.ComissaoPlataformaPct = 10
	}
	if in.ComissaoPlataformaPct > 50 {
		writeErr(w, http.StatusBadRequest, "comissaoPlataformaPct maxima 50")
		return
	}
	hash, err := auth.HashSenha(in.Senha)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "hash")
		return
	}
	id := uuid.NewString()
	_, err = db.Conn.Exec(`
		INSERT INTO cs_prestador
		(id, tipo, nome, nome_fantasia, cnpj, email, telefone, senha_hash, cidade, uf,
		 regiao_atendimento, bio, experiencia_anos, comissao_plataforma_pct, disponibilidade, ativo)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)`,
		id, in.Tipo, in.Nome, nullStr(in.NomeFantasia), nullStr(in.CNPJ), in.Email, nullStr(in.Telefone),
		hash, strings.TrimSpace(in.Cidade), in.UF, nullStr(in.RegiaoAtendimento), nullStr(in.Bio),
		in.ExperienciaAnos, in.ComissaoPlataformaPct, nullStr(in.Disponibilidade),
	)
	if err != nil {
		writeErr(w, http.StatusConflict, "nao foi possivel cadastrar (email ja existe?)")
		return
	}
	vincularCategorias(id, in.Categorias)
	vincularFabricantes(id, in.Fabricantes)
	writeJSON(w, http.StatusCreated, map[string]any{
		"ok": true,
		"prestador": map[string]any{
			"id": id, "tipo": in.Tipo, "nome": in.Nome, "email": in.Email,
			"comissaoPlataformaPct": in.ComissaoPlataformaPct,
		},
	})
}

func (h *Handler) MarketplaceFabricanteRegistrar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	var in struct {
		Nome                   string  `json:"nome"`
		Slug                   string  `json:"slug"`
		Descricao              string  `json:"descricao"`
		LogoURL                string  `json:"logoUrl"`
		ComissaoDisponivelPct  float64 `json:"comissaoDisponivelPct"`
		Email                  string  `json:"email"`
		Senha                  string  `json:"senha"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	in.Nome = strings.TrimSpace(in.Nome)
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.Slug = slugify(in.Slug)
	if in.Slug == "" {
		in.Slug = slugify(in.Nome)
	}
	if in.Nome == "" || in.Slug == "" || in.Email == "" || len(in.Senha) < 6 {
		writeErr(w, http.StatusBadRequest, "nome, email e senha(>=6) obrigatorios")
		return
	}
	if in.ComissaoDisponivelPct < 0 || in.ComissaoDisponivelPct > 50 {
		writeErr(w, http.StatusBadRequest, "comissaoDisponivelPct entre 0 e 50")
		return
	}
	hash, err := auth.HashSenha(in.Senha)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "hash")
		return
	}
	id := uuid.NewString()
	_, err = db.Conn.Exec(`
		INSERT INTO cs_fabricante
		(id, nome, slug, descricao, logo_url, comissao_disponivel_pct, email, senha_hash, ativo)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)`,
		id, in.Nome, in.Slug, nullStr(in.Descricao), nullStr(in.LogoURL),
		in.ComissaoDisponivelPct, in.Email, hash,
	)
	if err != nil {
		writeErr(w, http.StatusConflict, "nao foi possivel cadastrar (slug ou email ja existe?)")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"ok": true,
		"fabricante": map[string]any{
			"id": id, "nome": in.Nome, "slug": in.Slug,
			"comissaoDisponivelPct": in.ComissaoDisponivelPct,
		},
	})
}

func (h *Handler) MarketplaceOrcamento(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	var in struct {
		Texto        string `json:"texto"`
		Cidade       string `json:"cidade"`
		UF           string `json:"uf"`
		Categoria    string `json:"categoria"`
		NomeContato  string `json:"nomeContato"`
		Email        string `json:"email"`
		Telefone     string `json:"telefone"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	in.Texto = strings.TrimSpace(in.Texto)
	in.NomeContato = strings.TrimSpace(in.NomeContato)
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.UF = strings.ToUpper(strings.TrimSpace(in.UF))
	if in.Texto == "" || in.NomeContato == "" || in.Email == "" {
		writeErr(w, http.StatusBadRequest, "texto, nomeContato e email obrigatorios")
		return
	}
	var catID any
	if in.Categoria != "" {
		var id string
		err := db.Conn.QueryRow(`SELECT id FROM cs_categoria WHERE slug = ? OR id = ?`, in.Categoria, in.Categoria).Scan(&id)
		if err == nil {
			catID = id
		}
	}
	id := uuid.NewString()
	_, err := db.Conn.Exec(`
		INSERT INTO cs_orcamento_pedido
		(id, texto, cidade, uf, id_categoria, nome_contato, email, telefone, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'NOVO')`,
		id, in.Texto, strings.TrimSpace(in.Cidade), in.UF, catID,
		in.NomeContato, in.Email, nullStr(in.Telefone),
	)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "id": id})
}

func carregarCategoriasPrestadores(ids []string) map[string][]map[string]any {
	out := map[string][]map[string]any{}
	if len(ids) == 0 {
		return out
	}
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	q := `
		SELECT pc.id_prestador, c.id, c.slug, c.nome, c.icone
		FROM cs_prestador_categoria pc
		INNER JOIN cs_categoria c ON c.id = pc.id_categoria
		WHERE pc.id_prestador IN (` + placeholders + `)`
	rows, err := db.Conn.Query(q, args...)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var pid, id, slug, nome, icone string
		if err := rows.Scan(&pid, &id, &slug, &nome, &icone); err != nil {
			continue
		}
		out[pid] = append(out[pid], map[string]any{
			"id": id, "slug": slug, "nome": nome, "icone": icone,
		})
	}
	return out
}

func carregarFabricantesPrestadores(ids []string) map[string][]map[string]any {
	out := map[string][]map[string]any{}
	if len(ids) == 0 {
		return out
	}
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	q := `
		SELECT pf.id_prestador, f.id, f.nome, f.slug, f.comissao_disponivel_pct
		FROM cs_prestador_fabricante pf
		INNER JOIN cs_fabricante f ON f.id = pf.id_fabricante
		WHERE pf.id_prestador IN (` + placeholders + `)`
	rows, err := db.Conn.Query(q, args...)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var pid, id, nome, slug string
		var comissao float64
		if err := rows.Scan(&pid, &id, &nome, &slug, &comissao); err != nil {
			continue
		}
		out[pid] = append(out[pid], map[string]any{
			"id": id, "nome": nome, "slug": slug, "comissaoDisponivelPct": comissao,
		})
	}
	return out
}

func vincularCategorias(prestadorID string, refs []string) {
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			continue
		}
		var catID string
		err := db.Conn.QueryRow(`SELECT id FROM cs_categoria WHERE id = ? OR slug = ?`, ref, ref).Scan(&catID)
		if err != nil {
			continue
		}
		_, _ = db.Conn.Exec(`INSERT IGNORE INTO cs_prestador_categoria (id_prestador, id_categoria) VALUES (?, ?)`,
			prestadorID, catID)
	}
}

func vincularFabricantes(prestadorID string, refs []string) {
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			continue
		}
		var fabID string
		err := db.Conn.QueryRow(`SELECT id FROM cs_fabricante WHERE id = ? OR slug = ?`, ref, ref).Scan(&fabID)
		if err != nil {
			continue
		}
		_, _ = db.Conn.Exec(`INSERT IGNORE INTO cs_prestador_fabricante (id_prestador, id_fabricante) VALUES (?, ?)`,
			prestadorID, fabID)
	}
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash && b.Len() > 0 {
			b.WriteByte('-')
			prevDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	return out
}
