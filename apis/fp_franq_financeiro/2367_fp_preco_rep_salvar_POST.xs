// Representante salva valor de venda (>= piso Central)
query fp_preco_rep_salvar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text id_representante? filters=trim
    int fp_produto_catalogo_id?
    decimal valor_venda?
    text observacao? filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    precondition ($admin.userTipo == "REP" || $admin.userTipo == "CEN") {
      error = "Sem permissao para salvar preco de representante"
    }
  
    var $id_rep {
      value = $input.id_representante|first_notempty:""
    }
  
    conditional {
      if ($admin.userTipo == "REP") {
        var.update $id_rep {
          value = $admin.idVinculo
        }
      }
    }
  
    precondition (($id_rep|is_empty) == false) {
      error = "id_representante obrigatorio"
    }
  
    // CEN pode ajustar markup de um REP especifico; REP so o proprio
    conditional {
      if ($admin.userTipo == "REP" && $id_rep != $admin.idVinculo) {
        precondition (false) {
          error = "Representante so pode alterar o proprio preco"
        }
      }
    }
  
    function.run fn_fp_preco_rep_salvar {
      input = {
        id_representante       : $id_rep
        fp_produto_catalogo_id : $input.fp_produto_catalogo_id
        valor_venda            : $input.valor_venda
        observacao             : $input.observacao
        admin_usuario          : $input.admin_usuario|first_notempty:$admin.usuario
      }
    } as $resultado
  }

  response = $resultado
}

