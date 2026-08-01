// True se assinatura pertence a carteira do representante
function fn_fp_assinatura_na_carteira {
  input {
    json assinatura?
    text id_representante? filters=trim
    json ids_franqueado?
  }

  stack {
    var $ok {
      value = false
    }
  
    var $a {
      value = $input.assinatura
    }
  
    var $id_rep {
      value = $input.id_representante|first_notempty:""
    }
  
    conditional {
      if ($a != null && (($id_rep|is_empty) == false)) {
        conditional {
          if ($a.id_representante == $id_rep) {
            var.update $ok {
              value = true
            }
          }
        
          else {
            var $ids {
              value = $input.ids_franqueado|first_notempty:[]
            }
          
            foreach ($ids) {
              each as $id_fra {
                conditional {
                  if ($a.id_franqueado == $id_fra) {
                    var.update $ok {
                      value = true
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }

  response = {ok: $ok}
}