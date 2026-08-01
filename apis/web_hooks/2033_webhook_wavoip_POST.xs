query webhook_Wavoip verb=POST {
  api_group = "WebHooks"

  input {
  }

  stack {
    util.get_raw_input {
      encoding = "json"
      exclude_middleware = false
    } as $payload
  
    api.lambda {
      code = """
        const input = $var.payload;
        
        // 1. Iniciamos o objeto com TODOS os campos possíveis para evitar erros de 'undefined'
        let result = {
            success: false,
            whatsapp_call_id: "",
            status: "UNKNOWN",
            action: "NONE",
            caller: "",
            receiver: "",
            duration: 0,
            type: "CALL",
            direction: "OUTCOMING",
            transport: "sip",
            id_user: null,
            id_session: null,
            error_log: "" // Definimos aqui para o editor não marcar erro
        };
        
        try {
            // 2. Normalização para aceitar objeto direto ou array
            const body = Array.isArray(input) ? input[0] : input;
        
            if (body && (body.whatsapp_call_id || body.call_id)) {
                // Mapeamento principal
                result.whatsapp_call_id = body.whatsapp_call_id || body.call_id;
                result.status = body.status || "UNKNOWN";
                result.action = body.action || "UPDATE";
                result.duration = body.duration !== undefined ? body.duration : 0;
                
                // Mapeamento de campos de identificação (id_user vs idUser)
                result.id_user = body.id_user || body.idUser || null;
                result.id_session = body.id_session || null;
                
                // Campos que podem vir vazios em UPDATES
                result.caller = body.caller || "";
                result.receiver = body.receiver || "";
                result.direction = body.direction || "OUTCOMING";
                result.transport = body.transport || "sip";
        
                result.success = true;
            } else {
                result.error_log = "ID da ligação não encontrado no input.";
            }
        } catch (e) {
            // Captura o erro e joga na variável que definimos no início
            result.error_log = e.message;
        }
        
        return result;
        """
      timeout = 10
    } as $x1
  
    db.query LigacaoHistorico {
      where = $db.LigacaoHistorico.whatsapp_call_id == $x1.whatsapp_call_id
      sort = {LigacaoHistorico.id: "asc"}
      return = {type: "single"}
    } as $LigacaoHistorico2
  
    db.query LigacaoHistorico {
      where = $db.LigacaoHistorico.whatsapp_call_id == $x1.whatsapp_call_id && $db.LigacaoHistorico.status == $x1.status && $db.LigacaoHistorico.action == $x1.action
      sort = {LigacaoHistorico.id: "asc"}
      return = {type: "single"}
    } as $LigacaoHistoricoDuplic
  
    db.query whatsappLigarErro {
      where = $db.whatsappLigarErro.telefone == $x1.receiver
      sort = {whatsappLigarErro.id: "desc"}
      return = {type: "single"}
    } as $whatsappLigarErro1
  
    var $lAddDados {
      value = false
    }
  
    var $nIDWhatsLigarErro {
      value = 0
    }
  
    var $tCaller {
      value = ""
    }
  
    var $tReceiver {
      value = ""
    }
  
    conditional {
      if (($LigacaoHistorico2|is_empty) == false) {
        conditional {
          if ($LigacaoHistoricoDuplic|is_empty) {
            var.update $lAddDados {
              value = true
            }
          
            var.update $nIDWhatsLigarErro {
              value = $LigacaoHistorico2.whatsappligarerro_id
            }
          
            var.update $tCaller {
              value = $LigacaoHistorico2.caller
            }
          
            var.update $tReceiver {
              value = $LigacaoHistorico2.receiver
            }
          }
        }
      }
    
      else {
        conditional {
          if (($whatsappLigarErro1|is_empty) == false) {
            var.update $lAddDados {
              value = true
            }
          
            var.update $nIDWhatsLigarErro {
              value = $whatsappLigarErro1.id
            }
          
            var.update $tCaller {
              value = $x1.caller
            }
          
            var.update $tReceiver {
              value = $x1.receiver
            }
          }
        }
      }
    }
  
    conditional {
      if ($lAddDados) {
        db.add LigacaoHistorico {
          enforce_hidden_fields = false
          data = {
            created_at          : "now"
            success             : $x1.success
            whatsapp_call_id    : $x1.whatsapp_call_id
            id_session          : $x1.id_session
            caller              : $tCaller
            receiver            : $tReceiver
            status              : $x1.status
            action              : $x1.action
            type                : $x1.type
            direction           : $x1.direction
            id_user             : $x1.id_user
            duration            : $x1.duration
            whatsappligarerro_id: $nIDWhatsLigarErro
          }
        } as $LigacaoHistorico1
      }
    }
  }

  response = $x1
}