// Watchdog: destrava registros de ligacao presos em exec=true (crash/timeout do worker).
// Marca falha=true para os workers de retry (ligacao060/180/300/600) reprocessarem.
query "whatsappligarerro/destravar" verb=POST {
  api_group = "task"

  input {
    int minutos?=15 filters=min:1
  }

  stack {
    db.direct_query {
      sql = """
        UPDATE x1_66
        SET "exec" = false,
            "falha" = true
        WHERE COALESCE("exec", false) = true
          AND COALESCE("atendido", false) = false
          AND COALESCE("cancelado", false) = false
          AND COALESCE("tentativas", 0) < 99
          AND (
            CASE
              WHEN (COALESCE("dtUltimaTentativa", "created_at")::bigint) > 9999999999
                THEN (COALESCE("dtUltimaTentativa", "created_at")::bigint / 1000)
              ELSE (COALESCE("dtUltimaTentativa", "created_at")::bigint)
            END
          ) < (EXTRACT(EPOCH FROM NOW())::bigint - ({{ $input.minutos + 0 }} * 60))
        RETURNING "id"
        """
      parser = "template_engine"
      response_type = "list"
    } as $destravados
  
    var $qtd {
      value = 0
    }
  
    conditional {
      if (($destravados|is_empty) == false) {
        var.update $qtd {
          value = $destravados|count
        }
      }
    }
  }

  response = {
    dados: ""|set:"destravados":$qtd|set:"ids":$destravados
  }
}