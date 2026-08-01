table whatsappLigarErro {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text franqueado? filters=trim
    text telefone? filters=trim
    text nomecliente? filters=trim
    text empresanome? filters=trim
    text tipoevento? filters=trim
    text zona? filters=trim
    text local? filters=trim
    text datahorario? filters=trim
    int alarm_events_id? {
      table = "alarm_events"
    }
  
    text ideventgo? filters=trim
    bool falha?
    int tentativas?
    timestamp? dtUltimaTentativa?
    bool exec?
    bool atendido?
    text idProcesso? filters=trim
    text idDispositivo? filters=trim
    bool notificarsempre?
    text DispNome? filters=trim
    text DispDescricao? filters=trim
    text DispTipo? filters=trim
    int nroErr?
    bool cancelado?
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "telefone", op: "asc"}]}
    {type: "btree", field: [{name: "franqueado", op: "asc"}]}
    {type: "btree", field: [{name: "ideventgo", op: "asc"}]}
    {
      type : "btree"
      field: [{name: "alarm_events_id", op: "asc"}]
    }
    {type: "btree", field: [{name: "idProcesso", op: "asc"}]}
    {type: "btree", field: [{name: "idDispositivo", op: "asc"}]}
    {
      type : "btree"
      field: [
        {name: "idProcesso", op: "asc"}
        {name: "telefone", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [{name: "falha", op: "asc"}, {name: "exec", op: "asc"}]
    }
    {
      type : "btree"
      field: [
        {name: "falha", op: "asc"}
        {name: "tentativas", op: "asc"}
        {name: "dtUltimaTentativa", op: "asc"}
        {name: "exec", op: "asc"}
      ]
    }
  ]
}