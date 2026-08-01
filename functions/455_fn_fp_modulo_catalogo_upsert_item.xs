// Insere ou atualiza item do catalogo de modulos (a la carte) por escopo
function fn_fp_modulo_catalogo_upsert_item {
  input {
    text id_central?=CENTRAL filters=trim
    text id_representante? filters=trim
    text produto? filters=trim
    text chave? filters=trim
    text label? filters=trim
    text grupo? filters=trim
    text descricao? filters=trim
    text plano_minimo? filters=trim
    decimal valor_mensal?
    text ativo? filters=trim
    int ordem?
  }

  stack {
    var $id_central {
      value = $input.id_central|first_notempty:"CENTRAL"
    }
  
    var $id_representante {
      value = $input.id_representante|first_notempty:""
    }
  
    db.query fp_modulo_catalogo {
      where = $db.fp_modulo_catalogo.produto == $input.produto && $db.fp_modulo_catalogo.chave == $input.chave && $db.fp_modulo_catalogo.id_central == $id_central && $db.fp_modulo_catalogo.id_representante == $id_representante
      return = {type: "single"}
    } as $existe
  
    var $inserido {
      value = 0
    }
  
    conditional {
      if ($existe == null) {
        db.add fp_modulo_catalogo {
          data = {
            created_at            : "now"
            id_central            : $id_central
            id_representante      : $id_representante
            produto               : $input.produto
            chave                 : $input.chave
            label                 : $input.label
            grupo                 : $input.grupo
            descricao             : $input.descricao
            plano_minimo          : $input.plano_minimo|first_notempty:"lite"
            valor_mensal          : $input.valor_mensal
            valor_piso_breakglass : $input.valor_mensal
            ativo                 : $input.ativo|first_notempty:"S"
            ordem                 : $input.ordem
          }
        } as $novo
      
        var.update $inserido {
          value = 1
        }
      }
    
      else {
        db.patch fp_modulo_catalogo {
          field_name = "id"
          field_value = $existe.id
          data = ```
            {
              label        : $input.label
              grupo        : $input.grupo
              descricao    : $input.descricao
              plano_minimo : $input.plano_minimo|first_notempty:"lite"
              valor_mensal : $input.valor_mensal
              ativo        : $input.ativo|first_notempty:"S"
              ordem        : $input.ordem
            }
            ```
        } as $atualizado
      }
    }
  }

  response = {inserido: $inserido}
}
