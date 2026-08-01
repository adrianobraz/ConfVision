// Salva eventos da Linha do Tempo (upsert por chave — idempotente)
query co_timeline_salvar verb=POST {
  api_group = "centerOperacion"

  input {
    object[] eventos? {
      schema {
        text chave filters=trim
        int mapa_ambiente_id?
        text idCliente? filters=trim
        text idFranqueado? filters=trim
        text idProcesso? filters=trim
        text idSetor? filters=trim
        text idDispositivo? filters=trim
        text idOperador? filters=trim
        text nomeOperador? filters=trim
        text tipo? filters=trim
        text origem? filters=trim
        text codigo? filters=trim
        text zonaUser? filters=trim
        text particao? filters=trim
        text texto? filters=trim
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
      if ($input.eventos != null) {
        foreach ($input.eventos) {
          each as $ev {
            conditional {
              if (($ev.chave|is_empty) == false) {
                db.query co_timeline_evento {
                  where = $db.co_timeline_evento.chave == $ev.chave
                  return = {type: "single"}
                } as $existente
              
                conditional {
                  if ($existente == null) {
                    db.add co_timeline_evento {
                      data = {
                        created_at      : "now"
                        chave           : $ev.chave
                        mapa_ambiente_id: $ev.mapa_ambiente_id
                        idCliente       : $ev.idCliente
                        idFranqueado    : $ev.idFranqueado
                        idProcesso      : $ev.idProcesso
                        idSetor         : $ev.idSetor
                        idDispositivo   : $ev.idDispositivo
                        idOperador      : $ev.idOperador
                        nomeOperador    : $ev.nomeOperador
                        tipo            : $ev.tipo
                        origem          : $ev.origem
                        codigo          : $ev.codigo
                        zonaUser        : $ev.zonaUser
                        particao        : $ev.particao
                        texto           : $ev.texto
                        evento_ts       : $ev.evento_ts|first_notempty:"now"
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