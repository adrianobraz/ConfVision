// Seed idempotente dos precos de venda do Representante (markup inicial = piso da Central)
function fn_fp_preco_rep_seed {
  input {
    text id_representante? filters=trim
    text id_central?=CENTRAL filters=trim
    text produto? filters=trim
    text escopo? filters=trim
  }

  stack {
    precondition (($input.id_representante|is_empty) == false) {
      error = "id_representante obrigatorio"
    }
  
    var $id_central {
      value = $input.id_central|first_notempty:"CENTRAL"
    }
  
    db.query fp_produto_catalogo {
      where = $db.fp_produto_catalogo.id_central == $id_central && $db.fp_produto_catalogo.id_representante == "" && $db.fp_produto_catalogo.produto ==? $input.produto && $db.fp_produto_catalogo.ativo == "S"
      sort = {
        fp_produto_catalogo.produto     : "asc"
        fp_produto_catalogo.valor_mensal: "asc"
      }
    
      return = {type: "list"}
    } as $catalogo_cen
  
    precondition (($catalogo_cen|count) > 0) {
      error = "Catalogo da Central ainda nao existe para este modulo. Faca login como Central e rode o seed correspondente primeiro."
    }
  
    var $inseridos {
      value = 0
    }
  
    var $escopo {
      value = $input.escopo|first_notempty:""|to_lower
    }
  
    foreach ($catalogo_cen) {
      each as $cat {
        var $incluir {
          value = true
        }
      
        conditional {
          if ($escopo == "confvision_licenca" && $cat.produto != "confvision_licenca") {
            var.update $incluir {
              value = false
            }
          }
        
          elseif ($escopo == "assinaturas" && $cat.produto == "confvision_licenca") {
            var.update $incluir {
              value = false
            }
          }
        }
      
        conditional {
          if ($incluir) {
            var $piso {
              value = $cat.valor_mensal|first_notempty:0
            }
          
            db.query fp_preco_representante {
              where = $db.fp_preco_representante.id_representante == $input.id_representante && $db.fp_preco_representante.fp_produto_catalogo_id == $cat.id
              return = {type: "single"}
            } as $existe
          
            conditional {
              if ($existe == null) {
                db.add fp_preco_representante {
                  data = {
                    created_at            : "now"
                    id_representante      : $input.id_representante
                    fp_produto_catalogo_id: $cat.id
                    produto               : $cat.produto
                    plano                 : $cat.plano
                    valor_venda           : $piso
                    ativo                 : "S"
                    observacao            : "Seed automatico (valor = piso Central)"
                    atualizado_em         : "now"
                  }
                } as $novo_preco
              
                var.update $inseridos {
                  value = $inseridos + 1
                }
              }
            }
          }
        }
      }
    }
  }

  response = {
    inseridos       : $inseridos
    total_catalogo  : $catalogo_cen|count
    id_representante: $input.id_representante
    id_central      : $id_central
    produto         : $input.produto|first_notempty:""
    escopo          : $escopo
    status          : "OK"
  }
}