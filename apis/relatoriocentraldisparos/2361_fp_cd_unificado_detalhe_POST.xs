// Detalhe agregado de um evento (todas as fontes CD vinculadas)
query fp_cd_unificado_detalhe verb=POST {
  api_group = "Relatoriocentraldisparos"

  input {
    text idFranqueado filters=trim
    int alarmEventsId?
    text idProcesso? filters=trim
  }

  stack {
    precondition (($input.idFranqueado|is_empty) == false) {
      error = "idFranqueado obrigatorio"
    }
  
    precondition (($input.alarmEventsId > 0) || (($input.idProcesso|is_empty) == false)) {
      error = "alarmEventsId ou idProcesso obrigatorio"
    }
  
    var $ae {
      value = 0
    }
  
    var $proc {
      value = ""
    }
  
    conditional {
      if ($input.alarmEventsId > 0) {
        var.update $ae {
          value = $input.alarmEventsId
        }
      }
    }
  
    conditional {
      if (($input.idProcesso|is_empty) == false) {
        var.update $proc {
          value = $input.idProcesso|replace:"'":""|replace:'"':""
        }
      }
    }
  
    // Evento mestre
    db.direct_query {
      sql = """
        SELECT
          a."id",
          a."created_at",
          a."nomeCliente",
          a."idCliente",
          a."idDispositivo",
          a."idProcesso",
          a."ctiGrupo",
          a."ctiDescricao",
          a."codigo",
          a."particao",
          a."zonaUser",
          a."dataEntrada",
          a."conta",
          a."nivel",
          a."idFranqueado"
        FROM x1_3 a
        WHERE a."idFranqueado" = '{{ $input.idFranqueado }}'
          AND (
            ({{ $ae + 0 }} > 0 AND a."id" = {{ $ae + 0 }})
            OR ('{{ $proc }}' <> '' AND a."idProcesso" = '{{ $proc }}')
          )
        ORDER BY a."id" DESC
        LIMIT 1
        """
      parser = "template_engine"
      response_type = "list"
    } as $evento_rows
  
    var $evento {
      value = null
    }
  
    conditional {
      if (($evento_rows|is_empty) == false) {
        var.update $evento {
          value = $evento_rows|first
        }
      
        conditional {
          if ($ae == 0 && ($evento.id|is_empty) == false) {
            var.update $ae {
              value = $evento.id
            }
          }
        }
      
        conditional {
          if ($proc == "" && ($evento.idProcesso|is_empty) == false) {
            var.update $proc {
              value = $evento.idProcesso|replace:"'":""|replace:'"':""
            }
          }
        }
      }
    }
  
    // Pendentes
    db.direct_query {
      sql = """
        SELECT
          p."id",
          p."created_at",
          p."idDispositivo",
          p."zonaUser",
          p."particao",
          p."grupo",
          p."agendadoPara",
          p."cancelado",
          p."enviado"
        FROM x1_93 p
        WHERE {{ $ae + 0 }} > 0
          AND p."id_tblAlarm_events" = {{ $ae + 0 }}
        ORDER BY p."created_at" DESC
        LIMIT 50
        """
      parser = "template_engine"
      response_type = "list"
    } as $pendente
  
    // Finalizados robo
    db.direct_query {
      sql = """
        SELECT
          f."id",
          f."created_at",
          f."idprocesso",
          f."telefone",
          f."mensagem",
          f."Usuario"
        FROM x1_97 f
        WHERE '{{ $proc }}' <> ''
          AND f."idprocesso" = '{{ $proc }}'
        ORDER BY f."created_at" DESC
        LIMIT 50
        """
      parser = "template_engine"
      response_type = "list"
    } as $finalizadoRobo
  
    // Finalizados bot
    db.direct_query {
      sql = """
        SELECT
          b."id",
          b."created_at",
          b."status",
          b."motivo_final",
          b."acao_final",
          b."erro",
          b."tentativas",
          b."alarm_events_id",
          b."idProcesso"
        FROM x1_100 b
        WHERE (
            ({{ $ae + 0 }} > 0 AND b."alarm_events_id" = {{ $ae + 0 }})
            OR ('{{ $proc }}' <> '' AND b."idProcesso" = '{{ $proc }}')
          )
        ORDER BY b."created_at" DESC
        LIMIT 50
        """
      parser = "template_engine"
      response_type = "list"
    } as $finalizadoBot
  
    // WhatsApp enviados (texto, sem SMS/ligacao)
    db.direct_query {
      sql = """
        SELECT
          w."id",
          w."created_at",
          w."whats_sender",
          w."whats_text",
          w."ctiGrupo",
          w."DispNome",
          w."texto",
          w."SMS",
          w."ligacao",
          w."audio"
        FROM x1_7 w
        WHERE w."idFranqueado" = '{{ $input.idFranqueado }}'
          AND COALESCE(w."texto", false) = true
          AND COALESCE(w."SMS", false) = false
          AND COALESCE(w."ligacao", false) = false
          AND COALESCE(w."audio", false) = false
          AND (
            ({{ $ae + 0 }} > 0 AND w."id_tblAlarm_events" = {{ $ae + 0 }})
            OR ('{{ $proc }}' <> '' AND w."idProcesso" = '{{ $proc }}')
          )
        ORDER BY w."created_at" DESC
        LIMIT 50
        """
      parser = "template_engine"
      response_type = "list"
    } as $whatsappEnviados
  
    // Ligacoes (WhatsAppEnviados ligacao=true)
    db.direct_query {
      sql = """
        SELECT
          w."id",
          w."created_at",
          w."whats_sender",
          w."whats_text",
          w."ctiGrupo",
          w."DispNome"
        FROM x1_7 w
        WHERE w."idFranqueado" = '{{ $input.idFranqueado }}'
          AND COALESCE(w."ligacao", false) = true
          AND (
            ({{ $ae + 0 }} > 0 AND w."id_tblAlarm_events" = {{ $ae + 0 }})
            OR ('{{ $proc }}' <> '' AND w."idProcesso" = '{{ $proc }}')
          )
        ORDER BY w."created_at" DESC
        LIMIT 50
        """
      parser = "template_engine"
      response_type = "list"
    } as $ligacoes
  
    // SMS (texto + SMS + whats_ID=sms)
    db.direct_query {
      sql = """
        SELECT
          w."id",
          w."created_at",
          w."whats_sender",
          w."whats_text",
          w."ctiGrupo",
          w."DispNome"
        FROM x1_7 w
        WHERE w."idFranqueado" = '{{ $input.idFranqueado }}'
          AND COALESCE(w."texto", false) = true
          AND COALESCE(w."SMS", false) = true
          AND LOWER(COALESCE(w."whats_ID", '')) = 'sms'
          AND (
            ({{ $ae + 0 }} > 0 AND w."id_tblAlarm_events" = {{ $ae + 0 }})
            OR ('{{ $proc }}' <> '' AND w."idProcesso" = '{{ $proc }}')
          )
        ORDER BY w."created_at" DESC
        LIMIT 50
        """
      parser = "template_engine"
      response_type = "list"
    } as $sms
  
    // SMS historico
    db.direct_query {
      sql = """
        SELECT
          s."id",
          s."created_at",
          s."message_status",
          s."success",
          w."whats_sender",
          w."whats_text",
          w."DispNome",
          w."ctiGrupo"
        FROM x1_103 s
        INNER JOIN x1_7 w ON w."id" = s."whatsappenviados_id"
        WHERE w."idFranqueado" = '{{ $input.idFranqueado }}'
          AND (
            ({{ $ae + 0 }} > 0 AND w."id_tblAlarm_events" = {{ $ae + 0 }})
            OR ('{{ $proc }}' <> '' AND w."idProcesso" = '{{ $proc }}')
          )
        ORDER BY s."created_at" DESC
        LIMIT 50
        """
      parser = "template_engine"
      response_type = "list"
    } as $smsHistorico
  
    // Ligacao historico (via whatsappLigarErro)
    db.direct_query {
      sql = """
        SELECT
          l."id",
          l."created_at",
          l."conversation_id",
          l."receiver",
          l."duration",
          l."status",
          l."analysis_transcript_summary",
          w."nomecliente",
          w."tipoevento",
          w."telefone"
        FROM x1_67 l
        INNER JOIN x1_66 w ON w."id" = l."whatsappligarerro_id"
        WHERE w."franqueado" = '{{ $input.idFranqueado }}'
          AND (
            ({{ $ae + 0 }} > 0 AND w."alarm_events_id" = {{ $ae + 0 }})
            OR ('{{ $proc }}' <> '' AND w."idProcesso" = '{{ $proc }}')
          )
        ORDER BY l."created_at" DESC
        LIMIT 50
        """
      parser = "template_engine"
      response_type = "list"
    } as $ligacaoHistorico
  
    // Erros e tentativas
    db.direct_query {
      sql = """
        SELECT
          w."id",
          w."created_at",
          w."telefone",
          w."nomecliente",
          w."tipoevento",
          w."falha",
          w."tentativas",
          w."atendido",
          w."exec",
          w."dtUltimaTentativa",
          w."DispNome",
          w."nroErr"
        FROM x1_66 w
        WHERE w."franqueado" = '{{ $input.idFranqueado }}'
          AND (
            ({{ $ae + 0 }} > 0 AND w."alarm_events_id" = {{ $ae + 0 }})
            OR ('{{ $proc }}' <> '' AND w."idProcesso" = '{{ $proc }}')
          )
        ORDER BY w."created_at" DESC
        LIMIT 50
        """
      parser = "template_engine"
      response_type = "list"
    } as $ligacaoErros
  
    // Fila envio
    db.direct_query {
      sql = """
        SELECT
          f."id",
          f."created_at",
          f."numerowhatsapp",
          f."nomecliente",
          f."tipoevento",
          f."enviartexto",
          f."ligar",
          f."disparo",
          f."analise",
          f."DispNome",
          f."zona"
        FROM x1_92 f
        WHERE f."idFranqueado" = '{{ $input.idFranqueado }}'
          AND (
            ('{{ $proc }}' <> '' AND f."idProcesso" = '{{ $proc }}')
            OR ({{ $ae + 0 }} > 0 AND (
              f."ideventgo" = '{{ $ae + 0 }}'
              OR f."idevento" = '{{ $ae + 0 }}'
            ))
          )
        ORDER BY f."created_at" DESC
        LIMIT 50
        """
      parser = "template_engine"
      response_type = "list"
    } as $filaEnvio
  
    // Fila ligacao
    db.direct_query {
      sql = """
        SELECT
          f."id",
          f."created_at",
          w."telefone",
          w."nomecliente",
          w."tipoevento",
          w."tentativas",
          w."falha",
          w."exec",
          w."atendido",
          w."DispNome",
          w."zona"
        FROM x1_70 f
        INNER JOIN x1_66 w ON w."id" = f."whatsappligarerro_id"
        WHERE w."franqueado" = '{{ $input.idFranqueado }}'
          AND (
            ({{ $ae + 0 }} > 0 AND w."alarm_events_id" = {{ $ae + 0 }})
            OR ('{{ $proc }}' <> '' AND w."idProcesso" = '{{ $proc }}')
          )
        ORDER BY f."created_at" DESC
        LIMIT 50
        """
      parser = "template_engine"
      response_type = "list"
    } as $filaLigacao
  }

  response = {
    evento          : $evento
    pendente        : $pendente
    finalizadoRobo  : $finalizadoRobo
    finalizadoBot   : $finalizadoBot
    whatsappEnviados: $whatsappEnviados
    ligacoes        : $ligacoes
    sms             : $sms
    smsHistorico    : $smsHistorico
    ligacaoHistorico: $ligacaoHistorico
    ligacaoErros    : $ligacaoErros
    filaEnvio       : $filaEnvio
    filaLigacao     : $filaLigacao
  }
}