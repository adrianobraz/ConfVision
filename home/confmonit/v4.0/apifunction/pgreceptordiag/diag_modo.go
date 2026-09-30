package pgreceptordiag

import (
	"fmt"
	"strings"
)

func modoDiagnostico(conta, codigo string) string {
	conta = normalizarConta(conta)
	codigo = normalizarCodigoFisico(codigo)
	switch {
	case conta != "" && codigo != "":
		return "instalacao"
	case codigo != "":
		return "mac"
	case conta != "":
		return "conta"
	default:
		return "receptor"
	}
}

func journalTemConta(eventos []JournalEvent, conta string) bool {
	conta = normalizarConta(conta)
	if conta == "" {
		return false
	}
	for _, ev := range eventos {
		if normalizarConta(ev.Conta) == conta {
			return true
		}
	}
	return false
}

func journalTemCodigo(eventos []JournalEvent, codigo string) bool {
	codigo = normalizarCodigoFisico(codigo)
	if codigo == "" {
		return false
	}
	for _, ev := range eventos {
		if codigoMatch(normalizarCodigoFisico(ev.IMEI), codigo) {
			return true
		}
	}
	return false
}

func codigoMatch(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	if a == b {
		return true
	}
	return classificarCodigoMatch(a, b) == "similar"
}

func formatListaDispositivos(lista []DispositivoInfo) string {
	var parts []string
	for _, d := range lista {
		id := pickIdFisico(d.IdFisico1, d.IdFisico2)
		line := fmt.Sprintf("Conta %s | ID %s", d.Conta, id)
		if d.NomeCliente != "" {
			line += " | Cliente: " + d.NomeCliente
		}
		parts = append(parts, line)
	}
	return strings.Join(parts, " · ")
}

