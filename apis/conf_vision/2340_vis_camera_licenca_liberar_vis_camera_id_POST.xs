// Retira licenca da camera — volta disponivel (ou expirada se vencida); encerra gravacao se houver
query "vis_camera/licenca/liberar/{vis_camera_id}" verb=POST {
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
  
    var $licenca_cam_id {
      value = $camera.vis_licenca_id
    }
  
    conditional {
      if ($licenca_cam_id == null) {
        db.query vis_licenca {
          where = $db.vis_licenca.vis_camera_id == $input.vis_camera_id && ($db.vis_licenca.unidade == "camera" || ($db.vis_licenca.unidade|is_empty))
          return = {type: "single"}
        } as $licenca_orfa
      
        conditional {
          if ($licenca_orfa != null) {
            var.update $licenca_cam_id {
              value = $licenca_orfa.id
            }
          }
        }
      }
    }
  
    precondition ($licenca_cam_id != null) {
      error = "Camera nao possui licenca para retirar"
    }
  
    db.get vis_licenca {
      field_name = "id"
      field_value = $licenca_cam_id
    } as $licenca_cam
  
    precondition ($licenca_cam != null) {
      error = "Licenca vinculada nao encontrada"
    }
  
    conditional {
      if ($licenca_cam.valido_ate != null && $licenca_cam.valido_ate < now) {
        db.edit vis_licenca {
          field_name = "id"
          field_value = $licenca_cam_id
          data = {
            status        : "expirada"
            vis_camera_id : null
            id_dispositivo: null
            plano         : $licenca_cam.plano
          }
        } as $licenca_cam_expirada
      }
    
      else {
        db.edit vis_licenca {
          field_name = "id"
          field_value = $licenca_cam_id
          data = {
            status        : "disponivel"
            vis_camera_id : null
            id_dispositivo: null
            plano         : $licenca_cam.plano
          }
        } as $licenca_cam_liberada
      }
    }
  
    // Libera licenca de gravacao vinculada na camera
    var $licenca_grav_id {
      value = $camera.vis_licenca_gravacao_id
    }
  
    conditional {
      if ($licenca_grav_id == null) {
        db.query vis_licenca {
          where = $db.vis_licenca.vis_camera_id == $input.vis_camera_id && $db.vis_licenca.unidade == "gravacao"
          return = {type: "single"}
        } as $licenca_grav_orfa
      
        conditional {
          if ($licenca_grav_orfa != null) {
            var.update $licenca_grav_id {
              value = $licenca_grav_orfa.id
            }
          }
        }
      }
    }
  
    conditional {
      if ($licenca_grav_id != null) {
        db.get vis_licenca {
          field_name = "id"
          field_value = $licenca_grav_id
        } as $licenca_grav
      
        conditional {
          if ($licenca_grav != null) {
            conditional {
              if ($licenca_grav.valido_ate != null && $licenca_grav.valido_ate < now) {
                db.edit vis_licenca {
                  field_name = "id"
                  field_value = $licenca_grav_id
                  data = {
                    status        : "expirada"
                    vis_camera_id : null
                    id_dispositivo: null
                    plano         : $licenca_grav.plano
                  }
                } as $licenca_grav_expirada
              }
            
              else {
                db.edit vis_licenca {
                  field_name = "id"
                  field_value = $licenca_grav_id
                  data = {
                    status        : "disponivel"
                    vis_camera_id : null
                    id_dispositivo: null
                    plano         : $licenca_grav.plano
                  }
                } as $licenca_grav_liberada
              }
            }
          }
        }
      }
    }
  
    // Libera qualquer outra licenca de gravacao ainda apontando para a camera
    db.query vis_licenca {
      where = $db.vis_licenca.vis_camera_id == $input.vis_camera_id && $db.vis_licenca.unidade == "gravacao" && ($licenca_grav_id == null || $db.vis_licenca.id != $licenca_grav_id)
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
        |set:"vis_licenca_id":null
        |set:"plano":null
        |set:"ativo":false
        |set:"captura_sensor":false
        |set:"captura_analitico":false
        |set:"somente_armado":false
        |set:"deteccao_humano":false
        |set:"deteccao_veiculo":false
        |set:"evento_grava_foto":false
        |set:"evento_grava_video":false
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