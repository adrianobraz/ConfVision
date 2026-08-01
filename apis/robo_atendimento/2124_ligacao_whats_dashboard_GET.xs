query LigacaoWhatsDashboard verb=GET {
  api_group = "RoboAtendimento"

  input {
  }

  stack {
    db.direct_query {
      sql = """
        WITH base_66 AS (
            SELECT
                w."id",
                w."created_at",
                w."franqueado",
                w."telefone",
                w."nomecliente",
                w."empresanome",
                w."tipoevento",
                w."zona",
                w."local",
                w."datahorario",
                w."alarm_events_id",
                w."ideventgo",
                w."falha",
                w."tentativas",
                w."dtUltimaTentativa",
                w."exec",
                w."atendido",
                w."idProcesso",
                w."idDispositivo",
                w."notificarsempre",
                w."DispNome",
                w."DispDescricao",
                w."DispTipo"
            FROM x1_66 w
            ORDER BY w."created_at" DESC, w."id" DESC
            LIMIT 50
        )
        
        SELECT
            b."id"                  AS "id_x1_66",
            b."created_at"          AS "x1_66_created_at",
            b."franqueado",
            b."telefone",
            b."nomecliente",
            b."empresanome",
            b."tipoevento",
            b."zona",
            b."local",
            b."datahorario",
            b."alarm_events_id",
            b."ideventgo",
            b."falha"               AS "x1_66_falha",
            b."tentativas",
            b."dtUltimaTentativa",
            b."exec",
            b."atendido"            AS "x1_66_atendido",
            b."idProcesso",
            b."idDispositivo",
            b."notificarsempre",
            b."DispNome",
            b."DispDescricao",
            b."DispTipo",
        
            l."id"                  AS "id_x1_67",
            l."created_at"          AS "x1_67_created_at",
            l."conversation_id",
            l."success",
            l."whatsapp_call_id",
            l."id_session",
            l."caller",
            l."receiver",
            l."status",
            l."action",
            l."type",
            l."direction",
            l."id_user",
            l."duration",
            l."e_event_type",
            l."e_event_timestamp",
            l."e_agent_id",
            l."e_agent_name",
            l."e_user_id",
            l."e_failure_reason",
            l."e_sip_status_code",
            l."e_error_reason",
            l."e_twirp_code",
            l."e_sip_status",
            l."whatsappligarerro_id",
            l."call_sid",
            l."external_number",
            l."agent_number",
            l."termination_reason",
            l."analysis_success",
            l."analysis_summary_title",
            l."analysis_transcript_summary",
            l."cost_total",
            l."call_charge_credits",
            l."llm_charge_credits",
            l."llm_price_usd",
            l."tier",
            l."is_burst",
            l."free_minutes_consumed",
            l."free_llm_dollars_consumed",
            l."llm_input_tokens",
            l."llm_output_tokens",
            l."llm_input_price",
            l."llm_output_price"
        
        FROM base_66 b
        LEFT JOIN x1_67 l
            ON l."whatsappligarerro_id" = b."id"
        ORDER BY
            b."created_at" DESC,
            b."id" DESC,
            l."created_at" DESC,
            l."id" DESC;
        """
      parser = "template_engine"
      response_type = "list"
    } as $x1
  }

  response = {dados: $x1}
}