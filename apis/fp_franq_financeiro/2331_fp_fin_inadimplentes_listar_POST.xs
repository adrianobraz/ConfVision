// Lista faturas inadimplentes (abertas e vencidas) — escopo CEN/REP
query fp_fin_inadimplentes_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    db.query fp_fatura {
      where = $db.fp_fatura.status == "aberta" && $db.fp_fatura.vencimento_em != null && $db.fp_fatura.vencimento_em < now
      sort = {fp_fatura.vencimento_em: "asc"}
      return = {type: "list"}
    } as $faturas_raw
  
    var $faturas {
      value = []
    }
  
    conditional {
      if (($escopo|get:"userTipo":"") == "REP") {
        function.run fn_fp_franqueados_ids_representante {
          input = {id_representante: $escopo|get:"idRepresentante":""}
        } as $carteira
      
        var $ids {
          value = $carteira.ids|first_notempty:[]
        }
      
        foreach ($faturas_raw) {
          each as $f {
            function.run fn_fp_fatura_na_carteira {
              input = {
                fatura          : $f
                id_representante: $escopo|get:"idRepresentante":""
                ids_franqueado  : $ids
              }
            } as $chk
          
            conditional {
              if ($chk.ok) {
                array.push $faturas {
                  value = $f
                }
              }
            }
          }
        }
      }
    
      elseif ($escopo|get:"permite_global":false) {
        var.update $faturas {
          value = $faturas_raw
        }
      }
    
      else {
        foreach ($faturas_raw) {
          each as $f {
            conditional {
              if (($f.id_central|trim) == ($escopo|get:"idCentral":""|trim)) {
                array.push $faturas {
                  value = $f
                }
              }
            }
          }
        }
      }
    }
  
    var $total_valor {
      value = 0
    }
  
    var $franqueados {
      value = []
    }
  
    foreach ($faturas) {
      each as $f {
        var.update $total_valor {
          value = $total_valor + ($f.valor_total|first_notempty:0)
        }
      
        var $ja {
          value = false
        }
      
        foreach ($franqueados) {
          each as $idf {
            conditional {
              if ($idf == $f.id_franqueado) {
                var.update $ja {
                  value = true
                }
              }
            }
          }
        }
      
        conditional {
          if ($ja == false && ($f.id_franqueado|strlen) > 0) {
            var.update $franqueados {
              value = $franqueados|push:$f.id_franqueado
            }
          }
        }
      }
    }
  }

  response = {
    dados            : $faturas
    total_faturas    : $faturas|count
    total_franqueados: $franqueados|count
    total_valor      : $total_valor
    idCentral        : $escopo|get:"idCentral":""
  }
}
