// Regras padrao (modulos + limites) por produto/plano — fonte canonica
function fn_fp_plano_regras_default {
  input {
    text produto?=franqueadopro filters=trim
    text plano? filters=trim
  }

  stack {
    var $modulos {
      value = {}
    }
  
    var $limites {
      value = {}
    }
  
    var $retencao {
      value = 30
    }
  
    conditional {
      if ($input.produto == "franqueadopro" && $input.plano == "lite") {
        var.update $modulos {
          value = {
            "configuracao.dispositivo"            : true
            "configuracao.usuarios-alarme"        : true
            "configuracao.setores-alarme"         : true
            "configuracao.grade"                  : false
            "configuracao.procedimento"           : false
            "configuracao.contactid"              : false
            "configuracao.email-evento"           : false
            "atendimento.gerar-evento"            : true
            "atendimento.inteligencia-artificial" : false
            "minha-empresa.comercial.cliente"     : true
            "minha-empresa.comercial.pacotes"     : false
            "minha-empresa.operacional"           : false
            "minha-empresa.operacional.tecnico"   : false
            "minha-empresa.operacional.viatura"   : false
            "minha-empresa.gestao.dados"          : true
            "minha-empresa.gestao.responsaveis"   : true
            "minha-empresa.gestao.permissoes"     : false
            "whitelabel"                          : false
            "relatorio.atendimento"               : false
            "relatorio.ligacoes"                  : false
            "relatorio.eventos.lista"             : true
            "relatorio.eventos.config-grupo"      : false
            "relatorio.alarmes.armados"           : true
            "relatorio.alarmes.desarmados"        : true
            "relatorio.clientes.disp"             : true
            "relatorio.clientes.lista"            : true
            "relatorio.clientes.ativos"           : true
            "relatorio.clientes.inativos"         : true
            "dashboard.sem-comunicacao"           : true
            "dashboard.eventos"                   : false
            "franqueadopro.saas"                  : false
          }
        }
      
        var.update $limites {
          value = {
            usuarios_alarme_max: 5
            setores_alarme_max : 4
            clientes_max       : 50
            contas_max         : 100
          }
        }
      
        var.update $retencao {
          value = 30
        }
      }
    
      elseif ($input.produto == "franqueadopro" && $input.plano == "pro") {
        var.update $modulos {
          value = {
            "configuracao.dispositivo"            : true
            "configuracao.usuarios-alarme"        : true
            "configuracao.setores-alarme"         : true
            "configuracao.grade"                  : true
            "configuracao.procedimento"           : false
            "configuracao.contactid"              : false
            "configuracao.email-evento"           : true
            "atendimento.gerar-evento"            : true
            "atendimento.inteligencia-artificial" : false
            "minha-empresa.comercial.cliente"     : true
            "minha-empresa.comercial.pacotes"     : false
            "minha-empresa.operacional"           : true
            "minha-empresa.operacional.tecnico"   : true
            "minha-empresa.operacional.viatura"   : true
            "minha-empresa.gestao.dados"          : true
            "minha-empresa.gestao.responsaveis"   : true
            "minha-empresa.gestao.permissoes"     : false
            "whitelabel"                          : true
            "relatorio.atendimento"               : true
            "relatorio.ligacoes"                  : true
            "relatorio.eventos.lista"             : true
            "relatorio.eventos.config-grupo"      : true
            "relatorio.alarmes.armados"           : true
            "relatorio.alarmes.desarmados"        : true
            "relatorio.clientes.disp"             : true
            "relatorio.clientes.lista"            : true
            "relatorio.clientes.ativos"           : true
            "relatorio.clientes.inativos"         : true
            "dashboard.sem-comunicacao"           : true
            "dashboard.eventos"                   : true
            "franqueadopro.saas"                  : false
          }
        }
      
        var.update $limites {
          value = {
            usuarios_alarme_max: 20
            setores_alarme_max : 15
            clientes_max       : 500
            contas_max         : 1000
          }
        }
      
        var.update $retencao {
          value = 90
        }
      }
    
      elseif ($input.produto == "franqueadopro" && $input.plano == "pro_plus") {
        var.update $modulos {
          value = {
            "configuracao.dispositivo"            : true
            "configuracao.usuarios-alarme"        : true
            "configuracao.setores-alarme"         : true
            "configuracao.grade"                  : true
            "configuracao.procedimento"           : true
            "configuracao.contactid"              : true
            "configuracao.email-evento"           : true
            "atendimento.gerar-evento"            : true
            "atendimento.inteligencia-artificial" : true
            "minha-empresa.comercial.cliente"     : true
            "minha-empresa.comercial.pacotes"     : true
            "minha-empresa.operacional"           : true
            "minha-empresa.operacional.tecnico"   : true
            "minha-empresa.operacional.viatura"   : true
            "minha-empresa.gestao.dados"          : true
            "minha-empresa.gestao.responsaveis"   : true
            "minha-empresa.gestao.permissoes"     : true
            "whitelabel"                          : true
            "relatorio.atendimento"               : true
            "relatorio.ligacoes"                  : true
            "relatorio.eventos.lista"             : true
            "relatorio.eventos.config-grupo"      : true
            "relatorio.alarmes.armados"           : true
            "relatorio.alarmes.desarmados"        : true
            "relatorio.clientes.disp"             : true
            "relatorio.clientes.lista"            : true
            "relatorio.clientes.ativos"           : true
            "relatorio.clientes.inativos"         : true
            "dashboard.sem-comunicacao"           : true
            "dashboard.eventos"                   : true
            "franqueadopro.saas"                  : true
          }
        }
      
        var.update $limites {
          value = {
            usuarios_alarme_max: 999
            setores_alarme_max : 999
            clientes_max       : 9999
            contas_max         : 9999
          }
        }
      
        var.update $retencao {
          value = 180
        }
      }
    
      elseif ($input.produto == "webterminal" && $input.plano == "lite") {
        var.update $limites {
          value = {operadores_max: 3, filas_max: 1}
        }
      
        var.update $retencao {
          value = 30
        }
      }
    
      elseif ($input.produto == "webterminal" && $input.plano == "pro") {
        var.update $limites {
          value = {operadores_max: 15, filas_max: 5}
        }
      
        var.update $retencao {
          value = 90
        }
      }
    
      elseif ($input.produto == "webterminal" && $input.plano == "pro_plus") {
        var.update $limites {
          value = {operadores_max: 999, filas_max: 999}
        }
      
        var.update $retencao {
          value = 180
        }
      }
    
      elseif ($input.produto == "terminalmovel") {
        var.update $limites {
          value = {operadores_max: 20, dispositivos_max: 50}
        }
      
        var.update $retencao {
          value = 30
        }
      }
    
      elseif ($input.produto == "confvision") {
        var.update $limites {
          value = {}
        }
      
        var.update $retencao {
          value = 30
        }
      }
    
      elseif ($input.produto == "webambiente" && $input.plano == "pro") {
        var.update $limites {
          value = {clientes_portal_max: 500}
        }
      
        var.update $retencao {
          value = 30
        }
      }
    
      elseif ($input.produto == "webambiente" && $input.plano == "pro_plus") {
        var.update $limites {
          value = {clientes_portal_max: 9999}
        }
      
        var.update $modulos {
          value = {confvision_embed: true}
        }
      
        var.update $retencao {
          value = 90
        }
      }
    
      elseif ($input.produto == "dialyze") {
        var.update $limites {
          value = {
            usuarios_max       : 10
            canais_whatsapp_max: 3
            contatos_max       : 5000
          }
        }
      
        var.update $retencao {
          value = 30
        }
      }
    }
  }

  response = {
    modulos_json : $modulos
    limites_json : $limites
    retencao_dias: $retencao
  }
}