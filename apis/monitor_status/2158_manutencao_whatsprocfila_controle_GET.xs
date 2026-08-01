// Query all WhatsProcFila_Controle records
query "manutencao/whatsprocfila_controle" verb=GET {
  api_group = "MonitorStatus"

  input {
  }

  stack {
    db.query WhatsProcFila_Controle {
      return = {type: "single"}
    } as $model
  }

  response = {dados: $model}
}