query erro verb=POST {
  api_group = "testes"

  input {
  }

  stack {
    !db.query whatsappLigarErro {
      return = {
        type  : "list"
        paging: {page: 1, per_page: 25, metadata: false}
      }
    } as $whatsappLigarErro1
  
    db.query contactId_Erro {
      return = {
        type : "aggregate"
        group: {contactId_Erro_codigo1: $db.contactId_Erro.codigo}
      }
    } as $contactId_Erro1
  
    !api.realtime_event {
      channel = "GraficoEventos"
      data = $whatsappLigarErro1
      auth_table = "0"
      auth_id = ""
    }
  }

  response = $contactId_Erro1
}