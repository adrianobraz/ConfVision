// Update task record
query "task/{task_id}" verb=PUT {
  api_group = "task"

  input {
    int task_id? filters=min:1
    dblink {
      table = "task"
    }
  }

  stack {
    db.edit task {
      field_name = "id"
      field_value = $input.task_id
      enforce_hidden_fields = false
      data = {Running: $input.Running, update: now}
    } as $model
  }

  response = {dados: $model}
}