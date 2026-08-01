function FuncaoSistema_sendWhatsEventCentralAlarm {
  input {
    int idEvento?
  }

  stack {
    !return {
      value = ""
    }
  
    db.get alarm_events {
      field_name = "id"
      field_value = $input.idEvento
    } as $alarm_events1
  
    precondition (($alarm_events1.nivel|to_int) > 0 || $alarm_events1.ctiGrupo == "ARME" || $alarm_events1.ctiGrupo == "DESARME") {
      error = "Nivel menor que zero"
    }
  
    conditional {
      if ($alarm_events1.ctiGrupo == "FALHAS") {
        db.query alarm_events {
          where = $db.alarm_events.idDispositivo == $alarm_events1.idDispositivo && $db.alarm_events.ctiGrupo == "FALHAS" && $db.alarm_events.created_at >= ($alarm_events1.created_at|timestamp_add_minutes:-1)
          return = {type: "count"}
        } as $alarm_events2
      }
    
      else {
        var $alarm_events2 {
          value = 0
        }
      }
    }
  
    precondition ($alarm_events2 < 3) {
      error = "Erro Muitas Falhas"
    }
  
    // Cliente bloqueado (#107): cancela tudo (cliente, franqueado e IA)
    db.query WhatsEventCadBloq {
      where = $db.WhatsEventCadBloq.idFranqueado == `$alarm_events1.idFranqueado` && $db.WhatsEventCadBloq.idCliente == `$alarm_events1.idCliente`
      return = {type: "single"}
    } as $WhatsEventCadBloq1
  
    conditional {
      if (($WhatsEventCadBloq1|is_empty) == false) {
        return {
          value = {dados: "cliente_bloqueado"}
        }
      }
    }
  
    // ConfService: cliente com parceiro externo → webhook e para ConfMonit (WhatsApp/IA)
    var $confServiceUrl {
      value = "http://185.130.61.4:2020"
    }
  
    var $confServiceKey {
      value = "dev-confservice-api-key"
    }
  
    var $parceiroExterno {
      value = false
    }
  
    try_catch {
      try {
        api.request {
          url = $confServiceUrl ~ "/internal/vinculo/cliente?idFranqueado=" ~ ($alarm_events1.idFranqueado|url_encode) ~ "&idCliente=" ~ ($alarm_events1.idCliente|url_encode)
          method = "GET"
          headers = []
            |push:("X-Api-Key: "|concat:$confServiceKey)
            |push:"Accept: application/json"
          timeout = 8
        } as $csVinculo
      
        conditional {
          if ($csVinculo.response.result.ok == true && ($csVinculo.response.result.vinculo|is_empty) == false) {
            api.request {
              url = $confServiceUrl ~ "/internal/evento"
              method = "POST"
              params = {}
                |set:"idFranqueado":$alarm_events1.idFranqueado
                |set:"idCliente":$alarm_events1.idCliente
                |set:"alarmEventsId":$alarm_events1.id
                |set:"payload":({}
                  |set:"alarmEventsId":$alarm_events1.id
                  |set:"idEvento":$alarm_events1.idEvento
                  |set:"idProcesso":$alarm_events1.idProcesso
                  |set:"idFranqueado":$alarm_events1.idFranqueado
                  |set:"idCliente":$alarm_events1.idCliente
                  |set:"nomeCliente":$alarm_events1.nomeCliente
                  |set:"idDispositivo":$alarm_events1.idDispositivo
                  |set:"ctiGrupo":$alarm_events1.ctiGrupo
                  |set:"ctiDescricao":$alarm_events1.ctiDescricao
                  |set:"codigo":$alarm_events1.codigo
                  |set:"particao":$alarm_events1.particao
                  |set:"zonaUser":$alarm_events1.zonaUser
                  |set:"conta":$alarm_events1.conta
                  |set:"dataEntrada":$alarm_events1.dataEntrada
                  |set:"nivel":$alarm_events1.nivel
                )
              headers = []
                |push:("X-Api-Key: "|concat:$confServiceKey)
                |push:"Content-Type: application/json"
                |push:"Accept: application/json"
              timeout = 10
            } as $csEvento
          
            var.update $parceiroExterno {
              value = true
            }
          }
        }
      }
    
      catch {
        // ConfService indisponivel: segue fluxo ConfMonit (fail-open)
        var.update $parceiroExterno {
          value = false
        }
      }
    }
  
    conditional {
      if ($parceiroExterno) {
        return {
          value = {dados: "parceiro_externo"}
        }
      }
    }
  
    // Destinos do cliente (#5)
    conditional {
      if ($alarm_events1.codigo == "1M01" || $alarm_events1.codigo == "3M02") {
        db.query WhatsappEventoCad {
          where = $db.WhatsappEventoCad.IdCliente == `$alarm_events1.idCliente` && $db.WhatsappEventoCad.idDispositivo == $alarm_events1.idDispositivo && "GRADEHORARIO" in $db.WhatsappEventoCad.Tipo && $db.WhatsappEventoCad.bloqueado == false
          sort = {WhatsappEventoCad.ordemligacao: "asc"}
          return = {type: "list"}
        } as $WhatsappEventoCad1
      }
    
      else {
        db.query WhatsappEventoCad {
          where = $db.WhatsappEventoCad.IdCliente == `$alarm_events1.idCliente` && $db.WhatsappEventoCad.idDispositivo == $alarm_events1.idDispositivo && $alarm_events1.ctiGrupo in $db.WhatsappEventoCad.Tipo && $db.WhatsappEventoCad.bloqueado == false
          sort = {WhatsappEventoCad.ordemligacao: "asc"}
          return = {type: "list"}
        } as $WhatsappEventoCad1
      }
    }
  
    // Telefones do franqueado (#106): monitoramento / gerente / viatura
    db.query WhatsEventCadFranq {
      where = $db.WhatsEventCadFranq.idFranqueado == `$alarm_events1.idFranqueado`
      return = {type: "single"}
    } as $WhatsEventCadFranq1
  
    // Grupos: mensagem+ligacao vs somente texto; filtros por destino
    var $franqGrupoCritico {
      value = $alarm_events1.ctiGrupo == "ALARME" || $alarm_events1.ctiGrupo == "GERAL" || $alarm_events1.ctiGrupo == "GRADEHORARIO" || $alarm_events1.ctiGrupo == "INTERNET" || $alarm_events1.ctiGrupo == "ENERGIA" || $alarm_events1.ctiGrupo == "EMERGENCIA" || $alarm_events1.ctiGrupo == "MEDICO" || $alarm_events1.ctiGrupo == "PANICO"
    }
  
    var $franqGrupoViatura {
      value = $alarm_events1.ctiGrupo == "ALARME" || $alarm_events1.ctiGrupo == "EMERGENCIA" || $alarm_events1.ctiGrupo == "MEDICO" || $alarm_events1.ctiGrupo == "PANICO"
    }
  
    var $temDestinoFranq {
      value = false
    }
  
    conditional {
      if (($WhatsEventCadFranq1|is_empty) == false) {
        conditional {
          if ((($WhatsEventCadFranq1.telMonitoramento|is_empty) == false) && $franqGrupoCritico) {
            var.update $temDestinoFranq {
              value = true
            }
          }
        }
      
        conditional {
          if (($WhatsEventCadFranq1.telGerente|is_empty) == false) {
            var.update $temDestinoFranq {
              value = true
            }
          }
        }
      
        conditional {
          if ((($WhatsEventCadFranq1.telViatura|is_empty) == false) && $franqGrupoViatura) {
            var.update $temDestinoFranq {
              value = true
            }
          }
        }
      }
    }
  
    precondition ((($WhatsappEventoCad1|is_empty) == false) || $temDestinoFranq) {
      error = "Nenhum Whatsapp Cadastrado"
      payload = false
    }
  
    !conditional {
      if ($alarm_events1.ctiGrupo == "RESTAURE") {
        var $WhatsappEventoCad1 {
          value = []
        }
      }
    }
  
    // bot_finalizaeventoauto (IA)
  
    conditional {
      if ($alarm_events1.codigo != "1M01" && $alarm_events1.codigo != "3M02") {
        db.add_or_edit bot_finalizaeventoauto {
          field_name = "idProcesso"
          field_value = $alarm_events1.idProcesso
          enforce_hidden_fields = false
          data = {
            idDispositivo    : $alarm_events1.idDispositivo
            ultimo_evento_ts : now
            status           : "PENDENTE"
            rodar_em         : now
            tentativas       : 0
            lock_token       : ""
            lock_at          : now
            motivo_final     : ""
            acao_final       : ""
            erro             : ""
            payload_resultado: {}
            updated_at       : now
            alarm_events_id  : $alarm_events1.id
          }
        } as $bot_finalizaeventoauto1
      }
    }
  
    function.run WebLogarCache as $func1
    function.run ConfMonitCliente_DadosClienteID {
      input = {
        Authorization: `$func1`
        idcliente    : $alarm_events1.idCliente
      }
    } as $func2
  
    function.run ConfMonitEquipamento_DadosEquip {
      input = {
        Authorization: $func1
        idDispositivo: $alarm_events1.idDispositivo
      }
    } as $func3
  
    function.run ConfMonitEquipamento_DadosEquipSetor {
      input = {
        Authorization: $func1
        idDispositivo: $alarm_events1.idDispositivo
        particao     : $alarm_events1.particao
        zonauser     : $alarm_events1.zonaUser
      }
    } as $func5
  
    // REGRAS DE ENVIO WHATSAPP
  
    // REGRAS DE ENVIO DO WHATSAPP
    function.run fn_RegrasEnvioEvento {
      input = {
        ctiDescricao      : $alarm_events1.ctiDescricao
        ctiGrupo          : $alarm_events1.ctiGrupo
        codigo            : $alarm_events1.codigo
        idDispositivo     : $alarm_events1.idDispositivo
        zonaUser          : $alarm_events1.zonaUser
        id_tblAlarm_events: $alarm_events1.id
        particao          : $alarm_events1.particao
        idCliente         : $alarm_events1.idCliente
        nomeCliente       : $alarm_events1.nomeCliente
        emailCliente      : $alarm_events1.emailCliente
        idFranqueado      : $alarm_events1.idFranqueado
        GRADEHORARIO      : `$WhatsappEventoCad1.Tipo|in:"GRADEHORARIO"`
        idProcesso        : $alarm_events1.idProcesso
      }
    } as $REGRASENVIO
  
    precondition ($REGRASENVIO)
    security.random_number {
      min = 1
      max = 3
    } as $randomTexto
  
    conditional {
      if ($alarm_events1.ctiGrupo == "ARME" || $alarm_events1.ctiGrupo == "DESARME") {
        db.query status_link {
          where = $db.status_link.idDispositivo == $alarm_events1.idDispositivo && $db.status_link.alarm_events_id == $alarm_events1.id && $db.status_link.grupo == "ARME"
          sort = {status_link.expiresAt: "desc"}
          return = {type: "single"}
        } as $status_link1
      
        var $LinkAcesso {
          value = "https://admmonitoramento.com.br/finalizaevento/?dispositivo=" ~ $var.alarm_events1.idDispositivo ~ "&chave=" ~ $status_link1.token
        }
      }
    
      elseif ($alarm_events1.ctiGrupo == "ALARME") {
        db.query status_link {
          where = $db.status_link.idDispositivo == $alarm_events1.idDispositivo && $db.status_link.alarm_events_id == $alarm_events1.id && $db.status_link.grupo == "ALARME"
          sort = {status_link.expiresAt: "desc"}
          return = {type: "single"}
        } as $status_link1
      
        var $LinkAcesso {
          value = "https://admmonitoramento.com.br/finalizaevento/?dispositivoalarme=" ~ $var.alarm_events1.idDispositivo ~ "&chave=" ~ $status_link1.token
        }
      }
    
      else {
        var $LinkAcesso {
          value = ""
        }
      }
    }
  
    conditional {
      if ($func5|is_empty) {
        var $DispNome {
          value = ""
        }
      
        var $DispDescricao {
          value = ""
        }
      
        var $DispTipo {
          value = ""
        }
      
        // Tipos de Mensagens Whatsapp
      
        var $SendMsgWhats {
          value = ```
            $func3.nomeFranqueado ~ 
            "\nCentral de Monitoramento Informa:" ~
            "\nCliente: " ~ $alarm_events1.nomeCliente ~ 
            "\nData: " ~ $alarm_events1.dataEntrada ~
            "\nEvento da sua Central - " ~ $alarm_events1.ctiGrupo ~ " - " ~ $alarm_events1.ctiDescricao ~ 
            "\nDispositivo: " ~ $func3.nome ~ 
            "\nPartição: " ~ $func3.particao ~ " - Conta: " ~ $alarm_events1.conta ~ 
            "\nCordialmente, " ~ $func2.response.result.dados.fraRazao
            ```|text_unescape
        }
      
        !conditional {
          if ($randomTexto == 1) {
            var $SendMsgWhats {
              value = `🚨 Alerta da Central de Monitoramento`|text_unescape
            }
          }
        
          elseif ($randomTexto == 2) {
            var $SendMsgWhats {
              value = `Notificação da Central`|text_unescape
            }
          }
        
          else {
            var $SendMsgWhats {
              value = `📡 Evento registrado na central`|text_unescape
            }
          }
        }
      
        var $SendMsgWhatsAudio {
          value = ```
            "Cliente: " ~ $alarm_events1.nomeCliente ~
            ". A central de monitoramento informa: " ~
            "Identificamos o seguinte evento: " ~ $alarm_events1.ctiDescricao
            ```
        }
      }
    
      else {
        var $DispNome {
          value = $func5.nome|first
        }
      
        var $DispDescricao {
          value = $func5.descricao|first
        }
      
        var $DispTipo {
          value = $func5.tipo|first
        }
      
        var $SendMsgWhats {
          value = ```
            $func3.nomeFranqueado ~ 
            "\nCentral de Monitoramento Informa:" ~
            "\nCliente: " ~ $alarm_events1.nomeCliente ~ 
            "\nData: " ~ $alarm_events1.dataEntrada ~
            "\nEvento da sua Central - " ~ $alarm_events1.ctiGrupo ~ " - " ~ $alarm_events1.ctiDescricao ~ 
            "\nDispositivo: " ~ $func3.nome ~ 
            "\nPartição: " ~ $func3.particao ~ " - Conta: " ~ $alarm_events1.conta ~ 
            "\nLocal: " ~ $DispDescricao ~ " - " ~ $DispNome ~ " - " ~ $DispTipo ~
            "\nCordialmente, " ~ $func2.response.result.dados.fraRazao
            ```|text_unescape
        }
      
        !conditional {
          if ($randomTexto == 1) {
            var $SendMsgWhats {
              value = `\\\\nCliente:`|text_unescape
            }
          }
        
          elseif ($randomTexto == 2) {
            var $SendMsgWhats {
              value = `Notificação da Central`|text_unescape
            }
          }
        
          else {
            var $SendMsgWhats {
              value = `📡 Evento registrado na central`|text_unescape
            }
          }
        }
      
        var $SendMsgWhatsAudio {
          value = ```
            "Cliente: " ~ $alarm_events1.nomeCliente ~
            ". A central " ~ $func3.nomeFranqueado ~ ", informa:" ~
            " Identificamos o seguinte evento: " ~ $alarm_events1.ctiDescricao ~
            ". " ~ ($func5.descricao|first)
            ```
        }
      }
    }
  
    !var $SendMsgWhats {
      value = ```
        $func3.nomeFranqueado ~ 
        "\\nCentral de Monitoramento Informa:" ~
        "\\nCliente: " ~ $alarm_events1.nomeCliente ~ 
        "\\nData: " ~ $alarm_events1.dataEntrada ~
        "\\nEvento da sua Central - " ~ $alarm_events1.ctiGrupo ~ " - " ~ $alarm_events1.ctiDescricao ~ 
        "\\nDispositivo: " ~ $func3.nome ~ 
        "\\nPartição: " ~ $func3.particao ~ " - Conta: " ~ $alarm_events1.conta ~ 
        "\\nCordialmente, " ~ $func2.response.result.dados.fraRazao
        ```|text_unescape
    }
  
    !function.run OpenAi_Texto {
      input = {SendMsgWhats: $SendMsgWhats}
    } as $openAi_Texto
  
    !var $SendMsgWhats {
      value = $openAi_Texto.dados|text_unescape
    }
  
    var $varWhatsAppEnviadosLigar {
      value = 0
    }
  
    var $varWhatsAppEnviadosAudio {
      value = 0
    }
  
    var $varWhatsAppEnviadosTexto {
      value = 0
    }
  
    // Inclui destinos do franqueado (#106) junto com clientes (#5)
    // Monitoramento: so grupos criticos | Gerente: todos | Viatura: 4 grupos + horario
    // Grupos criticos: texto+ligar | demais (gerente): so texto
    conditional {
      if (($WhatsEventCadFranq1|is_empty) == false) {
        conditional {
          if ((($WhatsEventCadFranq1.telMonitoramento|is_empty) == false) && $franqGrupoCritico) {
            array.push $WhatsappEventoCad1 {
              value = {
                whatsapp       : $WhatsEventCadFranq1.telMonitoramento
                texto          : true
                audio          : false
                ligar          : true
                notificarsempre: true
              }
            }
          }
        }
      
        conditional {
          if (($WhatsEventCadFranq1.telGerente|is_empty) == false) {
            array.push $WhatsappEventoCad1 {
              value = {
                whatsapp       : $WhatsEventCadFranq1.telGerente
                texto          : true
                audio          : false
                ligar          : $franqGrupoCritico
                notificarsempre: true
              }
            }
          }
        }
      
        conditional {
          if ((($WhatsEventCadFranq1.telViatura|is_empty) == false) && $franqGrupoViatura) {
            var $horaAtual {
              value = now|format_timestamp:"H:i":"America/Sao_Paulo"
            }
          
            var $horaIni {
              value = $WhatsEventCadFranq1.horaIni
            }
          
            var $horaFin {
              value = $WhatsEventCadFranq1.horaFin
            }
          
            var $viaturaNoHorario {
              value = false
            }
          
            conditional {
              if (($horaIni|is_empty) || ($horaFin|is_empty)) {
                var.update $viaturaNoHorario {
                  value = false
                }
              }
            
              elseif ($horaIni <= $horaFin) {
                conditional {
                  if ($horaAtual >= $horaIni && $horaAtual <= $horaFin) {
                    var.update $viaturaNoHorario {
                      value = true
                    }
                  }
                }
              }
            
              else {
                // Janela atravessa meia-noite (ex.: 22:00 -> 06:00)
                conditional {
                  if ($horaAtual >= $horaIni || $horaAtual <= $horaFin) {
                    var.update $viaturaNoHorario {
                      value = true
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($viaturaNoHorario) {
                array.push $WhatsappEventoCad1 {
                  value = {
                    whatsapp       : $WhatsEventCadFranq1.telViatura
                    texto          : true
                    audio          : false
                    ligar          : true
                    notificarsempre: true
                  }
                }
              }
            }
          }
        }
      }
    }
  
    // Sem destino ativo (ex.: so viatura fora do horario)
    conditional {
      if ($WhatsappEventoCad1|is_empty) {
        return {
          value = {dados: "sem_destino_ativo"}
        }
      }
    }
  
    foreach ($WhatsappEventoCad1) {
      each as $item {
        conditional {
          if ($item.texto) {
            db.add WhatsAppEnviados {
              enforce_hidden_fields = false
              data = {
                created_at            : "now"
                whats_ID              : ""
                whats_messageTimestamp: ""
                whats_messageid       : ""
                whats_sender          : ""
                whats_senderName      : ""
                whats_text            : ""
                id_tblAlarm_events    : $alarm_events1.id
                id_Evento             : $alarm_events1.idEvento
                Codigo                : $alarm_events1.codigo
                Particao              : $alarm_events1.particao
                ZonaUser              : $alarm_events1.zonaUser
                DataEntrada           : $alarm_events1.dataEntrada
                idProcesso            : $alarm_events1.idProcesso
                idDispositivo         : $alarm_events1.idDispositivo
                idCliente             : $alarm_events1.idCliente
                ctiGrupo              : $alarm_events1.ctiGrupo
                idFranqueado          : $alarm_events1.idFranqueado
                codigoBenuvem         : $alarm_events1.codigoBenuvem
                Conta                 : $alarm_events1.conta
                CameraAtiva           : $alarm_events1.carmeraAtiva
                Data                  : `$alarm_events1.created_at|format_timestamp:"d-m-Y":timezone:"America/Sao_Paulo"`
                audio                 : false
                ligacao               : false
                texto                 : true
                DispNome              : $DispNome
                DispDescricao         : $DispDescricao
                DispTipo              : $DispTipo
              }
            } as $tblWhatsAppEnviadosTexto
          
            var.update $varWhatsAppEnviadosTexto {
              value = $tblWhatsAppEnviadosTexto.id
            }
          }
        }
      
        conditional {
          if ($item.audio) {
            db.add WhatsAppEnviados {
              enforce_hidden_fields = false
              data = {
                created_at            : "now"
                whats_ID              : ""
                whats_messageTimestamp: ""
                whats_messageid       : ""
                whats_sender          : ""
                whats_senderName      : ""
                whats_text            : ""
                id_tblAlarm_events    : $alarm_events1.id
                id_Evento             : $alarm_events1.idEvento
                Codigo                : $alarm_events1.codigo
                Particao              : $alarm_events1.particao
                ZonaUser              : $alarm_events1.zonaUser
                DataEntrada           : $alarm_events1.dataEntrada
                idProcesso            : $alarm_events1.idProcesso
                idDispositivo         : $alarm_events1.idDispositivo
                idCliente             : $alarm_events1.idCliente
                ctiGrupo              : $alarm_events1.ctiGrupo
                idFranqueado          : $alarm_events1.idFranqueado
                codigoBenuvem         : $alarm_events1.codigoBenuvem
                Conta                 : $alarm_events1.conta
                CameraAtiva           : $alarm_events1.carmeraAtiva
                Data                  : `$alarm_events1.created_at|format_timestamp:"d-m-Y":timezone:"America/Sao_Paulo"`
                audio                 : true
                ligacao               : false
                texto                 : false
                DispNome              : $DispNome
                DispDescricao         : $DispDescricao
                DispTipo              : $DispTipo
              }
            } as $tblWhatsAppEnviadosAudio
          
            var.update $varWhatsAppEnviadosAudio {
              value = $tblWhatsAppEnviadosAudio.id
            }
          }
        }
      
        conditional {
          if ($item.ligar) {
            db.add WhatsAppEnviados {
              enforce_hidden_fields = false
              data = {
                created_at            : "now"
                whats_ID              : ""
                whats_messageTimestamp: ""
                whats_messageid       : ""
                whats_sender          : ""
                whats_senderName      : ""
                whats_text            : ""
                id_tblAlarm_events    : $alarm_events1.id
                id_Evento             : $alarm_events1.idEvento
                Codigo                : $alarm_events1.codigo
                Particao              : $alarm_events1.particao
                ZonaUser              : $alarm_events1.zonaUser
                DataEntrada           : $alarm_events1.dataEntrada
                idProcesso            : $alarm_events1.idProcesso
                idDispositivo         : $alarm_events1.idDispositivo
                idCliente             : $alarm_events1.idCliente
                ctiGrupo              : $alarm_events1.ctiGrupo
                idFranqueado          : $alarm_events1.idFranqueado
                codigoBenuvem         : $alarm_events1.codigoBenuvem
                Conta                 : $alarm_events1.conta
                CameraAtiva           : $alarm_events1.carmeraAtiva
                Data                  : `$alarm_events1.created_at|format_timestamp:"d-m-Y":timezone:"America/Sao_Paulo"`
                audio                 : false
                ligacao               : true
                texto                 : false
                DispNome              : $DispNome
                DispDescricao         : $DispDescricao
                DispTipo              : $DispTipo
              }
            } as $tblWhatsAppEnviadosLigar
          
            var.update $varWhatsAppEnviadosLigar {
              value = $tblWhatsAppEnviadosLigar.id
            }
          }
        }
      
        !function.run FuncaoSis_WhatsappProcFila {
          runtime_mode = "async-shared"
          input = {
            SendMsgWhats            : $SendMsgWhats
            numerowhatsapp          : $item.whatsapp
            enviaSom                : $item.audio
            SendAudioWhats          : $SendMsgWhatsAudio
            tblWhatsAppEnviadosTexto: $varWhatsAppEnviadosTexto
            tblWhatsAppEnviadosAudio: $varWhatsAppEnviadosAudio
            tblWhatsAppEnviadosLigar: $varWhatsAppEnviadosLigar
            ligar                   : $item.ligar
            idFranqueado            : $alarm_events1.idFranqueado
            nomecliente             : $alarm_events1.nomeCliente
            empresanome             : $func3.nomeFranqueado
            tipoevento              : $alarm_events1.ctiDescricao
            zona                    : "Partição: " ~ $func3.particao ~ " Zona: " ~ $alarm_events1.zonaUser
            local                   : $func3.conta
            idevento                : $alarm_events1.id
            datahorario             : $alarm_events1.dataEntrada
            ideventgo               : $alarm_events1.idEvento
            enviartexto             : $item.texto
            idProcesso              : $alarm_events1.idProcesso
            idDispositivo           : $alarm_events1.idDispositivo
            notificarsempre         : $item.notificarsempre
            DispNome                : $DispNome
            DispDescricao           : $DispDescricao
            DispTipo                : $DispTipo
          }
        } as $func4
      
        conditional {
          if ($alarm_events1.ctiGrupo == "ALARME") {
            db.add WhatsappProcFila {
              enforce_hidden_fields = false
              data = {
                SendMsgWhats            : $SendMsgWhats
                numerowhatsapp          : $item.whatsapp
                enviaSom                : $item.audio
                SendAudioWhats          : $SendMsgWhatsAudio
                tblWhatsAppEnviadosTexto: $varWhatsAppEnviadosTexto
                tblWhatsAppEnviadosAudio: $varWhatsAppEnviadosAudio
                tblWhatsAppEnviadosLigar: $varWhatsAppEnviadosLigar
                ligar                   : $item.ligar
                idFranqueado            : $alarm_events1.idFranqueado
                nomecliente             : $alarm_events1.nomeCliente
                empresanome             : $func3.nomeFranqueado
                GRUPOFALHA              : $alarm_events1.ctiGrupo
                tipoevento              : $alarm_events1.ctiDescricao
                zona                    : "Partição: " ~ $func3.particao ~ " Zona: " ~ $alarm_events1.zonaUser
                local                   : $func3.conta
                idevento                : $alarm_events1.id
                datahorario             : $alarm_events1.dataEntrada
                ideventgo               : $alarm_events1.idEvento
                enviartexto             : $item.texto
                idProcesso              : $alarm_events1.idProcesso
                idDispositivo           : $alarm_events1.idDispositivo
                notificarsempre         : $item.notificarsempre
                DispNome                : $DispNome
                DispDescricao           : $DispDescricao
                DispTipo                : $DispTipo
                analise                 : true
                zonauser                : $alarm_events1.zonaUser
                particao                : $alarm_events1.particao
              }
            } as $func4
          }
        
          else {
            db.add WhatsappProcFila {
              enforce_hidden_fields = false
              data = {
                SendMsgWhats            : $SendMsgWhats
                numerowhatsapp          : $item.whatsapp
                enviaSom                : $item.audio
                SendAudioWhats          : $SendMsgWhatsAudio
                tblWhatsAppEnviadosTexto: $varWhatsAppEnviadosTexto
                tblWhatsAppEnviadosAudio: $varWhatsAppEnviadosAudio
                tblWhatsAppEnviadosLigar: $varWhatsAppEnviadosLigar
                ligar                   : $item.ligar
                idFranqueado            : $alarm_events1.idFranqueado
                nomecliente             : $alarm_events1.nomeCliente
                empresanome             : $func3.nomeFranqueado
                GRUPOFALHA              : $alarm_events1.ctiGrupo
                tipoevento              : $alarm_events1.ctiDescricao
                zona                    : "Partição: " ~ $func3.particao ~ " Zona: " ~ $alarm_events1.zonaUser
                local                   : $func3.conta
                idevento                : $alarm_events1.id
                datahorario             : $alarm_events1.dataEntrada
                ideventgo               : $alarm_events1.idEvento
                enviartexto             : $item.texto
                idProcesso              : $alarm_events1.idProcesso
                idDispositivo           : $alarm_events1.idDispositivo
                notificarsempre         : $item.notificarsempre
                DispNome                : $DispNome
                DispDescricao           : $DispDescricao
                DispTipo                : $DispTipo
                analise                 : false
                zonauser                : $alarm_events1.zonaUser
              }
            } as $func4
          }
        }
      }
    }
  
    // /notify fica a cargo do trigger whatsprocmsg (insert WhatsappProcFila)
    // e do trigger alarmeEventos_dispara — evita ciclo duplicado no taskxano.
  }

  response = "Dados"
    |set:"Data":`$var.func4alarm_events1.created_at|format_timestamp:"d-m-Y":timezone:"America/Sao_Paulo"`
    |set:"Evento":$alarm_events1
    |set:"Whatsapp":$func4
}