// Lista catalogo de planos (publico — franqueado ve preco do seu Representante ou piso)
query fp_catalogo_listar_publico verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text produto?=franqueadopro filters=trim
    text id_franqueado? filters=trim
  }

  stack {
    var $resultado {
      value = null
    }
  
    conditional {
      if (($input.id_franqueado|is_empty) == false) {
        // Representante para markup; central com fallback rapido (sem travar no legado)
        function.run fn_fp_franqueado_id_representante {
          input = {id_franqueado: $input.id_franqueado}
        } as $rep
      
        var $id_cen {
          value = "CENTRAL"
        }
      
        try_catch {
          try {
            function.run fn_fp_franqueado_id_central {
              input = {id_franqueado: $input.id_franqueado}
            } as $cen_fra
          
            var.update $id_cen {
              value = $cen_fra.id_central|first_notempty:"CENTRAL"
            }
          }
        
          catch {
            var.update $id_cen {
              value = "CENTRAL"
            }
          }
        }
      
        function.run fn_fp_catalogo_com_preco_rep {
          input = {
            id_central      : $id_cen
            id_representante: $rep.id_representante
            produto         : $input.produto
            ativo           : "S"
          }
        } as $resultado
      }
    
      else {
        db.query fp_produto_catalogo {
          where = $db.fp_produto_catalogo.id_central == "CENTRAL" && $db.fp_produto_catalogo.id_representante == "" && $db.fp_produto_catalogo.produto == $input.produto && $db.fp_produto_catalogo.ativo == "S"
          sort = {fp_produto_catalogo.valor_mensal: "asc"}
          return = {type: "list"}
        } as $lista
      
        var.update $resultado {
          value = {dados: $lista, total: $lista|count}
        }
      }
    }
  }

  response = $resultado
}
