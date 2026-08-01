query "autofim/context" verb=GET {
  api_group = "task"

  input {
    int idEvento
  }

  stack {
    precondition (($input.idEvento|is_empty) == false) {
      error = "idEvento obrigatório"
      payload = false
    }
  
    db.get alarm_events {
      field_name = "id"
      field_value = $input.idEvento
    } as $evt
  
    precondition (($evt.id|is_empty) == false) {
      error = "Evento não encontrado"
      payload = false
    }
  
    precondition (($evt.idProcesso|is_empty) == false) {
      error = "Evento sem idProcesso"
      payload = false
    }
  
    var $apiProcOk {
      value = false
    }
  
    var $dadosProcessoApi {
      value = {}
    }
  
    var $dataAtenFim {
      value = ""
    }
  
    function.run WebLogarCache as $tokenWeb
    api.request {
      url = "http://185.130.61.4:2010/v4/terminal/getDadosProcessoById"
      method = "POST"
      params = {}|set:"idProcesso":$evt.idProcesso
      headers = []
        |push:"Authorization: Bearer " ~ $tokenWeb
        |push:"Content-Type: application/json"
    } as $apiProc
  
    conditional {
      if (($apiProc.response.result.dados|is_empty) == false) {
        var.update $apiProcOk {
          value = true
        }
      
        var.update $dadosProcessoApi {
          value = $apiProc.response.result.dados
        }
      
        var.update $dataAtenFim {
          value = $apiProc.response.result.dados.dataAtenFim
        }
      }
    }
  
    var $processoJaFinalizado {
      value = false
    }
  
    conditional {
      if (($dataAtenFim|is_empty) == false && $dataAtenFim != "01/01/0001 00:00:00") {
        var.update $processoJaFinalizado {
          value = true
        }
      }
    }
  
    db.direct_query {
      sql = """
          WITH ultimos AS (
            SELECT
              e."idProcesso",
              e."ctiGrupo",
              CASE
                WHEN (e."created_at"::bigint) > 9999999999 THEN (e."created_at"::bigint / 1000)
                ELSE (e."created_at"::bigint)
              END AS ts_s
            FROM x1_3 e
            WHERE e."idDispositivo" = '{{$evt.idDispositivo}}'
            ORDER BY e."created_at"::bigint DESC
            LIMIT 150
          ),
          hist AS (
            SELECT
              COALESCE(COUNT(*) FILTER (WHERE "ctiGrupo" = 'ALARME'), 0) AS qtd_alarme_hist,
              COALESCE(COUNT(*) FILTER (WHERE "ctiGrupo" = 'RESTAURE'), 0) AS qtd_restaure_hist
            FROM ultimos
          ),
          janela3m AS (
            SELECT *
            FROM ultimos
            WHERE ts_s >= (EXTRACT(EPOCH FROM NOW())::bigint - 180)
          ),
          proc_flags AS (
            SELECT
              "idProcesso",
              MAX(CASE WHEN "ctiGrupo" = 'ALARME' THEN 1 ELSE 0 END) AS tem_alarme,
              MAX(CASE WHEN "ctiGrupo" = 'RESTAURE' THEN 1 ELSE 0 END) AS tem_restaure
            FROM janela3m
            GROUP BY "idProcesso"
          ),
          ciclos3m AS (
            SELECT COALESCE(COUNT(*), 0)::integer AS qtd_ciclos_3m_diff_proc
            FROM proc_flags
            WHERE tem_alarme = 1 AND tem_restaure = 1
          )
          SELECT
            COALESCE(h.qtd_alarme_hist, 0) AS qtd_alarme_hist,
            COALESCE(h.qtd_restaure_hist, 0) AS qtd_restaure_hist,
            COALESCE(c.qtd_ciclos_3m_diff_proc, 0) AS qtd_ciclos_3m_diff_proc,
            CASE
              WHEN COALESCE(h.qtd_alarme_hist, 0) >= 40
               AND COALESCE(h.qtd_restaure_hist, 0) >= (COALESCE(h.qtd_alarme_hist, 0) * 0.7)
              THEN 'PORTAO'
              ELSE 'NORMAL'
            END AS perfil_local
          FROM hist h
          CROSS JOIN ciclos3m c
        """
      parser = "template_engine"
      response_type = "single"
    } as $histAggRaw
  
    var $perfilLocal {
      value = "NORMAL"
    }
  
    conditional {
      if (($histAggRaw.perfil_local|is_empty) == false) {
        var.update $perfilLocal {
          value = $histAggRaw.perfil_local
        }
      }
    }
  
    var $histAgg {
      value = {}
        |set:"qtd_alarme_hist":$histAggRaw.qtd_alarme_hist + 0
        |set:"qtd_restaure_hist":$histAggRaw.qtd_restaure_hist + 0
        |set:"qtd_ciclos_3m_diff_proc":$histAggRaw.qtd_ciclos_3m_diff_proc + 0
        |set:"perfil_local":$perfilLocal
    }
  
    db.query alarm_events {
      where = $db.alarm_events.idProcesso == $evt.idProcesso
      sort = {alarm_events.created_at: "asc"}
      return = {type: "list"}
    } as $eventosProcesso
  }

  response = {
    dados: ""|set:"evento":$evt|set:"eventosProcesso":$eventosProcesso|set:"histAgg":$histAgg|set:"processoJaFinalizado":$processoJaFinalizado|set:"dataAtenFim":$dataAtenFim|set:"dadosProcessoApi":$dadosProcessoApi|set:"apiProcOk":$apiProcOk
  }
}