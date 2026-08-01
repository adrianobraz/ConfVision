// Preco efetivo: valor_venda do REP se >= piso; senao piso (margem 0)
// Inclui split Break-glass quando Central esta em modo "piso"
function fn_fp_preco_efetivo {
  input {
    text id_representante? filters=trim
    text id_franqueado? filters=trim
    text id_central? filters=trim
    text produto? filters=trim
    text plano? filters=trim
    int fp_produto_catalogo_id?
  }

  stack {
    var $id_rep {
      value = $input.id_representante|first_notempty:""
    }
  
    var $id_cen {
      value = $input.id_central|trim
    }
  
    conditional {
      if (($id_rep|is_empty) && (($input.id_franqueado|is_empty) == false)) {
        function.run fn_fp_franqueado_id_representante {
          input = {id_franqueado: $input.id_franqueado}
        } as $rep_res
      
        var.update $id_rep {
          value = $rep_res.id_representante|first_notempty:""
        }
      }
    }
  
    // Sem assinatura previa: resolve Central do franqueado (senao catalogo falha)
    conditional {
      if (($id_cen|is_empty) && (($input.id_franqueado|is_empty) == false)) {
        function.run fn_fp_franqueado_id_central {
          input = {id_franqueado: $input.id_franqueado}
        } as $cen_res
      
        var.update $id_cen {
          value = $cen_res.id_central|first_notempty:"CENTRAL"|trim
        }
      
        conditional {
          if (($id_rep|is_empty) && (($cen_res.id_representante|is_empty) == false)) {
            var.update $id_rep {
              value = $cen_res.id_representante|trim
            }
          }
        }
      }
    }
  
    conditional {
      if ($id_cen|is_empty) {
        var.update $id_cen {
          value = "CENTRAL"
        }
      }
    }
  
    var $cat {
      value = null
    }
  
    conditional {
      if ($input.fp_produto_catalogo_id != null && $input.fp_produto_catalogo_id > 0) {
        db.get fp_produto_catalogo {
          field_name = "id"
          field_value = $input.fp_produto_catalogo_id
        } as $cat
      
        conditional {
          if ($cat != null && ($id_cen|is_empty)) {
            var.update $id_cen {
              value = $cat.id_central|trim
            }
          }
        }
      }
    
      else {
        function.run fn_fp_catalogo_get_por_plano {
          input = {
            produto         : $input.produto
            plano           : $input.plano
            id_central      : $id_cen
            id_representante: ""
          }
        } as $cat
      
        // UUID da Central sem seed: tenta catalogo legado CENTRAL (mesmo da listagem publica)
        conditional {
          if ($cat == null && $id_cen != "CENTRAL") {
            function.run fn_fp_catalogo_get_por_plano {
              input = {
                produto         : $input.produto
                plano           : $input.plano
                id_central      : "CENTRAL"
                id_representante: ""
              }
            } as $cat
          
            conditional {
              if ($cat != null) {
                var.update $id_cen {
                  value = "CENTRAL"
                }
              }
            }
          }
        }
      }
    }
  
    precondition ($cat != null) {
      error = "Plano nao encontrado no catalogo"
    }
  
    var $piso {
      value = $cat.valor_mensal|first_notempty:0
    }
  
    var $valor_venda {
      value = $piso
    }
  
    var $tem_markup {
      value = false
    }
  
    var $preco_rep_id {
      value = null
    }
  
    conditional {
      if (($id_rep|is_empty) == false) {
        db.query fp_preco_representante {
          where = $db.fp_preco_representante.id_representante == $id_rep && $db.fp_preco_representante.fp_produto_catalogo_id == $cat.id && $db.fp_preco_representante.ativo == "S"
          return = {type: "single"}
        } as $preco_rep
      
        conditional {
          if ($preco_rep == null) {
            db.query fp_preco_representante {
              where = $db.fp_preco_representante.id_representante == $id_rep && $db.fp_preco_representante.produto == $cat.produto && $db.fp_preco_representante.plano == $cat.plano && $db.fp_preco_representante.ativo == "S"
              return = {type: "single"}
            } as $preco_rep
          }
        }
      
        conditional {
          if ($preco_rep != null && $preco_rep.valor_venda != null && $preco_rep.valor_venda >= $piso) {
            var.update $valor_venda {
              value = $preco_rep.valor_venda
            }
          
            var.update $tem_markup {
              value = $preco_rep.valor_venda > $piso
            }
          
            var.update $preco_rep_id {
              value = $preco_rep.id
            }
          }
        }
      }
    }
  
    function.run fn_fp_split_precos {
      input = {
        id_central            : $cat.id_central|first_notempty:"CENTRAL"
        produto               : $cat.produto
        plano                 : $cat.plano
        valor_piso_central    : $piso
        valor_venda           : $valor_venda
        valor_piso_breakglass : $cat.valor_piso_breakglass
      }
    } as $split
  }

  response = {
    catalogo_id           : $cat.id
    produto               : $cat.produto
    plano                 : $cat.plano
    nome_exibicao         : $cat.nome_exibicao
    id_representante      : $id_rep
    id_central            : $split.id_central
    modo_preco            : $split.modo_preco
    valor_piso_central    : $split.valor_piso_central
    valor_piso_breakglass : $split.valor_piso_breakglass
    margem_central        : $split.margem_central
    valor_venda           : $split.valor_venda
    margem_rep            : $split.margem_rep
    tem_markup            : $tem_markup
    preco_rep_id          : $preco_rep_id
    catalogo              : $cat
  }
}
