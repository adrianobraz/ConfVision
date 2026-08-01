// Normaliza lista de addons: chaves validas do catalogo (corrige legado franqueadopro -> franqueadopro.saas)
function fn_fp_alacarte_addons_normalizar {
  input {
    text produto?=franqueadopro filters=trim
    json addons?
  }

  stack {
    var $lista {
      value = []
    }
  
    foreach ($input.addons|first_notempty:[]) {
      each as $chave_raw {
        conditional {
          if (($chave_raw|is_empty) == false) {
            var $chave {
              value = $chave_raw
            }
          
            conditional {
              if ($chave == "franqueadopro") {
                var.update $chave {
                  value = "franqueadopro.saas"
                }
              }
            }
          
            db.query fp_modulo_catalogo {
              where = $db.fp_modulo_catalogo.id_central == "CENTRAL" && $db.fp_modulo_catalogo.id_representante == "" && $db.fp_modulo_catalogo.produto == $input.produto && $db.fp_modulo_catalogo.chave == $chave && $db.fp_modulo_catalogo.ativo == "S"
              return = {type: "single"}
            } as $item_cat
          
            conditional {
              if ($item_cat != null) {
                var $ja_tem {
                  value = false
                }
              
                foreach ($lista) {
                  each as $ch_existente {
                    conditional {
                      if ($ch_existente == $chave) {
                        var.update $ja_tem {
                          value = true
                        }
                      }
                    }
                  }
                }
              
                conditional {
                  if ($ja_tem == false) {
                    var.update $lista {
                      value = $lista|push:$chave
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

  response = {addons: $lista}
}