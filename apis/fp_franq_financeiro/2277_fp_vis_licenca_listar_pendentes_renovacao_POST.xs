// Worker — licencas em uso com renovacao em ate N dias
query fp_vis_licenca_listar_pendentes_renovacao verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text worker_key? filters=trim
    int dias_antecedencia?=5 filters=min:1
  }

  stack {
    function.run fn_fp_worker_validar {
      input = {worker_key: $input.worker_key}
    } as $worker_check
  
    var $limite {
      value = now
        |add_secs_to_timestamp:$input.dias_antecedencia * 86400
    }
  
    db.query vis_licenca {
      where = $db.vis_licenca.status == "em_uso" && $db.vis_licenca.valido_ate != null && $db.vis_licenca.valido_ate <= $limite && $db.vis_licenca.valido_ate >= now
      sort = {vis_licenca.valido_ate: "asc"}
      return = {type: "list"}
    } as $licencas
  
    var $por_franqueado {
      value = {}
    }
  
    foreach ($licencas) {
      each as $lic {
        var $ciclo_ref {
          value = `"VIS-" ~ $lic.id|to_text ~ "-" ~ ($lic.valido_ate|format_timestamp:"Ymd":"UTC")`
        }
      
        db.query fp_fatura {
          where = $db.fp_fatura.ciclo_ref == $ciclo_ref && $db.fp_fatura.status != "cancelada"
          return = {type: "single"}
        } as $fatura_existe
      
        conditional {
          if ($fatura_existe == null) {
            var $franq_lista {
              value = $por_franqueado|get:$lic.id_franqueado:null
            }
          
            var $entrada {
              value = {licenca: $lic, ciclo_ref: $ciclo_ref}
            }
          
            conditional {
              if ($franq_lista == null) {
                var.update $por_franqueado {
                  value = $por_franqueado
                    |set:$lic.id_franqueado:[$entrada]
                }
              }
            
              else {
                var.update $por_franqueado {
                  value = $por_franqueado
                    |set:$lic.id_franqueado:($franq_lista|push:$entrada)
                }
              }
            }
          }
        }
      }
    }
  }

  response = {agrupado: $por_franqueado}
}