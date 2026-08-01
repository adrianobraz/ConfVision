// Query all task records
query "manutencao/task" verb=GET {
  api_group = "MonitorStatus"

  input {
  }

  stack {
    db.query task {
      return = {type: "single"}
    } as $model
  }

  response = {dados: $model}
}