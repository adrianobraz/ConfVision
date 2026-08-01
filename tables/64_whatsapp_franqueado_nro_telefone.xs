table WhatsappFranqueadoNroTelefone {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text franqueado? filters=trim
    text telefone? filters=trim
    int ligacao?
    bool ModuloLigar?
    bool modulotexto?
    text eleven_agent? filters=trim
    text eleven_phone? filters=trim
    text txtApiToken? filters=trim
  
    // u(uazapi) - e(eleven) - g(evogo) - S(sms) - P(pabx)
    text tipoApi? filters=trim
  
    text Grupo? filters=trim
    text FranqueadoNome? filters=trim
    text instancia? filters=trim
    text observacao? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "franqueado", op: "asc"}]}
    {type: "btree", field: [{name: "telefone", op: "asc"}]}
    {
      type : "btree"
      field: [
        {name: "franqueado", op: "asc"}
        {name: "telefone", op: "asc"}
      ]
    }
    {type: "btree", field: [{name: "id", op: "asc"}]}
    {
      type : "btree"
      field: [
        {name: "franqueado", op: "asc"}
        {name: "modulotexto", op: "asc"}
        {name: "Grupo", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [
        {name: "franqueado", op: "asc"}
        {name: "ModuloLigar", op: "asc"}
        {name: "Grupo", op: "asc"}
      ]
    }
  ]
}