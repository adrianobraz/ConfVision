// Seed idempotente do catalogo de modulos a la carte (precos padrao)
function fn_fp_modulo_catalogo_seed {
  input {
    text user_tipo? filters=trim
    text id_vinculo? filters=trim
  }

  stack {
    var $id_central {
      value = "CENTRAL"
    }
  
    var $id_representante {
      value = ""
    }
  
    var $tipo {
      value = $input.user_tipo|first_notempty:"CEN"|to_upper
    }
  
    conditional {
      if ($tipo == "CEN" && ($input.id_vinculo|is_empty) == false) {
        var.update $id_central {
          value = $input.id_vinculo
        }
      }
    }
  
    precondition ($tipo == "CEN") {
      error = "Somente a Central pode criar o catalogo de modulos a la carte"
    }
  
    var $inseridos {
      value = 0
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "configuracao.dispositivo"
        label       : "Cadastro de Alarmes"
        grupo       : "Parâmetros Técnicos"
        descricao   : "Cadastro e gestão de dispositivos/contas de alarme."
        plano_minimo: "lite"
        valor_mensal: 0
        ordem       : 10
      }
    } as $m1
  
    var.update $inseridos {
      value = $inseridos + $m1.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "configuracao.usuarios-alarme"
        label       : "Usuários de Alarme"
        grupo       : "Parâmetros Técnicos"
        descricao   : "Usuários vinculados aos painéis de alarme."
        plano_minimo: "lite"
        valor_mensal: 0
        ordem       : 11
      }
    } as $m2
  
    var.update $inseridos {
      value = $inseridos + $m2.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "configuracao.setores-alarme"
        label       : "Setores de Alarme"
        grupo       : "Parâmetros Técnicos"
        descricao   : "Partições e setores dos alarmes."
        plano_minimo: "lite"
        valor_mensal: 0
        ordem       : 12
      }
    } as $m3
  
    var.update $inseridos {
      value = $inseridos + $m3.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "configuracao.grade"
        label       : "Grade Horário"
        grupo       : "Parâmetros Técnicos"
        descricao   : "Grades de horário de funcionamento e armado."
        plano_minimo: "pro"
        valor_mensal: 25
        ordem       : 13
      }
    } as $m4
  
    var.update $inseridos {
      value = $inseridos + $m4.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "configuracao.procedimento"
        label       : "Procedimento de Atendimento"
        grupo       : "Parâmetros Técnicos"
        descricao   : "Roteiros padrão para operadores no atendimento."
        plano_minimo: "pro_plus"
        valor_mensal: 45
        ordem       : 14
      }
    } as $m5
  
    var.update $inseridos {
      value = $inseridos + $m5.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "configuracao.contactid"
        label       : "Contact ID Personalizado"
        grupo       : "Parâmetros Técnicos"
        descricao   : "Códigos Contact ID customizados por central."
        plano_minimo: "pro_plus"
        valor_mensal: 45
        ordem       : 15
      }
    } as $m6
  
    var.update $inseridos {
      value = $inseridos + $m6.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "configuracao.email-evento"
        label       : "E-mail por Evento"
        grupo       : "Parâmetros Técnicos"
        descricao   : "Envio automático de e-mail quando eventos ocorrem."
        plano_minimo: "pro"
        valor_mensal: 25
        ordem       : 16
      }
    } as $m7
  
    var.update $inseridos {
      value = $inseridos + $m7.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "atendimento.gerar-evento"
        label       : "Gerar Evento"
        grupo       : "Atendimento"
        descricao   : "Geração manual de eventos para clientes."
        plano_minimo: "lite"
        valor_mensal: 0
        ordem       : 20
      }
    } as $m8
  
    var.update $inseridos {
      value = $inseridos + $m8.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "atendimento.inteligencia-artificial"
        label       : "Inteligência Artificial"
        grupo       : "Atendimento"
        descricao   : "Alertas por telefone, SMS e bloqueios inteligentes."
        plano_minimo: "pro_plus"
        valor_mensal: 55
        ordem       : 21
      }
    } as $m9
  
    var.update $inseridos {
      value = $inseridos + $m9.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "minha-empresa.comercial.cliente"
        label       : "Cadastro de Clientes"
        grupo       : "Comercial"
        descricao   : "Cadastro e gestão de clientes da central."
        plano_minimo: "lite"
        valor_mensal: 0
        ordem       : 30
      }
    } as $m10
  
    var.update $inseridos {
      value = $inseridos + $m10.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "minha-empresa.comercial.pacotes"
        label       : "Gerenciar Pacotes"
        grupo       : "Comercial"
        descricao   : "Pacotes comerciais com valor e quantidade de contas."
        plano_minimo: "pro_plus"
        valor_mensal: 45
        ordem       : 31
      }
    } as $m11
  
    var.update $inseridos {
      value = $inseridos + $m11.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "minha-empresa.operacional"
        label       : "Operacional / Campo"
        grupo       : "Operacional"
        descricao   : "Área de técnicos e viaturas em campo."
        plano_minimo: "pro"
        valor_mensal: 30
        ordem       : 40
      }
    } as $m12
  
    var.update $inseridos {
      value = $inseridos + $m12.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "minha-empresa.operacional.tecnico"
        label       : "Cadastro de Técnico"
        grupo       : "Operacional"
        descricao   : "Cadastro de técnicos que atendem em campo."
        plano_minimo: "pro"
        valor_mensal: 20
        ordem       : 41
      }
    } as $m13
  
    var.update $inseridos {
      value = $inseridos + $m13.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "minha-empresa.operacional.viatura"
        label       : "Cadastro de Viatura"
        grupo       : "Operacional"
        descricao   : "Veículos utilizados nas ordens de serviço."
        plano_minimo: "pro"
        valor_mensal: 20
        ordem       : 42
      }
    } as $m14
  
    var.update $inseridos {
      value = $inseridos + $m14.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "minha-empresa.gestao.dados"
        label       : "Dados do Monitoramento"
        grupo       : "Gestão Interna"
        descricao   : "Informações e parâmetros da empresa."
        plano_minimo: "lite"
        valor_mensal: 0
        ordem       : 50
      }
    } as $m15
  
    var.update $inseridos {
      value = $inseridos + $m15.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "minha-empresa.gestao.responsaveis"
        label       : "Gerenciar Responsáveis"
        grupo       : "Gestão Interna"
        descricao   : "Responsáveis legais e contatos da central."
        plano_minimo: "lite"
        valor_mensal: 0
        ordem       : 51
      }
    } as $m16
  
    var.update $inseridos {
      value = $inseridos + $m16.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "minha-empresa.gestao.meu-plano"
        label       : "Meu Plano"
        grupo       : "Gestão Interna"
        descricao   : "Plano contratado, licença e resumo."
        plano_minimo: "lite"
        valor_mensal: 0
        ordem       : 531
      }
    } as $m17a
  
    var.update $inseridos {
      value = $inseridos + $m17a.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "minha-empresa.gestao.montar-plano"
        label       : "Montar Meu Plano"
        grupo       : "Gestão Interna"
        descricao   : "Contratação à la carte de módulos."
        plano_minimo: "lite"
        valor_mensal: 0
        ordem       : 532
      }
    } as $m17b
  
    var.update $inseridos {
      value = $inseridos + $m17b.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "minha-empresa.gestao.faturas"
        label       : "Faturas e Cobrança"
        grupo       : "Gestão Interna"
        descricao   : "Histórico e faturas em aberto."
        plano_minimo: "lite"
        valor_mensal: 0
        ordem       : 533
      }
    } as $m17c
  
    var.update $inseridos {
      value = $inseridos + $m17c.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "minha-empresa.gestao.permissoes"
        label       : "Permissões de Acesso"
        grupo       : "Gestão Interna"
        descricao   : "Controle fino de menus por usuário."
        plano_minimo: "pro_plus"
        valor_mensal: 50
        ordem       : 53
      }
    } as $m18
  
    var.update $inseridos {
      value = $inseridos + $m18.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "whitelabel"
        label       : "White Label"
        grupo       : "Gestão Interna"
        descricao   : "Personalização de cores e logotipo da marca."
        plano_minimo: "pro"
        valor_mensal: 80
        ordem       : 54
      }
    } as $m19
  
    var.update $inseridos {
      value = $inseridos + $m19.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "franqueadopro.saas"
        label       : "Domínio Personalizado"
        grupo       : "Gestão Interna"
        descricao   : "Endereço próprio na nuvem com Apache e SSL automáticos."
        plano_minimo: "pro_plus"
        valor_mensal: 120
        ordem       : 55
      }
    } as $m20
  
    var.update $inseridos {
      value = $inseridos + $m20.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "relatorio.atendimento"
        label       : "Relatório de Atendimento"
        grupo       : "Relatórios"
        descricao   : "Histórico detalhado de atendimentos realizados."
        plano_minimo: "pro"
        valor_mensal: 25
        ordem       : 60
      }
    } as $m21
  
    var.update $inseridos {
      value = $inseridos + $m21.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "relatorio.ligacoes"
        label       : "Lista de Ligações"
        grupo       : "Relatórios"
        descricao   : "Relatório de ligações telefônicas."
        plano_minimo: "pro"
        valor_mensal: 25
        ordem       : 61
      }
    } as $m22
  
    var.update $inseridos {
      value = $inseridos + $m22.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "relatorio.eventos.lista"
        label       : "Relatório de Eventos"
        grupo       : "Relatórios"
        descricao   : "Lista de eventos recebidos no período."
        plano_minimo: "lite"
        valor_mensal: 0
        ordem       : 62
      }
    } as $m23
  
    var.update $inseridos {
      value = $inseridos + $m23.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "relatorio.eventos.config-grupo"
        label       : "Eventos por Grupo"
        grupo       : "Relatórios"
        descricao   : "Relatório avançado agrupado por grupo de eventos."
        plano_minimo: "pro"
        valor_mensal: 25
        ordem       : 63
      }
    } as $m24
  
    var.update $inseridos {
      value = $inseridos + $m24.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "relatorio.alarmes.armados"
        label       : "Alarmes Armados"
        grupo       : "Relatórios"
        descricao   : "Dispositivos com alarme armado."
        plano_minimo: "lite"
        valor_mensal: 0
        ordem       : 64
      }
    } as $m25
  
    var.update $inseridos {
      value = $inseridos + $m25.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "relatorio.alarmes.desarmados"
        label       : "Alarmes Desarmados"
        grupo       : "Relatórios"
        descricao   : "Dispositivos com alarme desarmado."
        plano_minimo: "lite"
        valor_mensal: 0
        ordem       : 65
      }
    } as $m26
  
    var.update $inseridos {
      value = $inseridos + $m26.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "relatorio.clientes.disp"
        label       : "Clientes e Dispositivos"
        grupo       : "Relatórios"
        descricao   : "Relação de clientes e suas contas."
        plano_minimo: "lite"
        valor_mensal: 0
        ordem       : 66
      }
    } as $m27
  
    var.update $inseridos {
      value = $inseridos + $m27.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "relatorio.clientes.lista"
        label       : "Lista de Clientes"
        grupo       : "Relatórios"
        descricao   : "Todos os clientes cadastrados."
        plano_minimo: "lite"
        valor_mensal: 0
        ordem       : 67
      }
    } as $m28
  
    var.update $inseridos {
      value = $inseridos + $m28.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "relatorio.clientes.ativos"
        label       : "Clientes Ativos"
        grupo       : "Relatórios"
        descricao   : "Clientes com situação ativa."
        plano_minimo: "lite"
        valor_mensal: 0
        ordem       : 68
      }
    } as $m29
  
    var.update $inseridos {
      value = $inseridos + $m29.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "relatorio.clientes.inativos"
        label       : "Clientes Inativos"
        grupo       : "Relatórios"
        descricao   : "Clientes inativos ou cancelados."
        plano_minimo: "lite"
        valor_mensal: 0
        ordem       : 69
      }
    } as $m30
  
    var.update $inseridos {
      value = $inseridos + $m30.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "dashboard.sem-comunicacao"
        label       : "Dashboard Sem Comunicação"
        grupo       : "Dashboard"
        descricao   : "Painel de contas sem comunicação por faixa de tempo."
        plano_minimo: "lite"
        valor_mensal: 0
        ordem       : 70
      }
    } as $m31
  
    var.update $inseridos {
      value = $inseridos + $m31.inserido
    }
  
    function.run fn_fp_modulo_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto     : "franqueadopro"
        chave       : "dashboard.eventos"
        label       : "Dashboard Eventos (7 dias)"
        grupo       : "Dashboard"
        descricao   : "KPIs de eventos por grupo nos últimos 7 dias."
        plano_minimo: "pro"
        valor_mensal: 30
        ordem       : 71
      }
    } as $m32
  
    var.update $inseridos {
      value = $inseridos + $m32.inserido
    }
  }

  response = {inseridos: $inseridos, ok: true, id_central: $id_central, user_tipo: $tipo}
}