// Update vis_camera record — flags do plano bloqueadas quando ha licenca; permite trocar licenca
query "vis_camera/{vis_camera_id}" verb=PUT {
  api_group = "confVision"

  input {
    int vis_camera_id? filters=min:1
    dblink {
      table = "vis_camera"
    }
  }

  stack {
    db.get vis_camera {
      field_name = "id"
      field_value = $input.vis_camera_id
    } as $atual
  
    precondition ($atual != null) {
      error_type = "notfound"
      error = "Camera nao encontrada"
    }
  
    var $vis_licenca_id {
      value = $atual.vis_licenca_id
    }
  
    var $captura_sensor {
      value = $input.captura_sensor
    }
  
    var $captura_analitico {
      value = $input.captura_analitico
    }
  
    var $somente_armado {
      value = $input.somente_armado
    }
  
    var $evento_grava_foto {
      value = $input.evento_grava_foto
    }
  
    var $evento_grava_video {
      value = $input.evento_grava_video
    }
  
    var $plano {
      value = $input.plano
    }
  
    conditional {
      if (($plano|is_empty) && ($atual.plano|is_empty) == false) {
        var.update $plano {
          value = $atual.plano
        }
      }
    }
  
    var $ativado_em {
      value = $atual.ativado_em
    }
  
    conditional {
      if (($input.vis_licenca_id|is_empty) == false && $input.vis_licenca_id != $atual.vis_licenca_id) {
        db.get vis_licenca {
          field_name = "id"
          field_value = $input.vis_licenca_id
        } as $nova_licenca
      
        precondition ($nova_licenca != null) {
          error = "Licenca selecionada nao encontrada"
        }
      
        precondition (($nova_licenca.plano|is_empty) == false) {
          error = "Licenca selecionada sem plano definido"
        }
      
        precondition ($nova_licenca.status == "disponivel") {
          error = "Licenca indisponivel ou ja em uso"
        }
      
        precondition ($nova_licenca.id_franqueado == $input.id_franqueado) {
          error = "Licenca nao pertence ao franqueado"
        }
      
        precondition ($nova_licenca.unidade == "camera" || ($nova_licenca.unidade|is_empty)) {
          error = "Licenca selecionada nao e de camera"
        }
      
        conditional {
          if ($nova_licenca.valido_ate != null) {
            precondition ($nova_licenca.valido_ate > now) {
              error = "Licenca expirada"
            }
          }
        }
      
        conditional {
          if ($atual.vis_licenca_id != null) {
            db.get vis_licenca {
              field_name = "id"
              field_value = $atual.vis_licenca_id
            } as $licenca_antiga
          
            db.edit vis_licenca {
              field_name = "id"
              field_value = $atual.vis_licenca_id
              data = {
                status        : "disponivel"
                vis_camera_id : null
                id_dispositivo: null
                plano         : $licenca_antiga.plano
              }
            } as $licenca_liberada
          }
        }
      
        db.edit vis_licenca {
          field_name = "id"
          field_value = $input.vis_licenca_id
          data = {
            status        : "em_uso"
            vis_camera_id : $input.vis_camera_id
            id_dispositivo: $input.id_dispositivo
            plano         : $nova_licenca.plano
          }
        } as $licenca_vinculada
      
        var.update $vis_licenca_id {
          value = $input.vis_licenca_id
        }
      
        function.run fn_vis_plano_flags {
          input = {plano: $nova_licenca.plano}
        } as $flags
      
        var.update $captura_sensor {
          value = $flags.captura_sensor
        }
      
        var.update $captura_analitico {
          value = $flags.captura_analitico
        }
      
        var.update $somente_armado {
          value = $flags.somente_armado
        }
      
        var.update $evento_grava_foto {
          value = $flags.evento_grava_foto
        }
      
        var.update $evento_grava_video {
          value = $flags.evento_grava_video
        }
      
        var.update $plano {
          value = $nova_licenca.plano
        }
      }
    
      else {
        conditional {
          if ($atual.vis_licenca_id != null) {
            db.get vis_licenca {
              field_name = "id"
              field_value = $atual.vis_licenca_id
            } as $licenca
          
            precondition ($licenca != null) {
              error = "Licenca vinculada nao encontrada"
            }
          
            function.run fn_vis_plano_flags {
              input = {plano: $licenca.plano}
            } as $flags
          
            var.update $captura_sensor {
              value = $flags.captura_sensor
            }
          
            var.update $captura_analitico {
              value = $flags.captura_analitico
            }
          
            var.update $somente_armado {
              value = $flags.somente_armado
            }
          
            var.update $evento_grava_foto {
              value = $flags.evento_grava_foto
            }
          
            var.update $evento_grava_video {
              value = $flags.evento_grava_video
            }
          
            var.update $plano {
              value = $licenca.plano
            }
          }
        }
      }
    }
  
    var $flags_atual {
      value = null
    }
  
    conditional {
      if (($plano|is_empty) == false) {
        function.run fn_vis_plano_flags {
          input = {plano: $plano}
        } as $flags_atual
      }
    }
  
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
      if ($plano == "online") {
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
    
      elseif ($flags_atual != null && $flags_atual.sem_ativo) {
        var.update $ativo {
          value = false
        }
      }
    }
  
    conditional {
      if ($flags_atual != null && $flags_atual.captura_analitico) {
        var.update $deteccao_humano {
          value = true
        }
      }
    }
  
    conditional {
      if ($ativo && $ativado_em == null) {
        var.update $ativado_em {
          value = now
        }
      }
    }
  
    var $limpar_gravacao {
      value = false
    }
  
    // Desativar analitico (ativo true -> false): libera licenca de gravacao e licenca de camera
    conditional {
      if ($atual.ativo && $ativo == false) {
        var $tem_gravacao {
          value = $atual.gravacao_status == "ativa" || $atual.vis_licenca_gravacao_id != null || $atual.grava_continua || $atual.grava_movimento || $atual.grava_timelapse
        }
      
        conditional {
          if ($tem_gravacao) {
            var.update $limpar_gravacao {
              value = true
            }
          
            var $licenca_gravacao_id {
              value = $atual.vis_licenca_gravacao_id
            }
          
            conditional {
              if ($licenca_gravacao_id == null) {
                db.query vis_licenca {
                  where = $db.vis_licenca.vis_camera_id == $input.vis_camera_id && $db.vis_licenca.unidade == "gravacao"
                  return = {type: "single"}
                } as $licenca_orfa
              
                conditional {
                  if ($licenca_orfa != null) {
                    var.update $licenca_gravacao_id {
                      value = $licenca_orfa.id
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($licenca_gravacao_id != null) {
                db.get vis_licenca {
                  field_name = "id"
                  field_value = $licenca_gravacao_id
                } as $licenca_grav
              
                conditional {
                  if ($licenca_grav != null) {
                    db.patch vis_licenca {
                      field_name = "id"
                      field_value = $licenca_gravacao_id
                      data = {}
                        |set:"status":"disponivel"
                        |set:"vis_camera_id":null
                        |set:"id_dispositivo":null
                        |set:"plano":$licenca_grav.plano
                    } as $licenca_grav_liberada
                  }
                }
              }
            }
          }
        }
      
        conditional {
          if ($vis_licenca_id != null) {
            db.get vis_licenca {
              field_name = "id"
              field_value = $vis_licenca_id
            } as $licenca_cam
          
            conditional {
              if ($licenca_cam != null) {
                db.patch vis_licenca {
                  field_name = "id"
                  field_value = $vis_licenca_id
                  data = {}
                    |set:"status":"disponivel"
                    |set:"vis_camera_id":null
                    |set:"id_dispositivo":null
                    |set:"plano":$licenca_cam.plano
                } as $licenca_cam_liberada
              }
            }
          
            var.update $vis_licenca_id {
              value = null
            }
          }
        }
      }
    }
  
    db.edit vis_camera {
      field_name = "id"
      field_value = $input.vis_camera_id
      enforce_hidden_fields = false
      data = {
        ativo             : $ativo
        bloqueado         : $atual.bloqueado
        nome              : $input.nome
        id_franqueado     : $input.id_franqueado
        id_cliente        : $input.id_cliente
        id_dispositivo    : $input.id_dispositivo
        conta             : $input.conta
        particao          : $input.particao
        canal             : $input.canal
        setor             : $input.setor
        protocolo         : $input.protocolo
        rtsp_url_sec      : $input.rtsp_url_sec
        onvif_host        : $input.onvif_host
        onvif_porta       : $input.onvif_porta
        onvif_usuario     : $input.onvif_usuario
        onvif_senha       : $input.onvif_senha
        confianca_min     : $input.confianca_min
        cooldown_seg      : $input.cooldown_seg
        modo_deteccao     : $input.modo_deteccao
        somente_armado    : $somente_armado
        deteccao_humano   : $deteccao_humano
        deteccao_veiculo  : $deteccao_veiculo
        status            : $input.status
        ultimo_evento_em  : $input.ultimo_evento_em
        worker_id         : $input.worker_id
        ultimo_ping_em    : $input.ultimo_ping_em
        zonauser          : $input.zonauser
        captura_sensor    : $captura_sensor
        captura_analitico : $captura_analitico
        analitico_pausado : $input.analitico_pausado|first_notempty:$atual.analitico_pausado|first_notempty:false
        evento_grava_foto : $evento_grava_foto
        evento_grava_video: $evento_grava_video
        id_setor          : $input.id_setor
        snapshot_url      : $input.snapshot_url
        vis_licenca_id    : $vis_licenca_id
        plano             : $plano
        ativado_em        : $ativado_em
      }
    } as $model
  
    conditional {
      if ($limpar_gravacao) {
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
        } as $camera_grav_limpa
      
        db.get vis_camera {
          field_name = "id"
          field_value = $input.vis_camera_id
        } as $model
      }
    }
  
    conditional {
      if ($vis_licenca_id != null) {
        conditional {
          if (($input.vis_licenca_id|is_empty) || $input.vis_licenca_id == $atual.vis_licenca_id) {
            db.get vis_licenca {
              field_name = "id"
              field_value = $vis_licenca_id
            } as $lic_sync
          
            db.edit vis_licenca {
              field_name = "id"
              field_value = $vis_licenca_id
              data = {
                id_dispositivo: $input.id_dispositivo
                plano         : $lic_sync.plano
              }
            } as $licenca_sync
          }
        }
      }
    }
  }

  response = $model
}