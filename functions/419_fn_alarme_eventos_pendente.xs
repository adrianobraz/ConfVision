function fn_AlarmeEventos_Pendente {
  input {
  }

  stack {
    db.query alarmeEventos_pendente {
      where = $db.alarmeEventos_pendente.cancelado == false && $db.alarmeEventos_pendente.enviado == false && $db.alarmeEventos_pendente.agendadoPara <= now
      sort = {alarmeEventos_pendente.agendadoPara: "asc"}
      return = {type: "list"}
    } as $alarmeEventos_pendente1
  
    var $processados {
      value = 0
    }
  
    foreach ($alarmeEventos_pendente1) {
      each as $item {
        db.query WhatsappProcFila {
          where = $db.WhatsappProcFila.idDispositivo == $item.idDispositivo && $db.WhatsappProcFila.zonauser == $item.zonaUser && $db.WhatsappProcFila.particao == $item.particao && $db.WhatsappProcFila.idevento == $item.id_tblAlarm_events
          sort = {WhatsappProcFila.id: "desc"}
          return = {type: "list"}
        } as $WhatsappProcFila1
      
        conditional {
          if (($WhatsappProcFila1|is_empty) == false) {
            db.edit alarmeEventos_pendente {
              field_name = "id"
              field_value = $item.id
              enforce_hidden_fields = false
              data = {enviado: true}
            } as $alarmeEventos_pendente2
          
            foreach ($WhatsappProcFila1) {
              each as $itemProcFila {
                db.edit WhatsappProcFila {
                  field_name = "id"
                  field_value = $itemProcFila.id
                  enforce_hidden_fields = false
                  data = {analise: false}
                } as $WhatsappProcFila3
              }
            }
          
            var.update $processados {
              value = $processados + 1
            }
          }
        
          else {
            // Vencido sem item na fila: encerra pendente para nao reprocessar indefinidamente
            db.edit alarmeEventos_pendente {
              field_name = "id"
              field_value = $item.id
              enforce_hidden_fields = false
              data = {enviado: true}
            } as $alarmeEventos_pendente_orfao
          
            var.update $processados {
              value = $processados + 1
            }
          }
        }
      }
    }
  }

  response = {
    processados: $processados
    lote       : $alarmeEventos_pendente1
  }
}