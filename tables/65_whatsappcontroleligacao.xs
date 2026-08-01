table whatsappcontroleligacao {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text franqueado? filters=trim
    text telefone? filters=trim
    timestamp? update?=now
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
  ]
}