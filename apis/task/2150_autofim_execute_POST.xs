query "autofim/execute" verb=POST {
  api_group = "task"

  input {
    int idEvento
  }

  stack {
    precondition (($input.idEvento|is_empty) == false) {
      error = "idEvento obrigatório"
      payload = false
    }
  
    function.run bot_finalizaeventoauto {
      input = {idEvento: $input.idEvento}
    } as $exec
  }

  response = {dados: $exec.dados}
}