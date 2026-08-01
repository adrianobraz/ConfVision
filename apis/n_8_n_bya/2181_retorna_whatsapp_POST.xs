query RetornaWhatsapp verb=POST {
  api_group = "n8nBya"

  input {
    text whatsapp? filters=trim
  }

  stack {
    db.direct_query {
      sql = """
          SELECT 
            w."IdCliente",
            w."nomeCliente",
            w."IdFranqueado",
            w."nomefranqueado",
            w."whatsapp",
            w."nome",
            w."idDispositivo",
            w."NomeDispositivo",
            MAX(e.created_at) as dataUltimoEvento
          FROM x1_5 w
          LEFT JOIN x1_3 e ON e."idDispositivo" = w."idDispositivo"
          WHERE w."whatsapp" = '{{$input.whatsapp}}' and w."bloqueado" = false
          GROUP BY w."IdCliente", w."nomeCliente", w."IdFranqueado", w."nomefranqueado", w."whatsapp", w."nome", w."idDispositivo", w."NomeDispositivo"
          ORDER BY dataUltimoEvento DESC NULLS LAST
        """
      parser = "template_engine"
      response_type = "list"
    } as $CadastroOrdenado
  
    var $TotalReg {
      value = $CadastroOrdenado|count
    }
  }

  response = {
    dados: {}|set:"Total":$TotalReg|set:"Cadastro":$CadastroOrdenado
  }
}