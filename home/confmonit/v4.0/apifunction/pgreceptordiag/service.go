package pgreceptordiag

import (
	"fmt"
	"strings"
	"time"

	"apifunction/config"
	"apifunction/pgreceptordns"
)

func Verificar(req ReqVerificar) (*Resultado, error) {
	modulo := strings.ToUpper(strings.TrimSpace(req.Modulo))
	fabricante := strings.ToUpper(strings.TrimSpace(req.Fabricante))
	conta := strings.TrimSpace(req.Conta)
	conta = normalizarConta(conta)
	codigo := normalizarCodigoFisico(req.Codigo)
	idFra := strings.TrimSpace(req.IDFranqueado)
	modo := modoDiagnostico(conta, codigo)

	if modulo == "" && fabricante != "" {
		modulo = ModuloForFabricante(fabricante)
	}
	if modulo == "" {
		return nil, fmt.Errorf("modulo ou fabricante obrigatorio")
	}
	if fabricante == "" {
		fabricante = modulo
	}

	porta, err := PortaModulo(modulo)
	if err != nil {
		return nil, err
	}

	res := &Resultado{
		Modulo:         modulo,
		Fabricante:     fabricante,
		Porta:          porta,
		IPTecnico:      strings.TrimSpace(req.IPTecnico),
		Checks:         []Check{},
		ModoReceptor:   modo == "receptor",
		ModoDiag:       modo,
		ContaDigitada:  conta,
		CodigoDigitado: codigo,
	}

	hostPainel := config.ReceptorDNSIP
	enderecoPainel := config.ReceptorDNSIP + ":" + porta
	if idFra != "" {
		reg, _ := pgreceptordns.Obter(idFra)
		if reg != nil && reg.Status == "ativo" && strings.TrimSpace(reg.FQDN) != "" {
			hostPainel = reg.FQDN
			enderecoPainel = reg.FQDN + ":" + porta
		}
	}
	res.HostPainel = hostPainel
	res.EnderecoPainel = enderecoPainel

	res.Checks = append(res.Checks, Check{
		ID:      "config_painel",
		OK:      true,
		Titulo:  "Endereco para configurar no painel",
		Detalhe: fmt.Sprintf("Host: %s | Porta: %s", hostPainel, porta),
	})

	if res.ModoReceptor {
		res.Checks = append(res.Checks, Check{
			ID:      "modo_receptor",
			OK:      true,
			Alerta:  true,
			Titulo:  "Modo teste do receptor",
			Detalhe: "Verificando DNS, porta e servico. Preencha conta e/ou MAC quando souber — o sistema localiza automaticamente.",
		})
	} else {
		titulos := map[string]string{
			"instalacao": "Modo instalacao — conta e MAC informados",
			"mac":        "Modo investigacao — buscando cadastro pelo MAC",
			"conta":      "Modo investigacao — buscando cadastro pela conta",
		}
		res.Checks = append(res.Checks, Check{
			ID:      "modo_diag",
			OK:      true,
			Alerta:  true,
			Titulo:  titulos[modo],
			Detalhe: "Cadastro exige conta + MAC juntos para match exato; abaixo o maximo de pistas com o que foi informado.",
		})
	}

	hostTCP := config.ReceptorDNSIP
	okTCP, cmdTCP := TCPPortaAcessivel(hostTCP, porta, 5*time.Second)
	res.Checks = append(res.Checks, checkFromBool("porta_externa", okTCP, false,
		fmt.Sprintf("Porta %s acessivel pela internet", porta),
		fmt.Sprintf("Porta %s bloqueada ou inacessivel — verifique firewall, DNS ou endereco no painel", porta),
		cmdTCP, ""))

	var idFisico string
	resolverCadastroVerificar(idFra, fabricante, conta, codigo, modo, res, &idFisico, &idFra)

	agent, agentErr := RunAgentChecks(modulo, porta, conta, idFisico)
	if agentErr != nil {
		res.Checks = append(res.Checks, Check{
			ID:      "agente",
			OK:      false,
			Alerta:  true,
			Titulo:  "Testes no servidor receptor indisponiveis",
			Detalhe: agentErr.Error(),
		})
	} else {
		res.Checks = append(res.Checks, checkFromBool("receptor_escuta", agent.ListenOK, false,
			fmt.Sprintf("Receptor %s online (porta %s)", fabricante, porta),
			fmt.Sprintf("Receptor nao escuta a porta %s — servico parado ou porta errada", porta),
			agent.ListenCmd, agent.ListenOut))

		eventos := enrichJournalDrivers(ParseJournalLines(agent.JournalLines), modulo)
		res.JournalEventos = eventos

		centrais := formatEstablishedList(agent.Established)
		if len(centrais) > 0 {
			titulo := fmt.Sprintf("%d centrais conectadas na porta %s", len(centrais), porta)
			detalhe := "Servidor recebendo conexoes. Com conta e MAC preenchidos, veja abaixo se a central ja comunicou."
			if modo == "instalacao" {
				if journalTemConta(eventos, conta) && journalTemCodigo(eventos, codigo) {
					titulo = "Central comunicando no log recente"
					detalhe = fmt.Sprintf("Conta %s e MAC %s apareceram no journal nas ultimas 24h", conta, codigo)
				} else if journalTemConta(eventos, conta) || journalTemCodigo(eventos, codigo) {
					detalhe = "Conta ou MAC apareceu parcialmente no log — confira se painel e cadastro batem"
				} else {
					detalhe = "Sua conta/MAC ainda nao apareceu no log — painel pode estar offline ou DNS errado"
				}
			} else if modo == "mac" && journalTemCodigo(eventos, codigo) {
				titulo = "MAC apareceu no log recente"
				detalhe = fmt.Sprintf("MAC %s comunicou nas ultimas 24h — veja eventos destacados abaixo", codigo)
			} else if modo == "conta" && journalTemConta(eventos, conta) {
				titulo = "Conta apareceu no log recente"
				detalhe = fmt.Sprintf("Conta %s teve eventos nas ultimas 24h — veja eventos destacados abaixo", conta)
			}
			res.Checks = append(res.Checks, Check{
				ID:      "conexao_ativa",
				OK:      true,
				Titulo:  titulo,
				Detalhe: detalhe,
				Comando: agent.EstablishedCmd,
				Saida:   agent.EstablishedOut,
			})
		} else {
			res.Checks = append(res.Checks, Check{
				ID:      "conexao_ativa",
				OK:      false,
				Alerta:  true,
				Titulo:  "Nenhuma central conectada agora",
				Detalhe: "Painel pode estar offline ou com host/porta incorretos no chip/GPRS",
				Comando: agent.EstablishedCmd,
				Saida:   agent.EstablishedOut,
			})
		}

		if len(eventos) > 0 {
			res.Checks = append(res.Checks, Check{
				ID:      "journal",
				OK:      true,
				Titulo:  fmt.Sprintf("Eventos recentes no receptor (%d)", len(eventos)),
				Detalhe: "Ultimas 24 horas — veja a lista completa abaixo",
				Comando: agent.JournalCmd,
				Saida:   strings.Join(agent.JournalLines, "\n"),
			})
		} else {
			res.Checks = append(res.Checks, Check{
				ID:      "journal",
				OK:      false,
				Alerta:  true,
				Titulo:  "Nenhum evento recente no log do receptor",
				Detalhe: fmt.Sprintf("Sem registros %s nas ultimas 24 horas", fabricante),
				Comando: agent.JournalCmd,
			})
		}

		if agent.ProtocolCmd != "" {
			res.Checks = append(res.Checks, checkFromBool("protocolo", agent.ProtocolOK, false,
				fmt.Sprintf("Protocolo %s respondendo", fabricante),
				fmt.Sprintf("Teste de protocolo %s falhou", fabricante),
				agent.ProtocolCmd, agent.ProtocolOut))
		}
	}

	if conta != "" {
		pres, _ := ObterPresenca(fabricante, conta)
		res.Presenca = pres
		if pres != nil && pres.IPRemoto != "" && res.IPCentral == "" {
			res.IPCentral = SanitizeIPRemoto(pres.IPRemoto)
		}
		if pres != nil && pres.UltimoSinal != nil {
			ago := time.Since(*pres.UltimoSinal)
			ok := ago < 5*time.Minute
			res.Checks = append(res.Checks, checkFromBool("ultimo_sinal", ok, !ok,
				fmt.Sprintf("Ultimo sinal ha %s", formatDuracao(ago)),
				fmt.Sprintf("Sem sinal ha %s", formatDuracao(ago)),
				"", pres.UltimoSinal.Format("02/01/2006 15:04:05")))
		}
	}

	var erros []ErroConexaoRow
	limiteErro := diagErroConexaoLimite
	if req.Lab {
		limiteErro = diagErroConexaoPageSize
	}
	erros, errQ := ListarErroConexao(idFra, fabricante, conta, codigo, limiteErro, 0)
	if errQ != nil {
		res.ErrosConexao = []ErroConexaoRow{}
		res.Checks = append(res.Checks, Check{
			ID:      "erro_conexao",
			OK:      false,
			Alerta:  true,
			Titulo:  "Consulta erroConexao indisponivel",
			Detalhe: errQ.Error(),
		})
	} else {
		erros = sanitizeErroConexaoRows(erros)
		if erros == nil {
			erros = []ErroConexaoRow{}
		}
		erros = EnrichErroConexaoMatch(idFra, fabricante, conta, codigo, erros)
		res.ErrosConexao = erros

		if modo == "receptor" {
			if len(erros) > 0 {
				res.Checks = append(res.Checks, Check{
					ID:      "erro_conexao",
					OK:      true,
					Alerta:  true,
					Titulo:  fmt.Sprintf("%d erros de conexao hoje", len(erros)),
					Detalhe: "Erros do fabricante hoje — role a tabela na aba Erros para carregar todos",
				})
			} else {
				detalhe := "Tabela erroConexao vazia para este fabricante hoje"
				if req.Lab {
					if total, porFab, errC := ContagemErroConexaoHoje(fabricante); errC == nil {
						detalhe = fmt.Sprintf(
							"%s | MySQL apifunction: %d erros hoje (total), %d deste fabricante",
							detalhe, total, porFab,
						)
					}
				}
				res.Checks = append(res.Checks, Check{
					ID:      "erro_conexao",
					OK:      true,
					Titulo:  "Nenhum erro de conexao hoje",
					Detalhe: detalhe,
				})
			}
		} else if len(erros) > 0 {
			e := pickErroRelevante(erros, conta, codigo, modo)
			if e != nil {
				if res.Dispositivo == nil && strings.TrimSpace(e.Codigo) != "" {
					if disp, _ := BuscarDispositivoPorImei(idFra, fabricante, e.Codigo); disp != nil {
						res.Dispositivo = disp
						idFisico = pickIdFisico(disp.IdFisico1, disp.IdFisico2)
						if idFra == "" {
							idFra = disp.IDFranqueado
						}
					}
				}

				if res.IPCentral == "" {
					res.IPCentral = SanitizeIPRemoto(e.IPRemoto)
				}
				detalhe := formatErroCadastroDetalhe(*e, res.Dispositivo)
				if e.MatchObservacao != "" {
					detalhe += " | " + e.MatchObservacao
				}
				res.Checks = append(res.Checks, Check{
					ID:      "erro_conexao",
					OK:      false,
					Titulo:  "Cadastro nao confere — central conectou com IMEI/conta errados",
					Detalhe: detalhe,
				})
			} else if codigo != "" {
				res.Checks = append(res.Checks, Check{
					ID:      "erro_conexao",
					OK:      true,
					Alerta:  true,
					Titulo:  fmt.Sprintf("%d erros hoje — MAC %s nao encontrado neles", len(erros), codigo),
					Detalhe: "Veja a tabela — linhas destacadas em verde limao batem com conta ou MAC informados",
				})
			}
		} else {
			msg := "Nenhum erro de cadastro hoje"
			if conta != "" && codigo != "" {
				msg = fmt.Sprintf("Nenhum erro hoje para conta %s — central pode ainda nao ter tentado conectar", conta)
			} else if codigo != "" {
				msg = fmt.Sprintf("Nenhum erro hoje para este fabricante com MAC %s", codigo)
			} else if conta != "" {
				msg = fmt.Sprintf("Nenhum erro hoje para conta %s", conta)
			}
			res.Checks = append(res.Checks, Check{
				ID:      "erro_conexao",
				OK:      true,
				Titulo:  "Nenhum erro de cadastro recente",
				Detalhe: msg,
			})
		}
	}

	if res.IPTecnico != "" && res.IPCentral != "" && modo == "instalacao" {
		mesma := res.IPTecnico == res.IPCentral
		res.MesmaRede = &mesma
		msg := "Seu IP e o IP da central sao iguais — mesma internet (rede local ou mesmo roteador)"
		if !mesma {
			msg = "IPs diferentes — normal se a central usa chip GPRS ou outra internet"
		}
		res.Checks = append(res.Checks, Check{
			ID:      "mesma_rede",
			OK:      mesma,
			Alerta:  !mesma,
			Titulo:  "Comparacao de rede",
			Detalhe: fmt.Sprintf("Seu IP: %s | IP da central: %s. %s", res.IPTecnico, res.IPCentral, msg),
		})
	} else if res.IPTecnico != "" && modo == "instalacao" {
		res.Checks = append(res.Checks, Check{
			ID:      "mesma_rede",
			OK:      false,
			Alerta:  true,
			Titulo:  "Conecte na mesma rede da central",
			Detalhe: fmt.Sprintf("Seu IP agora: %s. Ainda nao ha IP da central registrado para comparar.", res.IPTecnico),
		})
	}

	res.Status, res.Conclusao, res.Acao = concluir(res)
	_ = LogDiagnostico(idFra, fabricante, conta, res.IPTecnico, res.Status, res.Conclusao)
	return res, nil
}

