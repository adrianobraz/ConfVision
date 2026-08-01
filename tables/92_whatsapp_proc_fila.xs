table WhatsappProcFila {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text SendMsgWhats? filters=trim
    text numerowhatsapp? filters=trim
    bool enviaSom?
    text SendAudioWhats? filters=trim
    int tblWhatsAppEnviadosTexto?
    int tblWhatsAppEnviadosAudio?
    int tblWhatsAppEnviadosLigar?
    bool ligar?
    text idFranqueado? filters=trim
    text nomecliente? filters=trim
    text empresanome? filters=trim
    text GRUPOFALHA? filters=trim
    text tipoevento? filters=trim
    text zona? filters=trim
    text local? filters=trim
    int idevento?
    text datahorario? filters=trim
    text ideventgo? filters=trim
    bool enviartexto?
    text idProcesso? filters=trim
    text idDispositivo? filters=trim
    bool notificarsempre?
    text DispNome? filters=trim
    text DispDescricao? filters=trim
    text DispTipo? filters=trim
    bool analise?
    text zonauser? filters=trim
    text particao? filters=trim
    bool disparo?
    timestamp dtDisparo?=now
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {
      type : "btree"
      field: [
        {name: "created_at", op: "asc"}
        {name: "analise", op: "asc"}
      ]
    }
    {type: "btree", field: [{name: "analise", op: "asc"}]}
    {
      type : "btree"
      field: [
        {name: "analise", op: "asc"}
        {name: "disparo", op: "asc"}
        {name: "created_at", op: "desc"}
      ]
    }
    {type: "btree", field: [{name: "idDispositivo", op: "asc"}]}
    {
      type : "btree"
      field: [
        {name: "idDispositivo", op: "asc"}
        {name: "zonauser", op: "asc"}
        {name: "particao", op: "asc"}
        {name: "idevento", op: "asc"}
      ]
    }
    {type: "btree", field: [{name: "disparo", op: "asc"}]}
    {
      type : "btree"
      field: [{name: "analise", op: "asc"}, {name: "disparo", op: "asc"}]
    }
    {
      type : "btree"
      field: [
        {name: "idProcesso", op: "asc"}
        {name: "disparo", op: "asc"}
      ]
    }
  ]
}