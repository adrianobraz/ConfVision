// Retira licenca de gravacao da camera — volta disponivel (ou expirada se vencida)
query "vis_camera/gravacao/desativar/{vis_camera_id}" verb=POST {
  api_group = "confVision"

  input {
    int vis_camera_id? filters=min:1
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
  
    var $licenca_id {
      value = $camera.vis_licenca_gravacao_id
    }
  
    conditional {
      if ($licenca_id == null) {
        db.query vis_licenca {
          where = $db.vis_licenca.vis_camera_id == $input.vis_camera_id && $db.vis_licenca.unidade == "gravacao"
          return = {type: "single"}
        } as $licenca_orfa
      
        conditional {
          if ($licenca_orfa != null) {
            var.update $licenca_id {
              value = $licenca_orfa.id
            }
          }
        }
      }
    }
  
    var $tem_gravacao {
      value = $licenca_id != null || $camera.gravacao_status == "ativa" || $camera.grava_continua || $camera.grava_movimento || $camera.grava_timelapse || ($camera.retencao_dias != null && $camera.retencao_dias > 0)
    }
  
    precondition ($tem_gravacao) {
      error = "Camera nao possui licenca de gravacao para retirar"
    }
  
    conditional {
      if ($licenca_id != null) {
        db.get vis_licenca {
          field_name = "id"
          field_value = $licenca_id
        } as $licenca
      
        conditional {
          if ($licenca != null) {
            conditional {
              if ($licenca.valido_ate != null && $licenca.valido_ate < now) {
                db.edit vis_licenca {
                  field_name = "id"
                  field_value = $licenca_id
                  data = {
                    status        : "expirada"
                    vis_camera_id : null
                    id_dispositivo: null
                    plano         : $licenca.plano
                  }
                } as $licenca_expirada
              }
            
              else {
                db.edit vis_licenca {
                  field_name = "id"
                  field_value = $licenca_id
                  data = {
                    status        : "disponivel"
                    vis_camera_id : null
                    id_dispositivo: null
                    plano         : $licenca.plano
                  }
                } as $licenca_liberada
              }
            }
          }
        }
      }
    }
  
    // Libera qualquer outra licenca de gravacao ainda apontando para a camera
    db.query vis_licenca {
      where = $db.vis_licenca.vis_camera_id == $input.vis_camera_id && $db.vis_licenca.unidade == "gravacao" && ($licenca_id == null || $db.vis_licenca.id != $licenca_id)
      return = {type: "list"}
    } as $outras_grav
  
    foreach ($outras_grav) {
      each as $outra {
        conditional {
          if ($outra.valido_ate != null && $outra.valido_ate < now) {
            db.edit vis_licenca {
              field_name = "id"
              field_value = $outra.id
              data = {
                status        : "expirada"
                vis_camera_id : null
                id_dispositivo: null
                plano         : $outra.plano
              }
            } as $outra_expirada
          }
        
          else {
            db.edit vis_licenca {
              field_name = "id"
              field_value = $outra.id
              data = {
                status        : "disponivel"
                vis_camera_id : null
                id_dispositivo: null
                plano         : $outra.plano
              }
            } as $outra_liberada
          }
        }
      }
    }
  
    db.patch vis_camera {
      field_name = "id"
      field_value = $input.vis_camera_id
      data = {}
        |set:"vis_licenca_gravacao_id":null
        |set:"grava_continua":false
        |set:"grava_movimento":false
        |set:"grava_timelapse":false
        |set:"retencao_dias":null
        |set:"gravacao_ativada_em":null
        |set:"gravacao_status":"inativa"
    } as $camera_patch
  
    db.get vis_camera {
      field_name = "id"
      field_value = $input.vis_camera_id
    } as $model
  }

  response = $model
}