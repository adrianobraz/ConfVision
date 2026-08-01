// Worker — assinaturas com cobranca em ate N dias sem fatura do ciclo
query fp_assinatura_listar_pendentes_fatura verb=POST {
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
  
    db.query fp_assinatura_produto {
      where = $db.fp_assinatura_produto.status == "ativa" && $db.fp_assinatura_produto.proxima_cobranca_em != null && $db.fp_assinatura_produto.proxima_cobranca_em <= $limite
      sort = {fp_assinatura_produto.proxima_cobranca_em: "asc"}
      return = {type: "list"}
    } as $candidatas
  
    var $pendentes {
      value = []
    }
  
    foreach ($candidatas) {
      each as $item {
        var $ciclo_ref {
          value = `"ASS-" ~ $item.id|to_text ~ "-" ~ ($item.proxima_cobranca_em|format_timestamp:"Ymd":"UTC")`
        }
      
        db.query fp_fatura {
          where = $db.fp_fatura.id_franqueado == $item.id_franqueado && $db.fp_fatura.ciclo_ref == $ciclo_ref && $db.fp_fatura.status != "cancelada"
          return = {type: "single"}
        } as $fatura_existe
      
        conditional {
          if ($fatura_existe == null) {
            var.update $pendentes {
              value = $pendentes
                |push:```
                  {
                    assinatura: $item
                    ciclo_ref : $ciclo_ref
                  }
                  ```
            }
          }
        }
      }
    }
  }

  response = {dados: $pendentes, total: $pendentes|count}
}