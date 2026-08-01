query SendVideoWhatsBNuvem verb=POST {
  api_group = "FuncaoSistema"

  input {
    int ID?
  }

  stack {
    db.get WhatsAppEnviados {
      field_name = "id"
      field_value = $input.ID
    } as $WhatsAppEnviados1
  
    precondition ($WhatsAppEnviados1 != null) {
      payload = {success: false, message: "Cliente sem BENUVEM"}
    }
  
    conditional {
      if ($WhatsAppEnviados1.codigoBenuvem != null || ($WhatsAppEnviados1.codigoBenuvem|is_empty) != false && $WhatsAppEnviados1.ctiGrupo == "ALARME") {
        function.run FuncaoSistema_GeraTokenBeNuvem {
          input = {
            email   : "gustavo.reis@redeconfinet.com.br"
            password: "Confia2022"
          }
        } as $func_1
      
        api.request {
          url = "https://app.benuvem.com.br/api/v1/cameras/get-records"
          method = "POST"
          params = {}
            |set:"client_code":$WhatsAppEnviados1.Conta
            |set:"partition":$WhatsAppEnviados1.Particao
            |set:"company_code":$WhatsAppEnviados1.codigoBenuvem
            |set:"channel":$WhatsAppEnviados1.ZonaUser
            |set:"date_start":`$WhatsAppEnviados1.created_at|add_secs_to_timestamp:-300|format_timestamp:"Y-m-d H:i:s":timezone:"America/Sao_Paulo"`
            |set:"date_end":`$WhatsAppEnviados1.created_at|add_secs_to_timestamp:300|format_timestamp:"Y-m-d H:i:s":timezone:"America/Sao_Paulo"`
          headers = []
            |push:("Authorization: Bearer"|concat:$func_1:" ")
            |push:"Content-Type: application/x-www-form-urlencoded"
        } as $api1|set:"":`$api1.response.result`
      }
    
      else {
        var $api1 {
          value = {success: false, message: "erro ZONA"}
        }
      }
    }
  }

  response = $api1
}