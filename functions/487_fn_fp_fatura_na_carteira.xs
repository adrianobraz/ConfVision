// True se fatura pertence a carteira do representante
function fn_fp_fatura_na_carteira {
  input {
    json fatura?
    text id_representante? filters=trim
    json ids_franqueado?
  }

  stack {
    var $ok {
      value = false
    }
  
    var $id_rep {
      value = $input.id_representante|first_notempty:""
    }
  
    var $f {
      value = $input.fatura
    }
  
    conditional {
      if ($f != null && (($id_rep|is_empty) == false)) {
        conditional {
          if ($f.id_representante == $id_rep) {
            var.update $ok {
              value = true
            }
          }
        
          elseif ($f.tipo == "repasse_rep_central" && $f.id_representante == $id_rep) {
            var.update $ok {
              value = true
            }
          }
        
          elseif (($f.id_franqueado|is_empty) == false) {
            var $ids {
              value = $input.ids_franqueado|first_notempty:[]
            }
          
            foreach ($ids) {
              each as $id_fra {
                conditional {
                  if ($f.id_franqueado == $id_fra) {
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