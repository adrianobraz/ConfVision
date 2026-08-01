function fn_RegrasEnvioEvento {
  input {
    text ctiDescricao? filters=trim
    text ctiGrupo? filters=trim
    text codigo? filters=trim
    text idDispositivo? filters=trim
    text zonaUser? filters=trim
    int id_tblAlarm_events?
    text particao? filters=trim
    text idCliente? filters=trim
    text nomeCliente? filters=trim
    text emailCliente? filters=trim
    text idFranqueado? filters=trim
    bool GRADEHORARIO?
    text idProcesso? filters=trim
  }

  stack {
    var $PodeEnviar {
      value = false
    }
  
    conditional {
      // GRADE DE HORARIO
      if ($input.codigo == "3M02" || $input.codigo == "1M01") {
        conditional {
          if ($input.GRADEHORARIO) {
            conditional {
              if ($input.codigo == "3M02") {
                var $motivo {
                  value = "Sistema não Armado"
                }
              }
            
              else {
                var $motivo {
                  value = "Desarme fora do Horário"
                }
              }
            }
          
            function.run ProcessoEnd {
              runtime_mode = "async-shared"
              input = {
                idUsuario  : "0ROBOAUTO"
                Descricao  : "[AUTO] " ~ $motivo ~ " [" ~ $input.idProcesso ~ "]"
                UsuarioNome: "ROBO AUTO"
                idProcesso : $input.idProcesso
                telefone   : "ROBO AUTO"
                token      : "ROBO AUTO"
                ip         : ""
                ipcidade   : ""
                ipestado   : ""
                ipcep      : ""
                geolat     : 0
                geolon     : 0
                georua     : ""
                geonumero  : ""
                geobairro  : ""
                geocidade  : ""
                geoestado  : ""
                device     : ""
                browser    : ""
                sitema     : ""
                timezone   : ""
                fingerprint: "ROBO AUTO"
              }
            } as $procEnd
          
            var.update $PodeEnviar {
              value = true
            }
          }
        }
      }
    
      // RESTAURE
      elseif (($input.ctiDescricao|contains:"RESTAUR") || $input.ctiGrupo == "RESTAURE") {
        conditional {
          if ($input.ctiGrupo == "RESTAURE") {
            function.run fn_AlarmeRestaure {
              input = {
                idDispositivo     : $input.idDispositivo
                zonaUser          : $input.zonaUser
                id_tblAlarm_events: $input.id_tblAlarm_events
                grupo             : $input.ctiGrupo
                particao          : $input.particao
              }
            } as $lAlarmeRestaure
          }
        }
      }
    
      // SETUP
      elseif ($input.ctiGrupo == "SETUP") {
      }
    
      // NAO CADASTRADO - ALARME
      elseif ($input.ctiDescricao == "NAO CADASTRADO" && $input.ctiGrupo == "ALARME") {
        db.add contactId_Erro {
          enforce_hidden_fields = false
          data = {codigo: $input.codigo}
        } as $contactId_Erro1
      }
    
      // PANICO - EMERGENCIA - MEDICO
      elseif ($input.ctiGrupo == "PANICO" || $input.ctiGrupo == "EMERGENCIA" || $input.ctiGrupo == "MEDICO") {
        var.update $PodeEnviar {
          value = true
        }
      }
    
      // ALARME
      elseif ($input.ctiGrupo == "ALARME") {
        function.run fn_AlarmeRestaure {
          input = {
            idDispositivo     : $input.idDispositivo
            zonaUser          : $input.zonaUser
            id_tblAlarm_events: $input.id_tblAlarm_events
            grupo             : $input.ctiGrupo
            particao          : $input.particao
          }
        } as $lAlarmeRestaure
      
        var.update $PodeEnviar {
          value = $lAlarmeRestaure
        }
      
        conditional {
          if ($PodeEnviar) {
            db.query status_link {
              where = $db.status_link.idDispositivo == $input.idDispositivo && $db.status_link.particao == $input.particao && $db.status_link.zonauser == $input.zonaUser && $db.status_link.expiresAt > now && $db.status_link.grupo == "ALARME"
              sort = {status_link.id: "desc", status_link.created_at: "desc"}
              return = {type: "single"}
            } as $status_link2
          
            conditional {
              if ($status_link2|is_empty) {
                function.run statuslink_reUUID {
                  input = {dispositivo: $input.idDispositivo}
                } as $func1
              
                conditional {
                  if ($func1.dados|is_empty) {
                    db.add status_link {
                      enforce_hidden_fields = false
                      data = {
                        token          : `|uuid`
                        idDispositivo  : $input.idDispositivo
                        expiresAt      : now|add_secs_to_timestamp:300
                        alarm_events_id: $input.id_tblAlarm_events
                        particao       : $input.particao
                        zonauser       : $input.zonaUser
                        grupo          : "ALARME"
                      }
                    } as $status_link1
                  }
                
                  else {
                    db.edit status_link {
                      field_name = "id"
                      field_value = $func1.dados.id
                      enforce_hidden_fields = false
                      data = {
                        created_at     : now
                        expiresAt      : now|add_secs_to_timestamp:7200
                        alarm_events_id: $input.id_tblAlarm_events
                        particao       : $input.particao
                        zonauser       : $input.zonaUser
                        grupo          : "ALARME"
                      }
                    } as $status_link4
                  }
                }
              }
            
              else {
                db.edit status_link {
                  field_name = "id"
                  field_value = $status_link2.id
                  enforce_hidden_fields = false
                  data = {expiresAt: now|add_secs_to_timestamp:300}
                } as $status_link3
              }
            }
          }
        }
      }
    
      // ARME - DESARME
      elseif ($input.ctiGrupo == "ARME" || $input.ctiGrupo == "DESARME") {
        db.query status_link {
          where = $db.status_link.idDispositivo == $input.idDispositivo && $db.status_link.particao == $input.particao && $db.status_link.zonauser == $input.zonaUser && $db.status_link.expiresAt > now && $db.status_link.grupo == "ARME"
          sort = {status_link.id: "desc", status_link.created_at: "desc"}
          return = {type: "single"}
        } as $status_link2
      
        conditional {
          if ($status_link2|is_empty) {
            function.run statuslink_reUUID {
              input = {dispositivo: $input.idDispositivo}
            } as $func1
          
            conditional {
              if ($func1.dados|is_empty) {
                db.add status_link {
                  enforce_hidden_fields = false
                  data = {
                    token          : `|uuid`
                    idDispositivo  : $input.idDispositivo
                    expiresAt      : now|add_secs_to_timestamp:7200
                    alarm_events_id: $input.id_tblAlarm_events
                    particao       : $input.particao
                    zonauser       : $input.zonaUser
                    grupo          : "ARME"
                  }
                } as $status_link1
              }
            
              else {
                db.edit status_link {
                  field_name = "id"
                  field_value = $func1.dados.id
                  enforce_hidden_fields = false
                  data = {
                    created_at     : now
                    expiresAt      : now|add_secs_to_timestamp:7200
                    alarm_events_id: $input.id_tblAlarm_events
                    particao       : $input.particao
                    zonauser       : $input.zonaUser
                    grupo          : "ARME"
                  }
                } as $status_link4
              }
            }
          
            var.update $PodeEnviar {
              value = true
            }
          }
        }
      }
    
      // FALHAS - GERAL
      elseif ($input.ctiGrupo == "FALHAS" || $input.ctiGrupo == "GERAL") {
        var $cooldownHoras {
          value = 43200
        }
      
        conditional {
          if ($input.ctiGrupo == "GERAL") {
            var.update $cooldownHoras {
              value = 14400
            }
          }
        }
      
        db.query whatsapp_cooldown {
          where = $db.whatsapp_cooldown.idDispositivo == $input.idDispositivo && $db.whatsapp_cooldown.particao == $input.particao && $db.whatsapp_cooldown.zonaUser == $input.zonaUser && $db.whatsapp_cooldown.ctiDescricao == $input.ctiDescricao
          return = {type: "single"}
        } as $whatsapp_cooldown1
      
        conditional {
          if ($whatsapp_cooldown1|is_empty) {
            db.add whatsapp_cooldown {
              enforce_hidden_fields = false
              data = {
                idDispositivo: $input.idDispositivo
                ctiGrupo     : $input.ctiGrupo
                ctiDescricao : $input.ctiDescricao
                totalEventos : 1
                ultimoEnvio  : now|add_secs_to_timestamp:$cooldownHoras
                zonaUser     : $input.zonaUser
                particao     : $input.particao
                idCliente    : $input.idCliente
                nomeCliente  : $input.nomeCliente
                emailCliente : $input.emailCliente
                idFranqueado : $input.idFranqueado
              }
            } as $whatsapp_cooldown2
          }
        
          else {
            db.edit whatsapp_cooldown {
              field_name = "id"
              field_value = $whatsapp_cooldown1.id
              enforce_hidden_fields = false
              data = {totalEventos: $whatsapp_cooldown1.totalEventos + 1}
            } as $whatsapp_cooldown3
          }
        }
      
        conditional {
          if ($input.codigo == `"1301"` || $input.codigo == `"1361"`) {
            var.update $PodeEnviar {
              value = true
            }
          }
        }
      }
    
      else {
        var.update $PodeEnviar {
          value = true
        }
      }
    }
  }

  response = $PodeEnviar
}