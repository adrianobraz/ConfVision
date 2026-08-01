function OpenAi_Texto {
  input {
    text SendMsgWhats? filters=trim
  }

  stack {
    api.request {
      url = "https://api.openai.com/v1/chat/completions"
      method = "POST"
      params = {}
        |set:"model":"gpt-4.1-nano"
        |set:"messages":([]
          |push:({}
            |set:"role":"system"
            |set:"content":"""
        Você receberá dados de um evento de alarme.
        
        Mensagem de WhatsApp natural, profissional e variada, mas mantendo TODOS os dados obrigatórios.
        
        DADOS OBRIGATÓRIOS — nunca remover, resumir, cortar ou alterar:
        
        Cliente
        Data
        Evento
        Dispositivo
        Setor
        Partição / Conta
        Link completo
        Empresa Assinatura
        
        Monte a mensagem no WhatsApp exatamente nessa ordem:
        1. Os 6 campos de dados em negrito (ordem variada entre eles). Ordem Aleatoria
        2. Linha em branco
        3. Parágrafo curto humanizado (máximo 2 linhas)
        4. Linha em branco
        5. Link (sem alterar)
        6. Linha em branco
        7. Assinatura em itálico
        
        CAMPOS — todos obrigatórios, nenhum pode faltar:
        *Cliente:* {valor}
        *Data:* {valor}
        *Evento:* {valor}
        *Dispositivo:* {valor}
        *Setor:* {valor}
        *Partição:* {valor}
        *Link:* {valor}
        *Empresa:* {valor}
        
        REGRAS:
        * Nunca omita nenhum campo.
        * Negrito nos rótulos (*Campo:*).
        * Assinatura final: _Equipe de Monitoramento, (*Campo Empresa:*)
        * Link nunca alterado, sempre em linha separada.
        * Parágrafo sem repetir dados, sem assumir causa.
        * Varie o parágrafo a cada mensagem.
        
        O link deve ficar completo e exatamente igual.
        Não altere nomes, datas, evento, setor, partição, conta ou assinatura.
        Pode variar a ordem dos campos.
        Pode variar a frase humanizada.
        Não invente causa.
        Não use texto longo.
        
        Retorne SOMENTE a mensagem final.
        """
          )
          |push:({}
            |set:"role":"user"
            |set:"content":$input.SendMsgWhats
          )
        )
        |set:"temperature":1
        |set:"max_completion_tokens":300
      headers = []
        |push:"Content-Type: application/json"
        |push:"Authorization: Bearer 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"
    } as $api2
  }

  response = {dados: $api2.response.result.choices[0].message.content}
}