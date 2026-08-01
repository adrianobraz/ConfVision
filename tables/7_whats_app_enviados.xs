table WhatsAppEnviados {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text whats_ID? filters=trim
    text whats_messageTimestamp? filters=trim
    text whats_messageid? filters=trim
    text whats_sender? filters=trim
    text whats_senderName? filters=trim
    text whats_text? filters=trim
    int id_tblAlarm_events?
    text id_Evento? filters=trim
    text Codigo? filters=trim
    text Particao? filters=trim
    text ZonaUser? filters=trim
    text DataEntrada? filters=trim
    text idProcesso? filters=trim
    text idDispositivo? filters=trim
    text idCliente? filters=trim
    text ctiGrupo? filters=trim
    text idFranqueado? filters=trim
    text codigoBenuvem? filters=trim
    text Conta? filters=trim
    text CameraAtiva? filters=trim
    date? Data?
    bool audio?
    bool ligacao?
    bool texto?
    text DispNome? filters=trim
    text DispDescricao? filters=trim
    text DispTipo? filters=trim
    bool SMS?
    text sms_status? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {
      type : "btree"
      field: [
        {name: "idFranqueado", op: "asc"}
        {name: "created_at", op: "desc"}
      ]
    }
    {type: "btree", field: [{name: "idDispositivo", op: "asc"}]}
    {
      type : "btree"
      field: [{name: "id_tblAlarm_events", op: "asc"}]
    }
    {
      type : "btree"
      field: [
        {name: "idDispositivo", op: "asc"}
        {name: "ZonaUser", op: "asc"}
        {name: "id_tblAlarm_events", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [
        {name: "idDispositivo", op: "asc"}
        {name: "ZonaUser", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [
        {name: "idFranqueado", op: "asc"}
        {name: "created_at", op: "asc"}
      ]
    }
  ]
}