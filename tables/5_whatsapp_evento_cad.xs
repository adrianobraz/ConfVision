table WhatsappEventoCad {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text IdCliente? filters=trim
    text nomeCliente? filters=trim
    text IdFranqueado? filters=trim
    text nomefranqueado? filters=trim
    text whatsapp? filters=trim
    text nome? filters=trim
    enum[] Tipo? {
      values = [
        "ALARME"
        "ARME"
        "DESARME"
        "EMERGENCIA"
        "FALHAS"
        "GERAL"
        "MEDICO"
        "PANICO"
        "SETUP"
        "TESTE"
        "GRADEHORARIO"
        "INTERNET"
        "ENERGIA"
      ]
    }
  
    text idDispositivo? filters=trim
    text NomeDispositivo? filters=trim
    bool audio?
    bool ligar?
    bool texto?
    bool notificarsempre?
    int ordemligacao?
    bool bloqueado?
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "IdCliente", op: "asc"}]}
    {type: "btree", field: [{name: "IdFranqueado", op: "asc"}]}
    {type: "btree", field: [{name: "whatsapp", op: "asc"}]}
    {type: "btree", field: [{name: "idDispositivo", op: "asc"}]}
    {
      type : "btree"
      field: [
        {name: "IdCliente", op: "asc"}
        {name: "idDispositivo", op: "asc"}
      ]
    }
    {type: "btree", field: [{name: "ordemligacao", op: "desc"}]}
    {
      type : "btree"
      field: [
        {name: "IdCliente", op: "asc"}
        {name: "idDispositivo", op: "asc"}
        {name: "bloqueado", op: "asc"}
      ]
    }
  ]
}