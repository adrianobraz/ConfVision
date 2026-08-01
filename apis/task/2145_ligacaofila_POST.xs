// Add LigacaoFila record
query ligacaofila verb=POST {
  api_group = "task"

  input {
    dblink {
      table = "LigacaoFila"
    }
  }

  stack {
    db.add LigacaoFila {
      enforce_hidden_fields = false
      data = {
        created_at          : "now"
        whatsappligarerro_id: $input.whatsappligarerro_id
      }
    } as $model
  }

  response = {dados: $model}
}