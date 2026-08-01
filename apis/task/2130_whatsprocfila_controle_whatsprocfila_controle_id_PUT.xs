// Update WhatsProcFila_Controle record
query "whatsprocfila_controle/{whatsprocfila_controle_id}" verb=PUT {
  api_group = "task"

  input {
    int whatsprocfila_controle_id? filters=min:1
    dblink {
      table = "WhatsProcFila_Controle"
    }
  }

  stack {
    db.edit WhatsProcFila_Controle {
      field_name = "id"
      field_value = $input.whatsprocfila_controle_id
      enforce_hidden_fields = false
      data = {Lock: $input.Lock, update: now}
    } as $model
  }

  response = {dados: $model}
}