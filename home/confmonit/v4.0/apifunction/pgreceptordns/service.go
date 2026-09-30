package pgreceptordns

import (
	"errors"
	"fmt"
	"strings"

	"apifunction/config"
	"apifunction/db"
)

type Service struct {
	cf   *cfClient
	zona string
	ip   string
}

func NovoService() *Service {
	return &Service{
		cf:   newCFClient(config.CloudflareAPIToken, config.CloudflareZoneID),
		zona: config.ReceptorDNSZona,
		ip:   config.ReceptorDNSIP,
	}
}

func (s *Service) VerificarDisponibilidade(subdominio string) (Disponibilidade, error) {
	sub := normalizarSub(subdominio)
	out := Disponibilidade{Subdominio: sub, FQDN: montarFQDN(sub, s.zona)}
	if err := validarSubdominio(sub); err != nil {
		out.Disponivel = false
		out.Motivo = err.Error()
		return out, nil
	}
	existe, err := obterPorFQDN(out.FQDN)
	if err != nil {
		return out, err
	}
	if existe != nil {
		out.Disponivel = false
		out.Motivo = "nome ja cadastrado por outro franqueado"
		return out, nil
	}
	if s.cf.configured() {
		ok, _, err := s.cf.recordExists(out.FQDN)
		if err != nil {
			return out, err
		}
		if ok {
			out.Disponivel = false
			out.Motivo = "nome ja existe no Cloudflare"
			return out, nil
		}
	}
	out.Disponivel = true
	return out, nil
}

func (s *Service) Salvar(idFranqueado, subdominio string) (*Registro, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	sub := normalizarSub(subdominio)
	if idFranqueado == "" {
		return nil, errors.New("id_franqueado obrigatorio")
	}
	if err := validarSubdominio(sub); err != nil {
		return nil, err
	}
	if !s.cf.configured() {
		return nil, errors.New("cloudflare nao configurado no servidor")
	}

	atual, err := Obter(idFranqueado)
	if err != nil {
		return nil, err
	}
	if atual != nil && strings.EqualFold(atual.Subdominio, sub) && atual.Status == "ativo" {
		return atual, nil
	}
	if atual != nil {
		return nil, errors.New("dns ja configurado — remova antes de alterar")
	}

	fqdn := montarFQDN(sub, s.zona)
	emUso, err := fqdnEmUsoPorOutro(idFranqueado, fqdn)
	if err != nil {
		return nil, err
	}
	if emUso {
		return nil, errors.New("este nome ja esta em uso")
	}

	disp, err := s.VerificarDisponibilidade(sub)
	if err != nil {
		return nil, err
	}
	if !disp.Disponivel {
		return nil, errors.New(disp.Motivo)
	}

	recordID, err := s.cf.createA(sub, s.ip)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar dns: %w", err)
	}

	reg := Registro{
		IDFranqueado: idFranqueado,
		Subdominio:   sub,
		FQDN:         fqdn,
		Zona:         s.zona,
		IPDestino:    s.ip,
		CFRecordID:   recordID,
		Status:       "ativo",
	}
	if err := salvar(reg); err != nil {
		_ = s.cf.deleteRecord(recordID)
		return nil, err
	}
	return &reg, nil
}

func (s *Service) Remover(idFranqueado string) error {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return errors.New("id_franqueado obrigatorio")
	}
	atual, err := Obter(idFranqueado)
	if err != nil {
		return err
	}
	if atual == nil {
		return nil
	}
	if s.cf.configured() && strings.TrimSpace(atual.CFRecordID) != "" {
		if err := s.cf.deleteRecord(atual.CFRecordID); err != nil {
			return fmt.Errorf("falha ao remover dns no cloudflare: %w", err)
		}
	}
	return remover(idFranqueado)
}

