package paginacaoV4

import "fmt"

const TamanhoPagina = 100
const TamanhoMaximo = 500

// Normalizar ajusta limit/offset; retorna aplicar=false quando limit<=0 (lista completa).
func Normalizar(limit, offset int) (int, int, bool) {
	if limit <= 0 {
		return 0, 0, false
	}
	if limit > TamanhoMaximo {
		limit = TamanhoMaximo
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset, true
}

// Clausula retorna " LIMIT n OFFSET m" ou vazio se sem paginacao.
func Clausula(limit, offset int) string {
	l, o, ok := Normalizar(limit, offset)
	if !ok {
		return ""
	}
	return fmt.Sprintf(" LIMIT %d OFFSET %d", l, o)
}

func HasMore(total, limit, offset, retornados int) bool {
	if limit <= 0 {
		return false
	}
	if total > 0 {
		return offset+retornados < total
	}
	return retornados >= limit
}
