// Lista de eventos com falhas do franqueado
query fp_cd_eventos_falhas verb=POST {
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
          a."idFranqueado"
        FROM x1_3 a
        WHERE a."idFranqueado" = '{{ $input.idFranqueado }}'
          AND a."ctiGrupo" = 'FALHAS'
          AND ('{{ $input.idCliente }}' = '' OR '{{ $input.idCliente }}' = '0' OR a."idCliente" = '{{ $input.idCliente }}')
          AND ({{ $di + 0 }} = 0 OR a."created_at" >= {{ $di + 0 }})
          AND ({{ $df + 0 }} = 0 OR a."created_at" <= {{ $df + 0 }})
        ORDER BY a."created_at" DESC
        LIMIT {{ $per_page + 0 }} OFFSET {{ $offset + 0 }}
        """
      parser = "template_engine"
      response_type = "list"
    } as $dados
  }

  response = {dados: $dados}
}
