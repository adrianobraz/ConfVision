// REP salva markup (valor_venda >= piso). CEN nao usa — edita o catalogo.
function fn_fp_preco_rep_salvar {
  input {
    text id_representante? filters=trim
    int fp_produto_catalogo_id?
    decimal valor_venda?
    text observacao? filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    precondition (($input.id_representante|is_empty) == false) {
      error = "id_representante obrigatorio"
    }
  
    precondition ($input.fp_produto_catalogo_id != null && $input.fp_produto_catalogo_id > 0) {
      error = "fp_produto_catalogo_id obrigatorio"
    }
  
    precondition ($input.valor_venda != null && $input.valor_venda > 0) {
      error = "valor_venda deve ser maior que zero"
    }
  
    db.get fp_produto_catalogo {
      field_name = "id"
      field_value = $input.fp_produto_catalogo_id
    } as $cat
  
    precondition ($cat != null) {
      error = "Item do catalogo nao encontrado"
    }
  
    precondition ($cat.id_representante == "") {
      error = "Preco de venda deve referenciar item do catalogo piso da Central"
    }
  
    var $piso {
      value = $cat.valor_mensal|first_notempty:0
    }
  
    precondition ($input.valor_venda >= $piso) {
      error = "Valor de venda nao pode ser menor que o piso da Central (R$ " ~ ($piso|to_text) ~ ")"
    }
  
    db.query fp_preco_representante {
      where = $db.fp_preco_representante.id_representante == $input.id_representante && $db.fp_preco_representante.fp_produto_catalogo_id == $cat.id
      return = {type: "single"}
    } as $existe
  
    var $model {
      value = null
    }
  
    conditional {
      if ($existe != null) {
        db.patch fp_preco_representante {
          field_name = "id"
          field_value = $existe.id
          data = {
            produto      : $cat.produto
            plano        : $cat.plano
            valor_venda  : $input.valor_venda
            ativo        : "S"
            observacao   : $input.observacao
            atualizado_em: "now"
          }
        } as $model
      }
    
      else {
        db.add fp_preco_representante {
          data = {
            created_at            : "now"
            id_representante      : $input.id_representante
            fp_produto_catalogo_id: $cat.id
            produto               : $cat.produto
            plano                 : $cat.plano
            valor_venda           : $input.valor_venda
            ativo                 : "S"
            observacao            : $input.observacao
            atualizado_em         : "now"
          }
        } as $model
      }
    }
  
    function.run fn_fp_financeiro_log {
      input = {
        acao         : "preco_rep_salvar"
        id_franqueado: ""
        ref_tipo     : "fp_preco_representante"
        ref_id       : $model.id|to_text
        detalhe      : "rep=" ~ $input.id_representante ~ " " ~ $cat.produto ~ "/" ~ $cat.plano ~ " venda=" ~ ($input.valor_venda|to_text) ~ " piso=" ~ ($piso|to_text)
        origem       : "admin"
        valor        : $input.valor_venda
        produto      : $cat.produto
        plano        : $cat.plano
        admin_usuario: $input.admin_usuario
      }
    } as $log
  }

  response = {
    preco              : $model
    valor_piso_central : $piso
    margem_rep         : $input.valor_venda - $piso
    registro_financeiro: $log
  }
}