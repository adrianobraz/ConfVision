// Resumo: o que sobra para Break-glass e margem da Central (Centrais em modo piso)
query fp_margem_breakglass_resumo verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    precondition ($admin.breakglass) {
      error = "Somente Break-glass pode ver resumo de margem"
    }
  
    db.query fp_central_preco_config {
      where = $db.fp_central_preco_config.modo_preco == "piso"
      return = {type: "list"}
    } as $centrais_piso
  
    var $por_central {
      value = []
    }
  
    var $total_piso_bg {
      value = 0
    }
  
    var $total_margem_cen {
      value = 0
    }
  
    foreach ($centrais_piso) {
      each as $cfg {
        var $idc {
          value = $cfg.id_central|first_notempty:""
        }
      
        db.query fp_produto_catalogo {
          where = $db.fp_produto_catalogo.id_central == $idc && $db.fp_produto_catalogo.id_representante == "" && $db.fp_produto_catalogo.ativo == "S"
          return = {type: "list"}
        } as $itens
      
        var $itens_out {
          value = []
        }
      
        var $soma_bg {
          value = 0
        }
      
        var $soma_mc {
          value = 0
        }
      
        foreach ($itens) {
          each as $item {
            function.run fn_fp_piso_global_get {
              input = {produto: $item.produto, plano: $item.plano}
            } as $piso_g
          
            var $piso_bg {
              value = $piso_g.valor_minimo|first_notempty:0
            }
          
            var $piso_cen {
              value = $item.valor_mensal|first_notempty:0
            }
          
            var $margem_cen {
              value = $piso_cen - $piso_bg
            }
          
            conditional {
              if ($margem_cen < 0) {
                var.update $margem_cen {
                  value = 0
                }
              }
            }
          
            var.update $soma_bg {
              value = $soma_bg + $piso_bg
            }
          
            var.update $soma_mc {
              value = $soma_mc + $margem_cen
            }
          
            var.update $itens_out {
              value = $itens_out
                |push:```
                  {
                    catalogo_id           : $item.id
                    produto               : $item.produto
                    plano                 : $item.plano
                    nome_exibicao         : $item.nome_exibicao
                    valor_piso_central    : $piso_cen
                    valor_piso_breakglass : $piso_bg
                    margem_central        : $margem_cen
                  }
                  ```
            }
          }
        }
      
        var.update $total_piso_bg {
          value = $total_piso_bg + $soma_bg
        }
      
        var.update $total_margem_cen {
          value = $total_margem_cen + $soma_mc
        }
      
        var.update $por_central {
          value = $por_central
            |push:```
              {
                id_central            : $idc
                modo_preco            : "piso"
                total_piso_breakglass : $soma_bg
                total_margem_central  : $soma_mc
                itens                 : $itens_out
              }
              ```
        }
      }
    }
  
    db.query fp_fatura {
      where = $db.fp_fatura.tipo == "repasse_central_breakglass" && $db.fp_fatura.status == "aberta"
      return = {type: "list"}
    } as $repasses_abertos
  
    var $a_receber {
      value = 0
    }
  
    foreach ($repasses_abertos) {
      each as $r {
        var.update $a_receber {
          value = $a_receber + ($r.valor_total|first_notempty:0)
        }
      }
    }
  }

  response = {
    por_central            : $por_central
    total_piso_breakglass  : $total_piso_bg
    total_margem_central   : $total_margem_cen
    repasses_abertos       : $repasses_abertos
    total_a_receber_abertos: $a_receber
  }
}