table_trigger whatsprocmsg {
  table = "WhatsappProcFila"

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
        api.request {
          url = "http://185.130.61.4:8081/notify"
          method = "POST"
          headers = []
            |push:"X-Task-Key: 9c3d7f21a6b84e55aa91de73f0c2b118"
        } as $api1
      
        !api.request {
          url = "https://xpcy-oyme-lno7.b2.xano.io/api:Maqjhbl9/DisparaWhatsapp"
          method = "GET"
          headers = []
            |push:"Content-Type: application/json"
          timeout = 1
        } as $api1
      }
    }
  }

  actions = {insert: true}
  datasources = ["live"]
}