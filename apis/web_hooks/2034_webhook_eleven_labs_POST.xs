// wsec_0192c234b717798536f2ea00932a993b7d917c89a9e2b8eee50ea20e3aea74ff
query webhook_ElevenLabs verb=POST {
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
            success: false,
        
            conversation_id: "",
            call_sid: "",
            whatsapp_call_id: "",
        
            caller: "",
            receiver: "",
            direction: "",
            duration: null,
        
            status: "",
            action: "",
            type: "",
        
            id_user: null,
        
            e_event_type: "",
            e_event_timestamp: null,
            e_agent_id: "",
            e_agent_name: "",
            e_user_id: "",
            e_failure_reason: "",
            e_sip_status_code: null,
            e_error_reason: "",
            e_twirp_code: "",
            e_sip_status: "",
        
            termination_reason: "",
        
            analysis_success: "",
            analysis_summary_title: "",
            analysis_transcript_summary: "",
        
            cost_total: null,
            call_charge_credits: null,
            llm_charge_credits: null,
            llm_price_usd: null,
            tier: "",
            is_burst: false,
            free_minutes_consumed: null,
            free_llm_dollars_consumed: null,
        
            llm_input_tokens: null,
            llm_output_tokens: null,
            llm_input_price: null,
            llm_output_price: null
        };
        
        try {
            const body = Array.isArray(input) ? input[0] : input;
            if (!body || !body.data) throw new Error("Estrutura inválida");
        
            const data = body.data;
        
            // =====================================
            // DADOS GERAIS
            // =====================================
        
            result.type = body.type || "";
            result.e_event_type = body.type || "";
            result.e_event_timestamp = body.event_timestamp || null;
        
            result.conversation_id = data.conversation_id || "";
            result.e_agent_id = data.agent_id || "";
            result.e_agent_name = data.agent_name || "";
            result.e_user_id = data.user_id || "";
        
            // =====================================
            // POST CALL TRANSCRIPTION (SUCESSO)
            // =====================================
            if (body.type === "post_call_transcription") {
        
                const metadata = data.metadata || {};
                const charging = metadata.charging || {};
                const phoneCall = metadata.phone_call || {};
        
                result.status = data.status || "done";
                result.success = true;
        
                result.call_sid = phoneCall.call_sid || "";
                result.direction = phoneCall.direction || "";
                result.caller = phoneCall.agent_number || "";
                result.receiver = phoneCall.external_number || "";
        
                result.duration = metadata.call_duration_secs || null;
                result.termination_reason = metadata.termination_reason || "";
        
                // =====================
                // ANALYSIS
                // =====================
        
                result.analysis_success = data.analysis?.call_successful || "";
                result.analysis_summary_title = data.analysis?.call_summary_title || "";
                result.analysis_transcript_summary = data.analysis?.transcript_summary || "";
        
                // =====================
                // BILLING
                // =====================
        
                result.cost_total = metadata.cost || null;
                result.call_charge_credits = charging.call_charge || null;
                result.llm_charge_credits = charging.llm_charge || null;
                result.llm_price_usd = charging.llm_price || null;
                result.tier = charging.tier || "";
                result.is_burst = charging.is_burst || false;
                result.free_minutes_consumed = charging.free_minutes_consumed || null;
                result.free_llm_dollars_consumed = charging.free_llm_dollars_consumed || null;
        
                const modelUsage =
                    charging.llm_usage?.initiated_generation?.model_usage?.["gpt-4.1-nano"];
        
                if (modelUsage) {
                    result.llm_input_tokens = modelUsage.input?.tokens || null;
                    result.llm_output_tokens = modelUsage.output_total?.tokens || null;
                    result.llm_input_price = modelUsage.input?.price || null;
                    result.llm_output_price = modelUsage.output_total?.price || null;
                }
            }
        
            // =====================================
            // CALL INITIATION FAILURE
            // =====================================
            if (body.type === "call_initiation_failure") {
        
                const metadataBody = data.metadata?.body || {};
        
                result.status = "failed";
                result.success = false;
        
                result.caller = metadataBody.from_number || "";
                result.receiver = metadataBody.to_number || "";
        
                result.e_failure_reason = data.failure_reason || "";
                result.e_sip_status = metadataBody.sip_status || "";
                result.e_sip_status_code = metadataBody.sip_status_code || null;
                result.e_error_reason = metadataBody.error_reason || "";
                result.e_twirp_code = metadataBody.twirp_code || "";
            }
        
        } catch (error) {
            result.success = false;
        }
        
        return result;
        """
      timeout = 10
    } as $x1
  
    db.query whatsappLigarErro {
      where = $db.whatsappLigarErro.ideventgo == $x1.e_user_id && $db.whatsappLigarErro.telefone == $x1.receiver
      sort = {whatsappLigarErro.id: "desc"}
      return = {type: "single"}
    } as $whatsappLigarErro1
  
    conditional {
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
      
        conditional {
          // end_call tool e conclusao normal do agente — nao marcar falha/retry.
          // Retry so quando a chamada realmente falhou e sem duracao util.
          if ($x1.success == false && (($x1.duration|is_empty) || $x1.duration < 1)) {
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
          
            var $var_idProcesso {
              value = $whatsappLigarErro1.idProcesso
            }
          
            var $tentativasOriginais {
              value = $whatsappLigarErro1.tentativas
            }
          
            var $idLigarErroAtual {
              value = $whatsappLigarErro1.id
            }
          
            // SIP definitivo = numero inexistente/invalido (404/484/604): nao adianta religar
            var $sipDefinitivo {
              value = false
            }
          
            conditional {
              if ($x1.e_sip_status_code == 404 || $x1.e_sip_status_code == 484 || $x1.e_sip_status_code == 604) {
                var.update $sipDefinitivo {
                  value = true
                }
              }
            }
          
            conditional {
              if ($sipDefinitivo) {
                db.edit whatsappLigarErro {
                  field_name = "id"
                  field_value = $idLigarErroAtual
                  data = {
                    falha            : false
                    tentativas       : 99
                    dtUltimaTentativa: now
                    exec             : false
                    atendido         : false
                  }
                } as $whatsappLigarErroDefinitivo
              }
            }
          
            // Promocao imediata do proximo telefone: na 1a falha (tentativas=0)
            // ou quando o numero atual e invalido (SIP definitivo).
            // Antes o 2o contato so era promovido na 2a falha (~2-4 min depois).
            conditional {
              if ($sipDefinitivo || $tentativasOriginais == 0) {
                db.query whatsappLigarErro {
                  where = $db.whatsappLigarErro.idProcesso == $var_idProcesso && $db.whatsappLigarErro.atendido == true
                  return = {type: "single"}
                } as $ligAtendidoProc
              
                conditional {
                  if ($ligAtendidoProc|is_empty) {
                    db.query whatsappLigarErro {
                      where = $db.whatsappLigarErro.idProcesso == $var_idProcesso && $db.whatsappLigarErro.tentativas == 3333
                      sort = {whatsappLigarErro.id: "asc"}
                      return = {type: "single"}
                    } as $ligProximoTelefone
                  
                    conditional {
                      if (($ligProximoTelefone|is_empty) && $sipDefinitivo) {
                        db.query whatsappLigarErro {
                          where = $db.whatsappLigarErro.idProcesso == $var_idProcesso && $db.whatsappLigarErro.tentativas == 100
                          sort = {whatsappLigarErro.id: "asc"}
                          return = {type: "single"}
                        } as $ligProximoTelefone
                      }
                    }
                  
                    conditional {
                      if (($ligProximoTelefone|is_empty) && $sipDefinitivo) {
                        db.query whatsappLigarErro {
                          where = $db.whatsappLigarErro.idProcesso == $var_idProcesso && $db.whatsappLigarErro.tentativas == 200
                          sort = {whatsappLigarErro.id: "asc"}
                          return = {type: "single"}
                        } as $ligProximoTelefone
                      }
                    }
                  
                    conditional {
                      if (($ligProximoTelefone|is_empty) && $sipDefinitivo) {
                        db.query whatsappLigarErro {
                          where = $db.whatsappLigarErro.idProcesso == $var_idProcesso && $db.whatsappLigarErro.tentativas == 300
                          sort = {whatsappLigarErro.id: "asc"}
                          return = {type: "single"}
                        } as $ligProximoTelefone
                      }
                    }
                  
                    conditional {
                      if (($ligProximoTelefone|is_empty) == false) {
                        db.edit whatsappLigarErro {
                          field_name = "id"
                          field_value = $ligProximoTelefone.id
                          data = {
                            falha     : true
                            tentativas: $ligProximoTelefone.nroErr
                            exec      : false
                            atendido  : false
                          }
                        } as $ligProximoTelefonePromovido
                      }
                    }
                  }
                }
              }
            }
          
            // Funcao Controle Ligação
          
            conditional {
              if ($tentativasOriginais > 0 && $sipDefinitivo == false) {
                db.edit whatsappLigarErro {
                  field_name = "id"
                  field_value = $whatsappLigarErro1.id
                  enforce_hidden_fields = false
                  data = {
                    tentativas: ($var.whatsappLigarErro1.tentativas * 100)
                    nroErr    : $whatsappLigarErro1.tentativas
                  }
                } as $whatsappLigarErro2
              
                db.query whatsappLigarErro {
                  where = $db.whatsappLigarErro.idProcesso == $var_idProcesso && $db.whatsappLigarErro.tentativas == 3333
                  sort = {whatsappLigarErro.id: "asc"}
                  return = {type: "single"}
                } as $whatsappLigarErro1
              
                conditional {
                  if (($whatsappLigarErro1|is_empty) == false) {
                    db.edit whatsappLigarErro {
                      field_name = "id"
                      field_value = $whatsappLigarErro1.id
                      enforce_hidden_fields = false
                      data = {
                        falha     : true
                        tentativas: $whatsappLigarErro1.nroErr
                        exec      : false
                        atendido  : false
                      }
                    } as $whatsappLigarErro2
                  }
                
                  else {
                    db.query whatsappLigarErro {
                      where = $db.whatsappLigarErro.idProcesso == $var_idProcesso && $db.whatsappLigarErro.tentativas == 100
                      sort = {whatsappLigarErro.id: "asc"}
                      return = {type: "single"}
                    } as $whatsappLigarErro1
                  
                    conditional {
                      if (($whatsappLigarErro1|is_empty) == false) {
                        db.edit whatsappLigarErro {
                          field_name = "id"
                          field_value = $whatsappLigarErro1.id
                          enforce_hidden_fields = false
                          data = {
                            falha     : true
                            tentativas: $whatsappLigarErro1.nroErr
                            exec      : false
                            atendido  : false
                          }
                        } as $whatsappLigarErro2
                      }
                    
                      else {
                        db.query whatsappLigarErro {
                          where = $db.whatsappLigarErro.idProcesso == $var_idProcesso && $db.whatsappLigarErro.tentativas == 200
                          sort = {whatsappLigarErro.id: "asc"}
                          return = {type: "single"}
                        } as $whatsappLigarErro1
                      
                        conditional {
                          if (($whatsappLigarErro1|is_empty) == false) {
                            db.edit whatsappLigarErro {
                              field_name = "id"
                              field_value = $whatsappLigarErro1.id
                              enforce_hidden_fields = false
                              data = {
                                falha     : true
                                tentativas: $whatsappLigarErro1.nroErr
                                exec      : false
                                atendido  : false
                              }
                            } as $whatsappLigarErro2
                          }
                        
                          else {
                            db.query whatsappLigarErro {
                              where = $db.whatsappLigarErro.idProcesso == $var_idProcesso && $db.whatsappLigarErro.tentativas == 300
                              sort = {whatsappLigarErro.id: "asc"}
                              return = {type: "single"}
                            } as $whatsappLigarErro1
                          
                            conditional {
                              if (($whatsappLigarErro1|is_empty) == false) {
                                db.edit whatsappLigarErro {
                                  field_name = "id"
                                  field_value = $whatsappLigarErro1.id
                                  enforce_hidden_fields = false
                                  data = {
                                    falha     : true
                                    tentativas: $whatsappLigarErro1.nroErr
                                    exec      : false
                                    atendido  : false
                                  }
                                } as $whatsappLigarErro2
                              }
                            
                              else {
                                // cola aqui
                              
                                var $x2 {
                                  value = false
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
            }
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
          
            db.query whatsappLigarErro {
              where = $db.whatsappLigarErro.idProcesso == $whatsappLigarErro1.idProcesso && $db.whatsappLigarErro.id != $whatsappLigarErro1.id && $db.whatsappLigarErro.atendido == false && $db.whatsappLigarErro.tentativas != 99
              return = {type: "list"}
            } as $pendentesCancelar
          
            foreach ($pendentesCancelar) {
              each as $pendente {
                db.edit whatsappLigarErro {
                  field_name = "id"
                  field_value = $pendente.id
                  enforce_hidden_fields = false
                  data = {
                    falha            : false
                    tentativas       : 99
                    dtUltimaTentativa: now
                    exec             : false
                    atendido         : false
                    cancelado        : true
                  }
                } as $pendenteCancelado
              }
            }
          }
        }
      }
    }
  }

  response = $x1
}