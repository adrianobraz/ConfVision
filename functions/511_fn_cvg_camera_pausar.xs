// Pausa ou retoma deteccao analitica (patch direto vis_camera)
function fn_cvg_camera_pausar {
  input {
    int vis_camera_id? filters=min:1
    bool pausado?
  }

  stack {
    db.get vis_camera {
      field_name = "id"
      field_value = $input.vis_camera_id
    } as $camera

    precondition ($camera != null) {
      error = "Camera nao encontrada"
    }

    function.run fn_cvg_plano_tipo {
      input = {plano: $camera.plano}
    } as $pt

    precondition ($pt.captura_analitico) {
      error = "Plano sem deteccao analitica"
    }

    var $pausado {
      value = $input.pausado|first_notempty:false
    }

    db.patch vis_camera {
      field_name = "id"
      field_value = $input.vis_camera_id
      data = {}|set:"analitico_pausado":$pausado
    } as $result
  }

  response = {
    ok               : true
    vis_camera_id    : $input.vis_camera_id
    analitico_pausado: $pausado
  }
}
