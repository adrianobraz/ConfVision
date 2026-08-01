// Cria no padrao (servidor1) a partir de fp_config_financeiro se nenhum no existir
function fn_vis_mediamtx_node_ensure_default {
  input {
  }

  stack {
    var $primeiro {
      value = null
    }

    db.query vis_mediamtx_node {
      sort = {vis_mediamtx_node.ordem: "asc"}
      return = {type: "list"}
    } as $existentes

    conditional {
      if (($existentes|count) > 0) {
        var $primeiro {
          value = $existentes|first
        }
      }

      else {
        var $nome {
          value = "servidor1"
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_mediamtx_default_nome"
        } as $cfg_nome

        conditional {
          if ($cfg_nome != null && ($cfg_nome.valor|is_empty) == false) {
            var.update $nome {
              value = $cfg_nome.valor|trim
            }
          }
        }

        var $rtmp_public {
          value = ""
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_mediamtx_default_rtmp_public"
        } as $cfg_rtmp

        conditional {
          if ($cfg_rtmp != null && ($cfg_rtmp.valor|is_empty) == false) {
            var.update $rtmp_public {
              value = $cfg_rtmp.valor|trim
            }
          }
        }

        var $hls_public {
          value = ""
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_mediamtx_default_hls_public"
        } as $cfg_hls

        conditional {
          if ($cfg_hls != null && ($cfg_hls.valor|is_empty) == false) {
            var.update $hls_public {
              value = $cfg_hls.valor|trim
            }
          }
        }

        var $rtsp_internal {
          value = "rtsp://127.0.0.1:8554"
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_mediamtx_default_rtsp_internal"
        } as $cfg_rtsp

        conditional {
          if ($cfg_rtsp != null && ($cfg_rtsp.valor|is_empty) == false) {
            var.update $rtsp_internal {
              value = $cfg_rtsp.valor|trim
            }
          }
        }

        var $max_cameras {
          value = 200
        }

        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_mediamtx_default_max_cameras"
        } as $cfg_max

        conditional {
          if ($cfg_max != null && ($cfg_max.valor|is_empty) == false) {
            var.update $max_cameras {
              value = $cfg_max.valor|to_int
            }
          }
        }

        precondition (($rtmp_public|is_empty) == false) {
          error = "Nenhum no MediaMTX cadastrado — configure confvision_mediamtx_default_rtmp_public em Config Financeiro ou cadastre vis_mediamtx_node"
        }

        db.add vis_mediamtx_node {
          data = {
            created_at   : "now"
            nome         : $nome
            rtmp_public  : $rtmp_public
            hls_public   : $hls_public
            rtsp_internal: $rtsp_internal
            max_cameras  : $max_cameras
            ordem        : 1
            status       : "ativo"
            observacao   : "Provisionado automaticamente (fn_vis_mediamtx_node_ensure_default)"
          }
        } as $novo

        var.update $primeiro {
          value = $novo
        }
      }
    }
  }

  response = $primeiro
}