func resolverCadastroVerificar(idFra, fabricante, conta, codigo, modo string, res *Resultado, idFisico *string, idFraPtr *string) {
	if modo == "receptor" {
		return
	}

	switch modo {
	case "instalacao":
		disp, err := BuscarDispositivoExato(idFra, fabricante, conta, codigo)
		if err != nil {
			res.Checks = append(res.Checks, Check{
				ID:      "cadastro",
				OK:      false,
				Alerta:  true,
				Titulo:  "Consulta de cadastro indisponivel",
				Detalhe: err.Error(),
			})
			return
		}
		if disp != nil {
			res.Dispositivo = disp
			res.Dispositivos = []DispositivoInfo{*disp}
			*idFisico = pickIdFisico(disp.IdFisico1, disp.IdFisico2)
			if *idFraPtr == "" {
				*idFraPtr = disp.IDFranqueado
			}
			res.Checks = append(res.Checks, Check{
				ID:      "cadastro",
				OK:      true,
				Titulo:  "Cadastro confere — conta e MAC batem",
				Detalhe: formatDispositivoDetalhe(disp, *idFisico),
			})
			return
		}
		m := avaliarMatchCadastro(idFra, fabricante, conta, codigo)
		switch m.Tipo {
		case "similar":
			if dispConta, _ := BuscarDispositivo(idFra, fabricante, conta); dispConta != nil {
				res.Dispositivo = dispConta
				*idFisico = pickIdFisico(dispConta.IdFisico1, dispConta.IdFisico2)
				if *idFraPtr == "" {
					*idFraPtr = dispConta.IDFranqueado
				}
			}
			res.Checks = append(res.Checks, Check{
				ID:      "cadastro",
				OK:      false,
				Alerta:  true,
				Titulo:  "Possivel erro de digitacao no cadastro",
				Detalhe: m.Observacao,
			})
		case "mismatch":
			if dispConta, _ := BuscarDispositivo(idFra, fabricante, conta); dispConta != nil {
				res.Dispositivo = dispConta
				*idFisico = pickIdFisico(dispConta.IdFisico1, dispConta.IdFisico2)
				if *idFraPtr == "" {
					*idFraPtr = dispConta.IDFranqueado
				}
			}
			res.Checks = append(res.Checks, Check{
				ID:      "cadastro",
				OK:      false,
				Titulo:  "Conta existe, mas MAC nao confere",
				Detalhe: m.Observacao,
			})
		case "outra_conta":
			if dispCod, _ := BuscarDispositivoPorImei(idFra, fabricante, codigo); dispCod != nil {
				res.Dispositivo = dispCod
				*idFisico = pickIdFisico(dispCod.IdFisico1, dispCod.IdFisico2)
				if *idFraPtr == "" {
					*idFraPtr = dispCod.IDFranqueado
				}
			}
			res.Checks = append(res.Checks, Check{
				ID:      "cadastro",
				OK:      false,
				Titulo:  "MAC cadastrado em outra conta",
				Detalhe: m.Observacao,
			})
		default:
			res.Checks = append(res.Checks, Check{
				ID:      "cadastro",
				OK:      false,
				Titulo:  "Nenhum cadastro com conta e MAC informados",
				Detalhe: "Confira conta, MAC e fabricante — o receptor rejeita se nao bater exatamente",
			})
		}

	case "mac":
		disp, err := BuscarDispositivoPorImei(idFra, fabricante, codigo)
		if err != nil {
			res.Checks = append(res.Checks, Check{
				ID:      "cadastro",
				OK:      false,
				Alerta:  true,
				Titulo:  "Consulta de cadastro indisponivel",
				Detalhe: err.Error(),
			})
			return
		}
		if disp != nil {
			res.Dispositivo = disp
			res.Dispositivos = []DispositivoInfo{*disp}
			*idFisico = pickIdFisico(disp.IdFisico1, disp.IdFisico2)
			if *idFraPtr == "" {
				*idFraPtr = disp.IDFranqueado
			}
			res.Checks = append(res.Checks, Check{
				ID:      "cadastro",
				OK:      true,
				Titulo:  "Dispositivo encontrado pelo MAC",
				Detalhe: formatDispositivoDetalhe(disp, *idFisico),
			})
			return
		}
		res.Checks = append(res.Checks, Check{
			ID:      "cadastro",
			OK:      false,
			Alerta:  true,
			Titulo:  "MAC nao encontrado no cadastro",
			Detalhe: fmt.Sprintf("Nenhum dispositivo com ID %s para este fabricante", codigo),
		})

	case "conta":
		lista, err := ListarDispositivosPorConta(idFra, fabricante, conta, 10)
		if err != nil {
			res.Checks = append(res.Checks, Check{
				ID:      "cadastro",
				OK:      false,
				Alerta:  true,
				Titulo:  "Consulta de cadastro indisponivel",
				Detalhe: err.Error(),
			})
			return
		}
		res.Dispositivos = lista
		if len(lista) == 0 {
			res.Checks = append(res.Checks, Check{
				ID:      "cadastro",
				OK:      false,
				Alerta:  true,
				Titulo:  "Conta nao encontrada no cadastro",
				Detalhe: fmt.Sprintf("Nenhum dispositivo com conta %s para este fabricante", conta),
			})
			return
		}
		if len(lista) == 1 {
			res.Dispositivo = &lista[0]
			*idFisico = pickIdFisico(lista[0].IdFisico1, lista[0].IdFisico2)
			if *idFraPtr == "" {
				*idFraPtr = lista[0].IDFranqueado
			}
			res.Checks = append(res.Checks, Check{
				ID:      "cadastro",
				OK:      true,
				Titulo:  "Um dispositivo cadastrado nesta conta",
				Detalhe: formatDispositivoDetalhe(&lista[0], *idFisico),
			})
			return
		}
		res.Checks = append(res.Checks, Check{
			ID:      "cadastro",
			OK:      true,
			Alerta:  true,
			Titulo:  fmt.Sprintf("%d dispositivos cadastrados na conta %s", len(lista), conta),
			Detalhe: formatListaDispositivos(lista) + " — informe o MAC para identificar qual central esta instalada",
		})
	}
}

func pickErroRelevante(rows []ErroConexaoRow, conta, codigo, modo string) *ErroConexaoRow {
	if len(rows) == 0 {
		return nil
	}
	if codigo != "" {
		for i := range rows {
			if rows[i].CodigoMatch == "exact" || rows[i].CodigoMatch == "similar" {
				return &rows[i]
			}
		}
	}
	if conta != "" {
		for i := range rows {
			if rows[i].ContaMatch || normalizarConta(rows[i].Conta) == normalizarConta(conta) {
				return &rows[i]
			}
		}
	}
	if modo == "instalacao" || modo == "conta" || modo == "mac" {
		return &rows[0]
	}
	return nil
}
