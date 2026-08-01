query AlarmeEventos_Pendente verb=GET {
  api_group = "task"

  input {
  }

  stack {
    function.run fn_AlarmeEventos_Pendente as $func1
  }

  response = {dados: $func1}
}