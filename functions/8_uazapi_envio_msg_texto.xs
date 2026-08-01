function uazapi_EnvioMsgTexto {
  input {
    text number? filters=trim
    text textMensagem? filters=trim
    int tblWhatsAppEnviados?
    text InstanceToken? filters=trim
    text tipoAPi? filters=trim
    text instancia? filters=trim
  }

  stack {
    !return {
      value = ""
    }
  
    conditional {
      if ($input.InstanceToken|is_empty) {
        var $tokenzap {
          value = "b014aad8-99e1-4115-9873-7d004ab91dc7"
        }
      
        var $tipoApi {
          value = "U"
        }
      }
    
      else {
        var $tokenzap {
          value = $input.InstanceToken
        }
      
        var $tipoApi {
          value = $input.tipoAPi
        }
      }
    }
  
    conditional {
      if ($tipoApi == "U") {
        api.request {
          url = "https://monitoramento.uazapi.com/send/text"
          method = "POST"
          params = {}
            |set:"number":$input.number
            |set:"text":$input.textMensagem
          headers = []
            |push:"Accept: application/json"
            |push:"Content-Type: application/json"
            |push:`"token:"|concat:$var.tokenzap`
        } as $api1
      
        !api.request {
          url = "https://monitoramento.uazapi.com/send/menu"
          method = "POST"
          params = {}
            |set:"number":"5519992478859"
            |set:"type":"button"
            |set:"text":"Escolha um produto:"
            |set:"choices":([]
              |push:"Produto A|prod_a"
              |push:"Mais Info|prod_b"
              |push:"Produto B|prod_c"
            )
            |set:"footerText":"Produtos em destaque"
          headers = []
            |push:"Accept: application/json"
            |push:"Content-Type: application/json"
            |push:"token: b014aad8-99e1-4115-9873-7d004ab91dc7"
        } as $api1
      
        conditional {
          if ($input.tblWhatsAppEnviados > 0) {
            db.edit WhatsAppEnviados {
              field_name = "id"
              field_value = $input.tblWhatsAppEnviados
              enforce_hidden_fields = false
              data = {
                whats_ID              : $api1.response.result.id
                whats_messageTimestamp: $api1.response.result.messageTimestamp
                whats_messageid       : $api1.response.result.chatid
                whats_sender          : $api1.response.result.sender
                whats_text            : $api1.response.result.text
                texto                 : true
              }
            } as $WhatsAppEnviados1
          
            !db.edit WhatsAppEnviados {
              field_name = "id"
              field_value = $input.tblWhatsAppEnviados
              enforce_hidden_fields = false
              data = {
                whats_ID              : $api1.response.result.instanceId
                whats_messageTimestamp: $api1.response.result.messageTimestamp
                whats_messageid       : $api1.response.result.key.id
                whats_sender          : $api1.response.result.key.remoteJid
                whats_text            : $api1.response.result.message.conversation
                texto                 : true
              }
            } as $WhatsAppEnviados1
          }
        }
      }
    
      elseif ($tipoApi == "G") {
        // EvoGo / Dialyze: POST /send/text com apikey=instanceToken e instanceId
        api.request {
          url = "https://confianca-evolution-api.rkr351.easypanel.host/send/text"
          method = "POST"
          params = {}
            |set:"number":$input.number
            |set:"text":$input.textMensagem
          headers = []
            |push:"Content-Type: application/json"
            |push:("apikey: "|concat:$tokenzap)
            |push:("instanceId: "|concat:$input.instancia)
        } as $api1
      
        conditional {
          if ($input.tblWhatsAppEnviados > 0) {
            try_catch {
              try {
                db.edit WhatsAppEnviados {
                  field_name = "id"
                  field_value = $input.tblWhatsAppEnviados
                  enforce_hidden_fields = false
                  data = {
                    whats_ID  : $api1.response.result.data.id
                    whats_text: $input.textMensagem
                    texto     : true
                  }
                } as $WhatsAppEnviados1
              }
            
              catch {
                db.edit WhatsAppEnviados {
                  field_name = "id"
                  field_value = $input.tblWhatsAppEnviados
                  enforce_hidden_fields = false
                  data = {
                    whats_text: $input.textMensagem
                    texto     : true
                  }
                } as $WhatsAppEnviados1
              }
            }
          }
        }
      }
    
      else {
        !api.request {
          url = "https://confianca-evolution-api.rkr351.easypanel.host/message/sendText/monitoramento@confiancanet.com.br"
          method = "POST"
          params = {}
            |set:"number":$input.number
            |set:"text":$input.textMensagem
            |set:"delay":0
            |set:"linkPreview":false
            |set:"mentionsEveryOne":false
          headers = []
            |push:"Content-Type: application/json"
            |push:"apikey: F97CB3AE66BF-403D-A525-C53DFED3F76D"
        } as $api1
      
        conditional {
          if ($input.tblWhatsAppEnviados > 0) {
            !db.edit WhatsAppEnviados {
              field_name = "id"
              field_value = $input.tblWhatsAppEnviados
              enforce_hidden_fields = false
              data = {
                whats_ID              : $api1.response.result.instanceId
                whats_messageTimestamp: $api1.response.result.messageTimestamp
                whats_messageid       : $api1.response.result.key.id
                whats_sender          : $api1.response.result.key.remoteJid
                whats_text            : $api1.response.result.message.conversation
                texto                 : true
              }
            } as $WhatsAppEnviados1
          }
        }
      }
    }
  }

  response = $api1
}
