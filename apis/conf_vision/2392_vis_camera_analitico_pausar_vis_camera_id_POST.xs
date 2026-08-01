// Pausa ou retoma deteccao analitica (worker) sem liberar licenca
query "vis_camera/analitico/pausar/{vis_camera_id}" verb=POST {
  api_group = "confVision"

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

    function.run fn_vis_plano_flags {
      input = {plano: $camera.plano}
    } as $flags

    precondition ($flags.captura_analitico) {
      error = "Plano da camera nao possui deteccao analitica"
    }

    precondition ($camera.ativo) {
      error = "Ative a camera antes de pausar ou retomar a deteccao"
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
    success          : true
    vis_camera_id    : $input.vis_camera_id
    analitico_pausado: $pausado
    message          : $pausado ? "Deteccao analitica pausada" : "Deteccao analitica retomada"
  }
}
