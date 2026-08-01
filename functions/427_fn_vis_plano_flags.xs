// Deriva flags de captura, gravacao e valor a partir do plano da licenca (9 slugs camera + gravacao + legado)
function fn_vis_plano_flags {
  input {
    text plano? filters=trim
    text id_central? filters=trim
  }

  stack {
    var $plano_norm {
      value = $input.plano
    }
  
    conditional {
      if ($plano_norm == "sensor") {
        var.update $plano_norm {
          value = "sensor_foto_video"
        }
      }
    
      elseif ($plano_norm == "analitico_armado") {
        var.update $plano_norm {
          value = "analitico_armado_foto_video"
        }
      }
    
      elseif ($plano_norm == "analitico_24h") {
        var.update $plano_norm {
          value = "analitico_24h_foto_video"
        }
      }
    }
  
    var $captura_sensor {
      value = false
    }
  
    var $captura_analitico {
      value = false
    }
  
    var $somente_armado {
      value = false
    }
  
    var $evento_grava_foto {
      value = false
    }
  
    var $evento_grava_video {
      value = false
    }
  
    var $sem_ativo {
      value = false
    }
  
    var $grava_continua {
      value = false
    }
  
    var $grava_movimento {
      value = false
    }
  
    var $grava_timelapse {
      value = false
    }
  
    var $retencao_dias {
      value = 0
    }
  
    var $segmento_minutos {
      value = 5
    }
  
    var $unidade {
      value = "camera"
    }
  
    var $valor {
      value = 0
    }
  
    var $plano_label {
      value = "Nenhum"
    }
  
    conditional {
      if ($plano_norm == "online") {
        var.update $sem_ativo {
          value = true
        }
      
        var.update $valor {
          value = 2.99
        }
      
        var.update $plano_label {
          value = "Camera online"
        }
      }
    
      elseif ($plano_norm == "sensor_foto") {
        var.update $captura_sensor {
          value = true
        }
      
        var.update $evento_grava_foto {
          value = true
        }
      
        var.update $sem_ativo {
          value = true
        }
      
        var.update $valor {
          value = 7.99
        }
      
        var.update $plano_label {
          value = "Sensor foto"
        }
      }
    
      elseif ($plano_norm == "sensor_foto_video") {
        var.update $captura_sensor {
          value = true
        }
      
        var.update $evento_grava_foto {
          value = true
        }
      
        var.update $evento_grava_video {
          value = true
        }
      
        var.update $sem_ativo {
          value = true
        }
      
        var.update $valor {
          value = 9.99
        }
      
        var.update $plano_label {
          value = "Sensor foto + video"
        }
      }
    
      elseif ($plano_norm == "analitico_armado_evento") {
        var.update $captura_analitico {
          value = true
        }
      
        var.update $somente_armado {
          value = true
        }
      
        var.update $valor {
          value = 11.99
        }
      
        var.update $plano_label {
          value = "Analitico armado — so evento"
        }
      }
    
      elseif ($plano_norm == "analitico_armado_foto") {
        var.update $captura_analitico {
          value = true
        }
      
        var.update $somente_armado {
          value = true
        }
      
        var.update $evento_grava_foto {
          value = true
        }
      
        var.update $valor {
          value = 13.99
        }
      
        var.update $plano_label {
          value = "Analitico armado — foto"
        }
      }
    
      elseif ($plano_norm == "analitico_armado_foto_video") {
        var.update $captura_analitico {
          value = true
        }
      
        var.update $somente_armado {
          value = true
        }
      
        var.update $evento_grava_foto {
          value = true
        }
      
        var.update $evento_grava_video {
          value = true
        }
      
        var.update $valor {
          value = 14.99
        }
      
        var.update $plano_label {
          value = "Analitico armado — foto + video"
        }
      }
    
      elseif ($plano_norm == "analitico_24h_evento") {
        var.update $captura_analitico {
          value = true
        }
      
        var.update $valor {
          value = 16.99
        }
      
        var.update $plano_label {
          value = "Analitico 24h — so evento"
        }
      }
    
      elseif ($plano_norm == "analitico_24h_foto") {
        var.update $captura_analitico {
          value = true
        }
      
        var.update $evento_grava_foto {
          value = true
        }
      
        var.update $valor {
          value = 18.99
        }
      
        var.update $plano_label {
          value = "Analitico 24h — foto"
        }
      }
    
      elseif ($plano_norm == "analitico_24h_foto_video") {
        var.update $captura_analitico {
          value = true
        }
      
        var.update $evento_grava_foto {
          value = true
        }
      
        var.update $evento_grava_video {
          value = true
        }
      
        var.update $valor {
          value = 19.99
        }
      
        var.update $plano_label {
          value = "Analitico 24h — foto + video"
        }
      }
    
      elseif ($plano_norm == "gravacao_7d") {
        var.update $unidade {
          value = "gravacao"
        }
      
        var.update $grava_continua {
          value = true
        }
      
        var.update $retencao_dias {
          value = 7
        }
      
        var.update $segmento_minutos {
          value = 5
        }
      
        var.update $valor {
          value = 12.99
        }
      
        var.update $plano_label {
          value = "Gravacao continua 7 dias"
        }
      }
    
      elseif ($plano_norm == "gravacao_15d") {
        var.update $unidade {
          value = "gravacao"
        }
      
        var.update $grava_continua {
          value = true
        }
      
        var.update $retencao_dias {
          value = 15
        }
      
        var.update $segmento_minutos {
          value = 5
        }
      
        var.update $valor {
          value = 17.99
        }
      
        var.update $plano_label {
          value = "Gravacao continua 15 dias"
        }
      }
    
      elseif ($plano_norm == "gravacao_30d") {
        var.update $unidade {
          value = "gravacao"
        }
      
        var.update $grava_continua {
          value = true
        }
      
        var.update $retencao_dias {
          value = 30
        }
      
        var.update $segmento_minutos {
          value = 5
        }
      
        var.update $valor {
          value = 24.99
        }
      
        var.update $plano_label {
          value = "Gravacao continua 30 dias"
        }
      }
    
      elseif ($plano_norm == "gravacao_movimento_7d") {
        var.update $unidade {
          value = "gravacao"
        }
      
        var.update $grava_movimento {
          value = true
        }
      
        var.update $retencao_dias {
          value = 7
        }
      
        var.update $segmento_minutos {
          value = 5
        }
      
        var.update $valor {
          value = 9.99
        }
      
        var.update $plano_label {
          value = "Gravacao por movimento 7 dias"
        }
      }
    
      elseif ($plano_norm == "gravacao_movimento_15d") {
        var.update $unidade {
          value = "gravacao"
        }
      
        var.update $grava_movimento {
          value = true
        }
      
        var.update $retencao_dias {
          value = 15
        }
      
        var.update $segmento_minutos {
          value = 5
        }
      
        var.update $valor {
          value = 13.99
        }
      
        var.update $plano_label {
          value = "Gravacao por movimento 15 dias"
        }
      }
    
      elseif ($plano_norm == "gravacao_movimento_30d") {
        var.update $unidade {
          value = "gravacao"
        }
      
        var.update $grava_movimento {
          value = true
        }
      
        var.update $retencao_dias {
          value = 30
        }
      
        var.update $segmento_minutos {
          value = 5
        }
      
        var.update $valor {
          value = 19.99
        }
      
        var.update $plano_label {
          value = "Gravacao por movimento 30 dias"
        }
      }
    
      elseif ($plano_norm == "gravacao_timelapse_7d") {
        var.update $unidade {
          value = "gravacao"
        }
      
        var.update $grava_timelapse {
          value = true
        }
      
        var.update $retencao_dias {
          value = 7
        }
      
        var.update $segmento_minutos {
          value = 3
        }
      
        var.update $valor {
          value = 6.99
        }
      
        var.update $plano_label {
          value = "Gravacao timelapse inteligente 7 dias"
        }
      }
    
      elseif ($plano_norm == "gravacao_timelapse_15d") {
        var.update $unidade {
          value = "gravacao"
        }
      
        var.update $grava_timelapse {
          value = true
        }
      
        var.update $retencao_dias {
          value = 15
        }
      
        var.update $segmento_minutos {
          value = 3
        }
      
        var.update $valor {
          value = 9.99
        }
      
        var.update $plano_label {
          value = "Gravacao timelapse inteligente 15 dias"
        }
      }
    
      elseif ($plano_norm == "gravacao_timelapse_30d") {
        var.update $unidade {
          value = "gravacao"
        }
      
        var.update $grava_timelapse {
          value = true
        }
      
        var.update $retencao_dias {
          value = 30
        }
      
        var.update $segmento_minutos {
          value = 3
        }
      
        var.update $valor {
          value = 13.99
        }
      
        var.update $plano_label {
          value = "Gravacao timelapse inteligente 30 dias"
        }
      }
    }
  
    // Override de preco/nome editavel no adm (fp_produto_catalogo / confvision_licenca)
    conditional {
      if (($plano_norm|is_empty) == false && $plano_norm != "Nenhum") {
        function.run fn_fp_catalogo_get_por_plano {
          input = {
            produto   : "confvision_licenca"
            plano     : $plano_norm
            id_central: $input.id_central|first_notempty:""
          }
        } as $cat_preco
      
        conditional {
          if ($cat_preco != null) {
            conditional {
              if ($cat_preco.valor_mensal != null) {
                var.update $valor {
                  value = $cat_preco.valor_mensal
                }
              }
            }
          
            conditional {
              if (($cat_preco.nome_exibicao|is_empty) == false) {
                var.update $plano_label {
                  value = $cat_preco.nome_exibicao
                }
              }
            }
          
            conditional {
              if ($cat_preco.limites_json != null && ($cat_preco.limites_json.unidade|is_empty) == false) {
                var.update $unidade {
                  value = $cat_preco.limites_json.unidade
                }
              }
            }
          }
        }
      }
    }
  }

  response = {
    captura_sensor    : $captura_sensor
    captura_analitico : $captura_analitico
    somente_armado    : $somente_armado
    evento_grava_foto : $evento_grava_foto
    evento_grava_video: $evento_grava_video
    sem_ativo         : $sem_ativo
    grava_continua    : $grava_continua
    grava_movimento   : $grava_movimento
    grava_timelapse   : $grava_timelapse
    retencao_dias     : $retencao_dias
    segmento_minutos  : $segmento_minutos
    unidade           : $unidade
    valor             : $valor
    plano_label       : $plano_label
  }
}