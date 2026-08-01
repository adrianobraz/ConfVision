// Garante que REP so age sobre franqueado da propria carteira. CEN sempre ok.
function fn_fp_admin_assert_escopo_franqueado {
  input {
    text admin_token? filters=trim
    text id_franqueado? filters=trim
    text id_representante_registro? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    conditional {
      if ($admin.userTipo == "CEN") {
        var $ok {
          value = true
        }
      }
    
      elseif ($admin.userTipo == "REP") {
        var $id_rep {
          value = $admin|get:"idVinculo":($admin|get:"idRepresentante":"")|trim
        }
      
        precondition (($id_rep|is_empty) == false) {
          error = "Sessao Representante sem idVinculo — faca logout e login novamente"
        }
      
        var $ok {
          value = false
        }
      
        var $id_fra {
          value = $input.id_franqueado|trim
        }
      
        // Atalho: registro ja tem id_representante
        conditional {
          if (($input.id_representante_registro|is_empty) == false) {
            precondition (($input.id_representante_registro|trim) == $id_rep) {
              error = "Fora da carteira deste Representante"
            }
          
            var.update $ok {
              value = true
            }
          }
        }
      
        conditional {
          if ($ok == false) {
            precondition (($id_fra|is_empty) == false) {
              error = "id_franqueado obrigatorio para validar escopo"
            }
          
            // 1) Carteira oficial (mesmo endpoint da lista do combo)
            function.run fn_fp_franqueados_ids_representante {
              input = {id_representante: $id_rep}
            } as $carteira
          
            var $ids {
              value = $carteira|get:"ids":[]
            }
          
            foreach ($ids) {
              each as $idc {
                conditional {
                  if (($idc|trim) == $id_fra) {
                    var.update $ok {
                      value = true
                    }
                  }
                }
              }
            }
          }
        }
      
        // 2) Fallback: ID_Representante do franqueado na API legada
        conditional {
          if ($ok == false) {
            function.run fn_fp_franqueado_id_representante {
              input = {id_franqueado: $id_fra}
            } as $rep_fra
          
            var $rep_do_fra {
              value = $rep_fra|get:"id_representante":""|trim
            }
          
            conditional {
              if (($rep_do_fra|is_empty) == false && $rep_do_fra == $id_rep) {
                var.update $ok {
                  value = true
                }
              }
            }
          }
        }
      
        precondition ($ok == true) {
          error = "Franqueado fora da carteira deste Representante"
        }
      }
    
      else {
        precondition (false) {
          error = "Sem permissao"
        }
      }
    }
  }

  response = {
    ok       : true
    userTipo : $admin.userTipo
    idVinculo: $admin|get:"idVinculo":""
  }
}
