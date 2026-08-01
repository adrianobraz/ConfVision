function fn_DisparaWhatsapp {
  input {
  }

  stack {
    db.query WhatsProcFila_Controle {
      return = {type: "single"}
    } as $WhatsProcFila_Controle1
  
    precondition ($WhatsProcFila_Controle1.Lock == false)
    db.edit WhatsProcFila_Controle {
      field_name = "id"
      field_value = $WhatsProcFila_Controle1.id
      enforce_hidden_fields = false
      data = {Lock: true}
    } as $WhatsProcFila_Controle2
  
    function.run fn_AlarmeEventos_Pendente as $func1
  
    // Roda a Tabela enviando Whatsapp
  
    db.query WhatsappProcFila {
      where = $db.WhatsappProcFila.analise == false && $db.WhatsappProcFila.disparo == false
      sort = {WhatsappProcFila.created_at: "desc"}
      return = {type: "list"}
    } as $WhatsappProcFila1
  
    var $Contador {
      value = 0
    }
  
    foreach ($WhatsappProcFila1) {
      each as $item {
        // Claim atomico antes do envio — se outro worker ja pegou, pula
        function.run fn_ClaimWhatsappProcFila {
          input = {id: $item.id}
        } as $claim1
      
        conditional {
          if ($claim1.claimed == false) {
            continue
          }
        }
      
        var.update $Contador {
          value = $Contador + 1
        }
      
        conditional {
          if ($Contador > 5) {
            security.random_number {
              min = 15
              max = 20
            } as $random2
          
            util.sleep {
              value = $random2
            }
          
            var.update $Contador {
              value = 1
            }
          }
        }
      
        function.run FuncaoSis_WhatsappProcFila {
          input = {
            SendMsgWhats            : $item.SendMsgWhats
            numerowhatsapp          : $item.numerowhatsapp
            enviaSom                : $item.enviaSom
            SendAudioWhats          : $item.SendAudioWhats
            tblWhatsAppEnviadosTexto: $item.tblWhatsAppEnviadosTexto
            tblWhatsAppEnviadosAudio: $item.tblWhatsAppEnviadosAudio
            tblWhatsAppEnviadosLigar: $item.tblWhatsAppEnviadosLigar
            ligar                   : $item.ligar
            idFranqueado            : $item.idFranqueado
            nomecliente             : $item.nomecliente
            empresanome             : $item.empresanome
            tipoevento              : $item.tipoevento
            zona                    : $item.zona
            local                   : $item.local
            idevento                : $item.idevento
            datahorario             : $item.datahorario
            ideventgo               : $item.ideventgo
            enviartexto             : $item.enviartexto
            idProcesso              : $item.idProcesso
            idDispositivo           : $item.idDispositivo
            notificarsempre         : $item.notificarsempre
            DispNome                : $item.DispNome
            DispDescricao           : $item.DispDescricao
            DispTipo                : $item.DispTipo
          }
        } as $func4
      
        // Segurado (ex.: trava 15s): libera claim para reprocessar depois
        conditional {
          if ($func4.dados == false) {
            db.edit WhatsappProcFila {
              field_name = "id"
              field_value = $item.id
              enforce_hidden_fields = false
              data = {disparo: false}
            } as $WhatsappProcFilaRelease
          }
        }
      
        security.random_number {
          min = 1
          max = 3
        } as $random1
      
        util.sleep {
          value = $random1
        }
      }
    }
  
    db.edit WhatsProcFila_Controle {
      field_name = "id"
      field_value = $WhatsProcFila_Controle1.id
      enforce_hidden_fields = false
      data = {Lock: false}
    } as $WhatsProcFila_Controle2
  }

  response = $WhatsappProcFila1
}
