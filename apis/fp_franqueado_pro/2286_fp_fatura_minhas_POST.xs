// Faturas do franqueado (leitura)
query fp_fatura_minhas verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
    text status? filters=trim
    text tipo? filters=trim
    bool incluir_itens?
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    conditional {
      if (($input.status|is_empty) == false && ($input.tipo|is_empty) == false) {
        db.query fp_fatura {
          where = $db.fp_fatura.id_franqueado == $input.id_franqueado && $db.fp_fatura.status == $input.status && $db.fp_fatura.tipo == $input.tipo
          sort = {fp_fatura.vencimento_em: "desc"}
          return = {type: "list"}
        } as $lista
      }
    
      elseif (($input.status|is_empty) == false) {
        db.query fp_fatura {
          where = $db.fp_fatura.id_franqueado == $input.id_franqueado && $db.fp_fatura.status == $input.status
          sort = {fp_fatura.vencimento_em: "desc"}
          return = {type: "list"}
        } as $lista
      }
    
      elseif (($input.tipo|is_empty) == false) {
        db.query fp_fatura {
          where = $db.fp_fatura.id_franqueado == $input.id_franqueado && $db.fp_fatura.tipo == $input.tipo
          sort = {fp_fatura.vencimento_em: "desc"}
          return = {type: "list"}
        } as $lista
      }
    
      else {
        db.query fp_fatura {
          where = $db.fp_fatura.id_franqueado == $input.id_franqueado
          sort = {fp_fatura.vencimento_em: "desc"}
          return = {type: "list"}
        } as $lista
      }
    }

    var $dados {
      value = $lista
    }

    conditional {
      if ($input.incluir_itens == true) {
        function.run fn_fp_faturas_com_itens {
          input = {faturas: $lista}
        } as $com_itens

        var.update $dados {
          value = $com_itens.faturas|first_notempty:[]
        }
      }
    }
  }

  response = {dados: $dados, total: $dados|count}
}