// Claim atomico de item da fila WhatsappProcFila (evita envio duplicado)
query "whatsappprocfila/claim" verb=POST {
  api_group = "task"

  input {
    int id filters=min:1
  }

  stack {
    db.direct_query {
      sql = """
        UPDATE x1_92
        SET "disparo" = true,
            "dtDisparo" = NOW()
        WHERE "id" = {{ $input.id + 0 }}
          AND COALESCE("disparo", false) = false
          AND COALESCE("analise", false) = false
        RETURNING "id"
        """
      parser = "template_engine"
      response_type = "list"
    } as $claimed_rows
  
    var $claimed {
      value = false
    }
  
    conditional {
      if (($claimed_rows|is_empty) == false) {
        var.update $claimed {
          value = true
        }
      }
    }
  }

  response = {dados: ""|set:"claimed":$claimed|set:"id":$input.id}
}