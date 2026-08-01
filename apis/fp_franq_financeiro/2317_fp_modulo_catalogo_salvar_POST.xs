// Salvar modulo a la carte — mesma regra de piso Break-glass
query fp_modulo_catalogo_salvar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    int id?
    text produto? filters=trim
    text chave? filters=trim
    text label? filters=trim
    text grupo? filters=trim
    text descricao? filters=trim
    text plano_minimo? filters=trim
    decimal valor_mensal?
    text ativo? filters=trim
    int ordem?
    text id_central? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    precondition ($admin.userTipo == "CEN") {
      error = "Somente a Central pode alterar modulos a la carte"
    }
  
    var $id_central {
      value = $admin.idVinculo|trim
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
          error = "Sessao sem Central vinculada — faca logout e login novamente"
        }
      }
    }
  
    precondition (($input.produto|is_empty) == false && ($input.chave|is_empty) == false) {
      error = "produto e chave obrigatorios"
    }
  
    var $valor {
      value = $input.valor_mensal|first_notempty:0
    }
  
    function.run fn_fp_central_modo_preco {
      input = {id_central: $id_central}
    } as $modo_res
  
    var $piso_bg {
      value = 0
    }
  
    conditional {
      if ($input.id != null && $input.id > 0) {
        db.get fp_modulo_catalogo {
          field_name = "id"
          field_value = $input.id
        } as $atual
      
        precondition ($atual != null && $atual.id_central == $id_central && $atual.id_representante == "") {
          error = "Modulo fora do escopo da Central"
        }
      
        var.update $piso_bg {
          value = $atual.valor_piso_breakglass|first_notempty:0
        }
      
        conditional {
          if ($admin.breakglass == false && $modo_res.modo_preco == "piso") {
            precondition ($valor >= $piso_bg) {
              error = "Central em modo piso: valor minimo e R$ " ~ ($piso_bg|to_text)
            }
          }
        }
      
        conditional {
          if ($admin.breakglass == true) {
            var.update $piso_bg {
              value = $valor
            }
          }
        }
      
        db.patch fp_modulo_catalogo {
          field_name = "id"
          field_value = $input.id
          data = {
            produto               : $input.produto
            chave                 : $input.chave
            label                 : $input.label
            grupo                 : $input.grupo
            descricao             : $input.descricao
            plano_minimo          : $input.plano_minimo|first_notempty:"lite"
            valor_mensal          : $valor
            valor_piso_breakglass : $piso_bg
            ativo                 : $input.ativo|first_notempty:"S"
            ordem                 : $input.ordem
          }
        } as $model
      }
    
      else {
        conditional {
          if ($admin.breakglass == true || $modo_res.modo_preco == "piso") {
            var.update $piso_bg {
              value = $valor
            }
          }
        }
      
        db.add fp_modulo_catalogo {
          data = {
            created_at            : "now"
            id_central            : $id_central
            id_representante      : ""
            produto               : $input.produto
            chave                 : $input.chave
            label                 : $input.label
            grupo                 : $input.grupo
            descricao             : $input.descricao
            plano_minimo          : $input.plano_minimo|first_notempty:"lite"
            valor_mensal          : $valor
            valor_piso_breakglass : $piso_bg
            ativo                 : $input.ativo|first_notempty:"S"
            ordem                 : $input.ordem
          }
        } as $model
      }
    }
  }

  response = $model|set:"modo_preco":$modo_res.modo_preco|set:"valor_piso_minimo":$piso_bg
}
