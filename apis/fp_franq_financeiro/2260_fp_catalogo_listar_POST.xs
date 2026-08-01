// Lista catalogo — CEN na propria Central; Break-glass pode filtrar id_central
query fp_catalogo_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text produto? filters=trim
    text ativo? filters=trim
    text id_central? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    precondition ($admin.userTipo == "CEN") {
      error = "Representante deve usar fp_preco_rep_listar"
    }
  
    var $id_central {
      value = $admin.idCentral|first_notempty:($admin.idCentralCatalogo|first_notempty:($admin.idVinculo|trim))

    }
  
    conditional {
      if ($admin.breakglass == true) {
        conditional {
          if (($input.id_central|is_empty) == false) {
            var.update $id_central {
              value = $input.id_central|trim
            }
          }
        
          elseif (($id_central|is_empty) == true) {
            var.update $id_central {
              value = "CENTRAL"
            }
          }
        }
      }
    
      else {
        precondition (($id_central|is_empty) == false) {
          error = "Sessao sem Central vinculada — faca logout e login novamente (API Go precisa devolver idCentralCatalogo)"
        }
      }
    }
  
    function.run fn_fp_central_modo_preco {
      input = {id_central: $id_central}
    } as $modo_res
  
    db.query fp_produto_catalogo {
      where = $db.fp_produto_catalogo.id_central == $id_central && $db.fp_produto_catalogo.id_representante == "" && $db.fp_produto_catalogo.produto ==? $input.produto && $db.fp_produto_catalogo.ativo ==? $input.ativo
      sort = {
        fp_produto_catalogo.produto     : "asc"
        fp_produto_catalogo.valor_mensal: "asc"
      }
    
      return = {type: "list"}
    } as $lista_raw
  
    var $lista {
      value = []
    }
  
    foreach ($lista_raw) {
      each as $item {
        var $row {
          value = $item
            |set:"modo_preco":$modo_res.modo_preco
            |set:"valor_piso_minimo":($item.valor_piso_breakglass|first_notempty:0)
        }
      
        var $cota_id_item {
          value = $item|get:"fp_pacote_cota_id":0|first_notempty:0
        }
      
        conditional {
          if ($cota_id_item > 0) {
            db.get fp_pacote_cota {
              field_name = "id"
              field_value = $cota_id_item
            } as $pacote_row
          
            conditional {
              if ($pacote_row != null) {
                var $cota_valor {
                  value = $pacote_row|get:"valor":0|first_notempty:0
                }
              
                var $lic_valor {
                  value = $item.valor_mensal|first_notempty:0
                }
              
                var.update $row {
                  value = $row
                    |set:"pacote_cota_nome":($pacote_row|get:"nome":"")
                    |set:"pacote_cota_quantidade":($pacote_row|get:"quantidade":0)
                    |set:"valor_cota":$cota_valor
                    |set:"valor_total_mensal":($lic_valor + $cota_valor)
                }
              }
            }
          }
        }
      

        var.update $lista {
          value = $lista|push:$row
        }
      }
    }
  }

  response = {
    dados      : $lista
    total      : $lista|count
    id_central : $id_central
    modo_preco : $modo_res.modo_preco
    breakglass : $admin.breakglass
  }
}
