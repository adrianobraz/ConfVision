query "autofim/pending" verb=GET {
  api_group = "task"

  input {
    int limit?
  }

  stack {
    var $limite {
      value = 50
    }
  
    var $agora {
      value = "now"|to_ms
    }
  
    conditional {
      if (($input.limit|is_empty) == false && $input.limit > 0 && $input.limit <= 200) {
        var.update $limite {
          value = $input.limit
        }
      }
    }
  
    db.direct_query {
      sql = """
          SELECT
            id,
            "idProcesso",
            "idDispositivo",
            status,
            tentativas,
            "alarm_events_id",
            "rodar_em",
            "ultimo_evento_ts",
            "updated_at"
          FROM x1_100
          WHERE status = 'PENDENTE'
            AND "rodar_em" <= {{$agora + 0}}
          ORDER BY id ASC
          LIMIT {{$limite + 0}}
        """
      parser = "template_engine"
      response_type = "list"
    } as $rows
  }

  response = {dados: $rows}
}