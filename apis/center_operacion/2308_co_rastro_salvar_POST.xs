// Salva pontos do rastro de disparos (upsert por chave — idempotente)
query co_rastro_salvar verb=POST {
  api_group = "centerOperacion"

  input {
    object[] pontos? {
      schema {
        text chave filters=trim
        int mapa_ambiente_id?
        text idCliente? filters=trim
        text idFranqueado? filters=trim
        text idProcesso? filters=trim
        text idSetor? filters=trim
        text labelSetor? filters=trim
        decimal posX?
        decimal posY?
        int sequencia?
        text predito_id_setor? filters=trim
        text direcao? filters=trim
        timestamp evento_ts?
      }
    }
  }

  stack {
    var $inseridos {
      value = 0
    }
  
    var $ignorados {
      value = 0
    }
  
    conditional {
      if ($input.pontos != null) {
        foreach ($input.pontos) {
          each as $pt {
            conditional {
              if (($pt.chave|is_empty) == false) {
                db.query co_rastro_disparo {
                  where = $db.co_rastro_disparo.chave == $pt.chave
                  return = {type: "single"}
                } as $existente
              
                conditional {
                  if ($existente == null) {
                    db.add co_rastro_disparo {
                      data = {
                        created_at      : "now"
                        chave           : $pt.chave
                        mapa_ambiente_id: $pt.mapa_ambiente_id
                        idCliente       : $pt.idCliente
                        idFranqueado    : $pt.idFranqueado
                        idProcesso      : $pt.idProcesso
                        idSetor         : $pt.idSetor
                        labelSetor      : $pt.labelSetor
                        posX            : $pt.posX
                        posY            : $pt.posY
                        sequencia       : $pt.sequencia
                        predito_id_setor: $pt.predito_id_setor
                        direcao         : $pt.direcao
                        evento_ts       : $pt.evento_ts|first_notempty:"now"
                      }
                    } as $novo
                  
                    var.update $inseridos {
                      value = $inseridos + 1
                    }
                  }
                
                  else {
                    var.update $ignorados {
                      value = $ignorados + 1
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

  response = {
    status   : "OK"
    inseridos: $inseridos
    ignorados: $ignorados
  }
}