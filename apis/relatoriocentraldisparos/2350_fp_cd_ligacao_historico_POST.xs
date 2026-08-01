// Historico de ligacoes com transcricoes
query fp_cd_ligacao_historico verb=POST {
  api_group = "Relatoriocentraldisparos"

    input {
    text idFranqueado filters=trim
    text idCliente? filters=trim
    text nomeCliente? filters=trim
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
  
    var $nome_cli {
      value = ""
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
  
    conditional {
      if (($input.nomeCliente|is_empty) == false && $input.idCliente != "0" && ($input.idCliente|is_empty) == false) {
        var.update $nome_cli {
          value = $input.nomeCliente|replace:"'":""|replace:'"':""
        }
      }
    }
  
    db.direct_query {
      sql = """
        SELECT
          l."id",
          l."created_at",
          l."conversation_id",
          l."success",
          l."caller",
          l."receiver",
          l."status",
          l."duration",
          l."direction",
          l."analysis_success",
          l."analysis_summary_title",
          l."analysis_transcript_summary",
          l."termination_reason",
          l."cost_total",
          l."whatsappligarerro_id",
          w."franqueado",
          w."telefone",
          w."nomecliente",
          w."empresanome",
          w."tipoevento",
          w."zona",
          w."local",
          w."idDispositivo",
          w."idProcesso",
          w."DispNome",
          w."falha",
          w."tentativas",
          w."atendido"
        FROM x1_67 l
        INNER JOIN x1_66 w ON w."id" = l."whatsappligarerro_id"
        WHERE w."franqueado" = '{{ $input.idFranqueado }}'
          AND ({{ $di + 0 }} = 0 OR l."created_at" >= {{ $di + 0 }})
          AND ({{ $df + 0 }} = 0 OR l."created_at" <= {{ $df + 0 }})
          AND ('{{ $nome_cli }}' = '' OR LOWER(w."nomecliente") LIKE '%' || LOWER('{{ $nome_cli }}') || '%')
        ORDER BY l."created_at" DESC
        LIMIT {{ $per_page + 0 }} OFFSET {{ $offset + 0 }}
        """
      parser = "template_engine"
      response_type = "list"
    } as $dados
  }

  response = {dados: $dados}
}
