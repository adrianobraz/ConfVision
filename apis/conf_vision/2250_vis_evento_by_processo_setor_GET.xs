// Evento ConfVision de um setor no processo (terminal)
query vis_evento_by_processo_setor verb=GET {
  api_group = "confVision"

  input {
    text id_processo? filters=trim
    text id_dispositivo? filters=trim
    text particao? filters=trim
    text zonauser? filters=trim
    text id_evento? filters=trim
  }

  stack {
    var $evento {
      value = null
    }
  
    var $camera {
      value = null
    }
  
    conditional {
      if (($input.id_evento|is_empty) == false) {
        db.query vis_evento {
          where = $db.vis_evento.id_evento == $input.id_evento && $db.vis_evento.id_processo ==? $input.id_processo
          sort = {vis_evento.created_at: "desc"}
          return = {type: "list"}
        } as $por_id_evento
      
        var.update $evento {
          value = $por_id_evento|first
        }
      }
    }
  
    conditional {
      if ($evento == null) {
        db.query vis_camera {
          where = $db.vis_camera.id_dispositivo == $input.id_dispositivo && $db.vis_camera.particao == $input.particao && $db.vis_camera.zonauser == $input.zonauser
          sort = {vis_camera.id: "desc"}
          return = {type: "list"}
        } as $cameras
      
        var.update $camera {
          value = $cameras|first
        }
      
        precondition ($camera != null) {
          error = "Camera nao encontrada para o setor"
        }
      
        db.query vis_evento {
          where = $db.vis_evento.id_processo == $input.id_processo && $db.vis_evento.vis_camera_id == $camera.id && $db.vis_evento.particao ==? $input.particao
          sort = {vis_evento.created_at: "desc"}
          return = {type: "list"}
        } as $lista
      
        var.update $evento {
          value = $lista|first
        }
      }
    }
  
    conditional {
      if ($camera == null && $evento != null && $evento.vis_camera_id != null) {
        db.get vis_camera {
          field_name = "id"
          field_value = $evento.vis_camera_id
        } as $cam
      
        var.update $camera {
          value = $cam
        }
      }
    }
  }

  response = {dados: $evento, camera: $camera}
}