// Split 3 niveis: piso BG do item do catalogo (modo piso) + margem Central + margem REP
function fn_fp_split_precos {
  input {
    text id_central? filters=trim
    text produto? filters=trim
    text plano? filters=trim
    decimal valor_piso_central?
    decimal valor_venda?
    decimal valor_piso_breakglass?
  }

  stack {
    var $id_central {
      value = $input.id_central|first_notempty:"CENTRAL"|trim
    }
  
    var $piso_cen {
      value = $input.valor_piso_central|first_notempty:0
    }
  
    var $venda {
      value = $input.valor_venda|first_notempty:$piso_cen
    }
  
    function.run fn_fp_central_modo_preco {
      input = {id_central: $id_central}
    } as $modo_res
  
    var $modo {
      value = $modo_res.modo_preco|first_notempty:"livre"
    }
  
    var $piso_bg {
      value = 0
    }
  
    conditional {
      if ($modo == "piso") {
        // Preferencia: piso gravado no item; senao o proprio valor da Central (100% BG)
        var.update $piso_bg {
          value = $input.valor_piso_breakglass|first_notempty:$piso_cen
        }
      
        conditional {
          if ($piso_cen < $piso_bg) {
            var.update $piso_cen {
              value = $piso_bg
            }
          }
        }
      }
    }
  
    conditional {
      if ($venda < $piso_cen) {
        var.update $venda {
          value = $piso_cen
        }
      }
    }
  
    var $margem_cen {
      value = $piso_cen - $piso_bg
    }
  
    conditional {
      if ($margem_cen < 0) {
        var.update $margem_cen {
          value = 0
        }
      }
    }
  
    var $margem_rep {
      value = $venda - $piso_cen
    }
  
    conditional {
      if ($margem_rep < 0) {
        var.update $margem_rep {
          value = 0
        }
      }
    }
  }

  response = {
    id_central            : $id_central
    modo_preco            : $modo
    valor_venda           : $venda
    valor_piso_central    : $piso_cen
    valor_piso_breakglass : $piso_bg
    margem_central        : $margem_cen
    margem_rep            : $margem_rep
  }
}
