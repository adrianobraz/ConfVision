// Add vis_camera record — exige licenca prepaga disponivel
query vis_camera verb=POST {
  api_group = "confVision"

  input {
    dblink {
      table = "vis_camera"
    }
  }

  stack {
    precondition (($input.vis_licenca_id|is_empty) == false) {
      error = "Selecione uma licenca disponivel para cadastrar a camera"
    }
  
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    db.get vis_licenca {
      field_name = "id"
      field_value = $input.vis_licenca_id
    } as $licenca
  
    precondition ($licenca != null) {
      error = "Licenca nao encontrada"
    }
  
    precondition ($licenca.status == "disponivel") {
      error = "Licenca indisponivel ou ja em uso"
    }
  
    precondition (($licenca.plano|is_empty) == false) {
      error = "Licenca sem plano definido"
    }
  
    precondition ($licenca.id_franqueado == $input.id_franqueado) {
      error = "Licenca nao pertence ao franqueado"
    }
  
    precondition ($licenca.unidade == "camera" || ($licenca.unidade|is_empty)) {
      error = "Licenca informada nao e de camera — use licenca de plano online/sensor/analitico"
    }
  
    conditional {
      if ($licenca.valido_ate != null) {
        precondition ($licenca.valido_ate > now) {
          error = "Licenca expirada"
        }
      }
    }
  
    function.run fn_vis_plano_flags {
      input = {plano: $licenca.plano}
    } as $flags
  
    var $ativo {
      value = $input.ativo
    }
  
    var $deteccao_humano {
      value = $input.deteccao_humano
    }
  
    var $deteccao_veiculo {
      value = $input.deteccao_veiculo
    }
  
    conditional {
      if ($licenca.plano == "online") {
        var.update $ativo {
          value = false
        }
      
        var.update $deteccao_humano {
          value = false
        }
      
        var.update $deteccao_veiculo {
          value = false
        }
      }
    
      elseif ($flags.sem_ativo) {
        var.update $ativo {
          value = false
        }
      }
    }
  
    conditional {
      if ($flags.captura_analitico) {
        var.update $deteccao_humano {
          value = true
        }
      }
    }
  
    var $ativado_em {
      value = null
    }
  
    conditional {
      if ($ativo) {
        var.update $ativado_em {
          value = now
        }
      }
    }

    function.run fn_vis_mediamtx_pick_node {
    } as $mtx_node

    db.add vis_camera {
      enforce_hidden_fields = false
      data = {
        created_at          : "now"
        ativo               : $ativo
        bloqueado           : false
        nome                : $input.nome
        id_franqueado       : $input.id_franqueado
        id_cliente          : $input.id_cliente
        id_dispositivo      : $input.id_dispositivo
        conta               : $input.conta
        particao            : $input.particao
        canal               : $input.canal
        setor               : $input.setor
        protocolo           : $input.protocolo
        rtsp_url_sec        : $input.rtsp_url_sec
        onvif_host          : $input.onvif_host
        onvif_porta         : $input.onvif_porta
        onvif_usuario       : $input.onvif_usuario
        onvif_senha         : $input.onvif_senha
        confianca_min       : $input.confianca_min
        cooldown_seg        : $input.cooldown_seg
        modo_deteccao       : $input.modo_deteccao
        somente_armado      : $flags.somente_armado
        deteccao_humano     : $deteccao_humano
        deteccao_veiculo    : $deteccao_veiculo
        status              : $input.status
        ultimo_evento_em    : $input.ultimo_evento_em
        worker_id           : $input.worker_id
        ultimo_ping_em      : $input.ultimo_ping_em
        zonauser            : $input.zonauser
        captura_sensor      : $flags.captura_sensor
        captura_analitico   : $flags.captura_analitico
        analitico_pausado   : $input.analitico_pausado|first_notempty:false
        evento_grava_foto   : $flags.evento_grava_foto
        evento_grava_video  : $flags.evento_grava_video
        id_setor            : $input.id_setor
        snapshot_url        : $input.snapshot_url
        vis_licenca_id      : $input.vis_licenca_id
        plano               : $licenca.plano
        ativado_em          : $ativado_em
        vis_mediamtx_node_id: $mtx_node.id
      }
    } as $model

    function.run fn_vis_mediamtx_node_sync_status {
      input = {vis_mediamtx_node_id: $mtx_node.id}
    } as $mtx_sync

    db.edit vis_licenca {
      field_name = "id"
      field_value = $input.vis_licenca_id
      data = {
        status        : "em_uso"
        vis_camera_id : $model.id
        id_dispositivo: $input.id_dispositivo
        plano         : $licenca.plano
      }
    } as $licenca_atualizada
  }

  response = $model
}