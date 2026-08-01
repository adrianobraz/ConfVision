// Query all WhatsappProcFila records
query whatsappprocfila verb=GET {
  api_group = "task"

  input {
    bool analise?
    bool disparo?
  }

  stack {
    db.query WhatsappProcFila {
      where = $db.WhatsappProcFila.analise == $input.analise && $db.WhatsappProcFila.disparo == $input.disparo
      sort = {whatsappprocfila.created_at: "desc"}
      return = {type: "list"}
    } as $model
  }

  response = {dados: $model}
}