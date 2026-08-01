// Fila de eventos para enviar texto ou ligacao (WhatsappProcFila)
query fp_cd_fila_envio verb=POST {
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
          f."id",
          f."created_at",
          f."numerowhatsapp",
          f."nomecliente",
          f."empresanome",
          f."tipoevento",
          f."zona",
          f."local",
          f."datahorario",
          f."idevento",
          f."ideventgo",
          f."enviartexto",
          f."ligar",
          f."enviaSom",
          f."analise",
          f."disparo",
          f."dtDisparo",
          f."idDispositivo",
          f."idProcesso",
          f."DispNome",
          f."DispDescricao",
          f."DispTipo",
          f."idFranqueado",
          f."SendMsgWhats"
        FROM x1_92 f
        WHERE f."idFranqueado" = '{{ $input.idFranqueado }}'
          AND ({{ $di + 0 }} = 0 OR f."created_at" >= {{ $di + 0 }})
          AND ({{ $df + 0 }} = 0 OR f."created_at" <= {{ $df + 0 }})
        ORDER BY f."created_at" DESC
        LIMIT {{ $per_page + 0 }} OFFSET {{ $offset + 0 }}
        """
      parser = "template_engine"
      response_type = "list"
    } as $dados
  }

  response = {dados: $dados}
}
