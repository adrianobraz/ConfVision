query webhook_SMS_comtele verb=POST {
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
        
        let result = {
          success: true,
        
          message_id: "",
          message_tag: "",
          message_type: "",
          webhook_type: "",
          message_custom: "",
          message_status: "",
          message_status_details: "",
        
          id_referencia: null,
        
          queued_at: null,
          sent_at: null,
          delivered_at: null,
          failed_at: null,
        
          raw: input
        };
        
        try {
          const body = Array.isArray(input) ? input[0] : input;
        
          result.message_id = body.messageId || "";
          result.message_tag = body.messageTag || "";
          result.message_type = body.messageType || "";
          result.webhook_type = body.webhookType || "";
          result.message_custom = body.messageCustom || "";
          result.message_status = body.messageStatus || "";
          result.message_status_details = body.messageStatusDetails || "";
        
          result.id_referencia = body.messageCustom ? parseInt(body.messageCustom) : null;
        
          if (body.messageStatus === "Queued") {
            result.queued_at = new Date().toISOString();
          }
        
          if (body.messageStatus === "Sent") {
            result.sent_at = new Date().toISOString();
          }
        
          if (body.messageStatus === "Delivered") {
            result.delivered_at = new Date().toISOString();
          }
        
          if (body.messageStatus === "Failed") {
            result.failed_at = new Date().toISOString();
          }
        
        } catch (e) {
          result.success = false;
        }
        
        return result;
        """
      timeout = 10
    } as $x1
  
    db.add sms_Historico {
      enforce_hidden_fields = false
      data = {
        success               : $x1.success
        message_id            : $x1.message_id
        message_tag           : $x1.message_tag
        message_type          : $x1.message_type
        webhook_type          : $x1.webhook_type
        whatsappenviados_id   : $x1.id_referencia
        message_custom        : $x1.message_custom
        message_status        : $x1.message_status
        message_status_details: $x1.message_status_details
        id_referencia         : $x1.id_referencia
        queued_at             : $x1.queued_at
        sent_at               : $x1.sent_at
        delivered_at          : $x1.delivered_at
        failed_at             : $x1.failed_at
      }
    } as $sms_Historico1
  
    // B4: reflete o status da Comtele (Queued/Sent/Delivered/Failed) no registro do envio
    try_catch {
      try {
        conditional {
          if ($x1.id_referencia > 0 && (($x1.message_status|is_empty) == false)) {
            db.edit WhatsAppEnviados {
              field_name = "id"
              field_value = $x1.id_referencia
              data = {sms_status: $x1.message_status}
            } as $WhatsAppEnviadosStatus
          }
        }
      }
    
      catch {
        // id_referencia invalido nao pode derrubar o webhook
      }
    }
  
    !db.query whatsappLigarErro {
      where = $db.whatsappLigarErro.ideventgo == $x1.e_user_id && $db.whatsappLigarErro.telefone == $x1.receiver
      sort = {whatsappLigarErro.id: "desc"}
      return = {type: "single"}
    } as $whatsappLigarErro1
  
    !conditional {
      if ($x1.success == false) {
        db.edit whatsappLigarErro {
          field_name = "id"
          field_value = $whatsappLigarErro1.id
          enforce_hidden_fields = false
          data = {
            falha            : true
            dtUltimaTentativa: now
            exec             : false
            atendido         : false
          }
        } as $whatsappLigarErro2
      }
    
      else {
        db.edit whatsappLigarErro {
          field_name = "id"
          field_value = $whatsappLigarErro1.id
          enforce_hidden_fields = false
          data = {
            falha            : false
            dtUltimaTentativa: now
            exec             : false
            atendido         : true
          }
        } as $whatsappLigarErro2
      }
    }
  
    !conditional {
      if (($whatsappLigarErro1|is_empty) == false) {
        db.add LigacaoHistorico {
          enforce_hidden_fields = false
          data = {
            created_at                 : "now"
            conversation_id            : $x1.conversation_id
            success                    : $x1.success
            whatsapp_call_id           : $x1.whatsapp_call_id
            caller                     : $x1.caller
            receiver                   : $x1.receiver
            status                     : $x1.status
            action                     : $x1.action
            type                       : $x1.type
            direction                  : $x1.direction
            id_user                    : $x1.id_user
            duration                   : $x1.duration
            e_event_type               : $x1.e_event_type
            e_event_timestamp          : $x1.e_event_timestamp
            e_agent_id                 : $x1.e_agent_id
            e_agent_name               : $x1.e_agent_name
            e_user_id                  : $x1.e_user_id
            e_failure_reason           : $x1.e_failure_reason
            e_sip_status_code          : $x1.e_sip_status_code
            e_error_reason             : $x1.e_error_reason
            e_twirp_code               : $x1.e_twirp_code
            e_sip_status               : $x1.e_sip_status
            whatsappligarerro_id       : $whatsappLigarErro1.id
            call_sid                   : $x1.call_sid
            agent_number               : $x1.e_agent_name
            termination_reason         : $x1.termination_reason
            analysis_success           : $x1.analysis_success
            analysis_summary_title     : $x1.analysis_summary_title
            analysis_transcript_summary: $x1.analysis_transcript_summary
            cost_total                 : $x1.cost_total
            call_charge_credits        : $x1.call_charge_credits
            llm_charge_credits         : $x1.llm_charge_credits
            llm_price_usd              : $x1.llm_price_usd
            tier                       : $x1.tier
            is_burst                   : $x1.is_burst
            free_minutes_consumed      : $x1.free_minutes_consumed
            free_llm_dollars_consumed  : $x1.free_llm_dollars_consumed
            llm_input_tokens           : $x1.llm_input_tokens
            llm_output_tokens          : $x1.llm_output_tokens
            llm_input_price            : $x1.llm_input_price
            llm_output_price           : $x1.llm_output_price
          }
        } as $LigacaoHistorico1
      }
    }
  }

  response = $x1
}