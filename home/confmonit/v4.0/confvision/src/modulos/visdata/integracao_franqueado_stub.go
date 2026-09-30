package visdata

import (
	"context"
	"fmt"
)

func GetIntegracaoFranqueado(ctx context.Context, idFranqueado string) (map[string]any, error) {
	return ListIntegracoesByFranqueado(ctx, idFranqueado)
}

func SaveIntegracaoFranqueado(ctx context.Context, payload map[string]any) (map[string]any, error) {
	if id := intVal(payload, "id"); id > 0 {
		return UpdateIntegracao(ctx, id, payload)
	}
	return CreateIntegracao(ctx, payload)
}

func TestIntegracaoFranqueado(ctx context.Context, idFranqueado string, payload map[string]any) (map[string]any, error) {
	_ = ctx
	_ = idFranqueado
	_ = payload
	return map[string]any{"ok": true, "mensagem": "teste nao disponivel neste build"}, nil
}

func TestIntegracaoMoni(ctx context.Context, id int, payload map[string]any) (map[string]any, error) {
	_ = ctx
	_ = payload
	if id <= 0 {
		return nil, fmt.Errorf("id integracao invalido")
	}
	return map[string]any{"ok": true, "mensagem": "teste moni nao disponivel neste build"}, nil
}
