function bot_finalizaeventoauto {
  input {
    int idEvento?
  }

  stack {
    precondition (($input.idEvento|is_empty) == false) {
      error = "idEvento obrigatório"
      payload = false
    }
  
    db.get alarm_events {
      field_name = "id"
      field_value = $input.idEvento
    } as $evt
  
    precondition (($evt.id|is_empty) == false) {
      error = "Evento não encontrado"
      payload = false
    }
  
    precondition (($evt.idProcesso|is_empty) == false) {
      error = "Evento sem idProcesso"
      payload = false
    }
  
    var $perfilLocal {
      value = "NORMAL"
    }
  
    var $motivo {
      value = "SEM_REGRA"
    }
  
    var $acao {
      value = "NAO_FINALIZOU"
    }
  
    var $mostrarTela {
      value = true
    }
  
    var $finalizarAgora {
      value = false
    }
  
    var $retornoProcessoEnd {
      value = {}
    }
  
    var $gravouLog {
      value = false
    }
  
    var $temFalhas {
      value = false
    }
  
    var $temAlarme {
      value = false
    }
  
    var $temDesarme {
      value = false
    }
  
    var $temArme {
      value = false
    }
  
    var $temRestaure {
      value = false
    }
  
    var $temParAlarmeRestaure50 {
      value = false
    }
  
    var $qtdMonitorados {
      value = 0
    }
  
    var $qtdRestaure {
      value = 0
    }
  
    var $pendingAlarmeTs {
      value = 0
    }
  
    var $pendingAlarmeParticao {
      value = ""
    }
  
    var $pendingAlarmeZona {
      value = ""
    }
  
    var $pendingRestaureTs {
      value = 0
    }
  
    var $pendingRestaureParticao {
      value = ""
    }
  
    var $pendingRestaureZona {
      value = ""
    }
  
    var $nowTs {
      value = "now"|to_timestamp
    }
  
    var $qtdCiclos3mDiffProc {
      value = 0
    }
  
    function.run WebLogarCache as $func1
    api.request {
      url = "http://185.130.61.4:2010/v4/terminal/getDadosProcessoById"
      method = "POST"
      params = {}|set:"idProcesso":$evt.idProcesso
      headers = []
        |push:"Authorization: Bearer " ~ $func1
        |push:"Content-Type: application/json"
    } as $api1
  
    conditional {
      if (($api1.response.result.dados.dataAtenFim|is_empty) == false && $api1.response.result.dados.dataAtenFim != "01/01/0001 00:00:00") {
        var.update $motivo {
          value = "PROCESSO_JA_FINALIZADO_NO_INICIO"
        }
      
        var.update $acao {
          value = "JA_FINALIZADO"
        }
      
        var.update $mostrarTela {
          value = false
        }
      }
    
      else {
        db.direct_query {
          sql = """
              WITH ultimos AS (
                SELECT
                  e."idProcesso",
                  e."ctiGrupo",
                  CASE
                    WHEN (e."created_at"::bigint) > 9999999999 THEN (e."created_at"::bigint / 1000)
                    ELSE (e."created_at"::bigint)
                  END AS ts_s
                FROM x1_3 e
                WHERE e."idDispositivo" = '{{$evt.idDispositivo}}'
                ORDER BY e."created_at"::bigint DESC
                LIMIT 150
              ),
              hist AS (
                SELECT
                  COUNT(*) FILTER (WHERE "ctiGrupo" = 'ALARME') AS qtd_alarme_hist,
                  COUNT(*) FILTER (WHERE "ctiGrupo" = 'RESTAURE') AS qtd_restaure_hist
                FROM ultimos
              ),
              janela3m AS (
                SELECT *
                FROM ultimos
                WHERE ts_s >= (EXTRACT(EPOCH FROM NOW())::bigint - 180)
              ),
              proc_flags AS (
                SELECT
                  "idProcesso",
                  MAX(CASE WHEN "ctiGrupo" = 'ALARME' THEN 1 ELSE 0 END) AS tem_alarme,
                  MAX(CASE WHEN "ctiGrupo" = 'RESTAURE' THEN 1 ELSE 0 END) AS tem_restaure
                FROM janela3m
                GROUP BY "idProcesso"
              ),
              ciclos3m AS (
                SELECT COUNT(*)::integer AS qtd_ciclos_3m_diff_proc
                FROM proc_flags
                WHERE tem_alarme = 1 AND tem_restaure = 1
              )
              SELECT
                h.qtd_alarme_hist,
                h.qtd_restaure_hist,
                c.qtd_ciclos_3m_diff_proc,
                CASE
                  WHEN h.qtd_alarme_hist >= 40
                   AND h.qtd_restaure_hist >= (h.qtd_alarme_hist * 0.7)
                  THEN 'PORTAO'
                  ELSE 'NORMAL'
                END AS perfil_local
              FROM hist h
              CROSS JOIN ciclos3m c
            """
          parser = "template_engine"
          response_type = "single"
        } as $histAgg
      
        var.update $perfilLocal {
          value = $histAgg.perfil_local
        }
      
        var.update $qtdCiclos3mDiffProc {
          value = $histAgg.qtd_ciclos_3m_diff_proc
        }
      
        db.query alarm_events {
          where = $db.alarm_events.idProcesso == $evt.idProcesso
          sort = {alarm_events.created_at: "asc"}
          return = {type: "list"}
        } as $eventos
      
        precondition (($eventos|is_empty) == false) {
          error = "Sem eventos para análise"
          payload = false
        }
      
        foreach ($eventos) {
          each as $e {
            var $grupo {
              value = `$e.ctiGrupo|trim|upper`
            }
          
            conditional {
              if (($e.ctiDescricao|trim|upper)|contains:"RESTAUR") {
                var.update $grupo {
                  value = "RESTAURE"
                }
              }
            }
          
            var $tsEvento {
              value = $e.created_at + 0
            }
          
            conditional {
              if ($tsEvento > 9999999999) {
                var.update $tsEvento {
                  value = $tsEvento / 1000
                }
              }
            }
          
            conditional {
              if ($grupo == "ALARME" || $grupo == "ARME" || $grupo == "DESARME" || $grupo == "FALHAS" || $grupo == "RESTAURE") {
                var.update $qtdMonitorados {
                  value = $qtdMonitorados + 1
                }
              }
            }
          
            conditional {
              if ($grupo == "FALHAS") {
                var.update $temFalhas {
                  value = true
                }
              }
            }
          
            conditional {
              if ($grupo == "DESARME") {
                var.update $temDesarme {
                  value = true
                }
              }
            }
          
            conditional {
              if ($grupo == "ARME") {
                var.update $temArme {
                  value = true
                }
              }
            }
          
            conditional {
              if ($grupo == "ALARME") {
                var.update $temAlarme {
                  value = true
                }
              
                var $setorAlarmePart {
                  value = $e.particao
                }
              
                var $setorAlarmeZona {
                  value = $e.zonaUser
                }
              
                conditional {
                  if ($pendingRestaureTs > 0) {
                    var $mesmoSetorLagAR {
                      value = false
                    }
                  
                    conditional {
                      if (($setorAlarmePart == $pendingRestaureParticao) && ($setorAlarmeZona == $pendingRestaureZona)) {
                        var.update $mesmoSetorLagAR {
                          value = true
                        }
                      }
                    
                      elseif (($setorAlarmePart == $pendingRestaureParticao) && (($setorAlarmeZona|trim) != "") && (($pendingRestaureZona|trim) != "") && (((($setorAlarmeZona + 0) - ($pendingRestaureZona + 0)) == 10) || ((($setorAlarmeZona + 0) - ($pendingRestaureZona + 0)) == -10))) {
                        var.update $mesmoSetorLagAR {
                          value = true
                        }
                      }
                    }
                  
                    conditional {
                      if ($mesmoSetorLagAR) {
                        var $diffLagAR {
                          value = $tsEvento - $pendingRestaureTs
                        }
                      
                        conditional {
                          if (($diffLagAR >= 0) && ($diffLagAR <= 50)) {
                            var.update $temParAlarmeRestaure50 {
                              value = true
                            }
                          
                            var.update $pendingRestaureTs {
                              value = 0
                            }
                          
                            var.update $pendingRestaureParticao {
                              value = ""
                            }
                          
                            var.update $pendingRestaureZona {
                              value = ""
                            }
                          }
                        }
                      }
                    }
                  }
                }
              
                var.update $pendingAlarmeTs {
                  value = $tsEvento
                }
              
                var.update $pendingAlarmeParticao {
                  value = $e.particao
                }
              
                var.update $pendingAlarmeZona {
                  value = $e.zonaUser
                }
              }
            }
          
            conditional {
              if ($grupo == "RESTAURE") {
                var.update $temRestaure {
                  value = true
                }
              
                var.update $qtdRestaure {
                  value = $qtdRestaure + 1
                }
              
                conditional {
                  if ($pendingAlarmeTs > 0) {
                    var $mesmoSetorAR {
                      value = false
                    }
                  
                    conditional {
                      if (($e.particao == $pendingAlarmeParticao) && ($e.zonaUser == $pendingAlarmeZona)) {
                        var.update $mesmoSetorAR {
                          value = true
                        }
                      }
                    
                      elseif (($e.particao == $pendingAlarmeParticao) && (($e.zonaUser|trim) != "") && (($pendingAlarmeZona|trim) != "") && (((($e.zonaUser + 0) - ($pendingAlarmeZona + 0)) == 10) || ((($e.zonaUser + 0) - ($pendingAlarmeZona + 0)) == -10))) {
                        var.update $mesmoSetorAR {
                          value = true
                        }
                      }
                    }
                  
                    conditional {
                      if ($mesmoSetorAR) {
                        var $diffAR {
                          value = $tsEvento - $pendingAlarmeTs
                        }
                      
                        conditional {
                          if (($diffAR >= 0) && ($diffAR <= 50)) {
                            var.update $temParAlarmeRestaure50 {
                              value = true
                            }
                          
                            var.update $pendingAlarmeTs {
                              value = 0
                            }
                          
                            var.update $pendingAlarmeParticao {
                              value = ""
                            }
                          
                            var.update $pendingAlarmeZona {
                              value = ""
                            }
                          }
                        }
                      }
                    }
                  }
                }
              
                var.update $pendingRestaureTs {
                  value = $tsEvento
                }
              
                var.update $pendingRestaureParticao {
                  value = $e.particao
                }
              
                var.update $pendingRestaureZona {
                  value = $e.zonaUser
                }
              }
            }
          }
        }
      
        conditional {
          if ($temDesarme || $temArme) {
            var.update $motivo {
              value = "FINALIZA_ARME_OU_DESARME_COM_ALARME"
            }
          
            var.update $finalizarAgora {
              value = true
            }
          }
        
          elseif ($temAlarme && $temParAlarmeRestaure50) {
            conditional {
              if (($perfilLocal == "NORMAL") && ($qtdCiclos3mDiffProc > 1)) {
                var.update $motivo {
                  value = "NAO_FINALIZA_NORMAL_REPETICAO_DIFF_PROCESSOS_3M"
                }
              
                var.update $acao {
                  value = "NAO_FINALIZOU"
                }
              
                var.update $mostrarTela {
                  value = true
                }
              }
            
              else {
                var.update $motivo {
                  value = "FINALIZA_ALARME_RESTAURE_50S"
                }
              
                var.update $finalizarAgora {
                  value = true
                }
              }
            }
          }
        
          elseif ($temAlarme && ($temParAlarmeRestaure50 == false) && $temRestaure) {
            var.update $motivo {
              value = "NAO_FINALIZA_ALARME_RESTAURE_ACIMA_50S"
            }
          
            var.update $acao {
              value = "NAO_FINALIZOU"
            }
          
            var.update $mostrarTela {
              value = true
            }
          }
        
          elseif (($temAlarme == false) && $temFalhas) {
            var.update $motivo {
              value = "FINALIZA_FALHAS"
            }
          
            var.update $finalizarAgora {
              value = true
            }
          }
        
          elseif (($temAlarme == false) && ($qtdMonitorados > 0) && ($qtdMonitorados == $qtdRestaure)) {
            var.update $motivo {
              value = "FINALIZA_RESTAURE_ISOLADO"
            }
          
            var.update $finalizarAgora {
              value = true
            }
          }
        
          else {
            var.update $motivo {
              value = "SEM_REGRA"
            }
          
            var.update $acao {
              value = "NAO_FINALIZOU"
            }
          
            var.update $mostrarTela {
              value = true
            }
          }
        }
      
        conditional {
          if ($finalizarAgora) {
            function.run ProcessoEnd {
              input = {
                idUsuario  : "0ROBOAUTO"
                Descricao  : "[AUTO] " ~ $motivo ~ " [" ~ $perfilLocal ~ "]"
                UsuarioNome: "ROBO AUTO"
                idProcesso : $evt.idProcesso
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
          
            var.update $retornoProcessoEnd {
              value = $procEnd
            }
          
            conditional {
              if ($procEnd.dados|is_empty) {
                var.update $acao {
                  value = "JA_FINALIZADO"
                }
              
                var.update $mostrarTela {
                  value = false
                }
              }
            
              else {
                var.update $acao {
                  value = "FINALIZOU"
                }
              
                var.update $mostrarTela {
                  value = false
                }
              }
            }
          }
        }
      
        conditional {
          if ($acao == "FINALIZOU") {
            db.add alarmEvent_autofimlog {
              enforce_hidden_fields = false
              data = {
                created_at        : "now"
                idEvento          : $input.idEvento
                idProcesso        : $evt.idProcesso
                idDispositivo     : $evt.idDispositivo
                motivo            : $motivo
                acao              : $acao
                qtdCiclos5m       : $qtdCiclos3mDiffProc
                bloqueio3x5       : false
                temFalhas         : $temFalhas
                temAlarme         : $temAlarme
                temDesarme        : $temDesarme
                temRestaure       : $temRestaure
                temParAlarmeRest50: $temParAlarmeRestaure50
                retornoProcessoEnd: $retornoProcessoEnd
                regraVersao       : "vFinal_3_direct"
                erro              : ""
              }
            } as $log1
          
            var.update $gravouLog {
              value = true
            }
          }
        }
      }
    }
  }

  response = {
    dados: ""|set:"idEvento":$input.idEvento|set:"idProcesso":$evt.idProcesso|set:"idDispositivo":$evt.idDispositivo|set:"perfilLocal":$perfilLocal|set:"motivo":$motivo|set:"acao":$acao|set:"mostrarTela":$mostrarTela|set:"temAlarme":$temAlarme|set:"temRestaure":$temRestaure|set:"temDesarme":$temDesarme|set:"temArme":$temArme|set:"temFalhas":$temFalhas|set:"temParAlarmeRest50":$temParAlarmeRestaure50|set:"qtdCiclos3mDiffProc":$qtdCiclos3mDiffProc|set:"gravouLog":$gravouLog|set:"retornoProcessoEnd":$retornoProcessoEnd
  }
}