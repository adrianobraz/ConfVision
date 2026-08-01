// Copia valor_mensal -> valor_piso_breakglass em todos os itens do catalogo da Central
function fn_fp_catalogo_sync_piso_bg {
  input {
    text id_central? filters=trim
  }

  stack {
    var $id_central {
      value = $input.id_central|first_notempty:"CENTRAL"|trim
    }
  
    var $atualizados {
      value = 0
    }
  
    db.query fp_produto_catalogo {
      where = $db.fp_produto_catalogo.id_central == $id_central && $db.fp_produto_catalogo.id_representante == ""
      return = {type: "list"}
    } as $itens
  
    foreach ($itens) {
      each as $item {
        db.patch fp_produto_catalogo {
          field_name = "id"
          field_value = $item.id
          data = {valor_piso_breakglass: $item.valor_mensal|first_notempty:0}
        } as $p
      
        var.update $atualizados {
          value = $atualizados + 1
        }
      }
    }
  
    db.query fp_modulo_catalogo {
      where = $db.fp_modulo_catalogo.id_central == $id_central && $db.fp_modulo_catalogo.id_representante == ""
      return = {type: "list"}
    } as $mods
  
    foreach ($mods) {
      each as $m {
        db.patch fp_modulo_catalogo {
          field_name = "id"
          field_value = $m.id
          data = {valor_piso_breakglass: $m.valor_mensal|first_notempty:0}
        } as $pm
      
        var.update $atualizados {
          value = $atualizados + 1
        }
      }
    }
  }

  response = {id_central: $id_central, atualizados: $atualizados}
}