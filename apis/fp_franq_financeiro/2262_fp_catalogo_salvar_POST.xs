// Salvar item do catalogo. Break-glass define piso; Central em modo piso nao baixa.
query fp_catalogo_salvar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    int id?
    text produto? filters=trim
    text plano? filters=trim
    text nome_exibicao? filters=trim
    decimal valor_mensal?
    text periodicidade_padrao? filters=trim
    int retencao_dias?
    json limites_json?
    json modulos_json?
    int fp_pacote_cota_id?
    text ativo? filters=trim
    text observacao? filters=trim
    text id_central? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    precondition ($admin.userTipo == "CEN") {
      error = "Somente a Central pode alterar o preco piso do catalogo"
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
  
    precondition (($input.produto|is_empty) == false && ($input.plano|is_empty) == false) {
      error = "produto e plano obrigatorios"
    }
  
    var $cota_id {
      value = null
    }
  
    conditional {
      if ($input.produto == "franqueadopro" && $input.fp_pacote_cota_id != null && $input.fp_pacote_cota_id > 0) {
        db.get fp_pacote_cota {
          field_name = "id"
          field_value = $input.fp_pacote_cota_id
        } as $pacote_vinc
      
        precondition ($pacote_vinc != null && $pacote_vinc.ativo == "S" && $pacote_vinc.id_central == $id_central) {
          error = "Pacote de Cotas invalido para esta Central"
        }
      
        var.update $cota_id {
          value = $input.fp_pacote_cota_id
        }
      }
    
      elseif ($input.produto != "franqueadopro") {
        var.update $cota_id {
          value = null
        }
      }
    
      elseif ($input.id != null && $input.id > 0) {
        // Mantem vinculo atual se nao enviado
        db.get fp_produto_catalogo {
          field_name = "id"
          field_value = $input.id
        } as $cat_atual_cota
      
        conditional {
          if ($cat_atual_cota != null) {
            var.update $cota_id {
              value = $cat_atual_cota|get:"fp_pacote_cota_id":null
            }
          }
        }
      }
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
        db.get fp_produto_catalogo {
          field_name = "id"
          field_value = $input.id
        } as $atual
      
        precondition ($atual != null && $atual.id_central == $id_central && $atual.id_representante == "") {
          error = "Item do catalogo fora do escopo da Central"
        }
      
        var.update $piso_bg {
          value = $atual.valor_piso_breakglass|first_notempty:0
        }
      
        // Central (nao BG) em modo piso: nao pode baixar abaixo do piso BG
        conditional {
          if ($admin.breakglass == false && $modo_res.modo_preco == "piso") {
            precondition ($valor >= $piso_bg) {
              error = "Central em modo piso: valor minimo e R$ " ~ ($piso_bg|to_text)
            }
          }
        }
      
        // Break-glass ao editar: redefine o piso BG
        conditional {
          if ($admin.breakglass == true) {
            var.update $piso_bg {
              value = $valor
            }
          }
        }
      
        db.patch fp_produto_catalogo {
          field_name = "id"
          field_value = $input.id
          data = {
            produto               : $input.produto
            plano                 : $input.plano
            nome_exibicao         : $input.nome_exibicao
            valor_mensal          : $valor
            valor_piso_breakglass : $piso_bg
            periodicidade_padrao  : $input.periodicidade_padrao
            retencao_dias         : $input.retencao_dias
            limites_json          : $input.limites_json
            modulos_json          : $input.modulos_json
            fp_pacote_cota_id     : $cota_id
            ativo                 : $input.ativo|first_notempty:"S"
            observacao            : $input.observacao
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
      
        db.add fp_produto_catalogo {
          data = {
            created_at             : "now"
            id_central             : $id_central
            id_representante       : ""
            produto                : $input.produto
            plano                  : $input.plano
            nome_exibicao          : $input.nome_exibicao
            valor_mensal           : $valor
            valor_piso_breakglass  : $piso_bg
            periodicidade_padrao   : $input.periodicidade_padrao|first_notempty:"mensal"
            retencao_dias          : $input.retencao_dias
            limites_json           : $input.limites_json
            modulos_json           : $input.modulos_json
            fp_pacote_cota_id      : $cota_id
            ativo                  : $input.ativo|first_notempty:"S"
            observacao             : $input.observacao

          }
        } as $model
      }
    }
  }

  response = $model
    |set:"modo_preco":$modo_res.modo_preco
    |set:"valor_piso_minimo":$piso_bg
}