func formatDispositivoDetalhe(d *DispositivoInfo, idFisico string) string {
	parts := []string{
		fmt.Sprintf("Conta %s | IMEI/ID fisico: %s", d.Conta, idFisico),
	}
	if d.NomeCliente != "" {
		parts = append(parts, fmt.Sprintf("Cliente: %s", d.NomeCliente))
	}
	return strings.Join(parts, " | ")
}

func formatErroCadastroDetalhe(e ErroConexaoRow, disp *DispositivoInfo) string {
	parts := []string{
		fmt.Sprintf("Codigo recebido: %s", e.Codigo),
	}
	if s := strings.TrimSpace(e.Serial); s != "" && s != e.Codigo {
		parts = append(parts, "Serial: "+s)
	}
	if s := strings.TrimSpace(e.Imei); s != "" && s != e.Codigo {
		parts = append(parts, "Imei: "+s)
	}
	if s := strings.TrimSpace(e.Modelo); s != "" {
		parts = append(parts, "Modelo: "+s)
	}
	if s := strings.TrimSpace(e.Versao); s != "" {
		parts = append(parts, "Versao: "+s)
	}
	if s := strings.TrimSpace(e.Protocolo); s != "" {
		parts = append(parts, "Protocolo: "+s)
	}
	if s := strings.TrimSpace(e.Fabricante); s != "" {
		parts = append(parts, "Fabricante: "+s)
	}
	if s := strings.TrimSpace(e.Conta); s != "" {
		parts = append(parts, "Conta: "+s)
	}
	if s := strings.TrimSpace(e.IDFranqueado); s != "" {
		parts = append(parts, "ID franqueado: "+s)
	}
	if ip := strings.TrimSpace(e.IPRemoto); ip != "" {
		parts = append(parts, "IP da central: "+ip)
	} else {
		parts = append(parts, "IP da central: nao identificado (pode ser teste local)")
	}
	parts = append(parts, "Quando: "+e.DataTentativa)
	if disp != nil {
		cad := pickIdFisico(disp.IdFisico1, disp.IdFisico2)
		parts = append(parts, fmt.Sprintf("Cadastrado no sistema: %s", cad))
		if disp.NomeCliente != "" {
			parts = append(parts, fmt.Sprintf("Cliente: %s", disp.NomeCliente))
		}
	}
	return strings.Join(parts, " | ")
}

