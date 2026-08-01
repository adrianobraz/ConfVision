// Query all pro_adm_user records
query pro_adm_user_0 verb=GET {
  api_group = "ProGerencial GoFluxo"

  input {
    int idadm?
  }

  stack {
    api.realtime_event {
      channel = "GraficoEventos/1"
      data = "Realtime"
      auth_table = "0"
      auth_id = ""
    }
  }

  response = {data: "realtime"}
}