// Cria/edita pacote de cota (CEN/BG). Quantidade e valor piso.
query fp_pacote_cota_salvar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    int id?
    text nome? filters=trim
    int quantidade?
    decimal valor?
    text ativo?=S filters=trim
    text observacao? filters=trim
    text id_central? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    precondition ($admin.userTipo == "CEN") {
      error = "Somente a Central (ou Break-glass) pode cadastrar Pacotes de Cotas"
    }
  
    var $id_central {
      value = $admin.idVinculo|trim
    }
  
    conditional {
      if ($admin.breakglass) {
        conditional {
          if (($input.id_central|is_empty) == false) {
            var.update $id_central {
              value = $input.id_central|trim
            }
          }
        
          elseif ($id_central|is_empty) {
            var.update $id_central {
              value = "CENTRAL"
            }
          }
        }
      }
    
      else {
        precondition (($id_central|is_empty) == false) {
          error = "Sessao sem Central vinculada"
        }
      }
    }
  
    precondition (($input.nome|is_empty) == false) {
      error = "nome obrigatorio"
    }
  
    precondition ($input.quantidade != null && $input.quantidade > 0) {
      error = "quantidade deve ser maior que zero"
    }
  
    var $valor {
      value = $input.valor|first_notempty:0
    }
  
    precondition ($valor >= 0) {
      error = "valor invalido"
    }
  
    var $salvo {
      value = null
    }
  
    conditional {
      if ($input.id != null && $input.id > 0) {
        db.get fp_pacote_cota {
          field_name = "id"
          field_value = $input.id
        } as $atual
      
        precondition ($atual != null && $atual.id_central == $id_central) {
          error = "Pacote nao encontrado nesta Central"
        }
      
        db.patch fp_pacote_cota {
          field_name = "id"
          field_value = $input.id
          data = {
            nome      : $input.nome
            quantidade: $input.quantidade
            valor     : $valor
            ativo     : $input.ativo|first_notempty:"S"
            observacao: $input.observacao
          }
        } as $salvo
      }
    
      else {
        db.add fp_pacote_cota {
          data = {
            created_at: "now"
            id_central: $id_central
            nome      : $input.nome
            quantidade: $input.quantidade
            valor     : $valor
            ativo     : $input.ativo|first_notempty:"S"
            observacao: $input.observacao
          }
        } as $salvo
      }
    }
  
    var $qtd_salva {
      value = $salvo|get:"quantidade":0|first_notempty:0
    }
  
    function.run fn_fp_cota_limites_de_quantidade {
      input = {quantidade: $qtd_salva}
    } as $lim
  
    var $lim_json {
      value = $lim|get:"limites_json":{}
    }
  }

  response = $salvo
    |set:"limites_json":$lim_json
    |set:"clientes":($lim_json|get:"clientes_max":$qtd_salva)
    |set:"dispositivos":($lim_json|get:"contas_max":($qtd_salva * 2))
    |set:"usuarios_alarme":($lim_json|get:"usuarios_alarme_max":($qtd_salva * 4))
    |set:"setores_alarme":($lim_json|get:"setores_alarme_max":($qtd_salva * 10))
}