func formatEstablishedList(peers []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range peers {
		ip := formatEstablishedPeer(p)
		if ip == "" || seen[ip] {
			continue
		}
		seen[ip] = true
		out = append(out, ip)
	}
	return out
}

func concluir(r *Resultado) (status, conclusao, acao string) {
	hasFail := false
	hasAlert := false
	for _, c := range r.Checks {
		if !c.OK && !c.Alerta {
			hasFail = true
		}
		if c.Alerta {
			hasAlert = true
		}
	}
	for _, c := range r.Checks {
		if c.ID == "erro_conexao" && !c.OK {
			return "erro", "Central conecta, mas conta ou IMEI estao errados no painel ou cadastro",
				"Corrija o IMEI/conta no painel ou no cadastro do dispositivo"
		}
		if c.ID == "porta_externa" && !c.OK {
			return "erro", "Porta do receptor inacessivel pela internet",
				"Verifique firewall, DNS do franqueado e endereco configurado no painel: " + r.EnderecoPainel
		}
		if c.ID == "receptor_escuta" && !c.OK {
			return "erro", "Receptor nao esta escutando a porta do fabricante",
				"Acione a central de monitoramento — servico receptor4.0 pode estar parado"
		}
	}
	if r.ModoReceptor && !hasFail {
		return "alerta",
			"Modo so receptor — ultimos eventos e erros do fabricante. Com conta e/ou MAC preenchidos, a analise foca na central.",
			"Configure o painel: " + r.EnderecoPainel
	}
	if hasFail {
		return "erro", "Instalacao com problemas — revise os itens em vermelho", "Siga as acoes indicadas em cada item"
	}
	if hasAlert {
		return "alerta", "Receptor operacional com ressalvas — verifique itens em amarelo", ""
	}
	return "ok", "Central / receptor operando normalmente", "Configure o painel: " + r.EnderecoPainel
}

func checkFromBool(id string, ok, alerta bool, okTitulo, failTitulo, cmd, saida string) Check {
	c := Check{ID: id, OK: ok, Alerta: alerta && ok, Comando: cmd, Saida: saida}
	if ok {
		c.Titulo = okTitulo
	} else {
		c.Titulo = failTitulo
	}
	return c
}

func pickIdFisico(a, b string) string {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a != "" {
		return a
	}
	return b
}

func extractIP(s string) string {
	return SanitizeIPRemoto(s)
}

func formatDuracao(d time.Duration) string {
	if d < time.Minute {
		return "menos de 1 min"
	}
	if d < time.Hour {
		return fmt.Sprintf("%d min", int(d.Minutes()))
	}
	return fmt.Sprintf("%d h", int(d.Hours()))
}
