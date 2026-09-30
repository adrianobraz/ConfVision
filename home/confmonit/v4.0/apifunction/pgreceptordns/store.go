package pgreceptordns

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"apifunction/pgcredito"
)

func Obter(idFranqueado string) (*Registro, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return nil, errors.New("id_franqueado obrigatorio")
	}
	db, err := pgcredito.DB()
	if err != nil {
		return nil, err
	}
	var r Registro
	var cfID, erro sql.NullString
	err = db.QueryRow(`
SELECT id_franqueado, subdominio, fqdn, zona, ip_destino, cf_record_id, status, erro, created_at, updated_at
FROM fp_receptor_dns WHERE id_franqueado = $1`, idFranqueado).Scan(
		&r.IDFranqueado, &r.Subdominio, &r.FQDN, &r.Zona, &r.IPDestino,
		&cfID, &r.Status, &erro, &r.CreatedAt, &r.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.CFRecordID = cfID.String
	r.Erro = erro.String
	return &r, nil
}

func obterPorFQDN(fqdn string) (*Registro, error) {
	db, err := pgcredito.DB()
	if err != nil {
		return nil, err
	}
	var r Registro
	var cfID, erro sql.NullString
	err = db.QueryRow(`
SELECT id_franqueado, subdominio, fqdn, zona, ip_destino, cf_record_id, status, erro, created_at, updated_at
FROM fp_receptor_dns WHERE LOWER(fqdn) = LOWER($1)`, fqdn).Scan(
		&r.IDFranqueado, &r.Subdominio, &r.FQDN, &r.Zona, &r.IPDestino,
		&cfID, &r.Status, &erro, &r.CreatedAt, &r.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.CFRecordID = cfID.String
	r.Erro = erro.String
	return &r, nil
}

func salvar(r Registro) error {
	db, err := pgcredito.DB()
	if err != nil {
		return err
	}
	_, err = db.Exec(`
INSERT INTO fp_receptor_dns (
  id_franqueado, subdominio, fqdn, zona, ip_destino, cf_record_id, status, erro, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW(),NOW())
ON CONFLICT (id_franqueado) DO UPDATE SET
  subdominio = EXCLUDED.subdominio,
  fqdn = EXCLUDED.fqdn,
  zona = EXCLUDED.zona,
  ip_destino = EXCLUDED.ip_destino,
  cf_record_id = EXCLUDED.cf_record_id,
  status = EXCLUDED.status,
  erro = EXCLUDED.erro,
  updated_at = NOW()`, r.IDFranqueado, r.Subdominio, r.FQDN, r.Zona, r.IPDestino,
		nullStr(r.CFRecordID), r.Status, nullStr(r.Erro))
	return err
}

func remover(idFranqueado string) error {
	db, err := pgcredito.DB()
	if err != nil {
		return err
	}
	_, err = db.Exec(`DELETE FROM fp_receptor_dns WHERE id_franqueado = $1`, idFranqueado)
	return err
}

func nullStr(s string) any {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return s
}

func fqdnEmUsoPorOutro(idFranqueado, fqdn string) (bool, error) {
	db, err := pgcredito.DB()
	if err != nil {
		return false, err
	}
	var out string
	err = db.QueryRow(`
SELECT id_franqueado FROM fp_receptor_dns
WHERE LOWER(fqdn) = LOWER($1) AND id_franqueado <> $2
LIMIT 1`, fqdn, idFranqueado).Scan(&out)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return out != "", nil
}

func montarFQDN(sub, zona string) string {
	sub = strings.ToLower(strings.TrimSpace(sub))
	zona = strings.ToLower(strings.TrimSpace(zona))
	if sub == "" || zona == "" {
		return ""
	}
	return sub + "." + zona
}

var reservados = map[string]bool{
	"www": true, "mail": true, "ftp": true, "webmail": true, "ativa": true,
	"imagem": true, "confiseg": true, "api": true, "admin": true, "teste": true,
}

func validarSubdominio(sub string) error {
	sub = strings.ToLower(strings.TrimSpace(sub))
	if len(sub) < 3 || len(sub) > 32 {
		return errors.New("subdominio deve ter entre 3 e 32 caracteres")
	}
	for i, c := range sub {
		ok := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-'
		if !ok {
			return errors.New("subdominio aceita apenas letras, numeros e hifen")
		}
		if (i == 0 || i == len(sub)-1) && c == '-' {
			return errors.New("subdominio nao pode comecar ou terminar com hifen")
		}
	}
	if reservados[sub] {
		return fmt.Errorf("subdominio '%s' reservado pelo sistema", sub)
	}
	return nil
}

func normalizarSub(sub string) string {
	return strings.ToLower(strings.TrimSpace(sub))
}

func nowUTC() time.Time { return time.Now().UTC() }
