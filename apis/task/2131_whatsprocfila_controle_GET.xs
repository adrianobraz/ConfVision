// Query all WhatsProcFila_Controle records
query whatsprocfila_controle verb=GET {
  api_group = "task"

  input {
  }

  stack {
    db.query WhatsProcFila_Controle {
      where = $db.WhatsProcFila_Controle.Lock == true && $db.WhatsProcFila_Controle.update < (now|timestamp_add_minutes:-5)
      return = {type: "single"}
    } as $WhatsProcFila_Controle99
  
    conditional {
      if (($WhatsProcFila_Controle99|is_empty) == false) {
        db.edit WhatsProcFila_Controle {
          field_name = "id"
          field_value = $WhatsProcFila_Controle99.id
          enforce_hidden_fields = false
          data = {Lock: false, update: now}
        } as $task2
      }
    }
  
    db.query WhatsProcFila_Controle {
      return = {type: "list"}
    } as $model
  }

  response = {dados: $model}
}