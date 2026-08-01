// Registrar segmento de gravacao (worker apos upload S3)
query vis_gravacao_segmento verb=POST {
  api_group = "confVision"

  input {
    dblink {
      table = "vis_gravacao_segmento"
    }
  }

  stack {
    precondition (($input.vis_camera_id|is_empty) == false) {
      error = "vis_camera_id obrigatorio"
    }
  
    precondition (($input.inicio_em|is_empty) == false) {
      error = "inicio_em obrigatorio"
    }
  
    precondition (($input.s3_key|is_empty) == false) {
      error = "s3_key obrigatorio"
    }
  
    db.get vis_camera {
      field_name = "id"
      field_value = $input.vis_camera_id
    } as $camera
  
    precondition ($camera != null) {
      error = "Camera nao encontrada"
    }
  
    var $vis_gravacao_storage_id {
      value = $input.vis_gravacao_storage_id
    }
  
    conditional {
      if ($vis_gravacao_storage_id == null) {
        db.query vis_gravacao_storage {
          where = $db.vis_gravacao_storage.id_franqueado == $camera.id_franqueado && $db.vis_gravacao_storage.status == "ativo"
          return = {type: "single"}
        } as $storage_auto
      
        conditional {
          if ($storage_auto != null) {
            var.update $vis_gravacao_storage_id {
              value = $storage_auto.id
            }
          }
        }
      }
    }
  
    var $status {
      value = $input.status
    }
  
    conditional {
      if ($status|is_empty) {
        var.update $status {
          value = "disponivel"
        }
      }
    }
  
    var $expira_em {
      value = $input.expira_em
    }
  
    var $tipo {
      value = $input.tipo
    }
  
    var $flags_gravacao {
      value = null
    }
  
    conditional {
      if ($camera.vis_licenca_gravacao_id != null) {
        db.get vis_licenca {
          field_name = "id"
          field_value = $camera.vis_licenca_gravacao_id
        } as $licenca_gravacao
      
        conditional {
          if ($licenca_gravacao != null) {
            function.run fn_vis_plano_flags {
              input = {plano: $licenca_gravacao.plano}
            } as $flags_gravacao
          }
        }
      }
    }
  
    conditional {
      if ($tipo|is_empty) {
        var.update $tipo {
          value = "continua"
        }
      
        conditional {
          if ($flags_gravacao != null && $flags_gravacao.grava_movimento) {
            var.update $tipo {
              value = "movimento"
            }
          }
        
          elseif ($flags_gravacao != null && $flags_gravacao.grava_timelapse) {
            var.update $tipo {
              value = "timelapse"
            }
          }
        }
      }
    }
  
    conditional {
      if ($expira_em == null && $camera.retencao_dias != null && $camera.retencao_dias > 0) {
        var.update $expira_em {
          value = $input.inicio_em
            |add_secs_to_timestamp:$camera.retencao_dias * 86400
        }
      }
    }
  
    db.add vis_gravacao_segmento {
      enforce_hidden_fields = false
      data = {
        created_at             : "now"
        vis_camera_id          : $input.vis_camera_id
        vis_gravacao_storage_id: $vis_gravacao_storage_id
        id_franqueado          : $camera.id_franqueado
        id_cliente             : $camera.id_cliente
        inicio_em              : $input.inicio_em
        fim_em                 : $input.fim_em
        duracao_seg            : $input.duracao_seg
        s3_key                 : $input.s3_key
        s3_url                 : $input.s3_url
        tamanho_bytes          : $input.tamanho_bytes
        status                 : $status
        tipo                   : $tipo
        erro_msg               : $input.erro_msg
        uploaded_em            : $input.uploaded_em
        expira_em              : $expira_em
      }
    } as $model
  }

  response = $model
}