func (s *Service) Endpoints(idFranqueado, idFabricante string) ([]EndpointFabricante, *Registro, error) {
	reg, err := Obter(idFranqueado)
	if err != nil {
		return nil, nil, err
	}
	portas, err := listarPortasReceptor()
	if err != nil {
		return nil, reg, err
	}
	fabricantes, err := listarFabricantes()
	if err != nil {
		return nil, reg, err
	}
	portaPorModulo := map[string]string{}
	for _, p := range portas {
		mod := strings.ToUpper(strings.TrimSpace(p.Modulo))
		if mod == "" || p.Porta == "" {
			continue
		}
		if _, ok := portaPorModulo[mod]; !ok {
			portaPorModulo[mod] = p.Porta
		}
	}

	idFabricante = strings.TrimSpace(idFabricante)
	out := make([]EndpointFabricante, 0)
	for _, f := range fabricantes {
		if idFabricante != "" && f.ID != idFabricante {
			continue
		}
		if fabricanteEhCamera(f.Nome) {
			continue
		}
		mod := moduloForFabricante(f.Nome)
		porta := portaPorModulo[mod]
		if porta == "" {
			continue
		}
		host := s.hostParaEndpoint(reg, mod)
		if host == "" {
			continue
		}
		out = append(out, EndpointFabricante{
			IDFabricante: f.ID,
			Nome:         f.Nome,
			Modulo:       mod,
			Host:         host,
			Porta:        porta,
			Endereco:     host + ":" + porta,
		})
	}
	return out, reg, nil
}

// hostParaEndpoint: DNS proprio do franqueado (ativo) ou subdominio padrao por modulo.
func (s *Service) hostParaEndpoint(reg *Registro, mod string) string {
	if reg != nil && reg.Status == "ativo" {
		if fqdn := strings.TrimSpace(reg.FQDN); fqdn != "" {
			return fqdn
		}
	}
	sub := subdominioPadraoPorModulo(mod)
	if sub == "" {
		return ""
	}
	return montarFQDN(sub, s.zona)
}

func subdominioPadraoPorModulo(mod string) string {
	switch strings.ToUpper(strings.TrimSpace(mod)) {
	case "JFL":
		return "jfl"
	case "INTELBRAS":
		return "intelbras"
	case "VETTI":
		return "vetti"
	case "COMPATEC":
		return "compatec"
	default:
		return ""
	}
}

func fabricanteEhCamera(nome string) bool {
	return strings.TrimSpace(strings.ToUpper(nome)) == "CAMERA"
}

type portaReceptor struct {
	Modulo string
	Porta  string
	Nome   string
}

type fabricanteRow struct {
	ID   string
	Nome string
}

func listarPortasReceptor() ([]portaReceptor, error) {
	if db.Conn == nil {
		return nil, errors.New("mysql indisponivel")
	}
	rows, err := db.Conn.Query(`
SELECT receptorEvento.Modulo, receptorEvento.Porta, receptorEvento.Nome
FROM receptorEvento
WHERE receptorEvento.Ativo = 'S'
  AND receptorEvento.Producao = 'S'
  AND receptorEvento.Master = 'S'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []portaReceptor
	for rows.Next() {
		var p portaReceptor
		if err := rows.Scan(&p.Modulo, &p.Porta, &p.Nome); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func listarFabricantes() ([]fabricanteRow, error) {
	if db.Conn == nil {
		return nil, errors.New("mysql indisponivel")
	}
	rows, err := db.Conn.Query(`
SELECT fabricantes.ID_Fabricante, fabricantes.Nome
FROM fabricantes
LEFT JOIN listaBloqueio ON fabricantes.ID_Fabricante = listaBloqueio.ID_Alvo
WHERE listaBloqueio.ID_Alvo IS NULL
ORDER BY fabricantes.Nome`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []fabricanteRow
	for rows.Next() {
		var f fabricanteRow
		if err := rows.Scan(&f.ID, &f.Nome); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func moduloForFabricante(nome string) string {
	n := strings.ToUpper(strings.TrimSpace(nome))
	switch {
	case strings.Contains(n, "JFL"):
		return "JFL"
	case strings.Contains(n, "INTELBRAS"):
		return "INTELBRAS"
	case strings.Contains(n, "VETTI"):
		return "VETTI"
	case strings.Contains(n, "COMPATEC"), strings.Contains(n, "CONTINENTE"):
		return "COMPATEC"
	default:
		return n
	}
}

func ListarPortasPublicas() ([]map[string]string, error) {
	portas, err := listarPortasReceptor()
	if err != nil {
		return nil, err
	}
	out := make([]map[string]string, 0, len(portas))
	for _, p := range portas {
		out = append(out, map[string]string{
			"modulo": strings.ToUpper(strings.TrimSpace(p.Modulo)),
			"porta":  p.Porta,
			"nome":   p.Nome,
		})
	}
	return out, nil
}
