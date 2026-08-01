table WhatsappEventCadJanela {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    int whatsappeventocad_id? {
      table = "WhatsappEventoCad"
    }
  
    enum[] dias? {
      values = ["0", "1", "2", "3", "4", "5", "6", "7"]
    }
  
    text hora_inicio? filters=trim
    text hora_fim? filters=trim
    text IdCliente? filters=trim
    text IdFranqueado? filters=trim
    text whatsapp? filters=trim
    text nome? filters=trim
    text idDispositivo? filters=trim
    text NomeDispositivo? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {
      type : "btree"
      field: [{name: "whatsappeventocad_id", op: "asc"}]
    }
    {type: "btree", field: [{name: "id", op: "asc"}]}
    {type: "btree", field: [{name: "hora_inicio", op: "asc"}]}
    {type: "btree", field: [{name: "hora_fim", op: "asc"}]}
  ]
}