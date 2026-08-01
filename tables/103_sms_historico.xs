table sms_Historico {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    bool success?
    text message_id? filters=trim
    text message_tag? filters=trim
    text message_type? filters=trim
    text webhook_type? filters=trim
    int whatsappenviados_id? {
      table = "WhatsAppEnviados"
    }
  
    text message_custom?
    text message_status? filters=trim
    text message_status_details? filters=trim
    int id_referencia?
    timestamp? queued_at?
    timestamp? sent_at?
    timestamp? delivered_at?
    timestamp? failed_at?
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
  ]
}