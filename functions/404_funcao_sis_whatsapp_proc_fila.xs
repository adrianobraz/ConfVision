function FuncaoSis_WhatsappProcFila {
  input {
    text SendMsgWhats? filters=trim
    text numerowhatsapp? filters=trim
    bool enviaSom?
    text SendAudioWhats? filters=trim
    int tblWhatsAppEnviadosTexto?
    int tblWhatsAppEnviadosAudio?
    int tblWhatsAppEnviadosLigar?
    bool ligar?
    text idFranqueado? filters=trim
    text nomecliente? filters=trim
    text empresanome? filters=trim
    text tipoevento? filters=trim
    text zona? filters=trim
    text local? filters=trim
    int idevento?
    text datahorario? filters=trim
    text ideventgo? filters=trim
    bool enviartexto?
    text idProcesso? filters=trim
    text idDispositivo? filters=trim
    bool notificarsempre?
    text DispNome? filters=trim
    text DispDescricao? filters=trim
    text DispTipo? filters=trim
  }

  stack {
    var $retorno {
      value = true
    }
  
    // Manda Mensagem Texto
  
    conditional {
      if ($input.enviartexto) {
        db.query WhatsappProcFila {
          where = $db.WhatsappProcFila.idProcesso == $input.idProcesso && $db.WhatsappProcFila.disparo == true
          sort = {WhatsappProcFila.dtDisparo: "desc"}
          return = {type: "single"}
        } as $ProcFilaTempo
      
        var $podeEnviarTexto {
          value = true
        }
      
        conditional {
          if (($ProcFilaTempo|is_empty) == false) {
            conditional {
              if (($ProcFilaTempo.dtDisparo|add_secs_to_timestamp:15) > now) {
                var.update $podeEnviarTexto {
                  value = false
                }
              
                var.update $retorno {
                  value = false
                }
              }
            }
          }
        }
      
        conditional {
          if ($podeEnviarTexto) {
            db.query WhatsappFranqueadoNroTelefone {
              where = $db.WhatsappFranqueadoNroTelefone.franqueado == $input.idFranqueado && $db.WhatsappFranqueadoNroTelefone.modulotexto == true
              sort = {Grupo: "desc"}
              return = {
                type : "aggregate"
                group: {Grupo: $db.WhatsappFranqueadoNroTelefone.Grupo}
              }
            } as $WhatsappFranqueadoNroTelefone1
          
            var $FranqueadoID {
              value = $input.idFranqueado
            }
          
            conditional {
              if ($WhatsappFranqueadoNroTelefone1|is_empty) {
                db.query WhatsappFranqueadoNroTelefone {
                  where = $db.WhatsappFranqueadoNroTelefone.franqueado == `"019c2bb9-5a02-74b9-812a-a20cb4c477ae"` && $db.WhatsappFranqueadoNroTelefone.modulotexto == true
                  sort = {Grupo: "desc"}
                  return = {
                    type : "aggregate"
                    group: {Grupo: $db.WhatsappFranqueadoNroTelefone.Grupo}
                  }
                } as $WhatsappFranqueadoNroTelefone1
              
                var.update $FranqueadoID {
                  value = "019c2bb9-5a02-74b9-812a-a20cb4c477ae"
                }
              }
            }
          
            conditional {
              if (($WhatsappFranqueadoNroTelefone1|count) > 1) {
                var $modHora {
                  value = `(now|to_timestamp|format_timestamp:"H":"America/Sao_Paulo")|to_int)`
                }
              
                var $modulo {
                  value = `$var.modHora|mod:($var.WhatsappFranqueadoNroTelefone1|count)`
                }
              
                var $idWhatsFranq {
                  value = $var.WhatsappFranqueadoNroTelefone1.[$var.modulo].Grupo
                }
              
                db.query WhatsappFranqueadoNroTelefone {
                  where = $db.WhatsappFranqueadoNroTelefone.franqueado == $FranqueadoID && $db.WhatsappFranqueadoNroTelefone.modulotexto == true && $db.WhatsappFranqueadoNroTelefone.Grupo == $idWhatsFranq
                  sort = {WhatsappFranqueadoNroTelefone.id: "rand"}
                  return = {
                    type  : "list"
                    paging: {page: 1, per_page: 1, metadata: false}
                  }
                } as $WhatsappFranqueadoNroTelefone1
              }
            
              else {
                conditional {
                  if (($WhatsappFranqueadoNroTelefone1|is_empty) == false) {
                    db.query WhatsappFranqueadoNroTelefone {
                      where = $db.WhatsappFranqueadoNroTelefone.franqueado == $FranqueadoID && $db.WhatsappFranqueadoNroTelefone.modulotexto == true && $db.WhatsappFranqueadoNroTelefone.Grupo == $WhatsappFranqueadoNroTelefone1[0].Grupo
                      sort = {WhatsappFranqueadoNroTelefone.id: "rand"}
                      return = {
                        type  : "list"
                        paging: {page: 1, per_page: 1, metadata: false}
                      }
                    } as $WhatsappFranqueadoNroTelefone1
                  }
                }
              }
            }
          
            conditional {
              if (($WhatsappFranqueadoNroTelefone1|is_empty) == false) {
                conditional {
                  if ($WhatsappFranqueadoNroTelefone1[0].tipoApi == "U") {
                    api.request {
                      url = "https://monitoramento.uazapi.com/chat/details"
                      method = "POST"
                      params = {}
                        |set:"number":$input.numerowhatsapp
                        |set:"preview":false
                      headers = []
                        |push:"Accept: application/json"
                        |push:"Content-Type: application/json"
                        |push:("token:"
                          |concat:$WhatsappFranqueadoNroTelefone1[0].txtApiToken
                        )
                    } as $api1
                  
                    var $ContatoNome {
                      value = $api1.response.result.wa_contactName
                    }
                  
                    var $Nome {
                      value = $api1.response.result.name
                    }
                  
                    conditional {
                      if ($ContatoNome|is_empty) {
                        conditional {
                          if (($Nome|is_empty) == false) {
                            api.request {
                              url = "https://monitoramento.uazapi.com/contact/add"
                              method = "POST"
                              params = {}
                                |set:"number":$input.numerowhatsapp
                                |set:"name":($Nome|concat:$input.empresanome:" ")
                              headers = []
                                |push:"Accept: application/json"
                                |push:"Content-Type: application/json"
                                |push:("token:"
                                  |concat:$WhatsappFranqueadoNroTelefone1[0].txtApiToken
                                )
                            }
                          }
                        
                          else {
                            api.request {
                              url = "https://monitoramento.uazapi.com/contact/add"
                              method = "POST"
                              params = {}
                                |set:"number":$input.numerowhatsapp
                                |set:"name":($input.nomecliente|concat:$input.empresanome:" ")
                              headers = []
                                |push:"Accept: application/json"
                                |push:"Content-Type: application/json"
                                |push:("token:"
                                  |concat:$WhatsappFranqueadoNroTelefone1[0].txtApiToken
                                )
                            }
                          }
                        }
                      }
                    }
                  }
                }
              
                conditional {
                  if ($WhatsappFranqueadoNroTelefone1[0].tipoApi == "U" || $WhatsappFranqueadoNroTelefone1[0].tipoApi == "G") {
                    var $textoEnviado {
                      value = false
                    }
                  
                    try_catch {
                      try {
                        function.run uazapi_EnvioMsgTexto {
                          input = {
                            number             : $input.numerowhatsapp
                            textMensagem       : $input.SendMsgWhats
                            tblWhatsAppEnviados: $input.tblWhatsAppEnviadosTexto
                            InstanceToken      : $WhatsappFranqueadoNroTelefone1[0].txtApiToken
                            tipoAPi            : $WhatsappFranqueadoNroTelefone1[0].tipoApi
                            instancia          : $WhatsappFranqueadoNroTelefone1[0].instancia
                          }
                        } as $func4
                      
                        conditional {
                          if ($func4.response.status >= 200 && $func4.response.status < 300) {
                            var.update $textoEnviado {
                              value = true
                            }
                          }
                        }
                      }
                    
                      catch {
                        var.update $textoEnviado {
                          value = false
                        }
                      }
                    }
                  
                    // Fallback: WhatsApp falhou -> SMS (exclui U e G)
                    conditional {
                      if ($textoEnviado == false) {
                        db.query WhatsappFranqueadoNroTelefone {
                          where = $db.WhatsappFranqueadoNroTelefone.franqueado == $FranqueadoID && $db.WhatsappFranqueadoNroTelefone.modulotexto == true && $db.WhatsappFranqueadoNroTelefone.tipoApi != "U" && $db.WhatsappFranqueadoNroTelefone.tipoApi != "G"
                          sort = {WhatsappFranqueadoNroTelefone.id: "rand"}
                          return = {
                            type  : "list"
                            paging: {page: 1, per_page: 1, metadata: false}
                          }
                        } as $instanciaSMSFallback
                      
                        conditional {
                          if ($instanciaSMSFallback|is_empty) {
                            db.query WhatsappFranqueadoNroTelefone {
                              where = $db.WhatsappFranqueadoNroTelefone.franqueado == `"019c2bb9-5a02-74b9-812a-a20cb4c477ae"` && $db.WhatsappFranqueadoNroTelefone.modulotexto == true && $db.WhatsappFranqueadoNroTelefone.tipoApi != "U" && $db.WhatsappFranqueadoNroTelefone.tipoApi != "G"
                              sort = {WhatsappFranqueadoNroTelefone.id: "rand"}
                              return = {
                                type  : "list"
                                paging: {page: 1, per_page: 1, metadata: false}
                              }
                            } as $instanciaSMSFallback
                          }
                        }
                      
                        conditional {
                          if (($instanciaSMSFallback|is_empty) == false) {
                            try_catch {
                              try {
                                function.run sms {
                                  input = {
                                    number             : $input.numerowhatsapp
                                    textMensagem       : $input.SendMsgWhats
                                    tblWhatsAppEnviados: $input.tblWhatsAppEnviadosTexto
                                    InstanceToken      : $instanciaSMSFallback[0].txtApiToken
                                    tipoAPi            : $instanciaSMSFallback[0].tipoApi
                                  }
                                } as $funcSMSFallback
                              }
                            
                              catch {
                                // fallback SMS tambem falhou; WhatsAppEnviados fica sem confirmacao para auditoria
                              }
                            }
                          }
                        }
                      }
                    }
                  }
                
                  else {
                    function.run sms {
                      input = {
                        number             : $input.numerowhatsapp
                        textMensagem       : $input.SendMsgWhats
                        tblWhatsAppEnviados: $input.tblWhatsAppEnviadosTexto
                        InstanceToken      : $WhatsappFranqueadoNroTelefone1[0].txtApiToken
                        tipoAPi            : $WhatsappFranqueadoNroTelefone1[0].tipoApi
                      }
                    } as $func1
                  }
                }
              }
            }
          }
        }
      }
    }
  
    // Envia Audio
  
    conditional {
      if ($input.enviaSom) {
        try_catch {
          try {
            function.run uazapi_EnvioMsgAudio {
              input = {
                number             : $input.numerowhatsapp
                textMensagem       : $input.SendAudioWhats
                tblWhatsAppEnviados: $input.tblWhatsAppEnviadosAudio
              }
            } as $func5
          }
        
          catch {
            // falha no envio de audio nao pode travar texto/ligacao do mesmo item
          }
        }
      }
    }
  
    // Funcao Faz Ligação
  
    conditional {
      if ($input.ligar) {
        db.query whatsappLigarErro {
          where = $db.whatsappLigarErro.idProcesso == $input.idProcesso && $db.whatsappLigarErro.atendido == true
          return = {type: "single"}
        } as $whatsappLigarErro3
      
        db.query whatsappLigarErro {
          where = $db.whatsappLigarErro.idProcesso == $input.idProcesso && $db.whatsappLigarErro.telefone == $input.numerowhatsapp
          return = {type: "single"}
        } as $whatsappLigarErro1
      
        conditional {
          if (($whatsappLigarErro3|is_empty) || $input.notificarsempre) {
            conditional {
              if ($whatsappLigarErro1|is_empty) {
                db.query whatsappLigarErro {
                  where = $db.whatsappLigarErro.idProcesso == $input.idProcesso
                  return = {type: "single"}
                } as $whatsappLigarErro3333
              
                conditional {
                  if ($whatsappLigarErro3333|is_empty) {
                    db.add whatsappLigarErro {
                      enforce_hidden_fields = false
                      data = {
                        created_at       : "now"
                        franqueado       : $input.idFranqueado
                        telefone         : $input.numerowhatsapp
                        nomecliente      : $input.nomecliente
                        empresanome      : $input.empresanome
                        tipoevento       : $input.tipoevento
                        zona             : $input.zona
                        local            : $input.local
                        datahorario      : $input.datahorario
                        alarm_events_id  : $input.idevento
                        ideventgo        : $input.ideventgo
                        falha            : false
                        tentativas       : 0
                        dtUltimaTentativa: now
                        exec             : true
                        atendido         : false
                        idProcesso       : $input.idProcesso
                        idDispositivo    : $input.idDispositivo
                        notificarsempre  : $input.notificarsempre
                        DispNome         : $input.DispNome
                        DispDescricao    : $input.DispDescricao
                        DispTipo         : $input.DispTipo
                        nroErr           : "0"
                      }
                    } as $whatsappLigarErroAdd
                  
                    db.add LigacaoFila {
                      enforce_hidden_fields = false
                      data = {whatsappligarerro_id: $whatsappLigarErroAdd.id}
                    } as $LigacaoFila1
                  }
                
                  else {
                    db.add whatsappLigarErro {
                      enforce_hidden_fields = false
                      data = {
                        created_at       : "now"
                        franqueado       : $input.idFranqueado
                        telefone         : $input.numerowhatsapp
                        nomecliente      : $input.nomecliente
                        empresanome      : $input.empresanome
                        tipoevento       : $input.tipoevento
                        zona             : $input.zona
                        local            : $input.local
                        datahorario      : $input.datahorario
                        alarm_events_id  : $input.idevento
                        ideventgo        : $input.ideventgo
                        falha            : true
                        tentativas       : 3333
                        dtUltimaTentativa: now
                        exec             : false
                        atendido         : false
                        idProcesso       : $input.idProcesso
                        idDispositivo    : $input.idDispositivo
                        notificarsempre  : $input.notificarsempre
                        DispNome         : $input.DispNome
                        DispDescricao    : $input.DispDescricao
                        DispTipo         : $input.DispTipo
                        nroErr           : "0"
                      }
                    } as $whatsappLigarErroAdd
                  }
                }
              }
            }
          }
        }
      }
    }
  }

  response = {dados: $retorno}
}