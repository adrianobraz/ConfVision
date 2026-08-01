// Worker — expira licencas ConfVision vencidas
query fp_vis_licenca_expirar_vencidas verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text worker_key? filters=trim
  }

  stack {
    function.run fn_fp_worker_validar {
      input = {worker_key: $input.worker_key}
    } as $worker_check
  
    db.query vis_licenca {
      where = $db.vis_licenca.valido_ate != null && $db.vis_licenca.valido_ate < now && ($db.vis_licenca.status == "disponivel" || $db.vis_licenca.status == "em_uso")
      return = {type: "list"}
    } as $vencidas
  
    var $expiradas {
      value = 0
    }
  
    var $franqueados_sync {
      value = {}
    }
  
    foreach ($vencidas) {
      each as $lic {
        db.patch vis_licenca {
          field_name = "id"
          field_value = $lic.id
          data = {status: "expirada"}
        } as $lic_upd
      
        conditional {
          if ($lic.status == "em_uso") {
            conditional {
              if ($lic.vis_camera_id != null) {
                db.patch vis_camera {
                  field_name = "id"
                  field_value = $lic.vis_camera_id
                  data = {ativo: false}
                } as $cam_upd
              }
            }
          }
        }
      
        conditional {
          if ($franqueados_sync|get:$lic.id_franqueado:null == null) {
            var.update $franqueados_sync {
              value = $franqueados_sync|set:$lic.id_franqueado:true
            }
          }
        }
      
        var.update $expiradas {
          value = $expiradas + 1
        }
      }
    }
  
    foreach ($franqueados_sync|keys) {
      each as $fra_id {
        function.run "" {
          input = {id_franqueado: $fra_id}
        } as $sync
      }
    }
  
    function.run fn_fp_financeiro_log {
      input = {
        acao    : "vis_licenca_expirar"
        ref_tipo: "vis_licenca"
        ref_id  : $expiradas|to_text
        detalhe : "licencas expiradas"
        origem  : "worker"
      }
    } as $log
  }

  response = {
    expiradas       : $expiradas
    franqueados_sync: $franqueados_sync|keys
  }
}