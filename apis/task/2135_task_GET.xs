// Query all task records
query task verb=GET {
  api_group = "task"

  input {
  }

  stack {
    db.query task {
      where = $db.task.Running == true && $db.task.update < (now|timestamp_add_minutes:-5)
      return = {type: "single"}
    } as $task1
  
    conditional {
      if (($task1|is_empty) == false) {
        db.edit task {
          field_name = "id"
          field_value = $task1.id
          enforce_hidden_fields = false
          data = {Running: false, update: now}
        } as $task2
      }
    }
  
    db.query task {
      return = {type: "single"}
    } as $model
  }

  response = {dados: $model}
}