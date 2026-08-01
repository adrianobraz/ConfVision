table_trigger alarmeEventos_dispara {
  table = "alarmeEventos_pendente"

  input {
    json new
    json old
    enum action {
      values = ["insert", "update", "delete", "truncate"]
    }
  
    text datasource
  }

  stack {
    conditional {
      if (($input.new|is_empty) == false) {
        // Unificado: so acorda o taskxano (/notify). Nao chama DisparaWhatsapp direto no Xano
        // (evita race com o mutex do Go + ciclo /notify).
        api.request {
          url = "http://185.130.61.4:8081/notify"
          method = "POST"
          headers = []
            |push:"X-Task-Key: 9c3d7f21a6b84e55aa91de73f0c2b118"
        } as $api1
      }
    }
  }

  actions = {insert: true}
  datasources = ["live"]
}
