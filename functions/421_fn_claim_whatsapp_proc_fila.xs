// Claim atomico: so um worker processa o item (UPDATE WHERE disparo=false)
function fn_ClaimWhatsappProcFila {
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

  response = {claimed: $claimed, id: $input.id}
}
