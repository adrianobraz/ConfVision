function fn_AlarmeRestaure {
  input {
    text idDispositivo? filters=trim
    text zonaUser? filters=trim
    int id_tblAlarm_events?
    text grupo? filters=trim
    text particao? filters=trim
  }

  stack {
    var $PodeEnviar {
      value = false
    }
  
    // PARTE 2 - Buscar pendente da mesma zona
    db.query alarmeEventos_pendente {
      where = $db.alarmeEventos_pendente.idDispositivo == $input.idDispositivo && $db.alarmeEventos_pendente.zonaUser == $input.zonaUser && $db.alarmeEventos_pendente.particao == $input.particao && $db.alarmeEventos_pendente.cancelado == false && $db.alarmeEventos_pendente.enviado == false && $db.alarmeEventos_pendente.agendadoPara > now
      sort = {alarmeEventos_pendente.created_at: "desc"}
      return = {type: "single"}
    } as $alarmeEventos_pendente1
  
    // Pendente da mesma zona sem limitar por agendamento (usado em cancelamentos)
    db.query alarmeEventos_pendente {
      where = $db.alarmeEventos_pendente.idDispositivo == $input.idDispositivo && $db.alarmeEventos_pendente.zonaUser == $input.zonaUser && $db.alarmeEventos_pendente.particao == $input.particao && $db.alarmeEventos_pendente.cancelado == false && $db.alarmeEventos_pendente.enviado == false
      sort = {alarmeEventos_pendente.created_at: "desc"}
      return = {type: "single"}
    } as $alarmeEventos_pendente_cancel
  
    // PARTE 3 - Buscar pendente de zona diferente (intrusao)
    db.query alarmeEventos_pendente {
      where = $db.alarmeEventos_pendente.idDispositivo == $input.idDispositivo && $db.alarmeEventos_pendente.zonaUser != $input.zonaUser && $db.alarmeEventos_pendente.particao == $input.particao && $db.alarmeEventos_pendente.cancelado == false && $db.alarmeEventos_pendente.enviado == false && $db.alarmeEventos_pendente.agendadoPara > now
      sort = {alarmeEventos_pendente.created_at: "desc"}
      return = {type: "single"}
    } as $alarmeIntrusao
  
    // PARTE 4 - Contar disparos nos ultimos 5 minutos
    db.direct_query {
      sql = """
          SELECT COUNT(*) AS total_disparos
          FROM x1_3
          WHERE "idDispositivo" = '{{$input.idDispositivo}}'
          AND "zonaUser"        = '{{$input.zonaUser}}'
          AND "particao"        = '{{$input.particao}}'
          AND "ctiGrupo"        = 'ALARME'
          AND "ctiDescricao"    NOT ILIKE '%NAO CADASTRADO%'
          AND (
            "id" = {{$input.id_tblAlarm_events}}
            OR "dataEntrada"::timestamp >= (NOW() - INTERVAL '5 minutes')
          )
        """
      parser = "template_engine"
      response_type = "single"
    } as $contaDisparos
  
    // PARTE 4A - Contar disparos no mesmo setor nos ultimos 3 minutos
    db.direct_query {
      sql = """
          SELECT COUNT(*) AS total_disparos_3min
          FROM x1_3
          WHERE "idDispositivo" = '{{$input.idDispositivo}}'
          AND "zonaUser"        = '{{$input.zonaUser}}'
          AND "particao"        = '{{$input.particao}}'
          AND "ctiGrupo"        = 'ALARME'
          AND "ctiDescricao"    NOT ILIKE '%NAO CADASTRADO%'
          AND (
            "id" = {{$input.id_tblAlarm_events}}
            OR "dataEntrada"::timestamp >= (NOW() - INTERVAL '3 minutes')
          )
        """
      parser = "template_engine"
      response_type = "single"
    } as $contaDisparos3min
  
    // PARTE 4B - Historico da zona
    db.direct_query {
      sql = """
          SELECT 
              COUNT(*) AS total_pares,
              ROUND(AVG(sub.segundos)::numeric, 0)::integer AS media_segundos,
              MAX(sub.segundos)::integer AS maximo_segundos
          FROM (
              SELECT 
                  EXTRACT(EPOCH FROM (e2."dataEntrada"::timestamp - e1."dataEntrada"::timestamp)) AS segundos,
                  ROW_NUMBER() OVER (
                      PARTITION BY e1."idDispositivo", e1."zonaUser" 
                      ORDER BY e1."dataEntrada" DESC
                  ) AS rn
              FROM x1_3 e1
              INNER JOIN x1_3 e2 
                  ON e1."idDispositivo" = e2."idDispositivo"
                  AND e1."zonaUser"     = e2."zonaUser"
                  AND e1."particao"     = e2."particao"
                  AND e2."ctiGrupo"     = 'RESTAURE'
                  AND e2."dataEntrada"::timestamp > e1."dataEntrada"::timestamp
                  AND EXTRACT(EPOCH FROM (
                      e2."dataEntrada"::timestamp - e1."dataEntrada"::timestamp
                  )) <= 300
              WHERE e1."ctiGrupo"    = 'ALARME'
              AND e1."idDispositivo" = '{{$input.idDispositivo}}'
              AND e1."zonaUser"      = '{{$input.zonaUser}}'
              AND e1."particao"      = '{{$input.particao}}'
              AND e1."ctiDescricao"  NOT ILIKE '%NAO CADASTRADO%'
          ) sub
          WHERE sub.rn <= 50
        """
      parser = "template_engine"
      response_type = "single"
    } as $query_janela
  
    // PARTE 4C - Verificar em quantos segundos o ULTIMO disparo restaurou
    db.direct_query {
      sql = """
          SELECT 
              e1."dataEntrada" AS ultimo_disparo,
              MIN(EXTRACT(EPOCH FROM (e2."dataEntrada"::timestamp - e1."dataEntrada"::timestamp))) AS segundos_ate_restaure
          FROM x1_3 e1
          LEFT JOIN x1_3 e2
              ON e1."idDispositivo" = e2."idDispositivo"
              AND e1."zonaUser"     = e2."zonaUser"
              AND e1."particao"     = e2."particao"
              AND e2."ctiGrupo"     = 'RESTAURE'
              AND e2."dataEntrada"::timestamp > e1."dataEntrada"::timestamp
              AND EXTRACT(EPOCH FROM (
                  e2."dataEntrada"::timestamp - e1."dataEntrada"::timestamp
              )) <= 300
          WHERE e1."ctiGrupo"    = 'ALARME'
          AND e1."idDispositivo" = '{{$input.idDispositivo}}'
          AND e1."zonaUser"      = '{{$input.zonaUser}}'
          AND e1."particao"      = '{{$input.particao}}'
          AND e1."ctiDescricao"  NOT ILIKE '%NAO CADASTRADO%'
          AND e1."dataEntrada"::timestamp >= (NOW() - INTERVAL '5 minutes')
          GROUP BY e1."dataEntrada"
          ORDER BY e1."dataEntrada" DESC
          LIMIT 1
        """
      parser = "template_engine"
      response_type = "single"
    } as $ultimo_disparo
  
    // PARTE 4C2 - Tempo do ultimo ALARME no setor (segundos ate agora)
    db.direct_query {
      sql = """
          SELECT EXTRACT(EPOCH FROM (NOW() - e1."dataEntrada"::timestamp)) AS segundos_desde_ultimo_alarme
          FROM x1_3 e1
          WHERE e1."ctiGrupo"    = 'ALARME'
          AND e1."idDispositivo" = '{{$input.idDispositivo}}'
          AND e1."particao"      = '{{$input.particao}}'
          AND e1."zonaUser"      = '{{$input.zonaUser}}'
          AND e1."ctiDescricao"  NOT ILIKE '%NAO CADASTRADO%'
          ORDER BY e1."dataEntrada" DESC
          LIMIT 1
        """
      parser = "template_engine"
      response_type = "single"
    } as $ultimo_alarme
  
    function.run WebLogarCache as $func1WebLogar
    function.run ConfMonitEquipamento_DadosEquipSetor {
      input = {
        Authorization: $func1WebLogar
        idDispositivo: $input.idDispositivo
        particao     : ""
        zonauser     : ""
      }
    } as $func5DadosEquipSetor
  
    var $x1TotalDadosEquipSetor {
      value = $func5DadosEquipSetor|count
    }
  
    // PARTE 4D - Zonas distintas com ALARME recente (numerador da regra de porcentagem)
    db.direct_query {
      sql = """
          SELECT COUNT(DISTINCT (COALESCE("particao",'') || ':' || COALESCE("zonaUser",''))) AS zonas_distintas
          FROM x1_3
          WHERE "idDispositivo" = '{{$input.idDispositivo}}'
          AND "ctiGrupo"        = 'ALARME'
          AND "ctiDescricao"    NOT ILIKE '%NAO CADASTRADO%'
          AND (
            "id" = {{$input.id_tblAlarm_events}}
            OR "dataEntrada"::timestamp >= (NOW() - INTERVAL '120 seconds')
          )
        """
      parser = "template_engine"
      response_type = "single"
    } as $evolucao_setor
  
    // PARTE 4E - Perfil recente para classificar condominio x residencia
    db.direct_query {
      sql = """
          SELECT
              COUNT(*) FILTER (WHERE "ctiGrupo" = 'ALARME')   AS total_alarmes,
              COUNT(*) FILTER (WHERE "ctiGrupo" = 'RESTAURE') AS total_restaures
          FROM x1_3
          WHERE "idDispositivo" = '{{$input.idDispositivo}}'
          AND "particao"        = '{{$input.particao}}'
          AND "ctiDescricao"    NOT ILIKE '%NAO CADASTRADO%'
          AND "dataEntrada"::timestamp >= (NOW() - INTERVAL '15 minutes')
        """
      parser = "template_engine"
      response_type = "single"
    } as $perfil_recente
  
    // PARTE 4F - Janela dinamica (aprendizado) com fallback de 30s
    var $janela {
      value = 30
    }
  
    conditional {
      if (($query_janela|is_empty) == false) {
        conditional {
          if ($query_janela.total_pares >= 5 && $query_janela.maximo_segundos <= 50) {
            var.update $janela {
              value = $query_janela.maximo_segundos
            }
          }
        
          elseif ($query_janela.total_pares >= 5 && $query_janela.maximo_segundos > 50) {
            var.update $janela {
              value = $query_janela.media_segundos
            }
          }
        }
      }
    }
  
    conditional {
      if ($janela > 50) {
        var.update $janela {
          value = 50
        }
      }
    }
  
    var $teveRestaureRapido {
      value = false
    }
  
    conditional {
      if ((($ultimo_disparo|is_empty) == false) && (($ultimo_disparo.segundos_ate_restaure|is_empty) == false) && ($ultimo_disparo.segundos_ate_restaure <= $janela)) {
        var.update $teveRestaureRapido {
          value = true
        }
      }
    }
  
    var $zonasDistintas {
      value = 0
    }
  
    conditional {
      if (($evolucao_setor|is_empty) == false) {
        var.update $zonasDistintas {
          value = $evolucao_setor.zonas_distintas
        }
      }
    }
  
    conditional {
      if ($input.grupo == "ALARME") {
        conditional {
          if ($zonasDistintas < 1) {
            var.update $zonasDistintas {
              value = 1
            }
          }
        }
      }
    }
  
    var $percentualRoubo {
      value = 0
    }
  
    var $ehRouboPorcentagem {
      value = false
    }
  
    conditional {
      if ($x1TotalDadosEquipSetor > 0) {
        var.update $percentualRoubo {
          value = $zonasDistintas
            |divide:$x1TotalDadosEquipSetor
            |multiply:100
        }
      }
    }
  
    conditional {
      if (($x1TotalDadosEquipSetor > 0) && ($percentualRoubo >= 50)) {
        var.update $ehRouboPorcentagem {
          value = true
        }
      }
    }
  
    var $temEvolucaoSetor {
      value = false
    }
  
    conditional {
      if ($zonasDistintas >= 2) {
        var.update $temEvolucaoSetor {
          value = true
        }
      }
    }
  
    var $isCondominio {
      value = false
    }
  
    conditional {
      if ((($perfil_recente|is_empty) == false) && (($perfil_recente.total_alarmes + $perfil_recente.total_restaures) >= 8)) {
        var.update $isCondominio {
          value = true
        }
      }
    }
  
    // Das 22:00 as 07:00 (America/Sao_Paulo) desativa perfil condominio
    var $horaAtual {
      value = `(now|to_timestamp|format_timestamp:"H":"America/Sao_Paulo")|to_int)`
    }
  
    conditional {
      if ($horaAtual >= 21 || $horaAtual < 7) {
        var.update $isCondominio {
          value = false
        }
      }
    }
  
    var $Decidido {
      value = false
    }
  
    // PARTE 5 - Logica condicional principal
    conditional {
      // Quando o evento recebido for RESTAURE, cancela pendente (exceto evolucao de setor ou >= 50%)
      if ($input.grupo == "RESTAURE") {
        conditional {
          if (($temEvolucaoSetor == false) && ($ehRouboPorcentagem == false)) {
            conditional {
              if (($alarmeEventos_pendente_cancel|is_empty) == false) {
                db.edit alarmeEventos_pendente {
                  field_name = "id"
                  field_value = $alarmeEventos_pendente_cancel.id
                  enforce_hidden_fields = false
                  data = {cancelado: true}
                } as $alarmeEventos_pendente4
              }
            }
          }
        }
      }
    
      else {
        // 3) Evolucao de setor manda (ignora demais regras de supressao)
        conditional {
          if ($temEvolucaoSetor) {
            conditional {
              if ($alarmeEventos_pendente1|is_empty) {
                db.add alarmeEventos_pendente {
                  enforce_hidden_fields = false
                  data = {
                    idDispositivo     : $input.idDispositivo
                    zonaUser          : $input.zonaUser
                    id_tblAlarm_events: $input.id_tblAlarm_events
                    agendadoPara      : now
                    janelaSegundos    : 0
                    cancelado         : false
                    enviado           : false
                    grupo             : $input.grupo
                    particao          : $input.particao
                  }
                } as $alarmeEventos_pendente2
              }
            }
          
            var.update $PodeEnviar {
              value = true
            }
          
            var.update $Decidido {
              value = true
            }
          }
        }
      
        // 3B) >= 50% das zonas - manda imediato
        conditional {
          if (($Decidido == false) && $ehRouboPorcentagem) {
            conditional {
              if ($alarmeEventos_pendente1|is_empty) {
                db.add alarmeEventos_pendente {
                  enforce_hidden_fields = false
                  data = {
                    idDispositivo     : $input.idDispositivo
                    zonaUser          : $input.zonaUser
                    id_tblAlarm_events: $input.id_tblAlarm_events
                    agendadoPara      : now
                    janelaSegundos    : 0
                    cancelado         : false
                    enviado           : false
                    grupo             : $input.grupo
                    particao          : $input.particao
                  }
                } as $alarmeEventos_pendente_pct_imediato
              }
            }
          
            var.update $PodeEnviar {
              value = true
            }
          
            var.update $Decidido {
              value = true
            }
          }
        }
      
        // 2) Se alarme ja passou de 50s, manda mesmo que RESTAURE chegue depois
        conditional {
          if (($Decidido == false) && (($ultimo_alarme|is_empty) == false) && (($ultimo_alarme.segundos_desde_ultimo_alarme|is_empty) == false) && ($ultimo_alarme.segundos_desde_ultimo_alarme > 50)) {
            conditional {
              if ($alarmeEventos_pendente1|is_empty) {
                db.add alarmeEventos_pendente {
                  enforce_hidden_fields = false
                  data = {
                    idDispositivo     : $input.idDispositivo
                    zonaUser          : $input.zonaUser
                    id_tblAlarm_events: $input.id_tblAlarm_events
                    agendadoPara      : now
                    janelaSegundos    : 0
                    cancelado         : false
                    enviado           : false
                    grupo             : $input.grupo
                    particao          : $input.particao
                  }
                } as $alarmeEventos_pendente3
              }
            }
          
            var.update $PodeEnviar {
              value = true
            }
          
            var.update $Decidido {
              value = true
            }
          }
        }
      
        // 5) Condominio nao manda
        conditional {
          if (($Decidido == false) && $isCondominio) {
            conditional {
              if (($alarmeEventos_pendente_cancel|is_empty) == false) {
                db.edit alarmeEventos_pendente {
                  field_name = "id"
                  field_value = $alarmeEventos_pendente_cancel.id
                  enforce_hidden_fields = false
                  data = {cancelado: true}
                } as $alarmeEventos_pendente_condominio
              }
            }
          
            // C6: trilha de auditoria da supressao (cancelado=true nao entra em nenhuma fila)
            db.add alarmeEventos_pendente {
              data = {
                idDispositivo     : $input.idDispositivo
                zonaUser          : $input.zonaUser
                id_tblAlarm_events: $input.id_tblAlarm_events
                agendadoPara      : now
                janelaSegundos    : 0
                cancelado         : true
                enviado           : false
                grupo             : "SUPRIMIDO-CONDOMINIO"
                particao          : $input.particao
              }
            } as $alarmeEventos_pendente_log_condominio
          
            var.update $Decidido {
              value = true
            }
          }
        }
      
        // 4) Residencia: mesmo setor em menos de 3 min e >2x manda
        conditional {
          if (($Decidido == false) && ($contaDisparos3min.total_disparos_3min > 2)) {
            conditional {
              if ($alarmeEventos_pendente1|is_empty) {
                db.add alarmeEventos_pendente {
                  enforce_hidden_fields = false
                  data = {
                    idDispositivo     : $input.idDispositivo
                    zonaUser          : $input.zonaUser
                    id_tblAlarm_events: $input.id_tblAlarm_events
                    agendadoPara      : now
                    janelaSegundos    : 0
                    cancelado         : false
                    enviado           : false
                    grupo             : $input.grupo
                    particao          : $input.particao
                  }
                } as $alarmeEventos_pendente4
              }
            }
          
            var.update $PodeEnviar {
              value = true
            }
          
            var.update $Decidido {
              value = true
            }
          }
        }
      
        // 1) Janela aprendida (fallback 30s), maximo 50s
        conditional {
          if ($Decidido == false) {
            conditional {
              if ($alarmeEventos_pendente1|is_empty) {
                db.add alarmeEventos_pendente {
                  enforce_hidden_fields = false
                  data = {
                    idDispositivo     : $input.idDispositivo
                    zonaUser          : $input.zonaUser
                    id_tblAlarm_events: $input.id_tblAlarm_events
                    agendadoPara      : now|add_secs_to_timestamp:$janela
                    janelaSegundos    : $janela
                    cancelado         : false
                    enviado           : false
                    grupo             : $input.grupo
                    particao          : $input.particao
                  }
                } as $alarmeEventos_pendente5
              }
            }
          
            var.update $PodeEnviar {
              value = true
            }
          }
        }
      }
    }
  
    // Regra de porcentagem: fallback se PodeEnviar=false (ex: condominio)
    conditional {
      if (($PodeEnviar == false) && ($input.grupo != "RESTAURE") && $ehRouboPorcentagem) {
        conditional {
          if ($alarmeEventos_pendente1|is_empty) {
            db.add alarmeEventos_pendente {
              enforce_hidden_fields = false
              data = {
                idDispositivo     : $input.idDispositivo
                zonaUser          : $input.zonaUser
                id_tblAlarm_events: $input.id_tblAlarm_events
                agendadoPara      : now
                janelaSegundos    : 0
                cancelado         : false
                enviado           : false
                grupo             : $input.grupo
                particao          : $input.particao
              }
            } as $alarmeEventos_pendente_pct
          }
        }
      
        var.update $PodeEnviar {
          value = true
        }
      }
    }
  }

  response = $PodeEnviar
}