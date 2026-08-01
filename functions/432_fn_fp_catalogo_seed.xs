// Seed inicial do catalogo Lite/Pro/Pro+ (idempotente) - modulos espelham fn_fp_plano_regras_default
function fn_fp_catalogo_seed {
  input {
    text user_tipo? filters=trim
    text id_vinculo? filters=trim
    // assinaturas | confvision_licenca | tudo
    text escopo?=assinaturas filters=trim
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
      error = "Somente a Central pode criar o catalogo piso"
    }
  
    var $escopo {
      value = $input.escopo|first_notempty:"assinaturas"|to_lower
    }
  
    precondition ($escopo == "assinaturas" || $escopo == "confvision_licenca" || $escopo == "tudo") {
      error = "escopo invalido — use assinaturas, confvision_licenca ou tudo"
    }
  
    var $inseridos {
      value = 0
    }
  
    conditional {
      if ($escopo == "assinaturas" || $escopo == "tudo") {
        function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto      : "franqueadopro"
        plano        : "lite"
        nome_exibicao: "FranqueadoPro Lite"
        valor_mensal : 99
        retencao_dias: 30
        limites_json : {
        usuarios_alarme_max: 5
        setores_alarme_max : 4
        clientes_max       : 50
        contas_max         : 100
      }
        modulos_json : {
        configuracao.grade                   : false
        configuracao.procedimento            : false
        configuracao.contactid               : false
        "configuracao.email-evento"          : false
        "atendimento.inteligencia-artificial": false
        "minha-empresa.comercial.pacotes"    : false
        "minha-empresa.operacional"          : false
        "minha-empresa.operacional.tecnico"  : false
        "minha-empresa.operacional.viatura"  : false
        "minha-empresa.gestao.permissoes"    : false
        whitelabel                           : false
        relatorio.atendimento                : false
        relatorio.ligacoes                   : false
        "relatorio.eventos.config-grupo"     : false
        dashboard.eventos                    : false
        franqueadopro.saas                   : false
      }
      }
    } as $s1
  
    var.update $inseridos {
      value = $inseridos + $s1.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto      : "franqueadopro"
        plano        : "pro"
        nome_exibicao: "FranqueadoPro Pro"
        valor_mensal : 199
        retencao_dias: 90
        limites_json : {
        usuarios_alarme_max: 20
        setores_alarme_max : 15
        clientes_max       : 500
        contas_max         : 1000
      }
        modulos_json : {
        configuracao.procedimento            : false
        configuracao.contactid               : false
        "atendimento.inteligencia-artificial": false
        "minha-empresa.comercial.pacotes"    : false
        "minha-empresa.gestao.permissoes"    : false
        franqueadopro.saas                   : false
        whitelabel                           : true
      }
      }
    } as $s2
  
    var.update $inseridos {
      value = $inseridos + $s2.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto      : "franqueadopro"
        plano        : "pro_plus"
        nome_exibicao: "FranqueadoPro Pro+"
        valor_mensal : 299
        retencao_dias: 180
        limites_json : {
        usuarios_alarme_max: 999
        setores_alarme_max : 999
        clientes_max       : 9999
        contas_max         : 9999
      }
        modulos_json : {
        configuracao.grade                   : true
        configuracao.procedimento            : true
        configuracao.contactid               : true
        "configuracao.email-evento"          : true
        "atendimento.inteligencia-artificial": true
        "minha-empresa.gestao.permissoes"    : true
        whitelabel                           : true
        franqueadopro.saas                   : true
      }
      }
    } as $s3
  
    var.update $inseridos {
      value = $inseridos + $s3.inserido
    }
  
    // --- Ecossistema: assinaturas (somente_inserir = preserva edicao no adm) ---
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "webterminal"
        plano          : "lite"
        nome_exibicao  : "WebTerminal Lite"
        valor_mensal   : 49
        retencao_dias  : 30
        limites_json   : {operadores_max: 3, filas_max: 1}
        somente_inserir: true
      }
    } as $s4
  
    var.update $inseridos {
      value = $inseridos + $s4.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "webterminal"
        plano          : "pro"
        nome_exibicao  : "WebTerminal Pro"
        valor_mensal   : 79
        retencao_dias  : 90
        limites_json   : {operadores_max: 15, filas_max: 5}
        somente_inserir: true
      }
    } as $s5
  
    var.update $inseridos {
      value = $inseridos + $s5.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "webterminal"
        plano          : "pro_plus"
        nome_exibicao  : "WebTerminal Pro+"
        valor_mensal   : 129
        retencao_dias  : 180
        limites_json   : {operadores_max: 999, filas_max: 999}
        somente_inserir: true
      }
    } as $s6
  
    var.update $inseridos {
      value = $inseridos + $s6.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "terminalmovel"
        plano          : "padrao"
        nome_exibicao  : "Terminal Movel"
        valor_mensal   : 59
        retencao_dias  : 30
        limites_json   : {operadores_max: 20, dispositivos_max: 50}
        somente_inserir: true
      }
    } as $s7
  
    var.update $inseridos {
      value = $inseridos + $s7.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "confvision"
        plano          : "padrao"
        nome_exibicao  : "ConfVision - Plano"
        valor_mensal   : 79
        retencao_dias  : 30
        somente_inserir: true
      }
    } as $s8
  
    var.update $inseridos {
      value = $inseridos + $s8.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "webambiente"
        plano          : "pro"
        nome_exibicao  : "webAmbiente Pro"
        valor_mensal   : 69
        retencao_dias  : 30
        limites_json   : {clientes_portal_max: 500}
        somente_inserir: true
      }
    } as $s9
  
    var.update $inseridos {
      value = $inseridos + $s9.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "webambiente"
        plano          : "pro_plus"
        nome_exibicao  : "webAmbiente Pro+"
        valor_mensal   : 99
        retencao_dias  : 90
        limites_json   : {clientes_portal_max: 9999}
        modulos_json   : {confvision_embed: true}
        somente_inserir: true
      }
    } as $s10
  
    var.update $inseridos {
      value = $inseridos + $s10.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "dialyze"
        plano          : "padrao"
        nome_exibicao  : "Dialyze"
        valor_mensal   : 149
        retencao_dias  : 30
        limites_json   : {
        usuarios_max       : 10
        canais_whatsapp_max: 3
        contatos_max       : 5000
      }
        somente_inserir: true
      }
    } as $s11
  
    var.update $inseridos {
      value = $inseridos + $s11.inserido
    }
      }
    }
  
    conditional {
      if ($escopo == "confvision_licenca" || $escopo == "tudo") {
        // --- ConfVision licencas camera/gravacao (precos editaveis no adm) ---
        function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "confvision_licenca"
        plano          : "online"
        nome_exibicao  : "Camera online"
        valor_mensal   : 2.99
        retencao_dias  : 30
        limites_json   : {unidade: "camera"}
        somente_inserir: true
      }
    } as $cv1
  
    var.update $inseridos {
      value = $inseridos + $cv1.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "confvision_licenca"
        plano          : "sensor_foto"
        nome_exibicao  : "Sensor foto"
        valor_mensal   : 7.99
        retencao_dias  : 30
        limites_json   : {unidade: "camera"}
        somente_inserir: true
      }
    } as $cv2
  
    var.update $inseridos {
      value = $inseridos + $cv2.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "confvision_licenca"
        plano          : "sensor_foto_video"
        nome_exibicao  : "Sensor foto + video"
        valor_mensal   : 9.99
        retencao_dias  : 30
        limites_json   : {unidade: "camera"}
        somente_inserir: true
      }
    } as $cv3
  
    var.update $inseridos {
      value = $inseridos + $cv3.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "confvision_licenca"
        plano          : "analitico_armado_evento"
        nome_exibicao  : "Analitico armado - so evento"
        valor_mensal   : 11.99
        retencao_dias  : 30
        limites_json   : {unidade: "camera"}
        somente_inserir: true
      }
    } as $cv4
  
    var.update $inseridos {
      value = $inseridos + $cv4.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "confvision_licenca"
        plano          : "analitico_armado_foto"
        nome_exibicao  : "Analitico armado - foto"
        valor_mensal   : 13.99
        retencao_dias  : 30
        limites_json   : {unidade: "camera"}
        somente_inserir: true
      }
    } as $cv5
  
    var.update $inseridos {
      value = $inseridos + $cv5.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "confvision_licenca"
        plano          : "analitico_armado_foto_video"
        nome_exibicao  : "Analitico armado - foto + video"
        valor_mensal   : 14.99
        retencao_dias  : 30
        limites_json   : {unidade: "camera"}
        somente_inserir: true
      }
    } as $cv6
  
    var.update $inseridos {
      value = $inseridos + $cv6.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "confvision_licenca"
        plano          : "analitico_24h_evento"
        nome_exibicao  : "Analitico 24h - so evento"
        valor_mensal   : 16.99
        retencao_dias  : 30
        limites_json   : {unidade: "camera"}
        somente_inserir: true
      }
    } as $cv7
  
    var.update $inseridos {
      value = $inseridos + $cv7.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "confvision_licenca"
        plano          : "analitico_24h_foto"
        nome_exibicao  : "Analitico 24h - foto"
        valor_mensal   : 18.99
        retencao_dias  : 30
        limites_json   : {unidade: "camera"}
        somente_inserir: true
      }
    } as $cv8
  
    var.update $inseridos {
      value = $inseridos + $cv8.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "confvision_licenca"
        plano          : "analitico_24h_foto_video"
        nome_exibicao  : "Analitico 24h - foto + video"
        valor_mensal   : 19.99
        retencao_dias  : 30
        limites_json   : {unidade: "camera"}
        somente_inserir: true
      }
    } as $cv9
  
    var.update $inseridos {
      value = $inseridos + $cv9.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "confvision_licenca"
        plano          : "gravacao_7d"
        nome_exibicao  : "Gravacao continua 7 dias"
        valor_mensal   : 12.99
        retencao_dias  : 30
        limites_json   : {unidade: "gravacao"}
        somente_inserir: true
      }
    } as $cv10
  
    var.update $inseridos {
      value = $inseridos + $cv10.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "confvision_licenca"
        plano          : "gravacao_15d"
        nome_exibicao  : "Gravacao continua 15 dias"
        valor_mensal   : 17.99
        retencao_dias  : 30
        limites_json   : {unidade: "gravacao"}
        somente_inserir: true
      }
    } as $cv11
  
    var.update $inseridos {
      value = $inseridos + $cv11.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "confvision_licenca"
        plano          : "gravacao_30d"
        nome_exibicao  : "Gravacao continua 30 dias"
        valor_mensal   : 24.99
        retencao_dias  : 30
        limites_json   : {unidade: "gravacao"}
        somente_inserir: true
      }
    } as $cv12
  
    var.update $inseridos {
      value = $inseridos + $cv12.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "confvision_licenca"
        plano          : "gravacao_movimento_7d"
        nome_exibicao  : "Gravacao por movimento 7 dias"
        valor_mensal   : 9.99
        retencao_dias  : 30
        limites_json   : {unidade: "gravacao"}
        somente_inserir: true
      }
    } as $cv13
  
    var.update $inseridos {
      value = $inseridos + $cv13.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "confvision_licenca"
        plano          : "gravacao_movimento_15d"
        nome_exibicao  : "Gravacao por movimento 15 dias"
        valor_mensal   : 13.99
        retencao_dias  : 30
        limites_json   : {unidade: "gravacao"}
        somente_inserir: true
      }
    } as $cv14
  
    var.update $inseridos {
      value = $inseridos + $cv14.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "confvision_licenca"
        plano          : "gravacao_movimento_30d"
        nome_exibicao  : "Gravacao por movimento 30 dias"
        valor_mensal   : 19.99
        retencao_dias  : 30
        limites_json   : {unidade: "gravacao"}
        somente_inserir: true
      }
    } as $cv15
  
    var.update $inseridos {
      value = $inseridos + $cv15.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "confvision_licenca"
        plano          : "gravacao_timelapse_7d"
        nome_exibicao  : "Gravacao timelapse inteligente 7 dias"
        valor_mensal   : 6.99
        retencao_dias  : 30
        limites_json   : {unidade: "gravacao"}
        somente_inserir: true
      }
    } as $cv16
  
    var.update $inseridos {
      value = $inseridos + $cv16.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "confvision_licenca"
        plano          : "gravacao_timelapse_15d"
        nome_exibicao  : "Gravacao timelapse inteligente 15 dias"
        valor_mensal   : 9.99
        retencao_dias  : 30
        limites_json   : {unidade: "gravacao"}
        somente_inserir: true
      }
    } as $cv17
  
    var.update $inseridos {
      value = $inseridos + $cv17.inserido
    }
  
    function.run fn_fp_catalogo_upsert_item {
      input = {
        id_central      : $id_central
        id_representante: $id_representante
        produto        : "confvision_licenca"
        plano          : "gravacao_timelapse_30d"
        nome_exibicao  : "Gravacao timelapse inteligente 30 dias"
        valor_mensal   : 13.99
        retencao_dias  : 30
        limites_json   : {unidade: "gravacao"}
        somente_inserir: true
      }
    } as $cv18
  
    var.update $inseridos {
      value = $inseridos + $cv18.inserido
    }
      }
    }
  
    conditional {
      if ($escopo == "assinaturas" || $escopo == "tudo") {
        // Desativa itens legados do catalogo (escopo da Central)
        db.query fp_produto_catalogo {
      where = $db.fp_produto_catalogo.id_central == $id_central && $db.fp_produto_catalogo.id_representante == "" && (($db.fp_produto_catalogo.produto == "webambiente" && $db.fp_produto_catalogo.plano == "padrao") || ($db.fp_produto_catalogo.produto == "confvision" && $db.fp_produto_catalogo.plano == "modulo"))
      return = {type: "list"}
    } as $legados
  
    foreach ($legados) {
      each as $leg {
        db.patch fp_produto_catalogo {
          field_name = "id"
          field_value = $leg.id
          data = {ativo: "N"}
        } as $leg_off
      }
    }
  
    db.get fp_config_financeiro {
      field_name = "chave"
      field_value = "worker_secret"
    } as $wk
  
    conditional {
      if ($escopo == "assinaturas" || $escopo == "tudo") {
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "franqueadopro"
            plano           : "lite"
            nome_exibicao   : "FranqueadoPro Lite"
            valor_mensal    : 99
            retencao_dias   : 30
            limites_json    : {
            usuarios_alarme_max: 5
            setores_alarme_max : 4
            clientes_max       : 50
            contas_max         : 100
          }
            modulos_json    : {
            configuracao.grade                   : false
            configuracao.procedimento            : false
            configuracao.contactid               : false
            "configuracao.email-evento"          : false
            "atendimento.inteligencia-artificial": false
            "minha-empresa.comercial.pacotes"    : false
            "minha-empresa.operacional"          : false
            "minha-empresa.operacional.tecnico"  : false
            "minha-empresa.operacional.viatura"  : false
            "minha-empresa.gestao.permissoes"    : false
            whitelabel                           : false
            relatorio.atendimento                : false
            relatorio.ligacoes                   : false
            "relatorio.eventos.config-grupo"     : false
            dashboard.eventos                    : false
            franqueadopro.saas                   : false
          }
          }
        } as $s1
      
        var.update $inseridos {
          value = $inseridos + $s1.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "franqueadopro"
            plano           : "pro"
            nome_exibicao   : "FranqueadoPro Pro"
            valor_mensal    : 199
            retencao_dias   : 90
            limites_json    : {
            usuarios_alarme_max: 20
            setores_alarme_max : 15
            clientes_max       : 500
            contas_max         : 1000
          }
            modulos_json    : {
            configuracao.procedimento            : false
            configuracao.contactid               : false
            "atendimento.inteligencia-artificial": false
            "minha-empresa.comercial.pacotes"    : false
            "minha-empresa.gestao.permissoes"    : false
            franqueadopro.saas                   : false
            whitelabel                           : true
          }
          }
        } as $s2
      
        var.update $inseridos {
          value = $inseridos + $s2.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "franqueadopro"
            plano           : "pro_plus"
            nome_exibicao   : "FranqueadoPro Pro+"
            valor_mensal    : 299
            retencao_dias   : 180
            limites_json    : {
            usuarios_alarme_max: 999
            setores_alarme_max : 999
            clientes_max       : 9999
            contas_max         : 9999
          }
            modulos_json    : {
            configuracao.grade                   : true
            configuracao.procedimento            : true
            configuracao.contactid               : true
            "configuracao.email-evento"          : true
            "atendimento.inteligencia-artificial": true
            "minha-empresa.gestao.permissoes"    : true
            whitelabel                           : true
            franqueadopro.saas                   : true
          }
          }
        } as $s3
      
        var.update $inseridos {
          value = $inseridos + $s3.inserido
        }
      
        // --- Ecossistema: assinaturas (somente_inserir = preserva edicao no adm) ---
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "webterminal"
            plano           : "lite"
            nome_exibicao   : "WebTerminal Lite"
            valor_mensal    : 49
            retencao_dias   : 30
            limites_json    : {operadores_max: 3, filas_max: 1}
            somente_inserir : true
          }
        } as $s4
      
        var.update $inseridos {
          value = $inseridos + $s4.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "webterminal"
            plano           : "pro"
            nome_exibicao   : "WebTerminal Pro"
            valor_mensal    : 79
            retencao_dias   : 90
            limites_json    : {operadores_max: 15, filas_max: 5}
            somente_inserir : true
          }
        } as $s5
      
        var.update $inseridos {
          value = $inseridos + $s5.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "webterminal"
            plano           : "pro_plus"
            nome_exibicao   : "WebTerminal Pro+"
            valor_mensal    : 129
            retencao_dias   : 180
            limites_json    : {operadores_max: 999, filas_max: 999}
            somente_inserir : true
          }
        } as $s6
      
        var.update $inseridos {
          value = $inseridos + $s6.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "terminalmovel"
            plano           : "padrao"
            nome_exibicao   : "Terminal Movel"
            valor_mensal    : 59
            retencao_dias   : 30
            limites_json    : {operadores_max: 20, dispositivos_max: 50}
            somente_inserir : true
          }
        } as $s7
      
        var.update $inseridos {
          value = $inseridos + $s7.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "confvision"
            plano           : "padrao"
            nome_exibicao   : "ConfVision - Plano"
            valor_mensal    : 79
            retencao_dias   : 30
            somente_inserir : true
          }
        } as $s8
      
        var.update $inseridos {
          value = $inseridos + $s8.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "webambiente"
            plano           : "pro"
            nome_exibicao   : "webAmbiente Pro"
            valor_mensal    : 69
            retencao_dias   : 30
            limites_json    : {clientes_portal_max: 500}
            somente_inserir : true
          }
        } as $s9
      
        var.update $inseridos {
          value = $inseridos + $s9.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "webambiente"
            plano           : "pro_plus"
            nome_exibicao   : "webAmbiente Pro+"
            valor_mensal    : 99
            retencao_dias   : 90
            limites_json    : {clientes_portal_max: 9999}
            modulos_json    : {confvision_embed: true}
            somente_inserir : true
          }
        } as $s10
      
        var.update $inseridos {
          value = $inseridos + $s10.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "dialyze"
            plano           : "padrao"
            nome_exibicao   : "Dialyze"
            valor_mensal    : 149
            retencao_dias   : 30
            limites_json    : {
            usuarios_max       : 10
            canais_whatsapp_max: 3
            contatos_max       : 5000
          }
            somente_inserir : true
          }
        } as $s11
      
        var.update $inseridos {
          value = $inseridos + $s11.inserido
        }
      }
    }
  
    conditional {
      if ($escopo == "confvision_licenca" || $escopo == "tudo") {
        // --- ConfVision licencas camera/gravacao (precos editaveis no adm) ---
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "confvision_licenca"
            plano           : "online"
            nome_exibicao   : "Camera online"
            valor_mensal    : 2.99
            retencao_dias   : 30
            limites_json    : {unidade: "camera"}
            somente_inserir : true
          }
        } as $cv1
      
        var.update $inseridos {
          value = $inseridos + $cv1.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "confvision_licenca"
            plano           : "sensor_foto"
            nome_exibicao   : "Sensor foto"
            valor_mensal    : 7.99
            retencao_dias   : 30
            limites_json    : {unidade: "camera"}
            somente_inserir : true
          }
        } as $cv2
      
        var.update $inseridos {
          value = $inseridos + $cv2.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "confvision_licenca"
            plano           : "sensor_foto_video"
            nome_exibicao   : "Sensor foto + video"
            valor_mensal    : 9.99
            retencao_dias   : 30
            limites_json    : {unidade: "camera"}
            somente_inserir : true
          }
        } as $cv3
      
        var.update $inseridos {
          value = $inseridos + $cv3.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "confvision_licenca"
            plano           : "analitico_armado_evento"
            nome_exibicao   : "Analitico armado - so evento"
            valor_mensal    : 11.99
            retencao_dias   : 30
            limites_json    : {unidade: "camera"}
            somente_inserir : true
          }
        } as $cv4
      
        var.update $inseridos {
          value = $inseridos + $cv4.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "confvision_licenca"
            plano           : "analitico_armado_foto"
            nome_exibicao   : "Analitico armado - foto"
            valor_mensal    : 13.99
            retencao_dias   : 30
            limites_json    : {unidade: "camera"}
            somente_inserir : true
          }
        } as $cv5
      
        var.update $inseridos {
          value = $inseridos + $cv5.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "confvision_licenca"
            plano           : "analitico_armado_foto_video"
            nome_exibicao   : "Analitico armado - foto + video"
            valor_mensal    : 14.99
            retencao_dias   : 30
            limites_json    : {unidade: "camera"}
            somente_inserir : true
          }
        } as $cv6
      
        var.update $inseridos {
          value = $inseridos + $cv6.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "confvision_licenca"
            plano           : "analitico_24h_evento"
            nome_exibicao   : "Analitico 24h - so evento"
            valor_mensal    : 16.99
            retencao_dias   : 30
            limites_json    : {unidade: "camera"}
            somente_inserir : true
          }
        } as $cv7
      
        var.update $inseridos {
          value = $inseridos + $cv7.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "confvision_licenca"
            plano           : "analitico_24h_foto"
            nome_exibicao   : "Analitico 24h - foto"
            valor_mensal    : 18.99
            retencao_dias   : 30
            limites_json    : {unidade: "camera"}
            somente_inserir : true
          }
        } as $cv8
      
        var.update $inseridos {
          value = $inseridos + $cv8.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "confvision_licenca"
            plano           : "analitico_24h_foto_video"
            nome_exibicao   : "Analitico 24h - foto + video"
            valor_mensal    : 19.99
            retencao_dias   : 30
            limites_json    : {unidade: "camera"}
            somente_inserir : true
          }
        } as $cv9
      
        var.update $inseridos {
          value = $inseridos + $cv9.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "confvision_licenca"
            plano           : "gravacao_7d"
            nome_exibicao   : "Gravacao continua 7 dias"
            valor_mensal    : 12.99
            retencao_dias   : 30
            limites_json    : {unidade: "gravacao"}
            somente_inserir : true
          }
        } as $cv10
      
        var.update $inseridos {
          value = $inseridos + $cv10.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "confvision_licenca"
            plano           : "gravacao_15d"
            nome_exibicao   : "Gravacao continua 15 dias"
            valor_mensal    : 17.99
            retencao_dias   : 30
            limites_json    : {unidade: "gravacao"}
            somente_inserir : true
          }
        } as $cv11
      
        var.update $inseridos {
          value = $inseridos + $cv11.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "confvision_licenca"
            plano           : "gravacao_30d"
            nome_exibicao   : "Gravacao continua 30 dias"
            valor_mensal    : 24.99
            retencao_dias   : 30
            limites_json    : {unidade: "gravacao"}
            somente_inserir : true
          }
        } as $cv12
      
        var.update $inseridos {
          value = $inseridos + $cv12.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "confvision_licenca"
            plano           : "gravacao_movimento_7d"
            nome_exibicao   : "Gravacao por movimento 7 dias"
            valor_mensal    : 9.99
            retencao_dias   : 30
            limites_json    : {unidade: "gravacao"}
            somente_inserir : true
          }
        } as $cv13
      
        var.update $inseridos {
          value = $inseridos + $cv13.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "confvision_licenca"
            plano           : "gravacao_movimento_15d"
            nome_exibicao   : "Gravacao por movimento 15 dias"
            valor_mensal    : 13.99
            retencao_dias   : 30
            limites_json    : {unidade: "gravacao"}
            somente_inserir : true
          }
        } as $cv14
      
        var.update $inseridos {
          value = $inseridos + $cv14.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "confvision_licenca"
            plano           : "gravacao_movimento_30d"
            nome_exibicao   : "Gravacao por movimento 30 dias"
            valor_mensal    : 19.99
            retencao_dias   : 30
            limites_json    : {unidade: "gravacao"}
            somente_inserir : true
          }
        } as $cv15
      
        var.update $inseridos {
          value = $inseridos + $cv15.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "confvision_licenca"
            plano           : "gravacao_timelapse_7d"
            nome_exibicao   : "Gravacao timelapse inteligente 7 dias"
            valor_mensal    : 6.99
            retencao_dias   : 30
            limites_json    : {unidade: "gravacao"}
            somente_inserir : true
          }
        } as $cv16
      
        var.update $inseridos {
          value = $inseridos + $cv16.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "confvision_licenca"
            plano           : "gravacao_timelapse_15d"
            nome_exibicao   : "Gravacao timelapse inteligente 15 dias"
            valor_mensal    : 9.99
            retencao_dias   : 30
            limites_json    : {unidade: "gravacao"}
            somente_inserir : true
          }
        } as $cv17
      
        var.update $inseridos {
          value = $inseridos + $cv17.inserido
        }
      
        function.run fn_fp_catalogo_upsert_item {
          input = {
            id_central      : $id_central
            id_representante: $id_representante
            produto         : "confvision_licenca"
            plano           : "gravacao_timelapse_30d"
            nome_exibicao   : "Gravacao timelapse inteligente 30 dias"
            valor_mensal    : 13.99
            retencao_dias   : 30
            limites_json    : {unidade: "gravacao"}
            somente_inserir : true
          }
        } as $cv18
      
        var.update $inseridos {
          value = $inseridos + $cv18.inserido
        }
      }
    }
  
    conditional {
      if ($escopo == "assinaturas" || $escopo == "tudo") {
        // Desativa itens legados do catalogo (escopo da Central)
        db.query fp_produto_catalogo {
          where = $db.fp_produto_catalogo.id_central == $id_central && $db.fp_produto_catalogo.id_representante == "" && (($db.fp_produto_catalogo.produto == "webambiente" && $db.fp_produto_catalogo.plano == "padrao") || ($db.fp_produto_catalogo.produto == "confvision" && $db.fp_produto_catalogo.plano == "modulo"))
          return = {type: "list"}
        } as $legados
      
        foreach ($legados) {
          each as $leg {
            db.patch fp_produto_catalogo {
              field_name = "id"
              field_value = $leg.id
              data = {ativo: "N"}
            } as $leg_off
          }
        }
      
        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "worker_secret"
        } as $wk
      
        conditional {
          if ($escopo == "assinaturas" || $escopo == "tudo") {
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "franqueadopro"
                plano           : "lite"
                nome_exibicao   : "FranqueadoPro Lite"
                valor_mensal    : 99
                retencao_dias   : 30
                limites_json    : {
                usuarios_alarme_max: 5
                setores_alarme_max : 4
                clientes_max       : 50
                contas_max         : 100
              }
                modulos_json    : {
                configuracao.grade                   : false
                configuracao.procedimento            : false
                configuracao.contactid               : false
                "configuracao.email-evento"          : false
                "atendimento.inteligencia-artificial": false
                "minha-empresa.comercial.pacotes"    : false
                "minha-empresa.operacional"          : false
                "minha-empresa.operacional.tecnico"  : false
                "minha-empresa.operacional.viatura"  : false
                "minha-empresa.gestao.permissoes"    : false
                whitelabel                           : false
                relatorio.atendimento                : false
                relatorio.ligacoes                   : false
                "relatorio.eventos.config-grupo"     : false
                dashboard.eventos                    : false
                franqueadopro.saas                   : false
              }
              }
            } as $s1
          
            var.update $inseridos {
              value = $inseridos + $s1.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "franqueadopro"
                plano           : "pro"
                nome_exibicao   : "FranqueadoPro Pro"
                valor_mensal    : 199
                retencao_dias   : 90
                limites_json    : {
                usuarios_alarme_max: 20
                setores_alarme_max : 15
                clientes_max       : 500
                contas_max         : 1000
              }
                modulos_json    : {
                configuracao.procedimento            : false
                configuracao.contactid               : false
                "atendimento.inteligencia-artificial": false
                "minha-empresa.comercial.pacotes"    : false
                "minha-empresa.gestao.permissoes"    : false
                franqueadopro.saas                   : false
                whitelabel                           : true
              }
              }
            } as $s2
          
            var.update $inseridos {
              value = $inseridos + $s2.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "franqueadopro"
                plano           : "pro_plus"
                nome_exibicao   : "FranqueadoPro Pro+"
                valor_mensal    : 299
                retencao_dias   : 180
                limites_json    : {
                usuarios_alarme_max: 999
                setores_alarme_max : 999
                clientes_max       : 9999
                contas_max         : 9999
              }
                modulos_json    : {
                configuracao.grade                   : true
                configuracao.procedimento            : true
                configuracao.contactid               : true
                "configuracao.email-evento"          : true
                "atendimento.inteligencia-artificial": true
                "minha-empresa.gestao.permissoes"    : true
                whitelabel                           : true
                franqueadopro.saas                   : true
              }
              }
            } as $s3
          
            var.update $inseridos {
              value = $inseridos + $s3.inserido
            }
          
            // --- Ecossistema: assinaturas (somente_inserir = preserva edicao no adm) ---
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "webterminal"
                plano           : "lite"
                nome_exibicao   : "WebTerminal Lite"
                valor_mensal    : 49
                retencao_dias   : 30
                limites_json    : {operadores_max: 3, filas_max: 1}
                somente_inserir : true
              }
            } as $s4
          
            var.update $inseridos {
              value = $inseridos + $s4.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "webterminal"
                plano           : "pro"
                nome_exibicao   : "WebTerminal Pro"
                valor_mensal    : 79
                retencao_dias   : 90
                limites_json    : {operadores_max: 15, filas_max: 5}
                somente_inserir : true
              }
            } as $s5
          
            var.update $inseridos {
              value = $inseridos + $s5.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "webterminal"
                plano           : "pro_plus"
                nome_exibicao   : "WebTerminal Pro+"
                valor_mensal    : 129
                retencao_dias   : 180
                limites_json    : {operadores_max: 999, filas_max: 999}
                somente_inserir : true
              }
            } as $s6
          
            var.update $inseridos {
              value = $inseridos + $s6.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "terminalmovel"
                plano           : "padrao"
                nome_exibicao   : "Terminal Movel"
                valor_mensal    : 59
                retencao_dias   : 30
                limites_json    : {operadores_max: 20, dispositivos_max: 50}
                somente_inserir : true
              }
            } as $s7
          
            var.update $inseridos {
              value = $inseridos + $s7.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "confvision"
                plano           : "padrao"
                nome_exibicao   : "ConfVision - Plano"
                valor_mensal    : 79
                retencao_dias   : 30
                somente_inserir : true
              }
            } as $s8
          
            var.update $inseridos {
              value = $inseridos + $s8.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "webambiente"
                plano           : "pro"
                nome_exibicao   : "webAmbiente Pro"
                valor_mensal    : 69
                retencao_dias   : 30
                limites_json    : {clientes_portal_max: 500}
                somente_inserir : true
              }
            } as $s9
          
            var.update $inseridos {
              value = $inseridos + $s9.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "webambiente"
                plano           : "pro_plus"
                nome_exibicao   : "webAmbiente Pro+"
                valor_mensal    : 99
                retencao_dias   : 90
                limites_json    : {clientes_portal_max: 9999}
                modulos_json    : {confvision_embed: true}
                somente_inserir : true
              }
            } as $s10
          
            var.update $inseridos {
              value = $inseridos + $s10.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "dialyze"
                plano           : "padrao"
                nome_exibicao   : "Dialyze"
                valor_mensal    : 149
                retencao_dias   : 30
                limites_json    : {
                usuarios_max       : 10
                canais_whatsapp_max: 3
                contatos_max       : 5000
              }
                somente_inserir : true
              }
            } as $s11
          
            var.update $inseridos {
              value = $inseridos + $s11.inserido
            }
          }
        }
      
        conditional {
          if ($escopo == "confvision_licenca" || $escopo == "tudo") {
            // --- ConfVision licencas camera/gravacao (precos editaveis no adm) ---
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "confvision_licenca"
                plano           : "online"
                nome_exibicao   : "Camera online"
                valor_mensal    : 2.99
                retencao_dias   : 30
                limites_json    : {unidade: "camera"}
                somente_inserir : true
              }
            } as $cv1
          
            var.update $inseridos {
              value = $inseridos + $cv1.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "confvision_licenca"
                plano           : "sensor_foto"
                nome_exibicao   : "Sensor foto"
                valor_mensal    : 7.99
                retencao_dias   : 30
                limites_json    : {unidade: "camera"}
                somente_inserir : true
              }
            } as $cv2
          
            var.update $inseridos {
              value = $inseridos + $cv2.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "confvision_licenca"
                plano           : "sensor_foto_video"
                nome_exibicao   : "Sensor foto + video"
                valor_mensal    : 9.99
                retencao_dias   : 30
                limites_json    : {unidade: "camera"}
                somente_inserir : true
              }
            } as $cv3
          
            var.update $inseridos {
              value = $inseridos + $cv3.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "confvision_licenca"
                plano           : "analitico_armado_evento"
                nome_exibicao   : "Analitico armado - so evento"
                valor_mensal    : 11.99
                retencao_dias   : 30
                limites_json    : {unidade: "camera"}
                somente_inserir : true
              }
            } as $cv4
          
            var.update $inseridos {
              value = $inseridos + $cv4.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "confvision_licenca"
                plano           : "analitico_armado_foto"
                nome_exibicao   : "Analitico armado - foto"
                valor_mensal    : 13.99
                retencao_dias   : 30
                limites_json    : {unidade: "camera"}
                somente_inserir : true
              }
            } as $cv5
          
            var.update $inseridos {
              value = $inseridos + $cv5.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "confvision_licenca"
                plano           : "analitico_armado_foto_video"
                nome_exibicao   : "Analitico armado - foto + video"
                valor_mensal    : 14.99
                retencao_dias   : 30
                limites_json    : {unidade: "camera"}
                somente_inserir : true
              }
            } as $cv6
          
            var.update $inseridos {
              value = $inseridos + $cv6.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "confvision_licenca"
                plano           : "analitico_24h_evento"
                nome_exibicao   : "Analitico 24h - so evento"
                valor_mensal    : 16.99
                retencao_dias   : 30
                limites_json    : {unidade: "camera"}
                somente_inserir : true
              }
            } as $cv7
          
            var.update $inseridos {
              value = $inseridos + $cv7.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "confvision_licenca"
                plano           : "analitico_24h_foto"
                nome_exibicao   : "Analitico 24h - foto"
                valor_mensal    : 18.99
                retencao_dias   : 30
                limites_json    : {unidade: "camera"}
                somente_inserir : true
              }
            } as $cv8
          
            var.update $inseridos {
              value = $inseridos + $cv8.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "confvision_licenca"
                plano           : "analitico_24h_foto_video"
                nome_exibicao   : "Analitico 24h - foto + video"
                valor_mensal    : 19.99
                retencao_dias   : 30
                limites_json    : {unidade: "camera"}
                somente_inserir : true
              }
            } as $cv9
          
            var.update $inseridos {
              value = $inseridos + $cv9.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "confvision_licenca"
                plano           : "gravacao_7d"
                nome_exibicao   : "Gravacao continua 7 dias"
                valor_mensal    : 12.99
                retencao_dias   : 30
                limites_json    : {unidade: "gravacao"}
                somente_inserir : true
              }
            } as $cv10
          
            var.update $inseridos {
              value = $inseridos + $cv10.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "confvision_licenca"
                plano           : "gravacao_15d"
                nome_exibicao   : "Gravacao continua 15 dias"
                valor_mensal    : 17.99
                retencao_dias   : 30
                limites_json    : {unidade: "gravacao"}
                somente_inserir : true
              }
            } as $cv11
          
            var.update $inseridos {
              value = $inseridos + $cv11.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "confvision_licenca"
                plano           : "gravacao_30d"
                nome_exibicao   : "Gravacao continua 30 dias"
                valor_mensal    : 24.99
                retencao_dias   : 30
                limites_json    : {unidade: "gravacao"}
                somente_inserir : true
              }
            } as $cv12
          
            var.update $inseridos {
              value = $inseridos + $cv12.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "confvision_licenca"
                plano           : "gravacao_movimento_7d"
                nome_exibicao   : "Gravacao por movimento 7 dias"
                valor_mensal    : 9.99
                retencao_dias   : 30
                limites_json    : {unidade: "gravacao"}
                somente_inserir : true
              }
            } as $cv13
          
            var.update $inseridos {
              value = $inseridos + $cv13.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "confvision_licenca"
                plano           : "gravacao_movimento_15d"
                nome_exibicao   : "Gravacao por movimento 15 dias"
                valor_mensal    : 13.99
                retencao_dias   : 30
                limites_json    : {unidade: "gravacao"}
                somente_inserir : true
              }
            } as $cv14
          
            var.update $inseridos {
              value = $inseridos + $cv14.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "confvision_licenca"
                plano           : "gravacao_movimento_30d"
                nome_exibicao   : "Gravacao por movimento 30 dias"
                valor_mensal    : 19.99
                retencao_dias   : 30
                limites_json    : {unidade: "gravacao"}
                somente_inserir : true
              }
            } as $cv15
          
            var.update $inseridos {
              value = $inseridos + $cv15.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "confvision_licenca"
                plano           : "gravacao_timelapse_7d"
                nome_exibicao   : "Gravacao timelapse inteligente 7 dias"
                valor_mensal    : 6.99
                retencao_dias   : 30
                limites_json    : {unidade: "gravacao"}
                somente_inserir : true
              }
            } as $cv16
          
            var.update $inseridos {
              value = $inseridos + $cv16.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "confvision_licenca"
                plano           : "gravacao_timelapse_15d"
                nome_exibicao   : "Gravacao timelapse inteligente 15 dias"
                valor_mensal    : 9.99
                retencao_dias   : 30
                limites_json    : {unidade: "gravacao"}
                somente_inserir : true
              }
            } as $cv17
          
            var.update $inseridos {
              value = $inseridos + $cv17.inserido
            }
          
            function.run fn_fp_catalogo_upsert_item {
              input = {
                id_central      : $id_central
                id_representante: $id_representante
                produto         : "confvision_licenca"
                plano           : "gravacao_timelapse_30d"
                nome_exibicao   : "Gravacao timelapse inteligente 30 dias"
                valor_mensal    : 13.99
                retencao_dias   : 30
                limites_json    : {unidade: "gravacao"}
                somente_inserir : true
              }
            } as $cv18
          
            var.update $inseridos {
              value = $inseridos + $cv18.inserido
            }
          }
        }
      
        conditional {
          if ($escopo == "assinaturas" || $escopo == "tudo") {
            // Desativa itens legados do catalogo (escopo da Central)
            db.query fp_produto_catalogo {
              where = $db.fp_produto_catalogo.id_central == $id_central && $db.fp_produto_catalogo.id_representante == "" && (($db.fp_produto_catalogo.produto == "webambiente" && $db.fp_produto_catalogo.plano == "padrao") || ($db.fp_produto_catalogo.produto == "confvision" && $db.fp_produto_catalogo.plano == "modulo"))
              return = {type: "list"}
            } as $legados
          
            foreach ($legados) {
              each as $leg {
                db.patch fp_produto_catalogo {
                  field_name = "id"
                  field_value = $leg.id
                  data = {ativo: "N"}
                } as $leg_off
              }
            }
          
            db.get fp_config_financeiro {
              field_name = "chave"
              field_value = "worker_secret"
            } as $wk
          
            conditional {
              if ($escopo == "assinaturas" || $escopo == "tudo") {
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "franqueadopro"
                    plano           : "lite"
                    nome_exibicao   : "FranqueadoPro Lite"
                    valor_mensal    : 99
                    retencao_dias   : 30
                    limites_json    : {
                    usuarios_alarme_max: 5
                    setores_alarme_max : 4
                    clientes_max       : 50
                    contas_max         : 100
                  }
                    modulos_json    : {
                    configuracao.grade                   : false
                    configuracao.procedimento            : false
                    configuracao.contactid               : false
                    "configuracao.email-evento"          : false
                    "atendimento.inteligencia-artificial": false
                    "minha-empresa.comercial.pacotes"    : false
                    "minha-empresa.operacional"          : false
                    "minha-empresa.operacional.tecnico"  : false
                    "minha-empresa.operacional.viatura"  : false
                    "minha-empresa.gestao.permissoes"    : false
                    whitelabel                           : false
                    relatorio.atendimento                : false
                    relatorio.ligacoes                   : false
                    "relatorio.eventos.config-grupo"     : false
                    dashboard.eventos                    : false
                    franqueadopro.saas                   : false
                  }
                  }
                } as $s1
              
                var.update $inseridos {
                  value = $inseridos + $s1.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "franqueadopro"
                    plano           : "pro"
                    nome_exibicao   : "FranqueadoPro Pro"
                    valor_mensal    : 199
                    retencao_dias   : 90
                    limites_json    : {
                    usuarios_alarme_max: 20
                    setores_alarme_max : 15
                    clientes_max       : 500
                    contas_max         : 1000
                  }
                    modulos_json    : {
                    configuracao.procedimento            : false
                    configuracao.contactid               : false
                    "atendimento.inteligencia-artificial": false
                    "minha-empresa.comercial.pacotes"    : false
                    "minha-empresa.gestao.permissoes"    : false
                    franqueadopro.saas                   : false
                    whitelabel                           : true
                  }
                  }
                } as $s2
              
                var.update $inseridos {
                  value = $inseridos + $s2.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "franqueadopro"
                    plano           : "pro_plus"
                    nome_exibicao   : "FranqueadoPro Pro+"
                    valor_mensal    : 299
                    retencao_dias   : 180
                    limites_json    : {
                    usuarios_alarme_max: 999
                    setores_alarme_max : 999
                    clientes_max       : 9999
                    contas_max         : 9999
                  }
                    modulos_json    : {
                    configuracao.grade                   : true
                    configuracao.procedimento            : true
                    configuracao.contactid               : true
                    "configuracao.email-evento"          : true
                    "atendimento.inteligencia-artificial": true
                    "minha-empresa.gestao.permissoes"    : true
                    whitelabel                           : true
                    franqueadopro.saas                   : true
                  }
                  }
                } as $s3
              
                var.update $inseridos {
                  value = $inseridos + $s3.inserido
                }
              
                // --- Ecossistema: assinaturas (somente_inserir = preserva edicao no adm) ---
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "webterminal"
                    plano           : "lite"
                    nome_exibicao   : "WebTerminal Lite"
                    valor_mensal    : 49
                    retencao_dias   : 30
                    limites_json    : {operadores_max: 3, filas_max: 1}
                    somente_inserir : true
                  }
                } as $s4
              
                var.update $inseridos {
                  value = $inseridos + $s4.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "webterminal"
                    plano           : "pro"
                    nome_exibicao   : "WebTerminal Pro"
                    valor_mensal    : 79
                    retencao_dias   : 90
                    limites_json    : {operadores_max: 15, filas_max: 5}
                    somente_inserir : true
                  }
                } as $s5
              
                var.update $inseridos {
                  value = $inseridos + $s5.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "webterminal"
                    plano           : "pro_plus"
                    nome_exibicao   : "WebTerminal Pro+"
                    valor_mensal    : 129
                    retencao_dias   : 180
                    limites_json    : {operadores_max: 999, filas_max: 999}
                    somente_inserir : true
                  }
                } as $s6
              
                var.update $inseridos {
                  value = $inseridos + $s6.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "terminalmovel"
                    plano           : "padrao"
                    nome_exibicao   : "Terminal Movel"
                    valor_mensal    : 59
                    retencao_dias   : 30
                    limites_json    : {operadores_max: 20, dispositivos_max: 50}
                    somente_inserir : true
                  }
                } as $s7
              
                var.update $inseridos {
                  value = $inseridos + $s7.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "confvision"
                    plano           : "padrao"
                    nome_exibicao   : "ConfVision - Plano"
                    valor_mensal    : 79
                    retencao_dias   : 30
                    somente_inserir : true
                  }
                } as $s8
              
                var.update $inseridos {
                  value = $inseridos + $s8.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "webambiente"
                    plano           : "pro"
                    nome_exibicao   : "webAmbiente Pro"
                    valor_mensal    : 69
                    retencao_dias   : 30
                    limites_json    : {clientes_portal_max: 500}
                    somente_inserir : true
                  }
                } as $s9
              
                var.update $inseridos {
                  value = $inseridos + $s9.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "webambiente"
                    plano           : "pro_plus"
                    nome_exibicao   : "webAmbiente Pro+"
                    valor_mensal    : 99
                    retencao_dias   : 90
                    limites_json    : {clientes_portal_max: 9999}
                    modulos_json    : {confvision_embed: true}
                    somente_inserir : true
                  }
                } as $s10
              
                var.update $inseridos {
                  value = $inseridos + $s10.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "dialyze"
                    plano           : "padrao"
                    nome_exibicao   : "Dialyze"
                    valor_mensal    : 149
                    retencao_dias   : 30
                    limites_json    : {
                    usuarios_max       : 10
                    canais_whatsapp_max: 3
                    contatos_max       : 5000
                  }
                    somente_inserir : true
                  }
                } as $s11
              
                var.update $inseridos {
                  value = $inseridos + $s11.inserido
                }
              }
            }
          
            conditional {
              if ($escopo == "confvision_licenca" || $escopo == "tudo") {
                // --- ConfVision licencas camera/gravacao (precos editaveis no adm) ---
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "confvision_licenca"
                    plano           : "online"
                    nome_exibicao   : "Camera online"
                    valor_mensal    : 2.99
                    retencao_dias   : 30
                    limites_json    : {unidade: "camera"}
                    somente_inserir : true
                  }
                } as $cv1
              
                var.update $inseridos {
                  value = $inseridos + $cv1.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "confvision_licenca"
                    plano           : "sensor_foto"
                    nome_exibicao   : "Sensor foto"
                    valor_mensal    : 7.99
                    retencao_dias   : 30
                    limites_json    : {unidade: "camera"}
                    somente_inserir : true
                  }
                } as $cv2
              
                var.update $inseridos {
                  value = $inseridos + $cv2.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "confvision_licenca"
                    plano           : "sensor_foto_video"
                    nome_exibicao   : "Sensor foto + video"
                    valor_mensal    : 9.99
                    retencao_dias   : 30
                    limites_json    : {unidade: "camera"}
                    somente_inserir : true
                  }
                } as $cv3
              
                var.update $inseridos {
                  value = $inseridos + $cv3.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "confvision_licenca"
                    plano           : "analitico_armado_evento"
                    nome_exibicao   : "Analitico armado - so evento"
                    valor_mensal    : 11.99
                    retencao_dias   : 30
                    limites_json    : {unidade: "camera"}
                    somente_inserir : true
                  }
                } as $cv4
              
                var.update $inseridos {
                  value = $inseridos + $cv4.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "confvision_licenca"
                    plano           : "analitico_armado_foto"
                    nome_exibicao   : "Analitico armado - foto"
                    valor_mensal    : 13.99
                    retencao_dias   : 30
                    limites_json    : {unidade: "camera"}
                    somente_inserir : true
                  }
                } as $cv5
              
                var.update $inseridos {
                  value = $inseridos + $cv5.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "confvision_licenca"
                    plano           : "analitico_armado_foto_video"
                    nome_exibicao   : "Analitico armado - foto + video"
                    valor_mensal    : 14.99
                    retencao_dias   : 30
                    limites_json    : {unidade: "camera"}
                    somente_inserir : true
                  }
                } as $cv6
              
                var.update $inseridos {
                  value = $inseridos + $cv6.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "confvision_licenca"
                    plano           : "analitico_24h_evento"
                    nome_exibicao   : "Analitico 24h - so evento"
                    valor_mensal    : 16.99
                    retencao_dias   : 30
                    limites_json    : {unidade: "camera"}
                    somente_inserir : true
                  }
                } as $cv7
              
                var.update $inseridos {
                  value = $inseridos + $cv7.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "confvision_licenca"
                    plano           : "analitico_24h_foto"
                    nome_exibicao   : "Analitico 24h - foto"
                    valor_mensal    : 18.99
                    retencao_dias   : 30
                    limites_json    : {unidade: "camera"}
                    somente_inserir : true
                  }
                } as $cv8
              
                var.update $inseridos {
                  value = $inseridos + $cv8.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "confvision_licenca"
                    plano           : "analitico_24h_foto_video"
                    nome_exibicao   : "Analitico 24h - foto + video"
                    valor_mensal    : 19.99
                    retencao_dias   : 30
                    limites_json    : {unidade: "camera"}
                    somente_inserir : true
                  }
                } as $cv9
              
                var.update $inseridos {
                  value = $inseridos + $cv9.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "confvision_licenca"
                    plano           : "gravacao_7d"
                    nome_exibicao   : "Gravacao continua 7 dias"
                    valor_mensal    : 12.99
                    retencao_dias   : 30
                    limites_json    : {unidade: "gravacao"}
                    somente_inserir : true
                  }
                } as $cv10
              
                var.update $inseridos {
                  value = $inseridos + $cv10.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "confvision_licenca"
                    plano           : "gravacao_15d"
                    nome_exibicao   : "Gravacao continua 15 dias"
                    valor_mensal    : 17.99
                    retencao_dias   : 30
                    limites_json    : {unidade: "gravacao"}
                    somente_inserir : true
                  }
                } as $cv11
              
                var.update $inseridos {
                  value = $inseridos + $cv11.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "confvision_licenca"
                    plano           : "gravacao_30d"
                    nome_exibicao   : "Gravacao continua 30 dias"
                    valor_mensal    : 24.99
                    retencao_dias   : 30
                    limites_json    : {unidade: "gravacao"}
                    somente_inserir : true
                  }
                } as $cv12
              
                var.update $inseridos {
                  value = $inseridos + $cv12.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "confvision_licenca"
                    plano           : "gravacao_movimento_7d"
                    nome_exibicao   : "Gravacao por movimento 7 dias"
                    valor_mensal    : 9.99
                    retencao_dias   : 30
                    limites_json    : {unidade: "gravacao"}
                    somente_inserir : true
                  }
                } as $cv13
              
                var.update $inseridos {
                  value = $inseridos + $cv13.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "confvision_licenca"
                    plano           : "gravacao_movimento_15d"
                    nome_exibicao   : "Gravacao por movimento 15 dias"
                    valor_mensal    : 13.99
                    retencao_dias   : 30
                    limites_json    : {unidade: "gravacao"}
                    somente_inserir : true
                  }
                } as $cv14
              
                var.update $inseridos {
                  value = $inseridos + $cv14.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "confvision_licenca"
                    plano           : "gravacao_movimento_30d"
                    nome_exibicao   : "Gravacao por movimento 30 dias"
                    valor_mensal    : 19.99
                    retencao_dias   : 30
                    limites_json    : {unidade: "gravacao"}
                    somente_inserir : true
                  }
                } as $cv15
              
                var.update $inseridos {
                  value = $inseridos + $cv15.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "confvision_licenca"
                    plano           : "gravacao_timelapse_7d"
                    nome_exibicao   : "Gravacao timelapse inteligente 7 dias"
                    valor_mensal    : 6.99
                    retencao_dias   : 30
                    limites_json    : {unidade: "gravacao"}
                    somente_inserir : true
                  }
                } as $cv16
              
                var.update $inseridos {
                  value = $inseridos + $cv16.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "confvision_licenca"
                    plano           : "gravacao_timelapse_15d"
                    nome_exibicao   : "Gravacao timelapse inteligente 15 dias"
                    valor_mensal    : 9.99
                    retencao_dias   : 30
                    limites_json    : {unidade: "gravacao"}
                    somente_inserir : true
                  }
                } as $cv17
              
                var.update $inseridos {
                  value = $inseridos + $cv17.inserido
                }
              
                function.run fn_fp_catalogo_upsert_item {
                  input = {
                    id_central      : $id_central
                    id_representante: $id_representante
                    produto         : "confvision_licenca"
                    plano           : "gravacao_timelapse_30d"
                    nome_exibicao   : "Gravacao timelapse inteligente 30 dias"
                    valor_mensal    : 13.99
                    retencao_dias   : 30
                    limites_json    : {unidade: "gravacao"}
                    somente_inserir : true
                  }
                } as $cv18
              
                var.update $inseridos {
                  value = $inseridos + $cv18.inserido
                }
              }
            }
          
            conditional {
              if ($escopo == "assinaturas" || $escopo == "tudo") {
                // Desativa itens legados do catalogo (escopo da Central)
                db.query fp_produto_catalogo {
                  where = $db.fp_produto_catalogo.id_central == $id_central && $db.fp_produto_catalogo.id_representante == "" && (($db.fp_produto_catalogo.produto == "webambiente" && $db.fp_produto_catalogo.plano == "padrao") || ($db.fp_produto_catalogo.produto == "confvision" && $db.fp_produto_catalogo.plano == "modulo"))
                  return = {type: "list"}
                } as $legados
              
                foreach ($legados) {
                  each as $leg {
                    db.patch fp_produto_catalogo {
                      field_name = "id"
                      field_value = $leg.id
                      data = {ativo: "N"}
                    } as $leg_off
                  }
                }
              
                db.get fp_config_financeiro {
                  field_name = "chave"
                  field_value = "worker_secret"
                } as $wk
              
                conditional {
                  if ($wk == null) {
                    db.add fp_config_financeiro {
                      data = {
                        created_at: "now"
                        chave     : "worker_secret"
                        valor     : "fp-worker-change-me"
                        descricao : "Chave do fp-billing-worker (header X-Worker-Key)"
                      }
                    } as $cfg_wk
                  }
                }
              
                db.get fp_config_financeiro {
                  field_name = "chave"
                  field_value = "dias_antecedencia_fatura"
                } as $dias_cfg
              
                conditional {
                  if ($dias_cfg == null) {
                    db.add fp_config_financeiro {
                      data = {
                        created_at: "now"
                        chave     : "dias_antecedencia_fatura"
                        valor     : "5"
                        descricao : "Dias antes do vencimento para gerar fatura"
                      }
                    } as $cfg_dias
                  }
                }
              
                db.get fp_config_financeiro {
                  field_name = "chave"
                  field_value = "admin_usuario"
                } as $adm_u
              
                conditional {
                  if ($adm_u == null) {
                    db.add fp_config_financeiro {
                      data = {
                        created_at: "now"
                        chave     : "admin_usuario"
                        valor     : "admin"
                        descricao : "Usuario admConfmonit"
                      }
                    } as $cfg_adm_u
                  }
                }
              
                db.get fp_config_financeiro {
                  field_name = "chave"
                  field_value = "admin_senha"
                } as $adm_s
              
                conditional {
                  if ($adm_s == null) {
                    db.add fp_config_financeiro {
                      data = {
                        created_at: "now"
                        chave     : "admin_senha"
                        valor     : "Confmonit@2026"
                        descricao : "Senha admConfmonit"
                      }
                    } as $cfg_adm_s
                  }
                }
              
                db.get fp_config_financeiro {
                  field_name = "chave"
                  field_value = "admin_token"
                } as $adm_t
              
                conditional {
                  if ($adm_t == null) {
                    db.add fp_config_financeiro {
                      data = {
                        created_at: "now"
                        chave     : "admin_token"
                        valor     : "fp-admin-token-change-me"
                        descricao : "Token sessao admConfmonit (retornado no login)"
                      }
                    } as $cfg_adm_t
                  }
                }
              
                db.get fp_config_financeiro {
                  field_name = "chave"
                  field_value = "api_legada_url"
                } as $api_legada
              
                conditional {
                  if ($api_legada == null) {
                    db.add fp_config_financeiro {
                      data = {
                        created_at: "now"
                        chave     : "api_legada_url"
                        valor     : "http://185.130.61.4:2010"
                        descricao : "URL base API Go legada (sync UsaConfVision)"
                      }
                    } as $cfg_api_legada
                  }
                }
              
                db.get fp_config_financeiro {
                  field_name = "chave"
                  field_value = "confvision_s3_endpoint"
                } as $cv_s3_endpoint
              
                conditional {
                  if ($cv_s3_endpoint == null) {
                    db.add fp_config_financeiro {
                      data = {
                        created_at: "now"
                        chave     : "confvision_s3_endpoint"
                        valor     : "https://usc1.contabostorage.com"
                        descricao : "Endpoint S3 Contabo padrao para gravacao ConfVision"
                      }
                    } as $cfg_cv_s3_endpoint
                  }
                }
              
                db.get fp_config_financeiro {
                  field_name = "chave"
                  field_value = "confvision_s3_bucket"
                } as $cv_s3_bucket
              
                conditional {
                  if ($cv_s3_bucket == null) {
                    db.add fp_config_financeiro {
                      data = {
                        created_at: "now"
                        chave     : "confvision_s3_bucket"
                        valor     : "confvision"
                        descricao : "Bucket S3 padrao para gravacao ConfVision"
                      }
                    } as $cfg_cv_s3_bucket
                  }
                }
              
                db.get fp_config_financeiro {
                  field_name = "chave"
                  field_value = "confvision_s3_tenant_id"
                } as $cv_s3_tenant
              
                conditional {
                  if ($cv_s3_tenant == null) {
                    db.add fp_config_financeiro {
                      data = {
                        created_at: "now"
                        chave     : "confvision_s3_tenant_id"
                        valor     : ""
                        descricao : "Tenant ID Contabo (obrigatorio para bucket tenant:bucket)"
                      }
                    } as $cfg_cv_s3_tenant
                  }
                }
              
                db.get fp_config_financeiro {
                  field_name = "chave"
                  field_value = "confvision_s3_access_key"
                } as $cv_s3_access
              
                conditional {
                  if ($cv_s3_access == null) {
                    db.add fp_config_financeiro {
                      data = {
                        created_at: "now"
                        chave     : "confvision_s3_access_key"
                        valor     : ""
                        descricao : "Access key S3 Contabo para provisionamento automatico de gravacao"
                      }
                    } as $cfg_cv_s3_access
                  }
                }
              
                db.get fp_config_financeiro {
                  field_name = "chave"
                  field_value = "confvision_s3_secret_key"
                } as $cv_s3_secret
              
                conditional {
                  if ($cv_s3_secret == null) {
                    db.add fp_config_financeiro {
                      data = {
                        created_at: "now"
                        chave     : "confvision_s3_secret_key"
                        valor     : ""
                        descricao : "Secret key S3 Contabo para provisionamento automatico de gravacao"
                      }
                    } as $cfg_cv_s3_secret
                  }
                }
              
                db.get fp_config_financeiro {
                  field_name = "chave"
                  field_value = "confvision_s3_segmento_minutos"
                } as $cv_s3_segmento
              
                conditional {
                  if ($cv_s3_segmento == null) {
                    db.add fp_config_financeiro {
                      data = {
                        created_at: "now"
                        chave     : "confvision_s3_segmento_minutos"
                        valor     : "5"
                        descricao : "Duracao padrao do segmento de gravacao (minutos)"
                      }
                    } as $cfg_cv_s3_segmento
                  }
                }
              }
            }
          
            db.get fp_config_financeiro {
              field_name = "chave"
              field_value = "confvision_s3_endpoint"
            } as $cv_s3_endpoint
          
            conditional {
              if ($cv_s3_endpoint == null) {
                db.add fp_config_financeiro {
                  data = {
                    created_at: "now"
                    chave     : "confvision_s3_endpoint"
                    valor     : "https://usc1.contabostorage.com"
                    descricao : "Endpoint S3 Contabo padrao para gravacao ConfVision"
                  }
                } as $cfg_cv_s3_endpoint
              }
            }
          
            db.get fp_config_financeiro {
              field_name = "chave"
              field_value = "confvision_s3_bucket"
            } as $cv_s3_bucket
          
            conditional {
              if ($cv_s3_bucket == null) {
                db.add fp_config_financeiro {
                  data = {
                    created_at: "now"
                    chave     : "confvision_s3_bucket"
                    valor     : "confvision"
                    descricao : "Bucket S3 padrao para gravacao ConfVision"
                  }
                } as $cfg_cv_s3_bucket
              }
            }
          
            db.get fp_config_financeiro {
              field_name = "chave"
              field_value = "confvision_s3_tenant_id"
            } as $cv_s3_tenant
          
            conditional {
              if ($cv_s3_tenant == null) {
                db.add fp_config_financeiro {
                  data = {
                    created_at: "now"
                    chave     : "confvision_s3_tenant_id"
                    valor     : ""
                    descricao : "Tenant ID Contabo (obrigatorio para bucket tenant:bucket)"
                  }
                } as $cfg_cv_s3_tenant
              }
            }
          
            db.get fp_config_financeiro {
              field_name = "chave"
              field_value = "confvision_s3_access_key"
            } as $cv_s3_access
          
            conditional {
              if ($cv_s3_access == null) {
                db.add fp_config_financeiro {
                  data = {
                    created_at: "now"
                    chave     : "confvision_s3_access_key"
                    valor     : ""
                    descricao : "Access key S3 Contabo para provisionamento automatico de gravacao"
                  }
                } as $cfg_cv_s3_access
              }
            }
          
            db.get fp_config_financeiro {
              field_name = "chave"
              field_value = "confvision_s3_secret_key"
            } as $cv_s3_secret
          
            conditional {
              if ($cv_s3_secret == null) {
                db.add fp_config_financeiro {
                  data = {
                    created_at: "now"
                    chave     : "confvision_s3_secret_key"
                    valor     : ""
                    descricao : "Secret key S3 Contabo para provisionamento automatico de gravacao"
                  }
                } as $cfg_cv_s3_secret
              }
            }
          
            db.get fp_config_financeiro {
              field_name = "chave"
              field_value = "confvision_s3_segmento_minutos"
            } as $cv_s3_segmento
          
            conditional {
              if ($cv_s3_segmento == null) {
                db.add fp_config_financeiro {
                  data = {
                    created_at: "now"
                    chave     : "confvision_s3_segmento_minutos"
                    valor     : "5"
                    descricao : "Duracao padrao do segmento de gravacao (minutos)"
                  }
                } as $cfg_cv_s3_segmento
              }
            }
          
            db.get fp_config_financeiro {
              field_name = "chave"
              field_value = "confvision_mediamtx_default_nome"
            } as $cv_mtx_nome
          
            conditional {
              if ($cv_mtx_nome == null) {
                db.add fp_config_financeiro {
                  data = {
                    created_at: "now"
                    chave     : "confvision_mediamtx_default_nome"
                    valor     : "servidor1"
                    descricao : "Nome do no MediaMTX padrao (servidor1)"
                  }
                } as $cfg_cv_mtx_nome
              }
            }
          
            db.get fp_config_financeiro {
              field_name = "chave"
              field_value = "confvision_mediamtx_default_rtmp_public"
            } as $cv_mtx_rtmp
          
            conditional {
              if ($cv_mtx_rtmp == null) {
                db.add fp_config_financeiro {
                  data = {
                    created_at: "now"
                    chave     : "confvision_mediamtx_default_rtmp_public"
                    valor     : "rtmp://servidor1.dnsid.com.br:1935"
                    descricao : "URL RTMP publica do no MediaMTX padrao"
                  }
                } as $cfg_cv_mtx_rtmp
              }
            }
          
            db.get fp_config_financeiro {
              field_name = "chave"
              field_value = "confvision_mediamtx_default_hls_public"
            } as $cv_mtx_hls
          
            conditional {
              if ($cv_mtx_hls == null) {
                db.add fp_config_financeiro {
                  data = {
                    created_at: "now"
                    chave     : "confvision_mediamtx_default_hls_public"
                    valor     : "https://hls.dnsid.com.br"
                    descricao : "URL HLS publica do no MediaMTX padrao"
                  }
                } as $cfg_cv_mtx_hls
              }
            }
          
            db.get fp_config_financeiro {
              field_name = "chave"
              field_value = "confvision_mediamtx_default_rtsp_internal"
            } as $cv_mtx_rtsp
          
            conditional {
              if ($cv_mtx_rtsp == null) {
                db.add fp_config_financeiro {
                  data = {
                    created_at: "now"
                    chave     : "confvision_mediamtx_default_rtsp_internal"
                    valor     : "rtsp://127.0.0.1:8554"
                    descricao : "URL RTSP interna (workers na mesma VPS do MediaMTX)"
                  }
                } as $cfg_cv_mtx_rtsp
              }
            }
          
            db.get fp_config_financeiro {
              field_name = "chave"
              field_value = "confvision_mediamtx_default_max_cameras"
            } as $cv_mtx_max
          
            conditional {
              if ($cv_mtx_max == null) {
                db.add fp_config_financeiro {
                  data = {
                    created_at: "now"
                    chave     : "confvision_mediamtx_default_max_cameras"
                    valor     : "200"
                    descricao : "Limite de cameras por no MediaMTX padrao"
                  }
                } as $cfg_cv_mtx_max
              }
            }
          }
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_s3_endpoint"
        } as $cv_s3_endpoint

        conditional {
          if ($cv_s3_endpoint == null) {
            db.add fp_config_financeiro {
              data = {
                created_at: "now"
                chave     : "confvision_s3_endpoint"
                valor     : "https://usc1.contabostorage.com"
                descricao : "Endpoint S3 Contabo padrao para gravacao ConfVision"
              }
            } as $cfg_cv_s3_endpoint
          }
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_s3_bucket"
        } as $cv_s3_bucket

        conditional {
          if ($cv_s3_bucket == null) {
            db.add fp_config_financeiro {
              data = {
                created_at: "now"
                chave     : "confvision_s3_bucket"
                valor     : "confvision"
                descricao : "Bucket S3 padrao para gravacao ConfVision"
              }
            } as $cfg_cv_s3_bucket
          }
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_s3_tenant_id"
        } as $cv_s3_tenant

        conditional {
          if ($cv_s3_tenant == null) {
            db.add fp_config_financeiro {
              data = {
                created_at: "now"
                chave     : "confvision_s3_tenant_id"
                valor     : ""
                descricao : "Tenant ID Contabo (obrigatorio para bucket tenant:bucket)"
              }
            } as $cfg_cv_s3_tenant
          }
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_s3_access_key"
        } as $cv_s3_access

        conditional {
          if ($cv_s3_access == null) {
            db.add fp_config_financeiro {
              data = {
                created_at: "now"
                chave     : "confvision_s3_access_key"
                valor     : ""
                descricao : "Access key S3 Contabo para provisionamento automatico de gravacao"
              }
            } as $cfg_cv_s3_access
          }
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_s3_secret_key"
        } as $cv_s3_secret

        conditional {
          if ($cv_s3_secret == null) {
            db.add fp_config_financeiro {
              data = {
                created_at: "now"
                chave     : "confvision_s3_secret_key"
                valor     : ""
                descricao : "Secret key S3 Contabo para provisionamento automatico de gravacao"
              }
            } as $cfg_cv_s3_secret
          }
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_s3_segmento_minutos"
        } as $cv_s3_segmento

        conditional {
          if ($cv_s3_segmento == null) {
            db.add fp_config_financeiro {
              data = {
                created_at: "now"
                chave     : "confvision_s3_segmento_minutos"
                valor     : "5"
                descricao : "Duracao padrao do segmento de gravacao (minutos)"
              }
            } as $cfg_cv_s3_segmento
          }
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_mediamtx_default_nome"
        } as $cv_mtx_nome

        conditional {
          if ($cv_mtx_nome == null) {
            db.add fp_config_financeiro {
              data = {
                created_at: "now"
                chave     : "confvision_mediamtx_default_nome"
                valor     : "servidor1"
                descricao : "Nome do no MediaMTX padrao (servidor1)"
              }
            } as $cfg_cv_mtx_nome
          }
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_mediamtx_default_rtmp_public"
        } as $cv_mtx_rtmp

        conditional {
          if ($cv_mtx_rtmp == null) {
            db.add fp_config_financeiro {
              data = {
                created_at: "now"
                chave     : "confvision_mediamtx_default_rtmp_public"
                valor     : "rtmp://servidor1.dnsid.com.br:1935"
                descricao : "URL RTMP publica do no MediaMTX padrao"
              }
            } as $cfg_cv_mtx_rtmp
          }
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_mediamtx_default_hls_public"
        } as $cv_mtx_hls

        conditional {
          if ($cv_mtx_hls == null) {
            db.add fp_config_financeiro {
              data = {
                created_at: "now"
                chave     : "confvision_mediamtx_default_hls_public"
                valor     : "https://hls.dnsid.com.br"
                descricao : "URL HLS publica do no MediaMTX padrao"
              }
            } as $cfg_cv_mtx_hls
          }
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_mediamtx_default_rtsp_internal"
        } as $cv_mtx_rtsp

        conditional {
          if ($cv_mtx_rtsp == null) {
            db.add fp_config_financeiro {
              data = {
                created_at: "now"
                chave     : "confvision_mediamtx_default_rtsp_internal"
                valor     : "rtsp://127.0.0.1:8554"
                descricao : "URL RTSP interna (workers na mesma VPS do MediaMTX)"
              }
            } as $cfg_cv_mtx_rtsp
          }
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_mediamtx_default_max_cameras"
        } as $cv_mtx_max

        conditional {
          if ($cv_mtx_max == null) {
            db.add fp_config_financeiro {
              data = {
                created_at: "now"
                chave     : "confvision_mediamtx_default_max_cameras"
                valor     : "200"
                descricao : "Limite de cameras por no MediaMTX padrao"
              }
            } as $cfg_cv_mtx_max
          }
        }
      }
    }
      }
    }
  }

  response = {inseridos: $inseridos, status: "OK", id_central: $id_central, user_tipo: $tipo, escopo: $escopo}
}