query WhatsappProcFila verb=GET {
  api_group = "testes"

  input {
  }

  stack {
    db.query WhatsappProcFila {
      where = $db.WhatsappProcFila.enviartexto == true
      sort = {WhatsappProcFila.created_at: "desc"}
      return = {type: "list", paging: {page: 1, per_page: 25}}
    } as $WhatsappProcFila1
  }

  response = {dados: $WhatsappProcFila1}
}