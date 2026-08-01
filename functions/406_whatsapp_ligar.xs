function WhatsappLigar {
  input {
    text idfranqueado? filters=trim
    text telefone? filters=trim
    text nomecliente? filters=trim
    text empresanome? filters=trim
    text tipoevento? filters=trim
    text zona? filters=trim
    text local? filters=trim
    int idevento?
    text datahorario? filters=trim
    text ideventogo? filters=trim
    text idDispositivo? filters=trim
    text idProcesso? filters=trim
    bool notificarsempre?
    text DispNome? filters=trim
    text DispDescricao? filters=trim
    text DispTipo? filters=trim
  }

  stack {
    db.query WhatsappFranqueadoNroTelefone {
      where = $db.WhatsappFranqueadoNroTelefone.franqueado == $input.idfranqueado && $db.WhatsappFranqueadoNroTelefone.ModuloLigar == true
      return = {
        type : "aggregate"
        group: {Grupo: $db.WhatsappFranqueadoNroTelefone.Grupo}
      }
    } as $WhatsappFranqueadoNroTelefone1
  
    var $FranqueadoID {
      value = $input.idfranqueado
    }
  
    conditional {
      if ($WhatsappFranqueadoNroTelefone1|is_empty) {
        db.query WhatsappFranqueadoNroTelefone {
          where = $db.WhatsappFranqueadoNroTelefone.franqueado == `"019c2bb9-5a02-74b9-812a-a20cb4c477ae"` && $db.WhatsappFranqueadoNroTelefone.ModuloLigar == true
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
          where = $db.WhatsappFranqueadoNroTelefone.franqueado == $FranqueadoID && $db.WhatsappFranqueadoNroTelefone.ModuloLigar == true && $db.WhatsappFranqueadoNroTelefone.Grupo == $idWhatsFranq
          sort = {WhatsappFranqueadoNroTelefone.id: "rand"}
          return = {type: "single"}
        } as $WhatsappFranqueadoNroTelefone1
      }
    
      else {
        conditional {
          if (($WhatsappFranqueadoNroTelefone1|is_empty) == false) {
            db.query WhatsappFranqueadoNroTelefone {
              where = $db.WhatsappFranqueadoNroTelefone.franqueado == $FranqueadoID && $db.WhatsappFranqueadoNroTelefone.ModuloLigar == true && $db.WhatsappFranqueadoNroTelefone.Grupo == $WhatsappFranqueadoNroTelefone1[0].Grupo
              sort = {WhatsappFranqueadoNroTelefone.id: "rand"}
              return = {type: "single"}
            } as $WhatsappFranqueadoNroTelefone1
          }
        }
      }
    }
  
    conditional {
      if ($WhatsappFranqueadoNroTelefone1|is_empty) {
        return {
          value = ""
        }
      }
    }
  
    conditional {
      if ($WhatsappFranqueadoNroTelefone1.tipoApi == "E") {
        api.request {
          url = "https://monitoramento.uazapi.com/chat/details"
          method = "POST"
          params = {}
            |set:"number":$input.telefone
            |set:"preview":false
          headers = []
            |push:"Accept: application/json"
            |push:"Content-Type: application/json"
            |push:`"token:"|concat:$var.WhatsappFranqueadoNroTelefone1.txtApiToken`
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
                    |set:"number":$input.telefone
                    |set:"name":($Nome|concat:$input.empresanome:" ")
                  headers = []
                    |push:"Accept: application/json"
                    |push:"Content-Type: application/json"
                    |push:`"token:"|concat:$var.WhatsappFranqueadoNroTelefone1.txtApiToken`
                }
              }
            
              else {
                api.request {
                  url = "https://monitoramento.uazapi.com/contact/add"
                  method = "POST"
                  params = {}
                    |set:"number":$input.telefone
                    |set:"name":($input.nomecliente|concat:$input.empresanome:" ")
                  headers = []
                    |push:"Accept: application/json"
                    |push:"Content-Type: application/json"
                    |push:`"token:"|concat:$var.WhatsappFranqueadoNroTelefone1.txtApiToken`
                }
              }
            }
          }
        }
      }
    }
  
    var $x2 {
      value = false
    }
  
    var $loops {
      value = 0
    }
  
    db.query whatsappcontroleligacao {
      where = $db.whatsappcontroleligacao.update < (`now`|timestamp_add_minutes:-5)
      return = {type: "list"}
    } as $whatsappcontroleligacao1
  
    foreach ($whatsappcontroleligacao1) {
      each as $item {
        db.del whatsappcontroleligacao {
          field_name = "id"
          field_value = $item.id
        }
      }
    }
  
    while (true) {
      each {
        var.update $loops {
          value = $loops + 1
        }
      
        var $x1 {
          value = 0
        }
      
        db.query whatsappcontroleligacao {
          where = $db.whatsappcontroleligacao.franqueado == $WhatsappFranqueadoNroTelefone1.franqueado && $db.whatsappcontroleligacao.telefone == $WhatsappFranqueadoNroTelefone1.telefone
          return = {type: "list"}
        } as $whatsappcontroleligacao1
      
        conditional {
          if ($whatsappcontroleligacao1|is_empty) {
          }
        
          else {
            var.update $x1 {
              value = $whatsappcontroleligacao1.franqueado|count
            }
          }
        }
      
        conditional {
          if ($x1 < $WhatsappFranqueadoNroTelefone1.ligacao) {
            var.update $x2 {
              value = true
            }
          
            db.add whatsappcontroleligacao {
              enforce_hidden_fields = false
              data = {
                created_at: "now"
                franqueado: $WhatsappFranqueadoNroTelefone1.franqueado
                telefone  : $WhatsappFranqueadoNroTelefone1.telefone
                update    : "now"
              }
            } as $whatsappcontroleligacao2
          
            db.query whatsappLigarErro {
              where = $db.whatsappLigarErro.ideventgo == $input.ideventogo && $db.whatsappLigarErro.telefone == $input.telefone
              return = {type: "single"}
            } as $whatsappLigarErro2
          
            var $nTentativas {
              value = $whatsappLigarErro2.tentativas
            }
          
            var $nId {
              value = $whatsappLigarErro2.id
            }
          
            conditional {
              if ($WhatsappFranqueadoNroTelefone1.tipoApi == "E" || $WhatsappFranqueadoNroTelefone1.tipoApi == "P") {
                try_catch {
                  try {
                    api.request {
                      url = "https://xpcy-oyme-lno7.b2.xano.io/api:G5ijUd6Z/fnDispararLigacaoEleven"
                      method = "POST"
                      params = {}
                        |set:"telefone":$input.telefone
                        |set:"ideventgo":$input.ideventogo
                        |set:"nomecliente":$input.nomecliente
                        |set:"empresanome":$input.empresanome
                        |set:"tipoevento":$input.tipoevento
                        |set:"DispTipo":$input.DispTipo
                        |set:"DispNome":$input.DispNome
                        |set:"DispDescricao":$input.DispDescricao
                        |set:"datahorario":$input.datahorario
                        |set:"eleven_agent":$WhatsappFranqueadoNroTelefone1.eleven_agent
                        |set:"eleven_phone":$WhatsappFranqueadoNroTelefone1.eleven_phone
                      headers = []
                        |push:"Content-Type: application/json"
                      timeout = 10
                    } as $api1
                  }
                
                  catch {
                    var $erroDisparo {
                      value = true
                    }
                  }
                }
              }
            
              else {
                // outra plataforma de ligacao futura
              }
            }
          
            db.edit whatsappLigarErro {
              field_name = "id"
              field_value = $nId
              enforce_hidden_fields = false
              data = {
                franqueado       : $input.idfranqueado
                telefone         : $input.telefone
                nomecliente      : $input.nomecliente
                empresanome      : $input.empresanome
                tipoevento       : $input.tipoevento
                zona             : $input.zona
                local            : $input.local
                datahorario      : $input.datahorario
                alarm_events_id  : $input.idevento
                ideventgo        : $input.ideventogo
                tentativas       : $nTentativas
                dtUltimaTentativa: now
                exec             : false
              }
            } as $whatsappLigarErro3
          
            conditional {
              if ($nTentativas == 100) {
                var.update $nTentativas {
                  value = 1
                }
              }
            
              elseif ($nTentativas == 200) {
                var.update $nTentativas {
                  value = 2
                }
              }
            
              elseif ($nTentativas == 300) {
                var.update $nTentativas {
                  value = 3
                }
              }
            }
          
            conditional {
              if ($nTentativas >= 2 && $nTentativas < 200) {
                db.edit whatsappLigarErro {
                  field_name = "id"
                  field_value = $nId
                  enforce_hidden_fields = false
                  data = {
                    franqueado       : $input.idfranqueado
                    telefone         : $input.telefone
                    nomecliente      : $input.nomecliente
                    empresanome      : $input.empresanome
                    tipoevento       : $input.tipoevento
                    zona             : $input.zona
                    local            : $input.local
                    datahorario      : $input.datahorario
                    alarm_events_id  : $input.idevento
                    ideventgo        : $input.ideventogo
                    tentativas       : 99
                    dtUltimaTentativa: now
                    exec             : false
                  }
                } as $whatsappLigarErro3
              }
            }
          
            db.del whatsappcontroleligacao {
              field_name = "id"
              field_value = $whatsappcontroleligacao2.id
            }
          
            break
          }
        }
      
        conditional {
          if ($x2) {
            break
          }
        }
      
        conditional {
          if ($loops >= 50) {
            break
          }
        }
      }
    }
  
    // Canal ocupado apos 50 tentativas: marca falha para o retry (ligacao060) nao perder a ligacao
    conditional {
      if ($x2 == false) {
        db.query whatsappLigarErro {
          where = $db.whatsappLigarErro.ideventgo == $input.ideventogo && $db.whatsappLigarErro.telefone == $input.telefone
          sort = {whatsappLigarErro.id: "desc"}
          return = {type: "single"}
        } as $whatsappLigarErroPerdido
      
        conditional {
          if (($whatsappLigarErroPerdido|is_empty) == false) {
            conditional {
              if ($whatsappLigarErroPerdido.atendido == false) {
                db.edit whatsappLigarErro {
                  field_name = "id"
                  field_value = $whatsappLigarErroPerdido.id
                  data = {
                    falha            : true
                    dtUltimaTentativa: now
                    exec             : false
                  }
                } as $whatsappLigarErroPerdido2
              }
            }
          }
        }
      }
    }
  }

  response = $WhatsappFranqueadoNroTelefone1
}