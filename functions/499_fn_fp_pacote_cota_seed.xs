// Seed idempotente dos pacotes Cota 50 / 200 / 800
function fn_fp_pacote_cota_seed {
  input {
    text id_central?=CENTRAL filters=trim
  }

  stack {
    var $id_central {
      value = $input.id_central|first_notempty:"CENTRAL"|trim
    }
  
    var $itens {
      value = [
        {nome: "Cota 50", quantidade: 50, valor: 50}
        {nome: "Cota 200", quantidade: 200, valor: 100}
        {nome: "Cota 800", quantidade: 800, valor: 200}
      ]
    }
  
    var $inseridos {
      value = 0
    }
  
    foreach ($itens) {
      each as $item {
        db.query fp_pacote_cota {
          where = $db.fp_pacote_cota.id_central == $id_central && $db.fp_pacote_cota.quantidade == $item.quantidade
          return = {type: "single"}
        } as $existe
      
        conditional {
          if ($existe == null) {
            db.add fp_pacote_cota {
              data = {
                created_at: "now"
                id_central: $id_central
                nome      : $item.nome
                quantidade: $item.quantidade
                valor     : $item.valor
                ativo     : "S"
              }
            } as $novo
          
            var.update $inseridos {
              value = $inseridos + 1
            }
          }
        }
      }
    }
  
    db.query fp_pacote_cota {
      where = $db.fp_pacote_cota.id_central == $id_central
      return = {type: "list"}
    } as $lista
  }

  response = {
    inseridos : $inseridos
    total     : $lista|count
    id_central: $id_central
    dados     : $lista
  }
}