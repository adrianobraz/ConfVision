// Lista catalogo com piso Central + preco de venda do Representante
query fp_preco_rep_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text produto? filters=trim
    text id_representante? filters=trim
    text ativo?=S filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    var $id_rep {
      value = $input.id_representante|first_notempty:""
    }
  
    // Catalogo = Central da sessao (CEN: idVinculo; REP: idCentral — nunca o ID do representante)
    var $id_central {
      value = $admin.idCentral|first_notempty:($admin.idCentralCatalogo|first_notempty:"")
    }
  
    conditional {
      if ($admin.userTipo == "CEN" && ($id_central|is_empty) == true) {
        var.update $id_central {
          value = $admin.idVinculo|trim
        }
      }
    }
  
    // REP: forca o proprio vinculo (nao pode ver markup de outro)
    conditional {
      if ($admin.userTipo == "REP") {
        var.update $id_rep {
          value = $admin.idVinculo
        }
      
        precondition (($id_central|is_empty) == false) {
          error = "Sessao Representante sem Central — faca logout e login novamente (API Go precisa devolver idCentralCatalogo)"
        }
      
        precondition (($id_rep|is_empty) == false) {
          error = "Sessao Representante sem idVinculo — faca logout e login novamente"
        }
      }
    }
  
    precondition (($id_central|is_empty) == false) {
      error = "Sessao sem Central vinculada — faca logout e login novamente"
    }
  
    // CEN sem id_representante: lista so piso (sem coluna de markup)
    conditional {
      if ($admin.userTipo == "CEN" && ($id_rep|is_empty)) {
        db.query fp_produto_catalogo {
          where = $db.fp_produto_catalogo.id_central == $id_central && $db.fp_produto_catalogo.id_representante == "" && $db.fp_produto_catalogo.produto ==? $input.produto && $db.fp_produto_catalogo.ativo ==? $input.ativo
          sort = {
            fp_produto_catalogo.produto     : "asc"
            fp_produto_catalogo.valor_mensal: "asc"
          }
        
          return = {type: "list"}
        } as $lista_piso
      
        var $saida_piso {
          value = []
        }
      
        foreach ($lista_piso) {
          each as $c {
            array.push $saida_piso {
              value = $c
                |set:"valor_piso_central":$c.valor_mensal
                |set:"valor_venda":$c.valor_mensal
                |set:"margem_rep":0
                |set:"tem_markup":false
            }
          }
        }
      
        var $resultado {
          value = {
            dados           : $saida_piso
            total           : $saida_piso|count
            id_representante: ""
            id_central      : $id_central
            userTipo        : $admin.userTipo
          }
        }
      }
    
      else {
        function.run fn_fp_catalogo_com_preco_rep {
          input = {
            id_central      : $id_central
            id_representante: $id_rep
            produto         : $input.produto
            ativo           : $input.ativo
          }
        } as $com_preco
      
        var $resultado {
          value = $com_preco|set:"userTipo":$admin.userTipo
        }
      }
    }
  }

  response = $resultado
}
