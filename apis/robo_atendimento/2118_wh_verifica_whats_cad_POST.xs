// https://admmonitoramento.com.br/version-test/finalizaevento/?idProcesso=2026032701375698984026926
query wh_verificaWhatsCad verb=POST {
  api_group = "RoboAtendimento"

  input {
    text numeroWhats? filters=trim
    text idprocesso? filters=trim
    text token? filters=trim
  }

  stack {
    var $WhatsappEventoCad1 {
      value = []
    }
  
    var $x1 {
      value = ""
    }
  
    conditional {
      if ($input.idprocesso|is_empty) {
        return {
          value = {dados: $WhatsappEventoCad1|set:"token":$x1}
        }
      }
    
      elseif ($input.numeroWhats|is_empty) {
        return {
          value = {dados: $WhatsappEventoCad1|set:"token":$x1}
        }
      }
    }
  
    api.request {
      url = "http://185.130.61.4:2010/v4/terminal/getDadosProcessoById"
      method = "POST"
      params = {}
        |set:"idProcesso":$input.idprocesso
      headers = []
        |push:"Authorization: Bearer " ~ $func1
        |push:"Content-Type: application/json"
    } as $api1
  
    var $varProcesso {
      value = $api1.response.result.dados
    }
  
    db.query WhatsappEventoCad {
      where = $db.WhatsappEventoCad.whatsapp == $input.numeroWhats && $db.WhatsappEventoCad.idDispositivo == $varProcesso.idDispositivo && $db.WhatsappEventoCad.Tipo in "ALARME"
      return = {type: "single"}
    } as $WhatsappEventoCad1
  
    conditional {
      if (($WhatsappEventoCad1|is_empty) == false) {
        conditional {
          if ($input.token == "a") {
            security.random_number {
              min = 1000
              max = 9999
            } as $x1
          
            function.run uazapi_EnvioMsgTexto {
              input = {
                number             : $input.numeroWhats
                textMensagem       : `($var.x1|to_text) ~ "\\\\nNão compartilhe esse código"`|text_unescape
                tblWhatsAppEnviados: 0
              }
            } as $func2
          }
        
          else {
            var $x1 {
              value = $input.token
            }
          }
        }
      }
    }
  }

  response = {dados: $WhatsappEventoCad1|set:"token":$x1}
}