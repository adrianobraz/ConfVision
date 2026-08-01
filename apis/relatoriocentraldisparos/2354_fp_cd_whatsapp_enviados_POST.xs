// WhatsApp enviados do franqueado (somente texto; exclui SMS/audio/ligacao)
query fp_cd_whatsapp_enviados verb=POST {
  api_group = "Relatoriocentraldisparos"

  input {
    text idFranqueado filters=trim
    text idCliente? filters=trim
    int dataInicioMs?
    int dataFimMs?
    int limit?=100
    int offset?=0
  }

  stack {
    precondition (($input.idFranqueado|is_empty) == false) {
      error = "idFranqueado obrigatorio"
    }
  
    var $per_page {
      value = 100
    }
  
    var $offset {
      value = 0
    }
  
    var $di {
      value = 0
    }
  
    var $df {
      value = 0
    }
  
    conditional {
      if ($input.offset > 0) {
        var.update $offset {
          value = $input.offset
        }
      }
    }
  
    conditional {
      if ($input.limit > 0 && $input.limit < 2000) {
        var.update $per_page {
          value = $input.limit
        }
      }
    }
  
    conditional {
      if (($input.dataInicioMs|is_empty) == false && $input.dataInicioMs > 0) {
        var.update $di {
          value = $input.dataInicioMs
        }
      }
    }
  
    conditional {
      if (($input.dataFimMs|is_empty) == false && $input.dataFimMs > 0) {
        var.update $df {
          value = $input.dataFimMs
        }
      }
    }
  
    db.direct_query {
      sql = """
        SELECT
          w."id",
          w."created_at",
          w."whats_sender",
          w."whats_senderName",
          w."whats_text",
          w."idCliente",
          w."idDispositivo",
          w."idProcesso",
          w."ctiGrupo",
          w."Conta",
          w."Particao",
          w."ZonaUser",
          w."DispNome",
          w."DispDescricao",
          w."texto",
          w."idFranqueado",
          w."id_tblAlarm_events",
          a."nomeCliente",
          a."ctiDescricao"
        FROM x1_7 w
        LEFT JOIN x1_3 a ON a."id" = w."id_tblAlarm_events"
        WHERE w."idFranqueado" = '{{ $input.idFranqueado }}'
          AND w."texto" = true
          AND COALESCE(w."SMS", false) = false
          AND COALESCE(w."audio", false) = false
          AND COALESCE(w."ligacao", false) = false
          AND COALESCE(w."whats_sender", '') <> 'SMS'
          AND ('{{ $input.idCliente }}' = '' OR '{{ $input.idCliente }}' = '0' OR w."idCliente" = '{{ $input.idCliente }}')
          AND ({{ $di + 0 }} = 0 OR w."created_at" >= {{ $di + 0 }})
          AND ({{ $df + 0 }} = 0 OR w."created_at" <= {{ $df + 0 }})
        ORDER BY w."created_at" DESC
        LIMIT {{ $per_page + 0 }} OFFSET {{ $offset + 0 }}
        """
      parser = "template_engine"
      response_type = "list"
    } as $dados
  }

  response = {dados: $dados}
}
