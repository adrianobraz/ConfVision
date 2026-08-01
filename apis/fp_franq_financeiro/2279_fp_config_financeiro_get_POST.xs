// Config financeiro
query fp_config_financeiro_get verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text chave? filters=trim
  }

  stack {
    var $resultado {
      value = null
    }
  
    conditional {
      if (($input.chave|is_empty) == false) {
        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = $input.chave
        } as $item
      
        var.update $resultado {
          value = $item
        }
      }
    
      else {
        db.query fp_config_financeiro {
          sort = {fp_config_financeiro.chave: "asc"}
          return = {type: "list"}
        } as $lista
      
        var.update $resultado {
          value = {dados: $lista}
        }
      }
    }
  }

  response = $resultado
}