// Cadastrar no MediaMTX (VPS)
query vis_mediamtx_node verb=POST {
  api_group = "confVision"

  input {
    dblink {
      table = "vis_mediamtx_node"
    }
  }

  stack {
    precondition (($input.nome|is_empty) == false) {
      error = "nome obrigatorio"
    }

    precondition (($input.rtmp_public|is_empty) == false) {
      error = "rtmp_public obrigatorio"
    }

    var $max_cameras {
      value = $input.max_cameras|first_notempty:200
    }

    var $ordem {
      value = $input.ordem
    }

    conditional {
      if ($ordem == null || $ordem == 0) {
        db.query vis_mediamtx_node {
          return = {type: "count"}
        } as $qtd

        var.update $ordem {
          value = $qtd + 1
        }
      }
    }

    var $status {
      value = $input.status|first_notempty:"ativo"
    }

    var $rtsp_internal {
      value = $input.rtsp_internal|first_notempty:"rtsp://127.0.0.1:8554"
    }

    db.add vis_mediamtx_node {
      data = {
        created_at   : "now"
        nome         : $input.nome
        rtmp_public  : $input.rtmp_public
        hls_public   : $input.hls_public
        rtsp_internal: $rtsp_internal
        max_cameras  : $max_cameras
        ordem        : $ordem
        status       : $status
        observacao   : $input.observacao
      }
    } as $model
  }

  response = $model
}
