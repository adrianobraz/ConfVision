// Query all LigacaoFila records
query ligacaofila verb=GET {
  api_group = "task"

  input {
  }

  stack {
    db.query LigacaoFila {
      return = {type: "list"}
    } as $model
  }

  response = {dados: $model}
}