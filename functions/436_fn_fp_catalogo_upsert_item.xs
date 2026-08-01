// Insere ou atualiza item do catalogo por produto+plano+escopo (Central ou Representante)
function fn_fp_catalogo_upsert_item {
  input {
    text id_central?=CENTRAL filters=trim
    text id_representante? filters=trim
    text produto? filters=trim
    text plano? filters=trim
    text nome_exibicao? filters=trim
    decimal valor_mensal?
    int retencao_dias?
    json limites_json?
    json modulos_json?
  
    // Se true: nao sobrescreve item existente (preserva edicao no adm)
    bool somente_inserir?
  }

  stack {
    var $id_central {
      value = $input.id_central|first_notempty:"CENTRAL"
    }
  
    var $id_representante {
      value = $input.id_representante|first_notempty:""
    }
  
    db.query fp_produto_catalogo {
      where = $db.fp_produto_catalogo.produto == $input.produto && $db.fp_produto_catalogo.plano == $input.plano && $db.fp_produto_catalogo.id_central == $id_central && $db.fp_produto_catalogo.id_representante == $id_representante
      return = {type: "single"}
    } as $existe
  
    var $inserido {
      value = 0
    }
  
    var $retencao {
      value = 30
    }
  
    conditional {
      if ($input.retencao_dias != null && $input.retencao_dias > 0) {
        var.update $retencao {
          value = $input.retencao_dias
        }
      }
    }
  
    var $limites {
      value = {}
    }
  
    conditional {
      if ($input.limites_json != null) {
        var.update $limites {
          value = $input.limites_json
        }
      }
    }
  
    var $modulos {
      value = {}
    }
  
    conditional {
      if ($input.modulos_json != null) {
        var.update $modulos {
          value = $input.modulos_json
        }
      }
    }
  
    conditional {
      if ($existe == null) {
        db.add fp_produto_catalogo {
          data = {
            created_at             : "now"
            id_central             : $id_central
            id_representante       : $id_representante
            produto                : $input.produto
            plano                  : $input.plano
            nome_exibicao          : $input.nome_exibicao
            valor_mensal           : $input.valor_mensal
            valor_piso_breakglass  : $input.valor_mensal
            periodicidade_padrao   : "mensal"
            retencao_dias          : $retencao
            limites_json           : $limites
            modulos_json           : $modulos
            ativo                  : "S"
          }
        } as $novo
      
        var.update $inserido {
          value = 1
        }
      }
    
      elseif ($input.somente_inserir != true) {
        db.patch fp_produto_catalogo {
          field_name = "id"
          field_value = $existe.id
          data = {
            nome_exibicao: $input.nome_exibicao
            valor_mensal : $input.valor_mensal
            retencao_dias: $retencao
            limites_json : $limites
            modulos_json : $modulos
            ativo        : "S"
          }
        } as $atualizado
      }
    }
  }

  response = {inserido: $inserido}
}
