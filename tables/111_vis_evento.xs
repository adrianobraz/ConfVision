table vis_evento {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    int vis_camera_id? {
      table = "vis_camera"
    }
  
    text id_franqueado? filters=trim
    text id_cliente? filters=trim
    text id_dispositivo? filters=trim
    text conta? filters=trim
    text particao? filters=trim
    text canal? filters=trim
    text tipo_deteccao? filters=trim
    decimal confianca?
    text snapshot_url? filters=trim
    text video_url? filters=trim
    text bbox_json? filters=trim
    bool processado?
    int alarm_events_id? {
      table = "alarm_events"
    }
  
    bool ignorado?
    text status? filters=trim
    text id_evento? filters=trim
    text id_processo? filters=trim
    timestamp? started_at?
    timestamp? ended_at?
    int clip_count?
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "id_franqueado", op: "asc"}]}
    {type: "btree", field: [{name: "id_cliente", op: "asc"}]}
    {type: "btree", field: [{name: "vis_camera_id", op: "asc"}]}
    {type: "btree", field: [{name: "id_dispositivo", op: "asc"}]}
    {
      type : "btree"
      field: [{name: "alarm_events_id", op: "asc"}]
    }
  ]
}