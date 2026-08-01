query alarm_events_webhook verb=POST {
  api_group = "WebHooks"

  input {
  }

  stack {
    precondition ($env.$http_headers.Authorization == "Bearer 1e2d2eef75ccd7cfea2e06d7ce1c63c4") {
      payload = "erro"
    }
  
    util.get_raw_input {
      encoding = "json"
      exclude_middleware = false
    } as $payload
  
    var $duplicado {
      value = false
    }
  
    // Dedup 1: mesmo idEvento ja gravado (retransmissao exata do receptor)
    conditional {
      if (($payload.idEvento|is_empty) == false) {
        db.query alarm_events {
          where = $db.alarm_events.idEvento == $payload.idEvento
          return = {type: "single"}
        } as $dupIdEvento
      
        conditional {
          if (($dupIdEvento|is_empty) == false) {
            var.update $duplicado {
              value = true
            }
          }
        }
      }
    }
  
    // Dedup 2: mesmo dispositivo+codigo+particao+zona em menos de 10s (painel com via dupla GPRS/Ethernet)
    conditional {
      if ($duplicado == false && (($payload.idDispositivo|is_empty) == false) && (($payload.codigo|is_empty) == false)) {
        var $janelaDedup {
          value = now|add_secs_to_timestamp:-10
        }
      
        db.query alarm_events {
          where = $db.alarm_events.idDispositivo == $payload.idDispositivo && $db.alarm_events.codigo == $payload.codigo && $db.alarm_events.particao == $payload.particao && $db.alarm_events.zonaUser == $payload.zonaUser && $db.alarm_events.created_at > $janelaDedup
          return = {type: "single"}
        } as $dupJanela
      
        conditional {
          if (($dupJanela|is_empty) == false) {
            var.update $duplicado {
              value = true
            }
          }
        }
      }
    }
  
    conditional {
      if ($duplicado == false) {
        db.add alarm_events {
          enforce_hidden_fields = false
          data = {
            created_at    : "now"
            idEvento      : $payload.idEvento
            codigo        : $payload.codigo
            particao      : $payload.particao
            zonaUser      : $payload.zonaUser
            nivel         : $payload.nivel
            dataEntrada   : $payload.dataEntrada
            img           : $payload.img
            idProcesso    : $payload.idProcesso
            idDispositivo : $payload.idDispositivo
            idCliente     : $payload.idCliente
            nomeCliente   : $payload.nomeCliente
            emailCliente  : $payload.emailCliente
            ctiGrupo      : $payload.ctiGrupo
            ctiDescricao  : $payload.ctiDescricao
            idFranqueado  : $payload.idFranqueado
            codigoBenuvem : $payload.codigoBenuvem
            conta         : $payload.conta
            carmeraAtiva  : $payload.carmeraAtiva
            usa_confvision: $payload.usaConfvision
            provedor_video: $payload.provedorVideo
          }
        } as $alarm_events1
      }
    }
  }

  response = "ok"
}