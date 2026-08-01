// Valida cupom (preview) — FranqueadoPro. So cupom REP→FRA; nao fura piso.
query fp_cupom_validar verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
    text codigo? filters=trim
    text produto?=franqueadopro filters=trim
    decimal valor_base?
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    var $piso {
      value = 0
    }
  
    conditional {
      if ($input.valor_base != null && $input.valor_base > 0 && (($input.produto|is_empty) == false)) {
        // Sem plano no input: piso aproximado via menor markup (catalogo efetivo)
        function.run fn_fp_franqueado_id_representante {
          input = {id_franqueado: $input.id_franqueado}
        } as $rep
      
        function.run fn_fp_catalogo_com_preco_rep {
          input = {
            id_representante: $rep.id_representante
            produto         : $input.produto
            ativo           : "S"
          }
        } as $cat
      
        conditional {
          if (($cat.dados|count) > 0) {
            var.update $piso {
              value = $cat.dados[0].valor_piso_central|first_notempty:0
            }
          }
        }
      }
    }
  
    function.run fn_fp_cupom_validar {
      input = {
        codigo            : $input.codigo
        id_franqueado     : $input.id_franqueado
        produto           : $input.produto
        valor_base        : $input.valor_base
        valor_piso_minimo : $piso
        contexto          : "franqueado"
      }
    } as $resultado
  }

  response = $resultado
}
