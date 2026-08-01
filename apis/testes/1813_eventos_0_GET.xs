query eventos_0 verb=GET {
  api_group = "testes"

  input {
  }

  stack {
    api.realtime_event {
      channel = "tester"
      data = "ok"
      auth_table = "0"
      auth_id = ""
    }
  }

  response = "ok"
}