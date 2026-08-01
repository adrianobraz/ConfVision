// Ativar ou trocar licenca de gravacao na camera (continua, movimento ou timelapse)
query "vis_camera/gravacao/ativar/{vis_camera_id}" verb=POST {
  api_group = "confVision"

  input {
    int vis_camera_id? filters=min:1
    int vis_licenca_gravacao_id? filters=min:1
  }

  stack {
    db.get vis_camera {
      field_name = "id"
      field_value = $input.vis_camera_id
    } as $camera
  
    precondition ($camera != null) {
      error_type = "notfound"
      error = "Camera nao encontrada"
    }
  
    precondition (($input.vis_licenca_gravacao_id|is_empty) == false) {
      error = "Selecione uma licenca de gravacao"
    }
  
    // Bloqueia so se ja estiver ativa com a MESMA licenca
    conditional {
      if ($camera.gravacao_status == "ativa" && $camera.vis_licenca_gravacao_id == $input.vis_licenca_gravacao_id) {
        precondition (1 == 0) {
          error = "Camera ja possui esta licenca de gravacao ativa"
        }
      }
    }
  
    db.get vis_licenca {
      field_name = "id"
      field_value = $input.vis_licenca_gravacao_id
    } as $licenca
  
    precondition ($licenca != null) {
      error = "Licenca de gravacao nao encontrada"
    }
  
    precondition ($licenca.unidade == "gravacao") {
      error = "Licenca informada nao e de gravacao"
    }
  
    precondition ($licenca.status == "disponivel") {
      error = "Licenca de gravacao indisponivel ou ja em uso"
    }
  
    precondition ($licenca.id_franqueado == $camera.id_franqueado) {
      error = "Licenca nao pertence ao franqueado da camera"
    }
  
    conditional {
      if ($licenca.valido_ate != null) {
        precondition ($licenca.valido_ate > now) {
          error = "Licenca de gravacao expirada"
        }
      }
    }
  
    function.run fn_vis_gravacao_storage_ensure {
      input = {id_franqueado: $camera.id_franqueado}
    } as $storage
  
    function.run fn_vis_plano_flags {
      input = {plano: $licenca.plano}
    } as $flags
  
    precondition ($flags.grava_continua || $flags.grava_movimento || $flags.grava_timelapse) {
      error = "Plano de licenca invalido para gravacao"
    }
  
    // Libera licenca(s) antiga(s) antes de vincular a nova
    conditional {
      if ($camera.vis_licenca_gravacao_id != null && $camera.vis_licenca_gravacao_id != $input.vis_licenca_gravacao_id) {
        db.get vis_licenca {
          field_name = "id"
          field_value = $camera.vis_licenca_gravacao_id
        } as $licenca_antiga
      
        conditional {
          if ($licenca_antiga != null) {
            conditional {
              if ($licenca_antiga.valido_ate != null && $licenca_antiga.valido_ate < now) {
                db.edit vis_licenca {
                  field_name = "id"
                  field_value = $camera.vis_licenca_gravacao_id
                  data = {
                    status        : "expirada"
                    vis_camera_id : null
                    id_dispositivo: null
                    plano         : $licenca_antiga.plano
                  }
                } as $licenca_antiga_expirada
              }
            
              else {
                db.edit vis_licenca {
                  field_name = "id"
                  field_value = $camera.vis_licenca_gravacao_id
                  data = {
                    status        : "disponivel"
                    vis_camera_id : null
                    id_dispositivo: null
                    plano         : $licenca_antiga.plano
                  }
                } as $licenca_antiga_liberada
              }
            }
          }
        }
      }
    }
  
    db.query vis_licenca {
      where = $db.vis_licenca.vis_camera_id == $input.vis_camera_id && $db.vis_licenca.unidade == "gravacao" && $db.vis_licenca.id != $input.vis_licenca_gravacao_id
      return = {type: "list"}
    } as $licencas_orfas
  
    foreach ($licencas_orfas) {
      each as $orfa {
        conditional {
          if ($orfa.valido_ate != null && $orfa.valido_ate < now) {
            db.edit vis_licenca {
              field_name = "id"
              field_value = $orfa.id
              data = {
                status        : "expirada"
                vis_camera_id : null
                id_dispositivo: null
                plano         : $orfa.plano
              }
            } as $orfa_expirada
          }
        
          else {
            db.edit vis_licenca {
              field_name = "id"
              field_value = $orfa.id
              data = {
                status        : "disponivel"
                vis_camera_id : null
                id_dispositivo: null
                plano         : $orfa.plano
              }
            } as $orfa_liberada
          }
        }
      }
    }
  
    db.edit vis_camera {
      field_name = "id"
      field_value = $input.vis_camera_id
      data = {
        vis_licenca_gravacao_id: $input.vis_licenca_gravacao_id
        grava_continua         : $flags.grava_continua
        grava_movimento        : $flags.grava_movimento
        grava_timelapse        : $flags.grava_timelapse
        retencao_dias          : $flags.retencao_dias
        gravacao_ativada_em    : "now"
        gravacao_status        : "ativa"
      }
    } as $model
  
    db.edit vis_licenca {
      field_name = "id"
      field_value = $input.vis_licenca_gravacao_id
      data = {
        status        : "em_uso"
        vis_camera_id : $input.vis_camera_id
        id_dispositivo: $camera.id_dispositivo
        plano         : $licenca.plano
      }
    } as $licenca_atualizada
  }

  response = $model
}