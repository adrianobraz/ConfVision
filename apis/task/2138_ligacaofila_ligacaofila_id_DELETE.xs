// Delete LigacaoFila record
query "ligacaofila/{ligacaofila_id}" verb=DELETE {
  api_group = "task"

  input {
    int ligacaofila_id? filters=min:1
  }

  stack {
    db.del LigacaoFila {
      field_name = "id"
      field_value = $input.ligacaofila_id
    }
  }

  response = null
}