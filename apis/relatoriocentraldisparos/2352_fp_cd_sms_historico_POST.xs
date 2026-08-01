// Historico de SMS do franqueado
query fp_cd_sms_historico verb=POST {
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
          s."id",
          s."created_at",
          s."success",
          s."message_id",
          s."message_tag",
          s."message_type",
          s."webhook_type",
          s."message_status",
          s."message_status_details",
          s."queued_at",
          s."sent_at",
          s."delivered_at",
          s."failed_at",
          s."whatsappenviados_id",
          w."idFranqueado",
          w."idCliente",
          w."idDispositivo",
          w."DispNome",
          w."ctiGrupo",
          w."whats_sender",
          w."whats_text",
          w."Conta",
          a."nomeCliente"
        FROM x1_103 s
        INNER JOIN x1_7 w ON w."id" = s."whatsappenviados_id"
        LEFT JOIN x1_3 a ON a."id" = w."id_tblAlarm_events"
        WHERE w."idFranqueado" = '{{ $input.idFranqueado }}'
          AND ('{{ $input.idCliente }}' = '' OR '{{ $input.idCliente }}' = '0' OR w."idCliente" = '{{ $input.idCliente }}')
          AND ({{ $di + 0 }} = 0 OR s."created_at" >= {{ $di + 0 }})
          AND ({{ $df + 0 }} = 0 OR s."created_at" <= {{ $df + 0 }})
        ORDER BY s."created_at" DESC
        LIMIT {{ $per_page + 0 }} OFFSET {{ $offset + 0 }}
        """
      parser = "template_engine"
      response_type = "list"
    } as $dados
  }

  response = {dados: $dados}
}
