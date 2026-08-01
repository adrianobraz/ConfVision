// Lista cupons no escopo: CEN ve da propria Central; REP ve os seus + da Central
query fp_cupom_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text status? filters=trim
    text produto? filters=trim
    text codigo? filters=trim
    text id_franqueado? filters=trim
    int limite?=200 filters=min:1|max:500
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    var $codigo {
      value = $input.codigo
        |first_notempty:""
        |trim
        |to_upper
    }
  
    var $id_cen {
      value = $escopo|get:"idCentral":""
    }
  
    db.query fp_cupom_desconto {
      where = ($db.fp_cupom_desconto.id > 0) && ($db.fp_cupom_desconto.status ==? $input.status) && ($db.fp_cupom_desconto.produto ==? $input.produto) && ($db.fp_cupom_desconto.codigo ==? $codigo) && ($db.fp_cupom_desconto.id_franqueado ==? $input.id_franqueado)
      sort = {fp_cupom_desconto.created_at: "desc"}
      return = {type: "list"}
    } as $lista_raw
  
    var $lista {
      value = []
    }
  
    conditional {
      if ($escopo|get:"permite_global":false) {
        var.update $lista {
          value = $lista_raw
        }
      }
    
      elseif (($escopo|get:"userTipo":"") == "CEN") {
        foreach ($lista_raw) {
          each as $c {
            conditional {
              if (($c.id_central|trim) == $id_cen) {
                array.push $lista {
                  value = $c
                }
              }
            
              elseif (($c.id_central|is_empty) && ($c.criado_por|trim) == ($escopo|get:"usuario":""|trim)) {
                array.push $lista {
                  value = $c
                }
              }
            }
          }
        }
      }
    
      else {
        foreach ($lista_raw) {
          each as $c {
            var $incluir {
              value = false
            }
          
            var $mesma_central {
              value = (($c.id_central|is_empty) || (($id_cen|is_empty) == false && ($c.id_central|trim) == $id_cen))
            }
          
            conditional {
              if ($mesma_central && $c.criado_por_tipo == "REP" && $c.id_representante == $escopo|get:"idRepresentante":"") {
                var.update $incluir {
                  value = true
                }
              }
            
              elseif ($mesma_central && $c.criado_por_tipo == "CEN" && $c.alvo_nivel == "REP" && (($c.id_representante|is_empty) || $c.id_representante == $escopo|get:"idRepresentante":"")) {
                var.update $incluir {
                  value = true
                }
              }
            
              elseif ($mesma_central && ($c.criado_por_tipo|is_empty) && (($c.id_representante|is_empty) || $c.id_representante == $escopo|get:"idRepresentante":"")) {
                var.update $incluir {
                  value = true
                }
              }
            }
          
            conditional {
              if ($incluir) {
                array.push $lista {
                  value = $c
                }
              }
            }
          }
        }
      }
    }
  }

  response = {
    dados    : $lista
    total    : $lista|count
    userTipo : $escopo|get:"userTipo":""
    idCentral: $id_cen
  }
}
