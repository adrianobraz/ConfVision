query fnDispararLigacaoEleven verb=POST {
  api_group = "WebHooks"

  input {
    text telefone? filters=trim
    text ideventgo? filters=trim
    text nomecliente? filters=trim
    text empresanome? filters=trim
    text tipoevento? filters=trim
    text DispTipo? filters=trim
    text DispNome? filters=trim
    text DispDescricao? filters=trim
    text datahorario? filters=trim
    text eleven_agent? filters=trim
    text eleven_phone? filters=trim
  }

  stack {
    !api.lambda {
      code = """
        try {
          const response = await fetch(
            "https://api.elevenlabs.io/v1/convai/sip-trunk/outbound-call",
            {
              method: "POST",
              headers: {
                "xi-api-key": "b812beb864b947cbe92ff634e52056dec1eb75ef48b62760cbaf04bbf4d0f84b",
                "Content-Type": "application/json"
              },
              body: JSON.stringify({
                agent_id: "agent_0301kg5ad62veeer9h4fs6qqh3qz",
                agent_phone_number_id: "phnum_6401kr3qfq5wfr8vwmkgmvw3vrm5",
                to_number: $input.telefone,
                conversation_initiation_client_data: {
                  user_id: $input.ideventgo,
                  dynamic_variables: {
                    nome_cliente: $input.nomecliente,
                    empresa: $input.empresanome,
                    tipo_evento: $input.tipoevento,
                    zona: [$input.DispTipo, $input.DispNome].filter(Boolean).join(", "),
                    local: $input.DispDescricao,
                    horario_evento: $input.datahorario
                  }
                }
              })
            }
          );
        
          const data = await response.json().catch(() => null);
        
          let sipCode = null;
          if (data?.message) {
            const match = data.message.match(/sip status:\s*(\d+)/i);
            if (match) {
              sipCode = match[1];
            }
          }
        
          return {
            http_success: response.ok,
            call_success: data?.success === true,
            sip_code: sipCode,
            sip_message: data?.message || null,
            conversation_id: data?.conversation_id || null,
            sip_call_id: data?.sip_call_id || null,
            status: response.status,
            raw: data
          };
        
        } catch (error) {
          return {
            http_success: false,
            call_success: false,
            sip_code: null,
            sip_message: error.message,
            status: 500,
            raw: null
          };
        }
        """
      timeout = 1
    } as $x1
  
    api.lambda {
      code = """
        try {
          const response = await fetch(
            "https://api.elevenlabs.io/v1/convai/sip-trunk/outbound-call",
            {
              method: "POST",
              headers: {
                "xi-api-key": "b812beb864b947cbe92ff634e52056dec1eb75ef48b62760cbaf04bbf4d0f84b",
                "Content-Type": "application/json"
              },
              body: JSON.stringify({
                agent_id: $input.eleven_agent,
                agent_phone_number_id: $input.eleven_phone,
                to_number: $input.telefone,
                conversation_initiation_client_data: {
                  user_id: $input.ideventgo,
                  dynamic_variables: {
                    nome_cliente: $input.nomecliente,
                    empresa: $input.empresanome,
                    tipo_evento: $input.tipoevento,
                    zona: [$input.DispTipo, $input.DispNome].filter(Boolean).join(", "),
                    local: $input.DispDescricao,
                    horario_evento: $input.datahorario
                  }
                }
              })
            }
          );
        
          const data = await response.json().catch(() => null);
        
          let sipCode = null;
          if (data?.message) {
            const match = data.message.match(/sip status:\s*(\d+)/i);
            if (match) {
              sipCode = match[1];
            }
          }
        
          return {
            http_success: response.ok,
            call_success: data?.success === true,
            sip_code: sipCode,
            sip_message: data?.message || null,
            conversation_id: data?.conversation_id || null,
            sip_call_id: data?.sip_call_id || null,
            status: response.status,
            raw: data
          };
        
        } catch (error) {
          return {
            http_success: false,
            call_success: false,
            sip_code: null,
            sip_message: error.message,
            status: 500,
            raw: null
          };
        }
        """
      timeout = 1
    } as $x1
  }

  response = $x1
}